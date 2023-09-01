use crate::sample_processor::{
    kindling_p9x::{P9xData, P9xRequest},
    kindling_p9x_grpc::P9XServiceClient,
    sampler::SampleTrace,
};
use futures::executor;
use grpc::ClientStubExt;
use std::collections::HashMap;
use std::time::{SystemTime, UNIX_EPOCH};

pub struct PrometheusP9xCache {
    client: P9XServiceClient,
    p9x_cache: HashMap<String, PrometheusContainerP9x>,
    host_name: String,
    timeout: u64,
}

impl PrometheusP9xCache {
    pub fn new(host: &str, port: u16) -> Self {
        let client_conf = Default::default();
        let client = match P9XServiceClient::new_plain(host, port, client_conf) {
            Ok(client) => client,
            Err(err) => {
                panic!("Connect Error: {}", err);
            }
        };
        let host_name = if let Ok(value) = std::env::var("MY_NODE_IP") {
            value
        } else {
            match hostname::get() {
                Ok(hostname) => match hostname.into_string() {
                    Ok(s) => s,
                    Err(_) => "unknown".to_owned(),
                },
                Err(_) => "unknown".to_owned(),
            }
        };

        Self {
            client,
            p9x_cache: HashMap::new(),
            host_name,
            timeout: 24 * 3600,
        }
    }

    pub fn get_p9x(&self, trace: &SampleTrace) -> f64 {
        if let Some(p9x) = self.p9x_cache.get(&trace.get_container_id()) {
            if let Some(value) = p9x.p9x_values.get(&trace.get_content_key()) {
                return *value;
            }
        };
        0.0
    }

    pub fn update_p9x_by_grpc(&mut self) {
        let mut request = P9xRequest::new();
        request.ip = self.host_name.clone();

        if let Ok(response) = executor::block_on(
            self.client
                .query_p9x(grpc::RequestOptions::new(), request)
                .drop_metadata(),
        ) {
            let datas = response.datas;
            if datas.is_empty() {
                return;
            }

            let now = SystemTime::now()
                .duration_since(UNIX_EPOCH)
                .unwrap()
                .as_secs();

            for data in datas.iter() {
                let p9x = self
                    .p9x_cache
                    .entry(data.get_containerId().to_string())
                    .or_insert_with(|| PrometheusContainerP9x::new(now));
                p9x.update_p9x(data, now);
            }

            self.p9x_cache
                .retain(|_, p9x| (now - p9x.update_time) < self.timeout);
        }
    }
}

struct PrometheusContainerP9x {
    update_time: u64,
    p9x_values: HashMap<String, f64>,
}

impl PrometheusContainerP9x {
    fn new(update_time: u64) -> Self {
        Self {
            update_time,
            p9x_values: HashMap::new(),
        }
    }

    fn update_p9x(&mut self, data: &P9xData, time: u64) {
        if data.value > 0.0 {
            self.p9x_values.insert(data.url.to_string(), data.value);
            self.update_time = time;
        }
    }
}
