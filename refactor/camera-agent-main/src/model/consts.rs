pub(crate) const TIMESTAMP: &str = "timestamp";
pub(crate) const PID: &str = "pid";
pub(crate) const TID: &str = "tid";
pub(crate) const PROTOCOL: &str = "protocol";
pub(crate) const CONTENT_KEY: &str = "content_key";
pub(crate) const IS_SLOW: &str = "is_slow";
pub(crate) const IS_SERVER: &str = "is_server";
pub(crate) const IS_ERROR: &str = "is_error";
pub(crate) const IS_PROFILED: &str = "is_profiled";
pub(crate) const P90: &str = "p90";
pub(crate) const TRACE_ID: &str = "trace_id";
pub(crate) const APM_SPAN_ID: &str = "apm_span_id";
pub(crate) const CONTAINER_ID: &str = "container_id";
pub(crate) const CONTAINER_NAME: &str = "container_name";
pub(crate) const WORKLOAD_NAME: &str = "workload_name";
pub(crate) const WORKLOAD_KIND: &str = "workload_kind";
pub(crate) const POD_IP: &str = "pod_ip";
pub(crate) const POD_NAME: &str = "pod_name";
pub(crate) const NODE_NAME: &str = "node_name";
pub(crate) const NAMESPACE: &str = "namespace";
pub(crate) const DURATION: &str = "duration";
pub(crate) const END_TIME: &str = "end_time";

pub(crate) const TRACE_HISTOGRAM_NAME: &str = "kindling_span_trace_duration_nanoseconds";

pub(crate) const PROFILING_CPU_PREFIX: &str = "kindling_profiling_cpu";
pub(crate) const PROFILING_NET_PREFIX: &str = "kindling_profiling_net";
pub(crate) const PROFILING_FILE_PREFIX: &str = "kindling_profiling_file";
pub(crate) const PROFILING_FUTEX_PREFIX: &str = "kindling_profiling_futex";
pub(crate) const PROFILING_IDLE_PREFIX: &str = "kindling_profiling_idle";
pub(crate) const PROFILING_OTHER_PREFIX: &str = "kindling_profiling_other";
pub(crate) const PROFILING_EPOLL_PREFIX: &str = "kindling_profiling_epoll";
pub(crate) const PROFILING_RUNQ_PREFIX: &str = "kindling_profiling_runq";

pub(crate) const PROFILING_DURATION_SUFFIX: &str = "duration_nanoseconds";
pub(crate) const PROFILING_COUNT_SUFFIX: &str = "count";

pub(crate) const PROFILING_CPU_DURATION_METRIC_NAME: &str =
    "kindling_profiling_cpu_duration_nanoseconds";
pub(crate) const PROFILING_NET_DURATION_METRIC_NAME: &str =
    "kindling_profiling_net_duration_nanoseconds";
pub(crate) const PROFILING_FILE_DURATION_METRIC_NAME: &str =
    "kindling_profiling_file_duration_nanoseconds";
pub(crate) const PROFILING_FUTEX_DURATION_METRIC_NAME: &str =
    "kindling_profiling_futex_duration_nanoseconds";
pub(crate) const PROFILING_IDLE_DURATION_METRIC_NAME: &str =
    "kindling_profiling_idle_duration_nanoseconds";
pub(crate) const PROFILING_OTHER_DURATION_METRIC_NAME: &str =
    "kindling_profiling_other_duration_nanoseconds";
pub(crate) const PROFILING_EPOLL_DURATION_METRIC_NAME: &str =
    "kindling_profiling_epoll_duration_nanoseconds";
pub(crate) const PROFILING_RUNQ_DURATION_METRIC_NAME: &str =
    "kindling_profiling_runq_duration_nanoseconds";

pub(crate) const PROFILING_CPU_COUNT_METRIC_NAME: &str = "kindling_profiling_cpu_count";
pub(crate) const PROFILING_NET_COUNT_METRIC_NAME: &str = "kindling_profiling_net_count";
pub(crate) const PROFILING_FILE_COUNT_METRIC_NAME: &str = "kindling_profiling_file_count";
pub(crate) const PROFILING_FUTEX_COUNT_METRIC_NAME: &str = "kindling_profiling_futex_count";
pub(crate) const PROFILING_IDLE_COUNT_METRIC_NAME: &str = "kindling_profiling_idle_count";
pub(crate) const PROFILING_OTHER_COUNT_METRIC_NAME: &str = "kindling_profiling_other_count";
pub(crate) const PROFILING_EPOLL_COUNT_METRIC_NAME: &str = "kindling_profiling_epoll_count";
