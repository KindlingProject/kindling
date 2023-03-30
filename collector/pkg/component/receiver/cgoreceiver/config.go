package cgoreceiver

type Config struct {
	SubscribeInfo     []SubEvent    `mapstructure:"subscribe"`
	ProcessFilterInfo ProcessFilter `mapstructure:"process_filter"`
}

type SubEvent struct {
	Category string            `mapstructure:"category"`
	Name     string            `mapstructure:"name"`
	Params   map[string]string `mapstructure:"params"`
}

type ProcessFilter struct {
	Comms []string `mapstructure:"comms"`
}

func NewDefaultConfig() *Config {
	return &Config{
		SubscribeInfo: []SubEvent{
			{
				Name:     "syscall_exit-writev",
				Category: "net",
			},
			{
				Name:     "syscall_exit-readv",
				Category: "net",
			},
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
				Name:     "syscall_exit-recvfrom",
				Category: "net",
			},
			{
				Name:     "syscall_exit-sendmsg",
				Category: "net",
			},
			{
				Name:     "syscall_exit-recvmsg",
				Category: "net",
			},
			{
				Name:     "syscall_exit-sendmmsg",
				Category: "net",
			},
			{
				Name: "kprobe-tcp_close",
			},
			{
				Name: "kprobe-tcp_rcv_established",
			},
			{
				Name: "kprobe-tcp_drop",
			},
			{
				Name: "kprobe-tcp_retransmit_skb",
			},
			{
				Name: "syscall_exit-connect",
			},
			{
				Name: "kretprobe-tcp_connect",
			},
			{
				Name: "kprobe-tcp_set_state",
			},
			{
				Name: "tracepoint-procexit",
			},
		},
		ProcessFilterInfo: ProcessFilter{
			Comms: []string{"kindling-collec", "containerd", "dockerd", "containerd-shim"},
		},
	}
}
