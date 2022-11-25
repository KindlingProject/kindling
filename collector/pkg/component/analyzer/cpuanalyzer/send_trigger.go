package cpuanalyzer

import (
	"sync"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/filepathhelper"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"github.com/Kindling-project/kindling/collector/pkg/model/constvalues"
)

var eventsWindowsDuration = 6 * time.Second
var (
	enableProfile bool
	once          sync.Once
	sendChannel   chan SendTriggerEvent
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
	if data.Labels.GetBoolValue(constlabels.IsSlow) {
		duration, ok := data.GetMetric(constvalues.RequestTotalTime)
		if !ok {
			return
		}
		event := SendTriggerEvent{
			Pid:          uint32(data.Labels.GetIntValue("pid")),
			StartTime:    data.Timestamp,
			SpendTime:    uint64(duration.GetInt().Value),
			OriginalData: data.Clone(),
		}
		sendChannel <- event
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
		task := &SendEventsTask{0, ca, &sendContent}
		// "Load" and "Delete" could be executed concurrently
		value, ok := ca.sendEventsRoutineMap.Load(sendContent.Pid)
		if !ok {
			ca.putNewRoutine(task)
		} else {
			// The routine may be found before but stopped and deleted after entering this branch.
			// So we much check twice.
			routine, _ := value.(*ScheduledTaskRoutine)
			// TODO Always replacing the task may cause that some events are skipped when the "spendTime"
			// becomes much smaller than before.
			err := routine.ResetExpiredTimerWithNewTask(task)
			if err != nil {
				// The routine has been expired.
				ca.putNewRoutine(task)
			}
		}
	}
}

func (ca *CpuAnalyzer) putNewRoutine(task *SendEventsTask) {
	expiredCallback := func() {
		ca.sendEventsRoutineMap.Delete(task.triggerEvent.Pid)
	}
	// The expired duration should be windowDuration+1 because the ticker and the timer are not started together.
	routine := NewAndStartScheduledTaskRoutine(1*time.Second, eventsWindowsDuration+1, task, expiredCallback)
	ca.sendEventsRoutineMap.Store(task.triggerEvent.Pid, routine)
}

type SendEventsTask struct {
	tickerCount  int
	cpuAnalyzer  *CpuAnalyzer
	triggerEvent *SendTriggerEvent
}

// |________________|______________|_________________|
// 0      (5s)      1  (duration)  2      (5s)       3
// 0: The start time of the windows where the events we need are.
// 1: The start time of the "trace".
// 2: The end time of the "trace". This is nearly equal to the creating time of the task.
// 3: The end time of the windows where the events we need are.
func (t *SendEventsTask) run() {
	currentWindowsStartTime := uint64(t.tickerCount*1e9) + t.triggerEvent.StartTime - uint64(eventsWindowsDuration)
	currentWindowsEndTime := uint64(t.tickerCount*1e9) + t.triggerEvent.StartTime + t.triggerEvent.SpendTime
	t.tickerCount++
	// keyElements are used to correlate the cpuEvents with the trace.
	keyElements := filepathhelper.GetFilePathElements(t.triggerEvent.OriginalData, t.triggerEvent.StartTime)
	t.cpuAnalyzer.sendEvents(keyElements.ToAttributes(), t.triggerEvent.Pid, currentWindowsStartTime, currentWindowsEndTime)
}

func (ca *CpuAnalyzer) sendEvents(keyElements *model.AttributeMap, pid uint32, startTime uint64, endTime uint64) {
	ca.lock.Lock()
	defer ca.lock.Unlock()

	maxSegmentSize := ca.cfg.GetSegmentSize()
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
			ca.telemetry.Logger.Debugf("pid=%d tid=%d events are beyond the time windows. BaseTimeSecond=%d, "+
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
				dataGroup := segment.toDataGroup()
				dataGroup.Labels.Merge(keyElements)
				for _, nexConsumer := range ca.nextConsumers {
					_ = nexConsumer.Consume(dataGroup)
				}
				segment.IsSend = 1
			}
		}
	}
}
