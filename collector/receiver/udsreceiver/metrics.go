package udsreceiver

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"sync"
	"sync/atomic"
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
	grpcUprobe        int64
	tcpClose          int64
	tcpRcvEstablished int64
	tcpDrop           int64
	tcpRetransmitSkb  int64
	other             int64
}

func (i *stats) add(name string, value int64) {
	switch name {
	case "read":
		atomic.AddInt64(&i.read, value)
	case "write":
		atomic.AddInt64(&i.write, value)
	case "readv":
		atomic.AddInt64(&i.readv, value)
	case "writev":
		atomic.AddInt64(&i.writev, value)
	case "sendto":
		atomic.AddInt64(&i.sendTo, value)
	case "recvfrom":
		atomic.AddInt64(&i.recvFrom, value)
	case "sendmsg":
		atomic.AddInt64(&i.sendMsg, value)
	case "recvmsg":
		atomic.AddInt64(&i.recvMsg, value)
	case "grpc_uprobe":
		atomic.AddInt64(&i.grpcUprobe, value)
	case "tcp_close":
		atomic.AddInt64(&i.tcpClose, value)
	case "tcp_rcv_established":
		atomic.AddInt64(&i.tcpRcvEstablished, value)
	case "tcp_drop":
		atomic.AddInt64(&i.tcpDrop, value)
	case "tcp_retransmit_skb":
		atomic.AddInt64(&i.tcpRetransmitSkb, value)
	default:
		atomic.AddInt64(&i.other, value)
	}
}

func (i *stats) getStats() map[string]int64 {
	ret := make(map[string]int64, 14)
	ret["read"] = atomic.LoadInt64(&i.read)
	ret["write"] = atomic.LoadInt64(&i.write)
	ret["readv"] = atomic.LoadInt64(&i.readv)
	ret["writev"] = atomic.LoadInt64(&i.writev)
	ret["sendto"] = atomic.LoadInt64(&i.sendTo)
	ret["recvfrom"] = atomic.LoadInt64(&i.recvFrom)
	ret["sendmsg"] = atomic.LoadInt64(&i.sendMsg)
	ret["recvmsg"] = atomic.LoadInt64(&i.recvMsg)
	ret["grpc_uprobe"] = atomic.LoadInt64(&i.grpcUprobe)
	ret["tcp_close"] = atomic.LoadInt64(&i.tcpClose)
	ret["tcp_rcv_established"] = atomic.LoadInt64(&i.tcpRcvEstablished)
	ret["tcp_drop"] = atomic.LoadInt64(&i.tcpClose)
	ret["tcp_retransmit_skb"] = atomic.LoadInt64(&i.tcpRetransmitSkb)
	ret["other"] = atomic.LoadInt64(&i.other)
	return ret
}
