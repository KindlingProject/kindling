package model

import (
	"encoding/binary"
	"errors"
	"math"
	"net"
)

const (
	LOWER32 = 0x00000000FFFFFFFF
	LOWER16 = 0x000000000000FFFF
)

var (
	ErrMessageNotSocket = errors.New("not a network receive/send event")
	byteOrder           = getByteOrder()
)

func getByteOrder() binary.ByteOrder {
	// Check if littleendian or bigendian
	s := int16(0x1234)
	littleVal := byte(0x34)
	if littleVal == byte(int8(s)) {
		return binary.LittleEndian
	}
	return binary.BigEndian
}

func (x *KindlingEvent) GetData() []byte {
	keyValue := x.GetUserAttribute("data")
	if keyValue != nil {
		return keyValue.GetValue()
	}
	return nil
}

func (x *KindlingEvent) GetDataLen() int {
	keyValue := x.GetUserAttribute("data")
	if keyValue != nil {
		return len(keyValue.GetValue())
	}
	return 0
}

func (x *KindlingEvent) GetResVal() int64 {
	keyValue := x.GetUserAttribute("res")
	if keyValue != nil {
		return int64(byteOrder.Uint64(keyValue.Value))
	}
	return -1
}

func (x *KindlingEvent) GetLatency() uint64 {
	keyValue := x.GetUserAttribute("latency")
	if keyValue != nil {
		return byteOrder.Uint64(keyValue.Value)
	}
	return 0
}

func (x *KindlingEvent) GetUintUserAttribute(key string) uint64 {
	keyValue := x.GetUserAttribute(key)
	if keyValue != nil {
		return keyValue.GetUintValue()
	}
	return 0
}

func (x *KindlingEvent) GetIntUserAttribute(key string) int64 {
	keyValue := x.GetUserAttribute(key)
	if keyValue != nil {
		return keyValue.GetIntValue()
	}
	return 0
}

func (x *KindlingEvent) GetFloatUserAttribute(key string) float32 {
	keyValue := x.GetUserAttribute(key)
	if keyValue != nil && keyValue.ValueType == ValueType_FLOAT {
		return math.Float32frombits(byteOrder.Uint32(keyValue.Value))
	}
	return 0.0
}

func (x *KindlingEvent) GetDoubleUserAttribute(key string) float64 {
	keyValue := x.GetUserAttribute(key)
	if keyValue != nil && keyValue.ValueType == ValueType_FLOAT {
		return math.Float64frombits(byteOrder.Uint64(keyValue.Value))
	}
	return 0.0
}

func (x *KindlingEvent) GetStringUserAttribute(key string) string {
	keyValue := x.GetUserAttribute(key)
	if keyValue != nil {
		return string(keyValue.GetValue())
	}
	return ""
}

func (x *KindlingEvent) GetStartTime() uint64 {
	return x.Timestamp - x.GetLatency()
}

func (x *KindlingEvent) GetUserAttribute(key string) *KeyValue {
	if x.UserAttributes == nil {
		return nil
	}
	for _, keyValue := range x.UserAttributes {
		if keyValue.Key == key {
			return keyValue
		}
	}
	return nil
}

func (x *KindlingEvent) GetPid() uint32 {
	ctx := x.GetCtx()
	if ctx == nil {
		return 0
	}
	threadInfo := ctx.GetThreadInfo()
	if threadInfo == nil {
		return 0
	}
	return threadInfo.Pid
}

func (x *KindlingEvent) GetContainerId() string {
	ctx := x.GetCtx()
	if ctx == nil {
		return ""
	}
	threadInfo := ctx.GetThreadInfo()
	if threadInfo == nil {
		return ""
	}
	return threadInfo.ContainerId
}

func (x *KindlingEvent) GetFd() int32 {
	ctx := x.GetCtx()
	if ctx == nil {
		return 0
	}
	fdInfo := ctx.GetFdInfo()
	if fdInfo == nil {
		return 0
	}
	return fdInfo.Num
}

func (x *KindlingEvent) GetSip() string {
	ctx := x.GetCtx()
	if ctx == nil {
		return ""
	}
	fdInfo := ctx.GetFdInfo()
	if fdInfo == nil {
		return ""
	}
	// TODO Only support IPV4, may support IPV6 in future.
	return IPLong2String(fdInfo.Sip[0])
}

func (x *KindlingEvent) GetDip() string {
	ctx := x.GetCtx()
	if ctx == nil {
		return ""
	}
	fdInfo := ctx.GetFdInfo()
	if fdInfo == nil {
		return ""
	}
	// TODO Only support IPV4, may support IPV6 in future.
	return IPLong2String(fdInfo.Dip[0])
}

func IPLong2String(i uint32) string {
	if i > math.MaxUint32 {
		return ""
	}

	ip := make(net.IP, net.IPv4len)
	ip[3] = byte(i >> 24)
	ip[2] = byte(i >> 16)
	ip[1] = byte(i >> 8)
	ip[0] = byte(i)

	return ip.String()
}

func (x *KindlingEvent) GetSport() uint32 {
	ctx := x.GetCtx()
	if ctx == nil {
		return 0
	}
	fdInfo := ctx.GetFdInfo()
	if fdInfo == nil {
		return 0
	}
	return fdInfo.Sport
}

func (x *KindlingEvent) GetDport() uint32 {
	ctx := x.GetCtx()
	if ctx == nil {
		return 0
	}
	fdInfo := ctx.GetFdInfo()
	if fdInfo == nil {
		return 0
	}
	return fdInfo.Dport
}

func (x *KindlingEvent) IsUdp() uint32 {
	if x.GetCtx().GetFdInfo().GetProtocol() == L4Proto_UDP {
		return 1
	}
	return 0
}

func (x *KindlingEvent) IsConnect() bool {
	return x.Name == "connect"
}

func (x *KindlingEvent) IsRequest() (bool, error) {
	if x.Category == Category_CAT_NET {
		switch x.Name {
		case "read", "recvfrom", "recvmsg", "readv", "pread", "preadv":
			return x.isRequest(true)
		case "write", "sendto", "sendmsg", "writev", "pwrite", "pwritev":
			return x.isRequest(false)
		default:
			break
		}
	}
	return false, ErrMessageNotSocket
}

func (x *KindlingEvent) isRequest(in bool) (bool, error) {
	ctx := x.GetCtx()
	if ctx == nil {
		return false, ErrMessageNotSocket
	}
	fdInfo := ctx.GetFdInfo()
	if fdInfo == nil {
		return false, ErrMessageNotSocket
	}
	return fdInfo.Role == in, nil
}

func (x *KindlingEvent) GetSocketKey() uint64 {
	return uint64(int64(x.Ctx.ThreadInfo.Pid)<<32) | uint64(x.Ctx.FdInfo.Num)&LOWER32
}

func (kv *KeyValue) GetUintValue() uint64 {
	switch kv.ValueType {
	case ValueType_UINT8:
		return uint64(kv.Value[0])
	case ValueType_UINT16:
		return uint64(byteOrder.Uint16(kv.Value))
	case ValueType_UINT32:
		return uint64(byteOrder.Uint32(kv.Value))
	case ValueType_UINT64:
		return byteOrder.Uint64(kv.Value)
	}
	return 0
}

func (kv *KeyValue) GetIntValue() int64 {
	switch kv.ValueType {
	case ValueType_INT8:
		return int64(int8(kv.Value[0]))
	case ValueType_INT16:
		return int64(int16(byteOrder.Uint16(kv.Value)))
	case ValueType_INT32:
		return int64(int32(byteOrder.Uint32(kv.Value)))
	case ValueType_INT64:
		return int64(byteOrder.Uint64(kv.Value))
	}
	return 0
}
