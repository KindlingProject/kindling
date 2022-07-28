package cpuanalyzer

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/Kindling-project/kindling/collector/analyzer"
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/consumer"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/olivere/elastic/v6"
	"go.uber.org/zap"
)

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
	ca.telemetry.Logger.Sugar().Infof("Es return with code %d and version %s", code, info.Version.Number)
	esversionCode, err := ca.esClient.ElasticsearchVersion(ca.cfg.GetEsHost())
	if err != nil {
		ca.telemetry.Logger.Warn("new es client error", zap.Error(err))
	}
	ca.telemetry.Logger.Sugar().Infof("es version %s\n", esversionCode)
	//go ca.SendTest()
	go ca.SendCircle()
	return nil
}

func (ca *CpuAnalyzer) ConsumeEvent(event *model.KindlingEvent) error {
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
	for i := 0; i < int(event.ParamsNumber); i++ {
		userAttributes := event.UserAttributes[i]
		switch {
		case userAttributes.GetKey() == "end_time":
			ev.EndTime, _ = strconv.ParseUint(string(userAttributes.GetValue()), 10, 64)
			break
		case userAttributes.GetKey() == "data":
			ev.DataVal = string(userAttributes.GetValue())
			break
		default:
			break
		}
	}
	ca.telemetry.Logger.Sugar().Infof("Receive a java futex event, %v", ev)
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
		case event.UserAttributes[i].GetKey() == "stack":
			ev.Stack = string(userAttributes.GetValue())
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

			for i := 0; i < clearSize-1; i++ {
				val, _ := timeSegments.Segments.GetByIndex(i + ca.cfg.GetSegmentSize()/2)
				if val != nil {
					timeSegments.Segments.UpdateByIndex(i, val)
				}
				segmentTmp := &Segment{
					Pid:             pid,
					Tid:             tid,
					StartTime:       (timeSegments.BaseTime + uint64(i+clearSize)) * nanoToSeconds,
					EndTime:         (timeSegments.BaseTime + uint64(i+clearSize+1)) * nanoToSeconds,
					ThreadName:      threadName,
					CpuEvents:       make([]TimedEvent, 0),
					JavaFutexEvents: make([]TimedEvent, 0),
				}
				timeSegments.Segments.UpdateByIndex(i+clearSize, segmentTmp)
			}

		}
		if int(event.EndTimestamp()/nanoToSeconds-timeSegments.BaseTime) < 0 {
			return
		}
		// 开始时间基于baseTime的偏移量
		startOffset := event.StartTimestamp()/nanoToSeconds - timeSegments.BaseTime
		// 结束时间基于开始时间的偏移量
		endOffset := event.EndTimestamp()/nanoToSeconds - event.StartTimestamp()/nanoToSeconds
		if startOffset > 0 {
			for i := startOffset; i <= startOffset+endOffset; i++ {
				var segment *Segment
				val, _ := timeSegments.Segments.GetByIndex(int(i))
				if val == nil {
					segment = &Segment{
						Pid:             pid,
						Tid:             tid,
						StartTime:       (timeSegments.BaseTime + uint64(i)) * nanoToSeconds,
						EndTime:         (timeSegments.BaseTime + uint64(i+1)) * nanoToSeconds,
						ThreadName:      threadName,
						CpuEvents:       make([]TimedEvent, 0),
						JavaFutexEvents: make([]TimedEvent, 0),
					}
				} else {
					segment = val.(*Segment)
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
				Pid:             pid,
				Tid:             tid,
				StartTime:       (newTimeSegments.BaseTime + uint64(i)) * nanoToSeconds,
				EndTime:         (newTimeSegments.BaseTime + uint64(i+1)) * nanoToSeconds,
				ThreadName:      threadName,
				CpuEvents:       make([]TimedEvent, 0),
				JavaFutexEvents: make([]TimedEvent, 0),
			}
			newTimeSegments.Segments.Push(segment)
		}
		val, _ := newTimeSegments.Segments.GetByIndex(0)
		segment := val.(*Segment)
		segment.putTimedEvent(event)
		tidCpuEvents[tid] = newTimeSegments
	}
}

func (ca *CpuAnalyzer) Shutdown() error {
	// TODO: implement
	return nil
}
