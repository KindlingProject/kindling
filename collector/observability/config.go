package observability

import "time"

type Config struct {
	ExportKind  string            `mapstructure:"export_kind"`
	PromCfg     *PrometheusConfig `mapstructure:"prometheus"`
	OtlpGrpcCfg *OtlpGrpcConfig   `mapstructure:"otlp"`
	StdoutCfg   *StdoutConfig     `mapstructure:"stdout"`
}

type PrometheusConfig struct {
	Port string `mapstructure:"port,omitempty"`
}

type OtlpGrpcConfig struct {
	CollectPeriod time.Duration `mapstructure:"collect_period,omitempty"`
	Endpoint      string        `mapstructure:"endpoint,omitempty"`
}

type StdoutConfig struct {
	CollectPeriod time.Duration `mapstructure:"collect_period,omitempty"`
}

var DefaultConfig = Config{
	ExportKind:  "stdout",
	PromCfg:     nil,
	OtlpGrpcCfg: nil,
	StdoutCfg:   &StdoutConfig{CollectPeriod: 10 * time.Second},
}
