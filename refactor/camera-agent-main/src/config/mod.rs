use log::info;
use serde::{Deserialize, Serialize};
use std::fs::File;

#[derive(Serialize, Deserialize, Debug)]
pub struct GlobalConfig {
    pub analyzers: Analyzers,
    pub processors: Processors,
    pub exporters: Exporters,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct Analyzers {
    #[serde(rename = "traceidanalyzer")]
    pub trace_id_analyzer: TraceIdAnalyzerCfg,
    #[serde(rename = "cpuanalyzer")]
    pub cpu_analyzer: CpuAnalyzerCfg,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct CpuAnalyzerCfg {
    #[serde(default = "default_segment_size")]
    pub segment_size: Option<i32>,
}

fn default_segment_size() -> Option<i32> {
    Some(40)
}

#[derive(Serialize, Deserialize, Copy, Clone, Debug)]
pub struct TraceIdAnalyzerCfg {
    pub open_java_trace_sampling: bool,
    #[serde(default = "default_java_trace_slow_time")]
    pub java_trace_slow_time: Option<u64>,
    #[serde(default = "default_java_trace_wait_second")]
    pub java_trace_wait_second: Option<i32>,
}

fn default_java_trace_slow_time() -> Option<u64> {
    Some(500)
}

fn default_java_trace_wait_second() -> Option<i32> {
    Some(30)
}

#[derive(Serialize, Deserialize, Debug)]
pub struct Processors {
    #[serde(rename = "sampleprocessor")]
    pub sample_processor: SampleProcessorCfg,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct SampleProcessorCfg {
    #[serde(default = "default_sample_trace_repeat_num")]
    pub sample_trace_repeat_num: Option<i32>,
    #[serde(default = "default_sample_trace_wait_time")]
    pub sample_trace_wait_time: Option<u64>,
    #[serde(default = "default_sample_url_hit_duration")]
    pub sample_url_hit_duration: Option<u64>,
    #[serde(default = "default_sample_trace_ignore_threshold")]
    pub sample_trace_ignore_threshold: Option<u64>,
    #[serde(default = "default_sample_trace_threshold")]
    pub sample_trace_threshold: Option<u64>,
    #[serde(default = "default_promethues_query_interval")]
    pub promethues_query_interval: Option<i32>,
    #[serde(default = "default_query_p9x_increase_rate")]
    pub query_p9x_increase_rate: Option<f64>,
    pub store_profile_tailbase: bool,
    #[serde(default = "default_receiver_ip")]
    pub receiver_ip: Option<String>,
    #[serde(default = "default_receiver_port")]
    pub receiver_port: Option<u16>,
}

fn default_sample_trace_repeat_num() -> Option<i32> {
    Some(3)
}

fn default_sample_trace_wait_time() -> Option<u64> {
    Some(30)
}

fn default_sample_url_hit_duration() -> Option<u64> {
    Some(5)
}

fn default_sample_trace_ignore_threshold() -> Option<u64> {
    Some(100)
}

fn default_sample_trace_threshold() -> Option<u64> {
    Some(0)
}

fn default_promethues_query_interval() -> Option<i32> {
    Some(30)
}

fn default_query_p9x_increase_rate() -> Option<f64> {
    Some(1.1)
}

fn default_receiver_ip() -> Option<String> {
    Some("127.0.0.1".to_string())
}

fn default_receiver_port() -> Option<u16> {
    Some(29090)
}

#[derive(Serialize, Deserialize, Debug)]
pub struct Exporters {
    #[serde(rename = "cameraexporter")]
    pub camera_exporter: CameraExporterCfg,
    #[serde(rename = "metricexporter")]
    pub metric_exporter: MetricExporterCfg,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct CameraExporterCfg {
    #[serde(default = "default_storage")]
    pub storage: Option<String>,
    #[serde(rename = "file_config")]
    pub file_config: FileConfig,
    #[serde(rename = "es_config")]
    pub es_config: EsConfig,
}

fn default_storage() -> Option<String> {
    Some("file".to_string())
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct FileConfig {
    #[serde(default = "default_storage_path")]
    pub storage_path: Option<String>,
    #[serde(default = "default_max_file_count_each_process")]
    pub max_file_count_each_process: Option<u32>,
}

fn default_storage_path() -> Option<String> {
    Some("/tmp/kindling".to_string())
}

fn default_max_file_count_each_process() -> Option<u32> {
    Some(50)
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct EsConfig {
    #[serde(default = "default_es_host")]
    pub es_host: Option<String>,
    #[serde(default = "default_index_suffix")]
    pub index_suffix: Option<String>,
}

fn default_es_host() -> Option<String> {
    Some("http://10.10.10.10:9200".to_string())
}

fn default_index_suffix() -> Option<String> {
    Some("dev".to_string())
}

pub(crate) const LOG_EXPORTER: &str = "log";
pub(crate) const PROMETHEUS_EXPORTER: &str = "prometheus";
#[derive(Serialize, Deserialize, Debug)]
pub struct MetricExporterCfg {
    #[serde(rename = "type", default = "default_metric_exporter_type")]
    pub exporter_type: Option<String>,
    #[serde(rename = "prometheus")]
    pub prometheus: PrometheusCfg,
}
fn default_metric_exporter_type() -> Option<String> {
    Some("prometheus".to_string())
}

#[derive(Serialize, Deserialize, Debug)]
pub struct PrometheusCfg {
    #[serde(default = "default_prometheus_port")]
    pub port: Option<u16>,
}

fn default_prometheus_port() -> Option<u16> {
    Some(9500)
}

pub fn load_conf() -> GlobalConfig {
    read_config_file("config/camera-agent-config.yml")
}

fn read_config_file(file_path: &str) -> GlobalConfig {
    let file = File::open(file_path).expect("Failed to open config file");
    info!("Config File: {}", file_path);
    serde_yaml::from_reader(file).unwrap()
}

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn test_unset_config() {
        let config = read_config_file("tests/unset_config.yml");
        assert_eq!(
            config.analyzers.trace_id_analyzer.java_trace_slow_time,
            default_java_trace_slow_time()
        );
        assert_eq!(
            config.analyzers.trace_id_analyzer.java_trace_wait_second,
            default_java_trace_wait_second()
        );
        assert_eq!(
            config.analyzers.cpu_analyzer.segment_size,
            default_segment_size()
        );
    }
}
