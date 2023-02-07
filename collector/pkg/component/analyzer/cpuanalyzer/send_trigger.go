package cpuanalyzer

import (
	"sync"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/filepathhelper"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"github.com/Kindling-project/kindling/collector/pkg/model/constvalues"
)

var (
	enableProfile bool
	once          sync.Once
	sendChannel   chan SendTriggerEvent
	sampleMap     sync.Map
	isInstallApm  map[uint64]bool
)

// ReceiveDataGroupAsSignal receives model.DataGroup as a signal.
// Signal is used to trigger to send CPU on/off events
func ReceiveDataGroupAsSignal(data *model.DataGroup) {
	if !enableProfile {
		once.Do(func() {
			// We must close the channel at the sender-side.
			// Otherwise, we need complex codes to handle it.
			if sendChannel != nil {
				close(sendChannel)
			}
		})
		return
	}
	if data.Labels.GetBoolValue("isInstallApm") {
		isInstallApm[uint64(data.Labels.GetIntValue("pid"))] = true
	} else {
		if isInstallApm[uint64(data.Labels.GetIntValue("pid"))] {
			return
		}
	}
	if data.Labels.GetBoolValue(constlabels.IsSlow) {
		url, ok := sampleMap.Load(data.Labels.GetStringValue(constlabels.ContentKey) + string(data.Labels.GetIntValue("pid")))
		if ok && url != nil {
			sampleMap.Store(data.Labels.GetStringValue(constlabels.ContentKey)+string(data.Labels.GetIntValue("pid")), data)
		}
	}
}

type SendTriggerEvent struct {
	Pid          uint32           `json:"pid"`
	StartTime    uint64           `json:"startTime"`
	SpendTime    uint64           `json:"spendTime"`
	OriginalData *model.DataGroup `json:"originalData"`
}

func (ca *CpuAnalyzer) ReceiveSendSignal() {
	// Break the for loop if the channel is closed
	for sendContent := range sendChannel {
		for _, nexConsumer := range ca.nextConsumers {
			_ = nexConsumer.Consume(sendContent.OriginalData)
		}
		// Copy the value and then get its pointer to create a new task
		triggerEvent := sendContent
		task := &SendEventsTask{
			cpuAnalyzer:              ca,
			triggerEvent:             &triggerEvent,
			edgeEventsWindowDuration: time.Duration(ca.cfg.EdgeEventsWindowSize) * time.Second,
		}
		expiredCallback := func() {
			ca.routineSize.Dec()
		}
		// The expired duration should be windowDuration+1 because the ticker and the timer are not started together.
		NewAndStartScheduledTaskRoutine(1*time.Second, time.Duration(ca.cfg.EdgeEventsWindowSize)*time.Second+1, task, expiredCallback)
		ca.routineSize.Inc()
	}
}

type SendEventsTask struct {
	tickerCount              int
	cpuAnalyzer              *CpuAnalyzer
	triggerEvent             *SendTriggerEvent
	edgeEventsWindowDuration time.Duration
}

// |________________|______________|_________________|
// 0  (edgeWindow)  1  (duration)  2  (edgeWindow)   3
// 0: The start time of the windows where the events we need are.
// 1: The start time of the "trace".
// 2: The end time of the "trace". This is nearly equal to the creating time of the task.
// 3: The end time of the windows where the events we need are.
func (t *SendEventsTask) run() {
	currentWindowsStartTime := uint64(t.tickerCount)*uint64(time.Second) + t.triggerEvent.StartTime - uint64(t.edgeEventsWindowDuration)
	currentWindowsEndTime := uint64(t.tickerCount)*uint64(time.Second) + t.triggerEvent.StartTime + t.triggerEvent.SpendTime
	t.tickerCount++
	// keyElements are used to correlate the cpuEvents with the trace.
	keyElements := filepathhelper.GetFilePathElements(t.triggerEvent.OriginalData, t.triggerEvent.StartTime)
	t.cpuAnalyzer.sendEvents(keyElements.ToAttributes(), t.triggerEvent.Pid, currentWindowsStartTime, currentWindowsEndTime)
}

func (ca *CpuAnalyzer) sampleSend() {
	timer := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timer.C:
			sampleMap.Range(func(k, v interface{}) bool {
				data := v.(*model.DataGroup)
				duration, ok := data.GetMetric(constvalues.RequestTotalTime)
				if !ok {
					return false
				}
				event := SendTriggerEvent{
					Pid:          uint32(data.Labels.GetIntValue("pid")),
					StartTime:    data.Timestamp,
					SpendTime:    uint64(duration.GetInt().Value),
					OriginalData: data.Clone(),
				}
				sendChannel <- event
				sampleMap.Delete(k)
				return true
			})

		}
	}

}

func (ca *CpuAnalyzer) sendEvents(keyElements *model.AttributeMap, pid uint32, startTime uint64, endTime uint64) {
	ca.lock.RLock()
	defer ca.lock.RUnlock()

	maxSegmentSize := ca.cfg.SegmentSize
	tidCpuEvents, exist := ca.cpuPidEvents[pid]
	if !exist {
		ca.telemetry.Logger.Infof("Not found the cpu events with the pid=%d, startTime=%d, endTime=%d",
			pid, startTime, endTime)
		return
	}
	startTimeSecond := startTime / nanoToSeconds
	endTimeSecond := endTime / nanoToSeconds

	for _, timeSegments := range tidCpuEvents {
		if endTimeSecond < timeSegments.BaseTime || startTimeSecond > timeSegments.BaseTime+uint64(maxSegmentSize) {
			ca.telemetry.Logger.Infof("pid=%d tid=%d events are beyond the time windows. BaseTimeSecond=%d, "+
				"startTimeSecond=%d, endTimeSecond=%d", pid, timeSegments.Tid, timeSegments.BaseTime, startTimeSecond, endTimeSecond)
			continue
		}
		startIndex := int(startTimeSecond - timeSegments.BaseTime)
		if startIndex < 0 {
			startIndex = 0
		}
		endIndex := endTimeSecond - timeSegments.BaseTime
		if endIndex > timeSegments.BaseTime+uint64(maxSegmentSize) {
			endIndex = timeSegments.BaseTime + uint64(maxSegmentSize)
		}
		ca.telemetry.Logger.Infof("pid=%d tid=%d sends events. startSecond=%d, endSecond=%d",
			pid, timeSegments.Tid, startTimeSecond, endTimeSecond)
		for i := startIndex; i <= int(endIndex) && i < maxSegmentSize; i++ {
			val := timeSegments.Segments.GetByIndex(i)
			if val == nil {
				continue
			}
			segment := val.(*Segment)
			if len(segment.CpuEvents) != 0 {
				// Don't remove the duplicated one
				segment.IndexTimestamp = time.Now().String()
				dataGroup := segment.toDataGroup(timeSegments)
				dataGroup.Labels.Merge(keyElements)
				for _, nexConsumer := range ca.nextConsumers {
					_ = nexConsumer.Consume(dataGroup)
				}
				segment.IsSend = 1
			}
		}
	}
}
