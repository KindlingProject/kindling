package model


type CgoKindlingEvent struct {
	Timestamp      uint64
	Name           string
	Category       uint32
	UserAttributes [8]CgoKeyValue
	Context        EventContext
}
type EventContext struct {
	ThreadInfo ThreadInfo
	FdInfo     FdInfo
}
type ThreadInfo struct {
	Pid         uint32
	Tid         uint32
	Uid         uint32
	Gid         uint32
	Comm        string
	ContainerId string
}
type FdInfo struct {
	Protocol    uint32
	Num         uint32
	FdType      uint32
	Filename    string
	Directory   string
	Role        uint8
	Sip         uint32
	Dip         uint32
	Sport       uint32
	Dport       uint32
	Source      uint64
	Destination uint64
}

type CgoKeyValue struct {
	Key       string
	Value     string
	ValueType uint32
}