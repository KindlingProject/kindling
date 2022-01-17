package model

import (
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
)

func (x *KindlingEvent) GetData() []byte {
	keyValue := x.GetUserAttribute("data")
	if keyValue != nil {
		return keyValue.GetValue().GetBytesValue()
	}
	return nil
}

func (x *KindlingEvent) GetDataLen() int {
	keyValue := x.GetUserAttribute("data")
	if keyValue != nil {
		return len(keyValue.GetValue().GetBytesValue())
	}
	return 0
}

func (x *KindlingEvent) GetResVal() int64 {
	keyValue := x.GetUserAttribute("res")
	if keyValue != nil {
		return keyValue.GetValue().GetIntValue()
	}
	return -1
}

func (x *KindlingEvent) GetLatency() uint64 {
	keyValue := x.GetUserAttribute("latency")
	if keyValue != nil {
		return uint64(keyValue.GetValue().GetIntValue())
	}
	return 0
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
