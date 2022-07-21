package cpuanalyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
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
			data, _ := json.Marshal(segment)
			fmt.Println(string(data))
			fmt.Println("--------------------------------")
			ca.esClient.Index().Index("cpu_event").Type("_doc").BodyJson(segment).Do(context.Background())
		}
	}
}

func (ca *CpuAnalyzer) SendCpuEvent(pid uint32, startTime uint64, spendTime uint64) error {
	ca.lock.Lock()
	defer ca.lock.Unlock()

	profilePid, _ := strconv.Atoi(os.Getenv("profilepid"))
	profilePid2 := uint32(profilePid)
	fmt.Println(profilePid2)
	if pid == profilePid2 {
		ca.telemetry.Logger.Info("time", zap.Uint64("start_time", startTime))
		ca.telemetry.Logger.Info("time", zap.Uint64("end_time", spendTime))
	} else {
		return nil
	}
	tidCpuEvents, exist := ca.cpuPidEvents[pid]
	if !exist {
		return nil
	}
	for _, timeSegments := range tidCpuEvents {
		if pid == profilePid2 {
			ca.telemetry.Logger.Info("base_time", zap.Uint64("base_time", timeSegments.BaseTime))
		}
		if timeSegments.BaseTime+uint64(ca.cfg.GetSegmentSize()) < startTime/nanoToSeconds || timeSegments.BaseTime > startTime/nanoToSeconds || pid != profilePid2 {
			return nil
		}

		for i := 0; i < int(spendTime/nanoToSeconds)+1+2; i++ {
			if pid == profilePid2 {
				ca.telemetry.Logger.Info("base_time", zap.Int("key", i))
			}
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

func (ca *CpuAnalyzer) SendCircle() {
	for {
		sendContent := <-SendChannel
		data, _ := json.Marshal(sendContent)
		fmt.Println(string(data))
		ca.SendCpuEvent(sendContent.Pid, sendContent.StartTime, sendContent.SpendTime)
	}

}

func (ca *CpuAnalyzer) SendTest() {
	for {
		for i := 0; i < 100000; i++ {
			ca.SendCpuEventTest(uint32(i))
		}
		time.Sleep(5 * time.Second)
	}
}
