package cpuanalyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Kindling-project/kindling/collector/analyzer"
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/consumer"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/olivere/elastic/v6"
	"go.uber.org/zap"
)

var SendChannel chan SendContent

type SendContent struct {
	Pid       uint32 `json:"pid"`
	StartTime uint64 `json:"startTime"`
	SpendTime uint64 `json:"spendTime"`
	Sport     int64  `json:"sport"`
}

const (
	CpuProfile analyzer.Type = "cpuanalyzer"
)

type CpuAnalyzer struct {
	cfg          *Config
	cpuPidEvents map[uint32]map[uint32]TimeSegments
	lock         sync.Mutex
	esClient     *elastic.Client
	telemetry    *component.TelemetryTools
}

func (ca *CpuAnalyzer) Type() analyzer.Type {
	return CpuProfile
}

func (ca *CpuAnalyzer) ConsumableEvents() []string {
	return []string{constnames.CpuEvent, constnames.JavaFutexInfo}
}

func NewCpuAnalyzer(cfg interface{}, telemetry *component.TelemetryTools, consumers []consumer.Consumer) analyzer.Analyzer {
	config, _ := cfg.(*Config)
	ca := &CpuAnalyzer{
		cfg:       config,
		telemetry: telemetry,
	}
	ca.cpuPidEvents = make(map[uint32]map[uint32]TimeSegments, 100000)
	SendChannel = make(chan SendContent, 3e5)
	return ca
}

func (ca *CpuAnalyzer) Start() error {
	errorLog := log.New(os.Stdout, "app", log.LstdFlags)
	var err error
	ca.esClient, err = elastic.NewClient(elastic.SetErrorLog(errorLog), elastic.SetURL(ca.cfg.GetEsHost()), elastic.SetSniff(false))
	if err != nil {
		ca.telemetry.Logger.Warn("new es client error", zap.Error(err))
	}
	info, code, err := ca.esClient.Ping(ca.cfg.GetEsHost()).Do(context.Background())
	if err != nil {
		ca.telemetry.Logger.Warn("new es client error", zap.Error(err))
	}
	fmt.Printf("Es return with code %d and version %s \n", code, info.Version.Number)
	esversionCode, err := ca.esClient.ElasticsearchVersion(ca.cfg.GetEsHost())
	if err != nil {
		ca.telemetry.Logger.Warn("new es client error", zap.Error(err))
	}
	fmt.Printf("es version %s\n", esversionCode)
	//go ca.SendTest()
	go ca.SendCircle()
	return nil
}

func (ca *CpuAnalyzer) ConsumeEvent(event *model.KindlingEvent) error {
	if event.Name != "cpu_event" {
		fmt.Println(event.Name)
	}

	if event.Name == "cpu_event" {
		ca.ConsumeCpuEvent(event)
	} else if event.Name == "java_futex_info" {
		ca.ConsumeJavaFutexEvent(event)
	}
	return nil
}

func (ca *CpuAnalyzer) ConsumeJavaFutexEvent(event *model.KindlingEvent) {
	ev := new(JavaFutexEvent)
	ev.StartTime = event.Timestamp
	fmt.Println(event.ParamsNumber)
	for i := 0; i < int(event.ParamsNumber); i++ {
		userAttributes := event.UserAttributes[i]
		switch {
		case userAttributes.GetKey() == "end_time":
			fmt.Println("end_time:")
			fmt.Println(string(userAttributes.GetValue()))
			ev.EndTime, _ = strconv.ParseUint(string(userAttributes.GetValue()), 10, 64)
			break
		case userAttributes.GetKey() == "data":
			ev.DataVal = string(userAttributes.GetValue())
			break
		default:
			break
		}
	}
	ca.PutEventToSegments(event.GetPid(), event.Ctx.ThreadInfo.GetTid(), event.Ctx.ThreadInfo.Comm, ev)
}

func (ca *CpuAnalyzer) ConsumeCpuEvent(event *model.KindlingEvent) {
	ev := new(CpuEvent)
	for i := 0; i < int(event.ParamsNumber); i++ {
		userAttributes := event.UserAttributes[i]
		switch {
		case userAttributes.GetKey() == "start_time":
			ev.StartTime = userAttributes.GetUintValue()
			break
		case event.UserAttributes[i].GetKey() == "end_time":
			ev.EndTime = userAttributes.GetUintValue()
			break
		case event.UserAttributes[i].GetKey() == "type_specs":
			ev.TypeSpecs = string(userAttributes.GetValue())
			break
		case event.UserAttributes[i].GetKey() == "runq_latency":
			ev.RunqLatency = string(userAttributes.GetValue())
			break
		case event.UserAttributes[i].GetKey() == "time_type":
			ev.TimeType = string(userAttributes.GetValue())
			break
		case event.UserAttributes[i].GetKey() == "on_info":
			ev.OnInfo = string(userAttributes.GetValue())
			break
		case event.UserAttributes[i].GetKey() == "off_info":
			ev.OffInfo = string(userAttributes.GetValue())
			break
		case event.UserAttributes[i].GetKey() == "log":
			ev.Log = string(userAttributes.GetValue())
			break
		default:
			break
		}
	}
	ca.PutEventToSegments(event.GetPid(), event.Ctx.ThreadInfo.GetTid(), event.Ctx.ThreadInfo.Comm, ev)
}

var nanoToSeconds uint64 = 1e9

func (ca *CpuAnalyzer) PutEventToSegments(pid uint32, tid uint32, threadName string, event TimedEvent) {
	ca.lock.Lock()
	defer ca.lock.Unlock()
	tidCpuEvents, exist := ca.cpuPidEvents[pid]
	if !exist {
		tidCpuEvents = make(map[uint32]TimeSegments)
		ca.cpuPidEvents[pid] = tidCpuEvents
	}
	timeSegments, exist := tidCpuEvents[tid]
	if exist {
		// The current segment is full
		if timeSegments.BaseTime+uint64(ca.cfg.GetSegmentSize()) <= event.StartTimestamp()/nanoToSeconds {
			clearSize := 0

			clearSize = ca.cfg.GetSegmentSize() / 2
			timeSegments.BaseTime = timeSegments.BaseTime + uint64(ca.cfg.GetSegmentSize())/2

			for i := 0; i < clearSize; i++ {
				val, _ := timeSegments.Segments.GetByIndex(i)
				segment := val.(Segment)
				segment.makeTimedEventList(event.Kind())
				if clearSize > ca.cfg.GetSegmentSize()/2 && i < ca.cfg.GetSegmentSize()/2 {
					val, _ := timeSegments.Segments.GetByIndex(i + ca.cfg.GetSegmentSize()/2)
					segmentTmp := val.(Segment)
					timeSegments.Segments.UpdateByIndex(i, segmentTmp)
				} else if clearSize > ca.cfg.GetSegmentSize()/2 && i >= ca.cfg.GetSegmentSize()/2 {
					timeSegments.Segments.UpdateByIndex(i, segment)
				} else {
					timeSegments.Segments.UpdateByIndex(i, segment)
				}
			}
		}
		if int(event.EndTimestamp()/nanoToSeconds-timeSegments.BaseTime) < 0 {
			return
		}
		// 开始时间基于baseTime的偏移量
		startOffset := event.StartTimestamp()/nanoToSeconds - timeSegments.BaseTime
		// 结束时间基于开始时间的偏移量
		endOffset := event.EndTimestamp()/nanoToSeconds - event.StartTimestamp()/nanoToSeconds
		fmt.Println(startOffset)
		fmt.Println(endOffset)
		if startOffset > 0 {
			for i := startOffset; i <= startOffset+endOffset; i++ {
				var segment Segment
				val, _ := timeSegments.Segments.GetByIndex(int(i))
				if val == nil {
					segment = *new(Segment)
				} else {
					segment = val.(Segment)
				}
				segment.putTimedEvent(event)
				timeSegments.Segments.UpdateByIndex(int(i), segment)
				tidCpuEvents[tid] = timeSegments
			}
		}
	} else {
		newTimeSegments := TimeSegments{
			Pid:      pid,
			Tid:      tid,
			BaseTime: event.StartTimestamp() / nanoToSeconds,
			Segments: NewCircleQueue(ca.cfg.GetSegmentSize() + 1),
		}
		for i := 0; i < ca.cfg.GetSegmentSize(); i++ {
			segment := &Segment{
				Pid:        pid,
				Tid:        tid,
				ThreadName: threadName,
			}
			newTimeSegments.Segments.Push(segment)
		}
		val, _ := newTimeSegments.Segments.GetByIndex(0)
		segment := val.(Segment)
		segment.putTimedEvent(event)
		tidCpuEvents[pid] = newTimeSegments
	}
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
			segment := val.(Segment)
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
			segment := val.(Segment)
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

func (ca *CpuAnalyzer) Shutdown() error {
	// TODO: implement
	return nil
}
