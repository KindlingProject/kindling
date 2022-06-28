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
	"sync"
	"time"
	"unsafe"

	analyzerpackage "github.com/Kindling-project/kindling/collector/analyzer"
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/model"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Kindling-project/kindling/collector/receiver"
)

const (
	Cgo = "cgoreceiver"
)

type CKindlingEventForGo C.struct_kindling_event_t_for_go

type CgoReceiver struct {
	cfg             *Config
	analyzerManager *analyzerpackage.Manager
	shutdownWG      sync.WaitGroup
	telemetry       *component.TelemetryTools
	eventChannel    chan *model.KindlingEvent
	eventCount      int
	stopCh          chan interface{}
	stats           eventCounter
}

func NewCgoReceiver(config interface{}, telemetry *component.TelemetryTools, analyzerManager *analyzerpackage.Manager) receiver.Receiver {
	cfg, ok := config.(*Config)
	if !ok {
		telemetry.Logger.Sugar().Panicf("Cannot convert [%s] config", Cgo)
	}
	cgoReceiver := &CgoReceiver{
		cfg:             cfg,
		analyzerManager: analyzerManager,
		telemetry:       telemetry,
		eventChannel:    make(chan *model.KindlingEvent, 3e5),
		stopCh:          make(chan interface{}, 1),
	}
	cgoReceiver.stats = newDynamicStats(cfg.SubscribeInfo)
	newSelfMetrics(telemetry.MeterProvider, cgoReceiver.stats)
	return cgoReceiver
}

func (r *CgoReceiver) Start() error {
	r.telemetry.Logger.Info("Start CgoReceiver")
	C.runForGo()
	go r.printMetrics()
	time.Sleep(2 * time.Second)
	r.subEvent()
	// Wait for the C routine running
	time.Sleep(2 * time.Second)
	go r.consumeEvents()
	go r.startGetEvent()
	return nil
}

// TODO finish it using opentelemetry
func (r *CgoReceiver) printMetrics() {
	timer := time.NewTicker(1 * time.Second)
	r.shutdownWG.Add(1)
	for {
		select {
		case <-r.stopCh:
			r.shutdownWG.Done()
			return
		case <-timer.C:
			r.telemetry.Logger.Info("Total number events received: ", zap.Int("events", r.eventCount))
			r.eventCount = 0
			r.telemetry.Logger.Info("Current channel size: ", zap.Int("channel size", len(r.eventChannel)))
		}
	}
}

func (r *CgoReceiver) startGetEvent() {
	var pKindlingEvent unsafe.Pointer
	r.shutdownWG.Add(1)
	for {
		select {
		case <-r.stopCh:
			r.shutdownWG.Done()
			return
		default:
			res := int(C.getKindlingEvent(&pKindlingEvent))
			if res == 1 {
				r.eventCount++
				r.eventChannel <- convertEvent((*CKindlingEventForGo)(pKindlingEvent))
			}
		}
	}
}

func (r *CgoReceiver) consumeEvents() {
	r.shutdownWG.Add(1)
	for {
		select {
		case <-r.stopCh:
			r.shutdownWG.Done()
			return
		case ev := <-r.eventChannel:
			err := r.sendToNextConsumer(ev)
			if err != nil {
				r.telemetry.Logger.Info("Failed to send KindlingEvent: ", zap.Error(err))
			}
		}
	}
}

func (r *CgoReceiver) Shutdown() error {
	// TODO stop the C routine
	close(r.stopCh)
	r.shutdownWG.Wait()
	return nil
}

func convertEvent(cgoEvent *CKindlingEventForGo) *model.KindlingEvent {
	ev := new(model.KindlingEvent)
	ev.Timestamp = uint64(cgoEvent.timestamp)
	ev.Name = C.GoString(cgoEvent.name)
	ev.Category = model.Category(cgoEvent.category)
	ev.Ctx.ThreadInfo.Pid = uint32(cgoEvent.context.tinfo.pid)
	ev.Ctx.ThreadInfo.Tid = uint32(cgoEvent.context.tinfo.tid)
	ev.Ctx.ThreadInfo.Uid = uint32(cgoEvent.context.tinfo.uid)
	ev.Ctx.ThreadInfo.Gid = uint32(cgoEvent.context.tinfo.gid)
	ev.Ctx.ThreadInfo.Comm = C.GoString(cgoEvent.context.tinfo.comm)
	ev.Ctx.ThreadInfo.ContainerId = C.GoString(cgoEvent.context.tinfo.containerId)
	ev.Ctx.FdInfo.Protocol = model.L4Proto(cgoEvent.context.fdInfo.protocol)
	ev.Ctx.FdInfo.Num = int32(cgoEvent.context.fdInfo.num)
	ev.Ctx.FdInfo.TypeFd = model.FDType(cgoEvent.context.fdInfo.fdType)
	ev.Ctx.FdInfo.Filename = C.GoString(cgoEvent.context.fdInfo.filename)
	ev.Ctx.FdInfo.Directory = C.GoString(cgoEvent.context.fdInfo.directory)
	ev.Ctx.FdInfo.Role = If(cgoEvent.context.fdInfo.role != 0, true, false).(bool)
	ev.Ctx.FdInfo.Sip = []uint32{uint32(cgoEvent.context.fdInfo.sip)}
	ev.Ctx.FdInfo.Dip = []uint32{uint32(cgoEvent.context.fdInfo.dip)}
	ev.Ctx.FdInfo.Sport = uint32(cgoEvent.context.fdInfo.sport)
	ev.Ctx.FdInfo.Dport = uint32(cgoEvent.context.fdInfo.dport)
	ev.Ctx.FdInfo.Source = uint64(cgoEvent.context.fdInfo.source)
	ev.Ctx.FdInfo.Destination = uint64(cgoEvent.context.fdInfo.destination)

	ev.ParamsNumber = uint16(cgoEvent.paramsNumber)
	for i := 0; i < int(ev.ParamsNumber); i++ {
		ev.UserAttributes[i].Key = C.GoString(cgoEvent.userAttributes[i].key)
		userAttributesLen := cgoEvent.userAttributes[i].len
		ev.UserAttributes[i].Value = C.GoBytes(unsafe.Pointer(cgoEvent.userAttributes[i].value), C.int(userAttributesLen))
		ev.UserAttributes[i].ValueType = model.ValueType(cgoEvent.userAttributes[i].valueType)
	}
	return ev
}

func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

func (r *CgoReceiver) sendToNextConsumer(evt *model.KindlingEvent) error {

	if ce := r.telemetry.Logger.Check(zapcore.DebugLevel, "Receive Event"); ce != nil {
		ce.Write(
			zap.String("event", evt.String()),
		)
	}
	analyzers := r.analyzerManager.GetConsumableAnalyzers(evt.Name)
	if analyzers == nil || len(analyzers) == 0 {
		r.telemetry.Logger.Info("analyzer not found for event ", zap.String("eventName", evt.Name))
		return nil
	}
	for _, analyzer := range analyzers {
		err := analyzer.ConsumeEvent(evt)
		if err != nil {
			r.telemetry.Logger.Warn("Error sending event to next consumer: ", zap.Error(err))
		}
	}
	return nil
}

func (r *CgoReceiver) subEvent() {
	for _, value := range r.cfg.SubscribeInfo {
		C.subEventForGo(C.CString(value.Name), C.CString(value.Category))
	}

}
