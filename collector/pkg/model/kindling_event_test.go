package model

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestUnmarshalEvent(t *testing.T) {
	v, _ := base64.StdEncoding.DecodeString("sd4btBwgHxc=")
	att1 := &KeyValue{
		Key:       "start_time",
		ValueType: 8,
		Value:     v,
	}
	assert.Equal(t, uint64(1666085694803271345), att1.GetUintValue())
	v, _ = base64.StdEncoding.DecodeString("19uyLHcgHxc=")
	att2 := &KeyValue{
		Key:       "end_time",
		ValueType: 8,
		Value:     v,
	}
	assert.Equal(t, uint64(1666086083373489111), att2.GetUintValue())
}

func TestTextEvent(t *testing.T) {
	withLogger(t, zap.InfoLevel, nil, func(l *zap.Logger, ol *observer.ObservedLogs) {
		v, _ := base64.StdEncoding.DecodeString("sd4btBwgHxc=")

		evt := &KindlingEvent{
			Source:         Source_TRACEPOINT,
			Timestamp:      160000,
			Name:           constnames.ReadEvent,
			Category:       Category_CAT_NET,
			ParamsNumber:   3,
			Latency:        100,
			UserAttributes: [16]KeyValue{{Key: "starttime", ValueType: ValueType_UINT64, Value: v}, {Key: "data", ValueType: ValueType_BYTEBUF, Value: []byte("/var /docker.dock")}},
			Ctx:            Context{ThreadInfo: Thread{Pid: 22592, Tid: 22592, Uid: 0, Gid: 0, Comm: "node", ContainerId: "123123", ContainerName: "test"}, FdInfo: Fd{Num: 19, TypeFd: 3, Filename: "tmp", Directory: "/root", Protocol: 1, Role: true, Sip: []uint32{16777343}, Dip: []uint32{16777343}, Sport: 39620, Dport: 37093, Source: 0, Destination: 0}},
		}
		l.Info(fmt.Sprintf("Event: %+v", TextKindlingEvent(evt)))
		assert.Equal(
			t,
			[]observer.LoggedEntry{
				{
					Entry: zapcore.Entry{
						Level:   zap.InfoLevel,
						Message: `Event: {Name:read Category:3 ParamsNumber:3 UserAttributes:[starttime:UINT(1666085694803271345) data:BYTEBUF(/var /docker.dock)] Ctx:{ThreadInfo:{Pid:22592 Tid:22592 Uid:0 Gid:0 Comm:node ContainerId:123123 ContainerName:test} FdInfo:{Num:19 TypeFd:3 Filename:tmp Directory:/root Protocol:1 Role:true Sip:127.0.0.1 Dip:127.0.0.1 Sport:39620 Dport:37093 Source:0 Destination:0}} Latency:100 Timestamp:160000}`,
					},
					Context: []zap.Field{},
				},
			},
			ol.AllUntimed(),
			"Unexpected log output format from KindlingEvent",
		)
	})
}

func withLogger(t testing.TB, e zapcore.LevelEnabler, zapOpt []zap.Option, f func(*zap.Logger, *observer.ObservedLogs)) {
	fac, logs := observer.New(e)
	logger := zap.New(fac, zapOpt...)
	f(logger, logs)
}
