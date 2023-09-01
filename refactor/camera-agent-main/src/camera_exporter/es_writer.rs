use super::camera_exporter::Writer;
use crate::config::EsConfig;
use crate::model::{Trace, TraceProfiling};
use elasticsearch::{http::transport::Transport, Elasticsearch};
use log::info;
use log::warn;
use serde_json::Value;
use std::sync::Arc;
use tokio::runtime::Runtime;

pub(crate) struct EsWriter {
    es_client: Arc<Elasticsearch>,
    index_suffix: String,
    runtime: Runtime,
}

impl EsWriter {
    pub(crate) fn new(
        es_config: &EsConfig,
    ) -> Result<Self, Box<dyn std::error::Error + Send + Sync>> {
        let transport = Transport::single_node(es_config.es_host.as_ref().unwrap().as_str())?;
        let es_client = Elasticsearch::new(transport);
        let runtime = tokio::runtime::Builder::new_multi_thread()
            .enable_all()
            .build()
            .unwrap();
        let arc_client = Arc::new(es_client);
        Ok(EsWriter {
            es_client: arc_client,
            index_suffix: es_config.index_suffix.as_ref().unwrap().clone(),
            runtime,
        })
    }
}

impl Writer for EsWriter {
    fn write_trace(&self, data: Trace) {
        let index = "span_trace_group";
        let index = if !self.index_suffix.is_empty() {
            format!("{}_{}", index, self.index_suffix)
        } else {
            index.to_string()
        };
        let client = self.es_client.clone();
        self.runtime.spawn(async move {
            println!("##write_trace");
            info!("Write Es: {:?}", &data);

            let response = client
                .index(elasticsearch::IndexParts::Index(&index))
                .body(data)
                .send()
                .await;
            match response {
                Ok(response) => {
                    let successful = response.status_code().is_success();
                    if !successful {
                        let body_json: Value = response.json().await.unwrap();
                        warn!("Errors span_trace_group indexing. {}", body_json);
                    }
                }
                Err(e) => {
                    warn!("Elasticsearch write trace error: {:?}", e);
                }
            }
        });
    }

    fn write_profiling(&self, data: TraceProfiling) {
        let is_sent = data.labels.isSent;
        // The data has been sent before, so esExporter will not index it again.
        // But fileExporter will.
        if is_sent == 1 {
            return;
        }
        let index = "camera_event_group";
        let index = if !self.index_suffix.is_empty() {
            format!("{}_{}", index, self.index_suffix)
        } else {
            index.to_string()
        };
        let client = self.es_client.clone();
        self.runtime.spawn(async move {
            let response = client
                .index(elasticsearch::IndexParts::Index(&index))
                .body(data)
                .send()
                .await;
            match response {
                Ok(response) => {
                    let successful = response.status_code().is_success();
                    if !successful {
                        let body_json: Value = response.json().await.unwrap();
                        warn!("Errors span_trace_group indexing. {}", body_json);
                    }
                }
                Err(e) => {
                    warn!("Elasticsearch write profiling error: {:?}", e);
                }
            }
        });
    }

    fn name(&self) -> &str {
        "elasticsearch"
    }
}

#[cfg(test)]
mod tests {
    use super::EsWriter;
    use super::Writer;
    use crate::config::EsConfig;
    use crate::model::Trace;
    use crate::traceid_analyzer::TraceEvent;
    #[test]
    fn write_es() {
        let trace_event = TraceEvent {
            pid: 1,
            tid: 1,
            protocol: "protocol".to_string(),
            content_key: "content_key".to_string(),
            http_url: "http_url".to_string(),
            is_slow: true,
            is_server: true,
            is_error: true,
            is_profiled: false,
            p90: 1000000.0,
            trace_id: "trace_id".to_string(),
            apm_type: "apm_type".to_string(),
            apm_span_id: "span_id".to_string(),
            container_id: "container_id".to_string(),
            container_name: "container_name".to_string(),
            workload_name: "workload_name".to_string(),
            workload_kind: "workload_kind".to_string(),
            pod_ip: "pod_ip".to_string(),
            pod_name: "pod_name".to_string(),
            namespace: "namespace".to_string(),
            duration: 1000000,
            end_time: 1000000000,
            node_ip: "10.10.10.10".to_string(),
            node_name: "Node-123".to_string(),
            offset_ts: 0,
        };
        let data = Trace::new(trace_event);
        let es_config = EsConfig {
            es_host: Some("http://10.10.103.96:9200".to_string()),
            index_suffix: Some("test".to_string()),
        };
        let writer = EsWriter::new(&es_config).unwrap();
        writer.write_trace(data);
    }
}
