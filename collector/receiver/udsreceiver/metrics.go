package udsreceiver

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/Kindling-project/kindling/collector/model/constnames"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var once sync.Once

const eventReceivedMetric = "kindling_telemetry_udsreceiver_events_total"

func newSelfMetrics(meterProvider metric.MeterProvider, counter eventCounter) {
	once.Do(func() {
		meter := metric.Must(meterProvider.Meter("kindling"))
		meter.NewInt64CounterObserver(eventReceivedMetric,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				for name, value := range counter.getStats() {
					result.Observe(value, attribute.String("name", name))
				}
			})
	})
}

type eventCounter interface {
	add(name string, value int64)
	getStats() map[string]int64
}

type stats struct {
	read              int64
	write             int64
	readv             int64
	writev            int64
	sendTo            int64
	recvFrom          int64
	sendMsg           int64
	recvMsg           int64
	connect           int64
	grpcUprobe        int64
	tcpClose          int64
	tcpRcvEstablished int64
	tcpDrop           int64
	tcpRetransmitSkb  int64
	tcpConnect        int64
	tcpSetState       int64
	other             int64
}

func (i *stats) add(name string, value int64) {
	switch name {
	case constnames.ReadEvent:
		atomic.AddInt64(&i.read, value)
	case constnames.WriteEvent:
		atomic.AddInt64(&i.write, value)
	case constnames.ReadvEvent:
		atomic.AddInt64(&i.readv, value)
	case constnames.WritevEvent:
		atomic.AddInt64(&i.writev, value)
	case constnames.SendToEvent:
		atomic.AddInt64(&i.sendTo, value)
	case constnames.RecvFromEvent:
		atomic.AddInt64(&i.recvFrom, value)
	case constnames.SendMsgEvent:
		atomic.AddInt64(&i.sendMsg, value)
	case constnames.RecvMsgEvent:
		atomic.AddInt64(&i.recvMsg, value)
	case constnames.GrpcUprobeEvent:
		atomic.AddInt64(&i.grpcUprobe, value)
	case constnames.TcpCloseEvent:
		atomic.AddInt64(&i.tcpClose, value)
	case constnames.TcpRcvEstablishedEvent:
		atomic.AddInt64(&i.tcpRcvEstablished, value)
	case constnames.TcpDropEvent:
		atomic.AddInt64(&i.tcpDrop, value)
	case constnames.TcpRetransmitSkbEvent:
		atomic.AddInt64(&i.tcpRetransmitSkb, value)
	case constnames.ConnectEvent:
		atomic.AddInt64(&i.connect, value)
	case constnames.TcpConnectEvent:
		atomic.AddInt64(&i.tcpConnect, value)
	case constnames.TcpSetStateEvent:
		atomic.AddInt64(&i.tcpSetState, value)
	default:
		atomic.AddInt64(&i.other, value)
	}
}

func (i *stats) getStats() map[string]int64 {
	ret := make(map[string]int64, 14)
	ret[constnames.ReadEvent] = atomic.LoadInt64(&i.read)
	ret[constnames.WriteEvent] = atomic.LoadInt64(&i.write)
	ret[constnames.ReadvEvent] = atomic.LoadInt64(&i.readv)
	ret[constnames.WritevEvent] = atomic.LoadInt64(&i.writev)
	ret[constnames.SendToEvent] = atomic.LoadInt64(&i.sendTo)
	ret[constnames.RecvFromEvent] = atomic.LoadInt64(&i.recvFrom)
	ret[constnames.SendMsgEvent] = atomic.LoadInt64(&i.sendMsg)
	ret[constnames.RecvMsgEvent] = atomic.LoadInt64(&i.recvMsg)
	ret[constnames.GrpcUprobeEvent] = atomic.LoadInt64(&i.grpcUprobe)
	ret[constnames.TcpCloseEvent] = atomic.LoadInt64(&i.tcpClose)
	ret[constnames.TcpRcvEstablishedEvent] = atomic.LoadInt64(&i.tcpRcvEstablished)
	ret[constnames.TcpCloseEvent] = atomic.LoadInt64(&i.tcpClose)
	ret[constnames.TcpRetransmitSkbEvent] = atomic.LoadInt64(&i.tcpRetransmitSkb)
	ret[constnames.ConnectEvent] = atomic.LoadInt64(&i.connect)
	ret[constnames.TcpConnectEvent] = atomic.LoadInt64(&i.tcpConnect)
	ret[constnames.TcpSetStateEvent] = atomic.LoadInt64(&i.tcpSetState)
	ret[constnames.OtherEvent] = atomic.LoadInt64(&i.other)
	return ret
}
