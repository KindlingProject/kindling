use tokio::runtime::Runtime;    

pub(crate) mod log_exporter;
use crate::model::consts;
use opentelemetry_api::metrics::{Counter, Histogram};
use opentelemetry_api::{Context, KeyValue};
use opentelemetry_prometheus::PrometheusExporter;
use opentelemetry_sdk::{
    export::metrics::aggregation,
    metrics::{controllers, processors, selectors},
};

pub trait MetricExporter: Send + Sync {
    fn record_metric(&self, name: &str, value: u64, attributes: &[KeyValue]);
}

pub struct PromExporter {
    pub exporter: PrometheusExporter,
    pub runtime: Runtime,
    // meter: opentelemetry_api::metrics::Meter,
    // instrument_map: HashMap<String, Counter<u64>>,
    profiling_on_duration: Histogram<u64>,
    profiling_net_duration: Histogram<u64>,
    profiling_file_duration: Histogram<u64>,
    profiling_futex_duration: Histogram<u64>,
    profiling_epoll_duration: Histogram<u64>,
    profiling_runq_duration: Histogram<u64>,
    profiling_on_counter: Counter<u64>,
    profiling_net_counter: Counter<u64>,
    profiling_file_counter: Counter<u64>,
    profiling_futex_counter: Counter<u64>,
    profiling_epoll_counter: Counter<u64>,
    trace_histogram: Histogram<u64>,
}

impl PromExporter {
    pub fn new() -> Self {
        let controller = controllers::basic(processors::factory(
            selectors::simple::histogram([
                5000000.0,
                10000000.0,
                20000000.0,
                30000000.0,
                50000000.0,
                80000000.0,
                100000000.0,
                150000000.0,
                200000000.0,
                300000000.0,
                400000000.0,
                500000000.0,
                800000000.0,
                1200000000.0,
                3000000000.0,
                5000000000.0,
            ]),
            aggregation::cumulative_temporality_selector(),
        ))
        .build();

        let exporter = opentelemetry_prometheus::exporter(controller).init();
        let meter = opentelemetry_api::global::meter("kindling");
        let runtime = tokio::runtime::Builder::new_multi_thread()
            .enable_all()
            .build()
            .unwrap();

        PromExporter {
            exporter,
            runtime,
            profiling_on_duration: meter
                .u64_histogram(consts::PROFILING_CPU_DURATION_METRIC_NAME)
                .init(),
            profiling_net_duration: meter
                .u64_histogram(consts::PROFILING_NET_DURATION_METRIC_NAME)
                .init(),
            profiling_file_duration: meter
                .u64_histogram(consts::PROFILING_FILE_DURATION_METRIC_NAME)
                .init(),
            profiling_futex_duration: meter
                .u64_histogram(consts::PROFILING_FUTEX_DURATION_METRIC_NAME)
                .init(),
            profiling_epoll_duration: meter
                .u64_histogram(consts::PROFILING_EPOLL_DURATION_METRIC_NAME)
                .init(),
            profiling_runq_duration: meter
                .u64_histogram(consts::PROFILING_RUNQ_DURATION_METRIC_NAME)
                .init(),
            profiling_on_counter: meter
                .u64_counter(consts::PROFILING_CPU_COUNT_METRIC_NAME)
                .init(),
            profiling_net_counter: meter
                .u64_counter(consts::PROFILING_NET_COUNT_METRIC_NAME)
                .init(),
            profiling_file_counter: meter
                .u64_counter(consts::PROFILING_FILE_COUNT_METRIC_NAME)
                .init(),
            profiling_futex_counter: meter
                .u64_counter(consts::PROFILING_FUTEX_COUNT_METRIC_NAME)
                .init(),
            profiling_epoll_counter: meter
                .u64_counter(consts::PROFILING_EPOLL_COUNT_METRIC_NAME)
                .init(),
            trace_histogram: meter.u64_histogram(consts::TRACE_HISTOGRAM_NAME).init(),
        }
    }
}

impl MetricExporter for PromExporter {
    fn record_metric(&self, name: &str, value: u64, attributes: &[KeyValue]) {
        let cx = Context::new();
        match name {
            consts::PROFILING_CPU_DURATION_METRIC_NAME => {
                self.profiling_on_duration.record(&cx, value, attributes)
            }
            consts::PROFILING_NET_DURATION_METRIC_NAME => {
                self.profiling_net_duration.record(&cx, value, attributes)
            }
            consts::PROFILING_FILE_DURATION_METRIC_NAME => {
                self.profiling_file_duration.record(&cx, value, attributes)
            }
            consts::PROFILING_FUTEX_DURATION_METRIC_NAME => {
                self.profiling_futex_duration.record(&cx, value, attributes)
            }
            consts::PROFILING_EPOLL_DURATION_METRIC_NAME => {
                self.profiling_epoll_duration.record(&cx, value, attributes)
            }
            consts::PROFILING_RUNQ_DURATION_METRIC_NAME => {
                self.profiling_runq_duration.record(&cx, value, attributes)
            }
            consts::PROFILING_CPU_COUNT_METRIC_NAME => {
                self.profiling_on_counter.add(&cx, value, attributes)
            }
            consts::PROFILING_NET_COUNT_METRIC_NAME => {
                self.profiling_net_counter.add(&cx, value, attributes)
            }
            consts::PROFILING_FILE_COUNT_METRIC_NAME => {
                self.profiling_file_counter.add(&cx, value, attributes)
            }
            consts::PROFILING_FUTEX_COUNT_METRIC_NAME => {
                self.profiling_futex_counter.add(&cx, value, attributes)
            }
            consts::PROFILING_EPOLL_COUNT_METRIC_NAME => {
                self.profiling_epoll_counter.add(&cx, value, attributes)
            }
            consts::TRACE_HISTOGRAM_NAME => self.trace_histogram.record(&cx, value, attributes),
            _ => {}
        }
    }
}

#[cfg(test)]
mod tests {
    use super::PromExporter;
    use crate::metric_exporter::MetricExporter;
    use crate::model::consts;
    use opentelemetry_api::KeyValue;
    use std::time::Duration;
    use std::{thread, time};

    #[test]
    fn prometheus() {
        let prom_exporter = PromExporter::new();
        let attributes = vec![KeyValue::new("a", "b"), KeyValue::new("c", 10)];
        for i in 0..10 {
            prom_exporter.record_metric(consts::PROFILING_CPU_DURATION_METRIC_NAME, i, &attributes);
            thread::sleep(time::Duration::from_millis(1));
        }
        println!("Hello, world!");
        thread::sleep(Duration::from_secs(20));
    }
}
