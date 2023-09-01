use log::{debug, info};
use std::collections::HashMap;
use std::sync::{Arc, Condvar, Mutex};
use std::thread;
use std::time::Duration;

use crate::kubernetes::k8scache;
use crate::model::consts;
use crate::probe_to_rust::KindlingEventForGo;
use crate::sample_processor::SampleProcessor;
use crate::{
    config::TraceIdAnalyzerCfg,
    traceid_analyzer::model::{SignalEvent, ThreadTraceIdEvent, TraceEvent, TraceIdEvents},
};

const CHECK_INTERVAL: Duration = Duration::from_secs(1);

fn read_string_value(val: Option<&[u8]>) -> String {
    String::from_utf8_lossy(val.unwrap_or_default()).to_string()
}

fn read_str_int_value(val: Option<&[u8]>) -> u64 {
    let str_val = read_string_value(val);
    str_val.parse().unwrap()
}

use crate::metric_exporter;
use crossbeam::channel::Sender;
pub struct TraceIdAnalyzer {
    #[allow(clippy::type_complexity)]
    java_traces: Arc<Mutex<HashMap<(String, u32, u32), TraceIdEvents>>>,
    batch_metrics: Arc<Mutex<Vec<SignalEvent>>>,
    java_trace_wait_second: i32,
    open_java_trace_sampling: bool,
    java_trace_slow_time: u64,
    node_name: String,
    node_ip: String,
    running: Arc<(Mutex<bool>, Condvar)>,
    thread: Mutex<Option<thread::JoinHandle<()>>>,
    sampler: SampleProcessor,
    metric_signal_sender: Arc<Sender<Vec<SignalEvent>>>,
    metric_exporter: Arc<dyn metric_exporter::MetricExporter>,
    k8s_cache: Arc<k8scache::K8sMetaDataCache>,
}

impl TraceIdAnalyzer {
    pub fn new(
        cfg: &TraceIdAnalyzerCfg,
        sampler: SampleProcessor,
        metric_signal_sender: Sender<Vec<SignalEvent>>,
        metric_exporter: Arc<dyn metric_exporter::MetricExporter>,
        k8s_cache: Arc<k8scache::K8sMetaDataCache>,
    ) -> Self {
        Self {
            java_traces: Arc::new(Mutex::new(HashMap::new())),
            batch_metrics: Arc::new(Mutex::new(Vec::with_capacity(1000))),
            open_java_trace_sampling: cfg.open_java_trace_sampling,
            java_trace_wait_second: cfg.java_trace_wait_second.unwrap(),
            java_trace_slow_time: cfg.java_trace_slow_time.unwrap() * 1000000,
            node_name: std::env::var("MY_NODE_NAME").unwrap_or_else(|_| "".to_string()),
            node_ip: std::env::var("MY_NODE_IP").unwrap_or_else(|_| "".to_string()),
            running: Arc::new((Mutex::new(false), Condvar::new())),
            thread: Mutex::new(None),
            sampler,
            metric_signal_sender: Arc::new(metric_signal_sender),
            metric_exporter,
            k8s_cache,
        }
    }

    pub fn start(&self) {
        // Start Sampler
        self.sampler.start();

        {
            let (started, _) = &*self.running;
            let mut started = started.lock().unwrap();
            if *started {
                return;
            }
            *started = true;
        }

        let java_traces = Arc::clone(&self.java_traces);
        let batch_mtrics = Arc::clone(&self.batch_metrics);
        let running = self.running.clone();
        let metric_signal_sender = Arc::clone(&self.metric_signal_sender);
        let thread = thread::Builder::new()
            .name("traceid-analyzer-check".to_owned())
            .spawn(move || {
                loop {
                    clear_expire_traces(&java_traces);
                    batch_send_metrics(&batch_mtrics, &metric_signal_sender);
                    let (running, timer) = &*running;
                    let mut running = running.lock().unwrap();
                    if !*running {
                        break;
                    }
                    running = timer.wait_timeout(running, CHECK_INTERVAL).unwrap().0;
                    if !*running {
                        break;
                    }
                }
                info!("TraceIdAnalyzer exited");
            })
            .unwrap();
        self.thread.lock().unwrap().replace(thread);
        info!("TraceIdAnalyzer started");
    }

    pub fn stop(&self) {
        self.sampler.stop();

        let (stopped, timer) = &*self.running;
        {
            let mut stopped = stopped.lock().unwrap();
            if !*stopped {
                return;
            }
            *stopped = false;
        }
        timer.notify_one();

        if let Some(thread) = self.thread.lock().unwrap().take() {
            let _ = thread.join();
        }
        info!("TraceIdAnalyzer stopped");
    }

    pub fn consume_traceid_event(&mut self, event: &KindlingEventForGo) {
        if !self.open_java_trace_sampling {
            // Skip datas if profiling is not start or sampling is not opened.
            return;
        }

        let pid = event.get_pid();
        let tid = event.get_tid();
        let trace_id = read_string_value(event.user_attributes[0].get_value());
        let is_enter: u64 = read_str_int_value(event.user_attributes[1].get_value());
        if is_enter > 0 {
            let is_sampled: bool = read_str_int_value(event.user_attributes[3].get_value()) == 1;
            let ev = ThreadTraceIdEvent {
                timestamp: event.timestamp,
                tid,
                url: read_string_value(event.user_attributes[2].get_value()),
                ..ThreadTraceIdEvent::default()
            };
            // debug!("is_sampled: {}, {:?}", is_sampled, &ev);
            self.process_request(ev, pid, tid, trace_id, is_sampled);
        } else {
            let ev = ThreadTraceIdEvent {
                timestamp: event.timestamp,
                tid,
                apm_type: read_string_value(event.user_attributes[2].get_value()),
                thread_type: read_str_int_value(event.user_attributes[3].get_value()),
                error: read_str_int_value(event.user_attributes[4].get_value()),
                span_id: read_string_value(event.user_attributes[5].get_value()),
                parent_span_id: read_string_value(event.user_attributes[6].get_value()),
                container_id: event.get_container_id(),
                ..ThreadTraceIdEvent::default()
            };
            //debug!("is_sampled aaa:  {:?}",  &ev);
            if let (Some(trace_event), is_sampled) = self.process_response(ev, pid, tid, trace_id) {
                // Cache Metrics or Send 1K datas to CpuAnalyzer to calculate metrics
                //debug!("is_sampled bbb: {}", is_sampled);
                self.cache_metric_signal(&trace_event);
                // Aggregate the trace metrics and export via prometheus exporter
                self.send_trace_metric_via_promethues(&trace_event);
                // SendTo Sampler
                self.sampler.consume(trace_event, is_sampled);
            }
        }
    }

    fn process_request(
        &mut self,
        event: ThreadTraceIdEvent,
        pid: u32,
        tid: u32,
        trace_id: String,
        is_sampled: bool,
    ) {
        let trace_id_clone = trace_id.clone();
        let key = (trace_id, pid, tid);
        let mut cache = self.java_traces.lock().unwrap();
        match cache.get_mut(&key) {
            Some(entry) => entry.update_trace_event(event, is_sampled),
            None => {
                cache.insert(
                    key,
                    TraceIdEvents::new(
                        pid,
                        trace_id_clone,
                        is_sampled,
                        self.java_trace_wait_second,
                        event,
                    ),
                );
            }
        }
    }

    fn process_response(
        &mut self,
        event: ThreadTraceIdEvent,
        pid: u32,
        tid: u32,
        trace_id: String,
    ) -> (Option<TraceEvent>, bool) {
        let mut cache = self.java_traces.lock().unwrap();
        let key = (trace_id.clone(), pid, tid);
        match cache.remove(&key) {
            None => {
                info!(
                    "Miss entry traceid event for TraceID={}, Pid={}, Tid={}",
                    key.0, pid, tid
                );
                (None, false)
            }
            Some(trace_id_events) => {
                if !event.is_business_thread() {
                    return (None, false);
                }
                let mut thread_trace_event = trace_id_events.event;
                thread_trace_event.merge_response(event);
                let is_sampled = trace_id_events.is_sampled;
                let k8s_pod_info = self
                    .k8s_cache
                    .get_k8s_pod_info(&self.node_ip, &thread_trace_event.container_id);
                (
                    Some(TraceEvent::new(
                        pid,
                        trace_id,
                        thread_trace_event,
                        self.java_trace_slow_time,
                        self.node_name.to_string(),
                        self.node_ip.to_string(),
                        k8s_pod_info,
                    )),
                    is_sampled,
                )
            }
        }
    }

    fn get_k8s_pod_info(&self, container_id: &str) -> k8scache::K8sPodInfo {
        self.k8s_cache.get_k8s_pod_info(&self.node_ip, container_id)
    }

    fn send_trace_metric_via_promethues(&self, trace_event: &TraceEvent) {
        self.metric_exporter.record_metric(
            consts::TRACE_HISTOGRAM_NAME,
            trace_event.duration,
            &trace_event.to_key_value(),
        );
    }
    fn cache_metric_signal(&self, trace_event: &TraceEvent) {
        let mut metrics = self.batch_metrics.lock().unwrap();
        metrics.push(SignalEvent::new(trace_event));
        if metrics.len() >= 1000 {
            let mut copied = Vec::with_capacity(1000);
            std::mem::swap(&mut copied, &mut metrics);
            // Release the lock
            drop(metrics);
            self.metric_signal_sender.send(copied).unwrap();
        }
    }
}

#[allow(clippy::type_complexity)]
fn clear_expire_traces(java_traces: &Arc<Mutex<HashMap<(String, u32, u32), TraceIdEvents>>>) {
    let mut map_guard = java_traces.lock().unwrap();

    let mut count: u32 = 0;
    map_guard.retain(|k, v| {
        let skip = v.is_expired();
        if skip {
            debug!("==>Clean Expired JavaTrace {}, pid: {}", k.0, k.1);
            count += 1;
        }
        !skip
    });
    if count > 0 {
        info!("Clean {} Expired JavaTraces", count);
    }
}

fn batch_send_metrics(
    batch_metrics: &Arc<Mutex<Vec<SignalEvent>>>,
    metric_signal_sender: &Arc<Sender<Vec<SignalEvent>>>,
) {
    let mut metrics = batch_metrics.lock().unwrap();
    if metrics.len() == 0 {
        return;
    }

    let mut copied = Vec::with_capacity(1000);
    std::mem::swap(&mut copied, &mut metrics);
    // Release the lock
    drop(metrics);

    metric_signal_sender.send(copied).unwrap();
}
