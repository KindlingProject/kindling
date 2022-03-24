module github.com/Kindling-project/kindling/collector

go 1.16

require (
	github.com/golang/protobuf v1.5.2
	github.com/hashicorp/go-multierror v1.1.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/mdlayher/netlink v1.5.0
	github.com/orcaman/concurrent-map v0.0.0-20210501183033-44dafcb38ecc
	github.com/pebbe/zmq4 v1.2.7
	github.com/spf13/viper v1.10.1
	github.com/stretchr/testify v1.7.0
	github.com/sugary199/collector-version v0.0.0-20220323082224-e39a32dfb492
	github.com/ti-mo/conntrack v0.4.0
	github.com/ti-mo/netfilter v0.3.1
	github.com/vishvananda/netns v0.0.0-20211101163701-50045581ed74 // indirect
	go.opentelemetry.io/otel v1.2.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.25.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.2.0
	go.opentelemetry.io/otel/exporters/prometheus v0.25.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v0.25.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.2.0
	go.opentelemetry.io/otel/metric v0.25.0
	go.opentelemetry.io/otel/sdk v1.2.0
	go.opentelemetry.io/otel/sdk/export/metric v0.25.0
	go.opentelemetry.io/otel/sdk/metric v0.25.0
	go.opentelemetry.io/otel/trace v1.2.0
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0
	go.uber.org/zap v1.17.0
	golang.org/x/sys v0.0.0-20211210111614-af8b64212486
	google.golang.org/protobuf v1.27.1
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.21.1
)
