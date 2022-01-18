package udsreceiver

import (
	"github.com/Kindling-project/kindling/collector/analyzer"
	"github.com/Kindling-project/kindling/collector/logger"
	"testing"
)

func TestUdsReceiver_Start(t *testing.T) {
	cfg := &Config{
		ZEROMQPULL: &ZeroMqPullSettings{
			Endpoint: "ipc:///home/kindling/0",
		},
		ZEROMQREQ: &ZeroMqReqSettings{
			Endpoint: "ipc:///home/kindling/0",
			SubcribeInfo: []SubEvent{
				{
					Name:     "syscall_exit-write",
					Category: "net",
				},
				{
					Name:     "syscall_exit-read",
					Category: "net",
				},
				{
					Name:     "syscall_exit-sendto",
					Category: "net",
				},
				{
					Name:     "syscall_exit-recvform",
					Category: "net",
				},
			},
		},
	}
	am, _ := analyzer.NewManager()
	r := NewUdsReceiver(cfg, logger.CreateDefaultLogger(), am)
	r.Start()
}
