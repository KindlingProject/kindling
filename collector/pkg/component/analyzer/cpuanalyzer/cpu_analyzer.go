package cpuanalyzer

import (
	"fmt"
	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
	"go.uber.org/zap/zapcore"
	"strconv"
	"sync"
)

const (
	CpuProfile analyzer.Type = "cpuanalyzer"
)

type CpuAnalyzer struct {
	cfg          *Config
	cpuPidEvents map[uint32]map[uint32]*TimeSegments
	// { pid: routine }
	sendEventsRoutineMap sync.Map
	lock                 sync.Mutex
	telemetry            *component.TelemetryTools

	nextConsumers []consumer.Consumer
}

func (ca *CpuAnalyzer) Type() analyzer.Type {
	return CpuProfile
}

func (ca *CpuAnalyzer) ConsumableEvents() []string {
	return []string{constnames.CpuEvent, constnames.JavaFutexInfo, constnames.TransactionIdEvent, constnames.ProcessExitEvent}
}

func NewCpuAnalyzer(cfg interface{}, telemetry *component.TelemetryTools, consumers []consumer.Consumer) analyzer.Analyzer {
	config, _ := cfg.(*Config)
	ca := &CpuAnalyzer{
		cfg:           config,
		telemetry:     telemetry,
		nextConsumers: consumers,
	}
	ca.cpuPidEvents = make(map[uint32]map[uint32]*TimeSegments, 100000)
	return ca
}

func (ca *CpuAnalyzer) Start() error {
	// Disable receiving and sending the profiling data by default.
	return nil
}

func (ca *CpuAnalyzer) Shutdown() error {
	ca.StopProfile()
	return nil
}

func (ca *CpuAnalyzer) ConsumeEvent(event *model.KindlingEvent) error {
	if !enableProfile {
		return nil
	}
	switch event.Name {
	case constnames.CpuEvent:
		ca.ConsumeCpuEvent(event)
	case constnames.JavaFutexInfo:
		ca.ConsumeJavaFutexEvent(event)
	case constnames.TransactionIdEvent:
		ca.ConsumeTransactionIdEvent(event)
	case constnames.ProcessExitEvent:
		pid := event.GetPid()
		tid := event.Ctx.ThreadInfo.GetTid()
		ca.trimExitedThread(pid, tid)
	}
	return nil
}

func (ca *CpuAnalyzer) ConsumeTransactionIdEvent(event *model.KindlingEvent) {
	isEntry, _ := strconv.Atoi(event.GetStringUserAttribute("is_enter"))
	ev := &TransactionIdEvent{
		Timestamp: event.Timestamp,
		TraceId:   event.GetStringUserAttribute("trace_id"),
		IsEntry:   uint32(isEntry),
	}
	//ca.sendEventDirectly(event.GetPid(), event.Ctx.ThreadInfo.GetTid(), event.Ctx.ThreadInfo.Comm, ev)
	ca.PutEventToSegments(event.GetPid(), event.Ctx.ThreadInfo.GetTid(), event.Ctx.ThreadInfo.Comm, ev)
}

func (ca *CpuAnalyzer) ConsumeJavaFutexEvent(event *model.KindlingEvent) {
	ev := new(JavaFutexEvent)
	ev.StartTime = event.Timestamp
	for i := 0; i < int(event.ParamsNumber); i++ {
		userAttributes := event.UserAttributes[i]
		switch {
		case userAttributes.GetKey() == "end_time":
			ev.EndTime, _ = strconv.ParseUint(string(userAttributes.GetValue()), 10, 64)
		case userAttributes.GetKey() == "data":
			ev.DataVal = string(userAttributes.GetValue())
		}
	}
	//ca.sendEventDirectly(event.GetPid(), event.Ctx.ThreadInfo.GetTid(), event.Ctx.ThreadInfo.Comm, ev)
	ca.PutEventToSegments(event.GetPid(), event.Ctx.ThreadInfo.GetTid(), event.Ctx.ThreadInfo.Comm, ev)
}

func (ca *CpuAnalyzer) ConsumeCpuEvent(event *model.KindlingEvent) {
	ev := new(CpuEvent)
	for i := 0; i < int(event.ParamsNumber); i++ {
		userAttributes := event.UserAttributes[i]
		switch {
		case userAttributes.GetKey() == "start_time":
			ev.StartTime = userAttributes.GetUintValue()
		case event.UserAttributes[i].GetKey() == "end_time":
			ev.EndTime = userAttributes.GetUintValue()
		case event.UserAttributes[i].GetKey() == "type_specs":
			ev.TypeSpecs = string(userAttributes.GetValue())
		case event.UserAttributes[i].GetKey() == "runq_latency":
			ev.RunqLatency = string(userAttributes.GetValue())
		case event.UserAttributes[i].GetKey() == "time_type":
			ev.TimeType = string(userAttributes.GetValue())
		case event.UserAttributes[i].GetKey() == "on_info":
			ev.OnInfo = string(userAttributes.GetValue())
		case event.UserAttributes[i].GetKey() == "off_info":
			ev.OffInfo = string(userAttributes.GetValue())
		case event.UserAttributes[i].GetKey() == "log":
			ev.Log = string(userAttributes.GetValue())
		case event.UserAttributes[i].GetKey() == "stack":
			ev.Stack = string(userAttributes.GetValue())
		}
	}
	if ce := ca.telemetry.Logger.Check(zapcore.DebugLevel, ""); ce != nil {
		ca.telemetry.Logger.Debug(fmt.Sprintf("Receive CpuEvent: pid=%d, tid=%d, comm=%s, %+v", event.GetPid(),
			event.Ctx.ThreadInfo.GetTid(), event.Ctx.ThreadInfo.Comm, ev))
	}
	//ca.sendEventDirectly(event.GetPid(), event.Ctx.ThreadInfo.GetTid(), event.Ctx.ThreadInfo.Comm, ev)
	ca.PutEventToSegments(event.GetPid(), event.Ctx.ThreadInfo.GetTid(), event.Ctx.ThreadInfo.Comm, ev)
}

var nanoToSeconds uint64 = 1e9

func (ca *CpuAnalyzer) PutEventToSegments(pid uint32, tid uint32, threadName string, event TimedEvent) {
	ca.lock.Lock()
	defer ca.lock.Unlock()
	tidCpuEvents, exist := ca.cpuPidEvents[pid]
	if !exist {
		tidCpuEvents = make(map[uint32]*TimeSegments)
		ca.cpuPidEvents[pid] = tidCpuEvents
	}
	timeSegments, exist := tidCpuEvents[tid]
	maxSegmentSize := ca.cfg.GetSegmentSize()
	if exist {
		endOffset := int(event.EndTimestamp()/nanoToSeconds - timeSegments.BaseTime)
		if endOffset < 0 {
			ca.telemetry.Logger.Debugf("EndOffset of the event is negative. EndTimestamp=%d, BaseTime=%d",
				event.EndTimestamp(), timeSegments.BaseTime)
			return
		}
		startOffset := int(event.StartTimestamp()/nanoToSeconds - timeSegments.BaseTime)
		if startOffset < 0 {
			startOffset = 0
		}
		// If the timeSegment is full, we clear half of its elements.
		// Note the offset will be times of maxSegmentSize when no events with this tid come for long time,
		// so the timeSegment will be cleared multiple times until it can accommodate the events.
		// TODO: clear the whole elements if startOffset>=1.5*maxSegmentSize
		if startOffset >= maxSegmentSize || endOffset > maxSegmentSize {
			if startOffset*2 >= 3*maxSegmentSize {
				// clear all elements
				ca.telemetry.Logger.Debugf("pid=%d, tid=%d, comm=%s, reset BaseTime from %d to %d", pid, tid,
					threadName, timeSegments.BaseTime, event.StartTimestamp()/nanoToSeconds)
				timeSegments.Segments.Clear()
				timeSegments.BaseTime = event.StartTimestamp() / nanoToSeconds
				endOffset = endOffset - startOffset
				startOffset = 0
			} else {
				// Clear half of the elements
				clearSize := maxSegmentSize / 2
				ca.telemetry.Logger.Debugf("pid=%d, tid=%d, comm=%s, update BaseTime from %d to %d", pid, tid,
					threadName, timeSegments.BaseTime, timeSegments.BaseTime+uint64(clearSize))
				timeSegments.BaseTime = timeSegments.BaseTime + uint64(clearSize)
				startOffset -= clearSize
				if startOffset < 0 {
					startOffset = 0
				}
				endOffset -= clearSize
				for i := 0; i < clearSize; i++ {
					movedIndex := i + clearSize
					val := timeSegments.Segments.GetByIndex(movedIndex)
					timeSegments.Segments.UpdateByIndex(i, val)
					segmentTmp := newSegment(pid, tid, threadName,
						(timeSegments.BaseTime+uint64(movedIndex))*nanoToSeconds,
						(timeSegments.BaseTime+uint64(movedIndex+1))*nanoToSeconds)
					timeSegments.Segments.UpdateByIndex(movedIndex, segmentTmp)
				}
			}
		}

		for i := startOffset; i <= endOffset && i < maxSegmentSize; i++ {
			val := timeSegments.Segments.GetByIndex(i)
			segment := val.(*Segment)
			segment.putTimedEvent(event)
			segment.IsSend = 0
			timeSegments.Segments.UpdateByIndex(i, segment)
		}

	} else {
		newTimeSegments := &TimeSegments{
			Pid:      pid,
			Tid:      tid,
			BaseTime: event.StartTimestamp() / nanoToSeconds,
			Segments: NewCircleQueue(maxSegmentSize),
		}
		for i := 0; i < maxSegmentSize; i++ {
			segment := newSegment(pid, tid, threadName,
				(newTimeSegments.BaseTime+uint64(i))*nanoToSeconds,
				(newTimeSegments.BaseTime+uint64(i+1))*nanoToSeconds)
			newTimeSegments.Segments.UpdateByIndex(i, segment)
		}
		val := newTimeSegments.Segments.GetByIndex(0)
		segment := val.(*Segment)
		segment.putTimedEvent(event)
		tidCpuEvents[tid] = newTimeSegments
	}
}

func (ca *CpuAnalyzer) trimExitedThread(pid uint32, tid uint32) {
	ca.lock.Lock()
	defer ca.lock.Unlock()
	tidEventsMap := ca.cpuPidEvents[pid]
	if tidEventsMap == nil {
		return
	}
	ca.telemetry.Logger.Debugf("Receive a procexit pid=%d, tid=%d, which will be deleted from map", pid, tid)
	delete(tidEventsMap, tid)
}
