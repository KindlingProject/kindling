package cpuanalyzer

import (
	"encoding/json"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
)

type TimedEventKind int

const (
	TimedCpuEventKind TimedEventKind = iota
	TimedJavaFutexEventKind
	TimedTransactionIdEventKind
)

const (
	CpuEventLabel           = "cpuEvents"
	JavaFutexEventLabel     = "javaFutexEvents"
	TransactionIdEventLabel = "transactionIds"
)

type TimedEvent interface {
	StartTimestamp() uint64
	EndTimestamp() uint64
	Kind() TimedEventKind
}

type TimeSegments struct {
	Pid      uint32       `json:"pid"`
	Tid      uint32       `json:"tid"`
	BaseTime uint64       `json:"baseTime"`
	Segments *CircleQueue `json:"segments"`
}

type Segment struct {
	Pid             uint32       `json:"pid"`
	Tid             uint32       `json:"tid"`
	ThreadName      string       `json:"threadName"`
	StartTime       uint64       `json:"startTime"`
	EndTime         uint64       `json:"endTime"`
	CpuEvents       []TimedEvent `json:"cpuEvents"`
	JavaFutexEvents []TimedEvent `json:"javaFutexEvents"`
	TransactionIds  []TimedEvent `json:"transactionIds"`
	IsSend          int
	IndexTimestamp  string `json:"indexTimestamp"`
}

func newSegment(pid uint32, tid uint32, threadName string, startTime uint64, endTime uint64) *Segment {
	return &Segment{
		Pid:             pid,
		Tid:             tid,
		ThreadName:      threadName,
		StartTime:       startTime,
		EndTime:         endTime,
		CpuEvents:       make([]TimedEvent, 0),
		JavaFutexEvents: make([]TimedEvent, 0),
		TransactionIds:  make([]TimedEvent, 0),
		IsSend:          0,
		IndexTimestamp:  "",
	}
}
func (s *Segment) putTimedEvent(event TimedEvent) {
	switch event.Kind() {
	case TimedCpuEventKind:
		s.CpuEvents = append(s.CpuEvents, event)
	case TimedJavaFutexEventKind:
		s.JavaFutexEvents = append(s.JavaFutexEvents, event)
	case TimedTransactionIdEventKind:
		s.TransactionIds = append(s.TransactionIds, event)
	}
}

func (s *Segment) toDataGroup() *model.DataGroup {
	labels := model.NewAttributeMap()
	labels.AddIntValue(constlabels.Pid, int64(s.Pid))
	labels.AddIntValue(constlabels.Tid, int64(s.Tid))
	labels.AddStringValue("threadName", s.ThreadName)
	labels.AddIntValue("startTime", int64(s.StartTime))
	labels.AddIntValue("endTime", int64(s.EndTime))
	cpuEventString, err := json.Marshal(s.CpuEvents)
	if err == nil {
		labels.AddStringValue(CpuEventLabel, string(cpuEventString))
	}
	javaFutexEventString, err := json.Marshal(s.JavaFutexEvents)
	if err == nil {
		labels.AddStringValue(JavaFutexEventLabel, string(javaFutexEventString))
	}
	transactionIdEventString, err := json.Marshal(s.TransactionIds)
	if err == nil {
		labels.AddStringValue(TransactionIdEventLabel, string(transactionIdEventString))
	}
	return model.NewDataGroup(constnames.CameraEventGroupName, labels, s.StartTime)
}

type CpuEvent struct {
	StartTime   uint64 `json:"startTime"`
	EndTime     uint64 `json:"endTime"`
	TypeSpecs   string `json:"typeSpecs"`
	RunqLatency string `json:"runqLatency"`
	TimeType    string `json:"timeType"`
	OnInfo      string `json:"onInfo"`
	OffInfo     string `json:"offInfo"`
	Log         string `json:"log"`
	Stack       string `json:"stack"`
}

func (c *CpuEvent) StartTimestamp() uint64 {
	return c.StartTime
}

func (c *CpuEvent) EndTimestamp() uint64 {
	return c.EndTime
}

func (c *CpuEvent) Kind() TimedEventKind {
	return TimedCpuEventKind
}

type JavaFutexEvent struct {
	StartTime uint64 `json:"startTime"`
	EndTime   uint64 `json:"endTime"`
	DataVal   string `json:"dataValue"`
}

func (j *JavaFutexEvent) StartTimestamp() uint64 {
	return j.StartTime
}

func (j *JavaFutexEvent) EndTimestamp() uint64 {
	return j.EndTime
}

func (j *JavaFutexEvent) Kind() TimedEventKind {
	return TimedJavaFutexEventKind
}

type TransactionIdEvent struct {
	Timestamp uint64 `json:"timestamp"`
	TraceId   string `json:"traceId"`
	IsEntry   uint32 `json:"isEntry"`
}

func (t *TransactionIdEvent) StartTimestamp() uint64 {
	return t.Timestamp
}

func (t *TransactionIdEvent) EndTimestamp() uint64 {
	return t.Timestamp
}

func (t *TransactionIdEvent) Kind() TimedEventKind {
	return TimedTransactionIdEventKind
}
