package cgoreceiver

/*
#cgo LDFLAGS: -L ./ -lkindling  -lstdc++ -ldl
#cgo CFLAGS: -I .
#include <stdlib.h>
#include <stdint.h>
#include <stdio.h>
#include "cgo_func.h"
*/
import "C"
import (
	"fmt"
	analyzerpackage "github.com/Kindling-project/kindling/collector/analyzer"
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/model"
	"time"

	//"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/receiver"
	"sync"
	"unsafe"
)

type Config struct {
}
const (
	Cgo = "cgoreceiver"
)
var cnt int
type CKindlingEventForGo C.struct_kindling_event_t_for_go

type CgoReceiver struct {
	cfg             *Config
	analyzerManager analyzerpackage.Manager
	shutdownWG      sync.WaitGroup
	shutdwonState   bool
	telemetry       *component.TelemetryTools
}
func (r *CgoReceiver) Start() error {
	var err error
	cnt = 0
	r.telemetry.Logger.Info("startCgoReceiver")
	go C.runForGo()
	go static()
	time.Sleep(2 * time.Second)
	r.startGetEvent()
	time.Sleep(1000*time.Second)
	if err != nil {
		return err
	}
	return err
}

func NewCgoReceiver(config interface{}, telemetry *component.TelemetryTools, analyzerManager analyzerpackage.Manager) receiver.Receiver {
	cfg, ok := config.(*Config)
	if !ok {
		telemetry.Logger.Sugar().Panicf("Cannot convert [%s] config", Cgo)
	}
	cgoReceiver := &CgoReceiver{
		cfg:             cfg,
		analyzerManager: analyzerManager,
		telemetry:       telemetry,
	}
	return cgoReceiver
}

func static() {

	timer := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timer.C:
			fmt.Println("cnt:")
			fmt.Println(cnt)
			cnt = 0
		}
	}

}

func(r *CgoReceiver) startGetEvent() {

	var pKindlingEvent unsafe.Pointer
	for {
		res := int(C.getKindlingEvent(&pKindlingEvent))
		if res != 0 {
			cnt++
			ev := new(model.CgoKindlingEvent)
			ev.Timestamp = uint64((*CKindlingEventForGo)(pKindlingEvent).timestamp)
			ev.Name = C.GoString((*CKindlingEventForGo)(pKindlingEvent).name)
			ev.Category = uint32((*CKindlingEventForGo)(pKindlingEvent).category)
			ev.Context.ThreadInfo.Pid = uint32((*CKindlingEventForGo)(pKindlingEvent).context.tinfo.pid)
			ev.Context.ThreadInfo.Tid = uint32((*CKindlingEventForGo)(pKindlingEvent).context.tinfo.tid)
			ev.Context.ThreadInfo.Uid = uint32((*CKindlingEventForGo)(pKindlingEvent).context.tinfo.uid)
			ev.Context.ThreadInfo.Gid = uint32((*CKindlingEventForGo)(pKindlingEvent).context.tinfo.gid)
			ev.Context.ThreadInfo.Comm = C.GoString((*CKindlingEventForGo)(pKindlingEvent).context.tinfo.comm)
			ev.Context.ThreadInfo.ContainerId = C.GoString((*CKindlingEventForGo)(pKindlingEvent).context.tinfo.containerId)
			ev.Context.FdInfo.Protocol = uint32((*CKindlingEventForGo)(pKindlingEvent).context.fdInfo.protocol)
			ev.Context.FdInfo.Num = uint32((*CKindlingEventForGo)(pKindlingEvent).context.fdInfo.num)
			ev.Context.FdInfo.FdType = uint32((*CKindlingEventForGo)(pKindlingEvent).context.fdInfo.fdType)
			ev.Context.FdInfo.Filename = C.GoString((*CKindlingEventForGo)(pKindlingEvent).context.fdInfo.filename)
			ev.Context.FdInfo.Directory = C.GoString((*CKindlingEventForGo)(pKindlingEvent).context.fdInfo.directory)
			ev.Context.FdInfo.Role = uint8((*CKindlingEventForGo)(pKindlingEvent).context.fdInfo.role)
			ev.Context.FdInfo.Sip = uint32((*CKindlingEventForGo)(pKindlingEvent).context.fdInfo.sip)
			ev.Context.FdInfo.Dip = uint32((*CKindlingEventForGo)(pKindlingEvent).context.fdInfo.dip)
			ev.Context.FdInfo.Sport = uint32((*CKindlingEventForGo)(pKindlingEvent).context.fdInfo.sport)
			ev.Context.FdInfo.Dport = uint32((*CKindlingEventForGo)(pKindlingEvent).context.fdInfo.dport)
			ev.Context.FdInfo.Source = uint64((*CKindlingEventForGo)(pKindlingEvent).context.fdInfo.source)
			ev.Context.FdInfo.Destination = uint64((*CKindlingEventForGo)(pKindlingEvent).context.fdInfo.destination)
			for i := 0; i < 8; i++ {
				ev.UserAttributes[i].Key = C.GoString((*CKindlingEventForGo)(pKindlingEvent).userAttributes[i].key)
				if len(ev.UserAttributes[i].Key) == 0 {
					break
				}
				ev.UserAttributes[i].Value = C.GoString((*CKindlingEventForGo)(pKindlingEvent).userAttributes[i].value)
				ev.UserAttributes[i].ValueType = uint32((*CKindlingEventForGo)(pKindlingEvent).userAttributes[i].valueType)
			}
		}
	}

}

func (r *CgoReceiver) Shutdown() error {
	var err error
	r.shutdownWG.Wait()
	return err
}

