use crate::camera_exporter::camera_exporter::CameraExporter;
use crate::config::SampleProcessorCfg;
use crate::cpu_analyzer::is_profiled_enabled;
use crate::model::Trace;
use crate::probe_to_rust::sampleUrl;
use crate::sample_processor::{
    kindling_traceid::TraceIds, kindling_traceid_grpc::TraceIdServiceClient,
    promethues::PrometheusP9xCache,
};
use crate::traceid_analyzer::TraceEvent;
use crossbeam::channel::Sender;
use futures::executor;
use grpc::ClientStubExt;
use log::{error, info};
use protobuf::RepeatedField;
use std::collections::HashMap;
use std::ffi::CString;
use std::sync::Arc;
use std::time::{Duration, SystemTime};

#[derive(Default, Debug, Clone)]

pub struct SampleTrace {
    pub repeat_num: i32,
    pub is_sampled: bool,
    pub event: TraceEvent,
}

impl SampleTrace {
    pub fn new(event: TraceEvent, is_sampled: bool, repeat_num: i32) -> Self {
        Self {
            event,
            is_sampled,
            repeat_num,
        }
    }

    fn get_pid_url(&self) -> String {
        format!("{}-{}", self.event.pid, self.event.content_key)
    }

    pub fn set_profiled(&mut self, is_profiled: bool) {
        self.event.is_profiled = is_profiled
    }

    pub fn set_normal(&mut self, is_normal: bool) {
        self.event.is_normal = is_normal
    }

    fn set_p90(&mut self, p9x: f64) {
        self.event.p90 = p9x
    }

    fn is_slow(&self) -> bool {
        self.event.is_slow
    }

    pub fn get_container_id(&self) -> String {
        self.event.container_id.clone()
    }

    pub fn get_content_key(&self) -> String {
        self.event.content_key.clone()
    }
}

pub struct SampleCache {
    trace_cache: Vec<SampleTrace>,
    un_send_ids: UnSendIds,
    sampled_trace_ids: HashMap<String, SystemTime>,
    url_hits: HashMap<String, SystemTime>,
    trace_hold_time: Duration,
    url_hit_duration: Duration,
    trace_threshold: u64,
    store_profile_tailbase: bool,
    p9x_increase_rate: f64,
    query_time: i64,
    client: TraceIdServiceClient,
    p9x_cache: PrometheusP9xCache,
    normal_sent_cache: HashMap<(String, String), SystemTime>,
    profiling_signal_sender: Sender<TraceEvent>,
    trace_exporter: Arc<CameraExporter>,
}

impl SampleCache {
    pub fn new(
        cfg: &SampleProcessorCfg,
        profiling_signal_sender: Sender<TraceEvent>,
        trace_exporter: Arc<CameraExporter>,
    ) -> Self {
        let client_conf = Default::default();
        let host = cfg.receiver_ip.as_ref().unwrap().as_str();
        let port = cfg.receiver_port.unwrap();
        let client = match TraceIdServiceClient::new_plain(host, port, client_conf) {
            Ok(client) => client,
            Err(err) => {
                panic!("Connect Error: {:?}", err);
            }
        };
        Self {
            trace_cache: Vec::new(),
            un_send_ids: UnSendIds::new(cfg.sample_trace_repeat_num.unwrap() as usize),
            sampled_trace_ids: HashMap::new(),
            url_hits: HashMap::new(),
            trace_hold_time: Duration::from_secs(cfg.sample_trace_wait_time.unwrap().to_owned()),
            url_hit_duration: Duration::from_secs(cfg.sample_url_hit_duration.unwrap()),
            trace_threshold: cfg.sample_trace_threshold.unwrap() * 1000000,
            store_profile_tailbase: cfg.store_profile_tailbase,
            p9x_increase_rate: cfg.query_p9x_increase_rate.unwrap(),
            client,
            query_time: 0,
            p9x_cache: PrometheusP9xCache::new(host, port),
            normal_sent_cache: HashMap::new(),
            profiling_signal_sender,
            trace_exporter,
        }
    }

    pub fn is_tail_base_sampled(&self, sample_trace: &SampleTrace) -> bool {
        let trace_id = &sample_trace.event.trace_id;
        match self.sampled_trace_ids.get(trace_id) {
            Some(_) => {
                info!("Trace is stored by tailBase: traceId[{}]", trace_id);
                true
            }
            None => false,
        }
    }

    pub fn is_sampled(&self, sample_trace: &mut SampleTrace) -> bool {
        match self.url_hits.get(&sample_trace.get_pid_url()) {
            Some(_) => false,
            None => self.is_slow(sample_trace),
        }
    }

    pub fn get_p9x(&self, sample_trace: &SampleTrace) -> f64 {
        self.p9x_cache.get_p9x(sample_trace)
    }

    pub fn is_normal_sampled(&mut self, sample_trace: &SampleTrace) -> bool {
        let key = (
            sample_trace.get_container_id(),
            sample_trace.get_content_key(),
        );
        match self.normal_sent_cache.get(&key) {
            Some(_) => false,
            None => {
                self.normal_sent_cache.insert(key, SystemTime::now());
                true
            }
        }
    }

    pub fn store_normal_profiling(&mut self, sample_trace: &SampleTrace) {
        // Store Profiling Data.
        self.profiling_signal_sender
            .send(sample_trace.event.clone())
            .unwrap();
    }

    fn is_slow(&self, sample_trace: &mut SampleTrace) -> bool {
        let p9x = self.p9x_cache.get_p9x(sample_trace);
        sample_trace.set_p90(p9x);
        if self.trace_threshold > 0 {
            sample_trace.event.duration >= self.trace_threshold
        } else if p9x == 0.0 {
            // Not Got P9x.
            sample_trace.is_slow()
        } else {
            sample_trace.event.duration as f64 >= p9x * self.p9x_increase_rate
        }
    }

    pub fn cache_sample_trace(&mut self, sample_trace: SampleTrace) {
        self.trace_cache.push(sample_trace);
    }

    pub fn tail_base_profiling(&mut self, sample_trace: &mut SampleTrace) {
        if self.is_slow(sample_trace) && self.store_profile_tailbase {
            // Store Profiling
            self.store_profiling(sample_trace, true);
        }
        // Store Trace
        self.store_trace(sample_trace);
    }

    pub fn store_profiling(&mut self, sample_trace: &mut SampleTrace, is_tail_base_profiled: bool) {
        if sample_trace.is_sampled {
            return;
        }
        sample_trace.set_profiled(true);

        let now = SystemTime::now();
        // Update Url HitTime when profile is stored.
        let pid_url = sample_trace.get_pid_url();
        let c_pid_url = CString::new(pid_url.clone()).unwrap().into_raw();
        self.url_hits.insert(pid_url, now);

        // 调用C代码
        unsafe {
            sampleUrl(c_pid_url, 1);
            drop(CString::from_raw(c_pid_url));
        }

        if !is_tail_base_profiled {
            let trace_id = &sample_trace.event.trace_id;

            self.sampled_trace_ids
                .entry(trace_id.to_string())
                .or_insert_with(|| {
                    // Record sampled traceIds and send them to receiver per second.
                    self.un_send_ids.cache_ids(trace_id.clone());
                    now
                });
        }
        // Store Profiling Data.
        self.profiling_signal_sender
            .send(sample_trace.event.clone())
            .unwrap();
    }

    pub fn store_trace(&mut self, sample_trace: &SampleTrace) {
        // 将采样的Trace传递给ES存储
        self.trace_exporter
            .consume_trace(Trace::new(sample_trace.event.clone()));
    }

    pub fn check_tail_base_traces(&mut self) {
        let now = SystemTime::now();

        // Delete expired traceIds.
        self.sampled_trace_ids.retain(|k, v| {
            let skip = now - self.trace_hold_time > *v;
            if skip {
                info!("Delete Old TraceId: {}", k);
            }
            !skip
        });

        // Delete expired urlHits.
        self.url_hits.retain(|k, v| {
            let skip = now - self.url_hit_duration > *v;
            if now - self.url_hit_duration > *v {
                info!("Delete Old Url: {}", k);
                let c_pid_url = CString::new((*k).clone()).unwrap().into_raw();
                // 调用C代码
                unsafe {
                    sampleUrl(c_pid_url, 0);
                    drop(CString::from_raw(c_pid_url));
                }
            }
            !skip
        });

        let size = self.trace_cache.len();
        if size == 0 {
            return;
        }

        let mut last_loop_traces: Vec<SampleTrace> = self.trace_cache.drain(..).collect();
        let mut new_size = 0;
        for sample_trace in last_loop_traces.iter_mut() {
            if self.is_tail_base_sampled(sample_trace) {
                // Store Profiling And Trace
                self.tail_base_profiling(sample_trace);
            } else if sample_trace.repeat_num > 0 {
                // Set sampleTrace times-1.
                sample_trace.repeat_num -= 1;
                self.trace_cache.push(sample_trace.clone());
                new_size += 1;
            }
            // Skip sampleTrace after N times.
        }
        info!("Clear Normal Traces[{}] => {}", size, new_size);
    }

    pub fn send_and_recv_sampled_traces(&mut self) {
        let notify_trace_count = self.un_send_ids.get_to_send_count();

        // Don't request traceIds when profiling is not started.
        if !is_profiled_enabled() {
            if notify_trace_count > 0 {
                self.un_send_ids.mark_sent(notify_trace_count);
            }
            return;
        }

        // Cache N seconds data when the server is not available, send them when the server is available.
        let mut request = TraceIds::new();
        request.queryTime = self.query_time;
        let mut trace_ids: RepeatedField<String> = RepeatedField::new();
        self.un_send_ids
            .get_to_send_ids(notify_trace_count as usize)
            .iter()
            .for_each(|trace_id| trace_ids.push(trace_id.clone()));
        request.traceIds = trace_ids;

        let result = executor::block_on(
            self.client
                .send_trace_ids(grpc::RequestOptions::new(), request)
                .drop_metadata(),
        );

        match result {
            Ok(response) => {
                if notify_trace_count > 0 {
                    self.un_send_ids.mark_sent(notify_trace_count);
                }
                // Record Last queryTime for server
                self.query_time = response.queryTime;
                for trace_id in response.traceIds.iter() {
                    // Store the tailbased traceIds.
                    self.sampled_trace_ids
                        .insert(trace_id.clone(), SystemTime::now());
                }
            }
            Err(err) => {
                error!("Send TraceIds failed: {}", err);
                self.un_send_ids.mark_unsent(notify_trace_count);
            }
        }
    }

    pub fn update_p9x_by_grpc(&mut self) {
        self.p9x_cache.update_p9x_by_grpc();
    }
}

struct UnSendIds {
    to_send_ids: Vec<String>,
    counts: Vec<i32>, // Record count per seconds
    last_num: i32,    // Record how many datas are not sent last time.
    size: usize,      // Record how many datas are sent
}

impl UnSendIds {
    fn new(size: usize) -> Self {
        Self {
            to_send_ids: Vec::new(),
            counts: vec![0; size],
            last_num: 0,
            size: 0,
        }
    }

    fn cache_ids(&mut self, id: String) {
        self.to_send_ids.push(id);
    }

    fn mark_unsent(&mut self, num: i32) {
        if self.size == 0 {
            self.counts[0] = num;
            self.size = 1;
            self.last_num = num;
        } else if self.size < self.counts.len() {
            self.counts[self.size] = num - self.last_num;
            self.size += 1;
            self.last_num = num;
        } else {
            let to_remove_size = self.counts[0];
            if to_remove_size > 0 {
                self.to_send_ids.drain(0..to_remove_size as usize);
            }
            // Clean the oldest record, add new recrod to make cache store N seconds.
            self.counts.rotate_left(1);
            let new_size = self.counts.len() - 1;
            self.counts[new_size] = num - self.last_num;
            self.last_num = num - to_remove_size;
        }
    }

    fn mark_sent(&mut self, num: i32) {
        self.to_send_ids.drain(0..num as usize);
        // Reset count
        self.size = 0;
        self.last_num = 0;
    }

    fn get_to_send_count(&self) -> i32 {
        self.to_send_ids.len() as i32
    }

    fn get_to_send_ids(&self, size: usize) -> Vec<String> {
        self.to_send_ids.iter().take(size).cloned().collect()
    }
}
