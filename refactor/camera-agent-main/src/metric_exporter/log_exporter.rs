use crate::metric_exporter::MetricExporter;
use log::info;
use opentelemetry_api::KeyValue;
pub struct LogExporter {}

impl LogExporter {
    pub fn new() -> Self {
        return LogExporter {};
    }
}
impl MetricExporter for LogExporter {
    fn record_metric(&self, _name: &str, _value: u64, _attributes: &[KeyValue]) {
        info!("Record the metric: name={}, value={}", _name, _value)
    }
}
