package cpuanalyzer

type TimedEventKind int

const (
	TimedCpuEventKind TimedEventKind = iota
	TimedJavaFutexEventKind
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
	IsSend          int
}

func (s *Segment) makeTimedEventList(kind TimedEventKind) {
	switch kind {
	case TimedCpuEventKind:
		s.CpuEvents = make([]TimedEvent, 0)
	case TimedJavaFutexEventKind:
		s.JavaFutexEvents = make([]TimedEvent, 0)
	}
}

func (s *Segment) putTimedEvent(event TimedEvent) {
	switch event.Kind() {
	case TimedCpuEventKind:
		s.CpuEvents = append(s.CpuEvents, event)
	case TimedJavaFutexEventKind:
		s.JavaFutexEvents = append(s.JavaFutexEvents, event)
	}
}

func (s *Segment) getTimedEventList(kind TimedEventKind) []TimedEvent {
	switch kind {
	case TimedCpuEventKind:
		return s.CpuEvents
	case TimedJavaFutexEventKind:
		return s.JavaFutexEvents
	default:
		return []TimedEvent{}
	}
}

type CpuEvent struct {
	Pid         uint32
	Tid         uint32
	Comm        string
	StartTime   uint64 `json:"startTime"`
	EndTime     uint64 `json:"endTime"`
	TypeSpecs   string `json:"typeSpecs"`
	RunqLatency string `json:"runqLatency"`
	TimeType    string `json:"timeType"`
	OnInfo      string `json:"onInfo"`
	OffInfo     string `json:"offInfo"`
	Log         string `json:"log"`
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
	Pid       uint32
	Tid       uint32
	Comm      string
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

func (c *JavaFutexEvent) Kind() TimedEventKind {
	return TimedJavaFutexEventKind
}
