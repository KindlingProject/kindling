package cpuanalyzer

import (
	"context"
	"encoding/json"
)

var SendChannel chan SendTriggerEvent

func init() {
	SendChannel = make(chan SendTriggerEvent, 3e5)
}

func ReceiveSendSignal(event SendTriggerEvent) {
	SendChannel <- event
}

type SendTriggerEvent struct {
	Pid       uint32 `json:"pid"`
	StartTime uint64 `json:"startTime"`
	SpendTime uint64 `json:"spendTime"`
}

func (ca *CpuAnalyzer) SendCpuEventTest(pid uint32) {
	ca.lock.Lock()
	defer ca.lock.Unlock()
	tidCpuEvents, exist := ca.cpuPidEvents[pid]
	if !exist {
		return
	}
	for _, timeSegments := range tidCpuEvents {
		for i := 0; i < 20; i++ {
			val, _ := timeSegments.Segments.GetByIndex(i)
			segment := val.(*Segment)
			ca.esClient.Index().Index("cpu_event").Type("_doc").BodyJson(segment).Do(context.Background())
		}
	}
}

func (ca *CpuAnalyzer) SendCircle() {
	for {
		sendContent := <-SendChannel
		data, _ := json.Marshal(sendContent)
		ca.telemetry.Logger.Sugar().Info("Receive a trace signal: %s", string(data))
		ca.SendCpuEvent(sendContent.Pid, sendContent.StartTime, sendContent.SpendTime)
	}
}

func (ca *CpuAnalyzer) SendCpuEvent(pid uint32, startTime uint64, spendTime uint64) error {
	ca.lock.Lock()
	defer ca.lock.Unlock()
	ca.telemetry.Logger.Sugar().Infof("Will send cpu events for pid=%d, start_time=%d, duration=%d", pid, startTime, spendTime)

	tidCpuEvents, exist := ca.cpuPidEvents[pid]
	if !exist {
		ca.telemetry.Logger.Sugar().Infof("Not found the cpu events with the pid=%d", pid)
		return nil
	}
	for _, timeSegments := range tidCpuEvents {
		if timeSegments.BaseTime+uint64(ca.cfg.GetSegmentSize()) < startTime/nanoToSeconds || timeSegments.BaseTime > startTime/nanoToSeconds {
			return nil
		}

		for i := 0; i < int(spendTime/nanoToSeconds)+1+2; i++ {
			val, _ := timeSegments.Segments.GetByIndex(int(startTime/nanoToSeconds-timeSegments.BaseTime) + i - 1)
			if val == nil {
				continue
			}
			segment := val.(*Segment)
			if segment.IsSend != 1 {
				ca.esClient.Index().Index("cpu_event").Type("_doc").BodyJson(segment).Do(context.Background())
			}
			segment.IsSend = 1
			timeSegments.Segments.UpdateByIndex(int(startTime/nanoToSeconds-timeSegments.BaseTime)+i-1, segment)
		}
	}
	return nil
}
