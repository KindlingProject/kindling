package kindlingformatprocessor

type Config struct {
	NeedTraceAsMetric  bool `mapstructure:"need_trace_as_metric"`
	NeedPodDetail      bool `mapstructure:"need_pod_detail"`
	StoreExternalSrcIP bool `mapstructure:"store_external_src_ip"`
}
