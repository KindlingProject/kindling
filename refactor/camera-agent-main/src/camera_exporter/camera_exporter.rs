use super::es_writer::EsWriter;
use crate::config::CameraExporterCfg;
use crate::model;
use log::debug;
use log::info;

const STORAGE_FILE: &str = "file";
const STORAGE_ELASTICSEARCH: &str = "elasticsearch";
const STORAGE_NOOP: &str = "noop";

pub struct CameraExporter {
    writer: Box<dyn Writer + Send + Sync>,
}

impl CameraExporter {
    pub(crate) fn new(
        config: &CameraExporterCfg,
    ) -> Result<Self, Box<dyn std::error::Error + Send + Sync>> {
        let storage = config.storage.clone().unwrap_or_else(|| {
            info!("No storage config, use default: {}", STORAGE_NOOP);
            STORAGE_NOOP.to_string()
        });
        info!("Create writer: {}", &storage);
        let writer: Box<dyn Writer + Send + Sync> = match storage.as_str() {
            STORAGE_ELASTICSEARCH => {
                let es_writer = EsWriter::new(&config.es_config)?;
                Box::new(es_writer)
            }
            STORAGE_NOOP => Box::new(NoopWriter {}),
            &_ => {
                let es_writer = EsWriter::new(&config.es_config)?;
                Box::new(es_writer)
            } // STORAGE_FILE => {
              //     let file_writer = FileWriter::new(&config.file_config)?;
              //     Box::new(file_writer)
              // }
        };
        Ok(CameraExporter { writer })
    }
    pub(crate) fn consume_trace(&self, data: model::Trace) {
        println!("Conusme_trace: {}", self.writer.name());
        let _ = self.writer.write_trace(data);
    }
    pub(crate) fn consume_profiling(&self, data: model::TraceProfiling) {
        let _ = self.writer.write_profiling(data);
    }
}
pub(crate) trait Writer {
    fn write_trace(&self, data_group: model::Trace);
    fn write_profiling(&self, profiling: model::TraceProfiling);
    fn name(&self) -> &str;
}

struct NoopWriter {}

impl Writer for NoopWriter {
    fn write_trace(&self, data_group: model::Trace) {
        let data_json: String = serde_json::to_string(&data_group).unwrap();
        debug!(
            "NoopWriter write_trace: pid={}, content_key={}, traceid={}, {}",
            data_group.labels.pid,
            data_group.labels.content_key,
            data_group.labels.trace_id,
            data_json
        )
    }
    fn write_profiling(&self, profiling: model::TraceProfiling) {
        let data_json: String = serde_json::to_string(&profiling).unwrap();
        debug!(
            "NoopWriter write_profiling: pid={}, thread_name={}, content_key={}, {}",
            profiling.labels.pid,
            profiling.labels.threadName,
            profiling.labels.content_key,
            data_json
        )
    }
    fn name(&self) -> &str {
        return "NoopWriter";
    }
}
