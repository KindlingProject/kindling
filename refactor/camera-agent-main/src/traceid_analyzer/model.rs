use crate::kubernetes::k8scache;
use crate::model::consts;
use crate::ntpsync::get_current_offset;
use serde::Serialize;
use std::fmt;

#[derive(Default, Debug)]
pub struct TraceIdEvents {
    pub pid: u32,
    pub trace_id: String,
    pub is_sampled: bool,
    check_times: i32,
    pub event: ThreadTraceIdEvent,
}

impl TraceIdEvents {
    pub fn new(
        pid: u32,
        trace_id: String,
        is_sampled: bool,
        check_times: i32,
        event: ThreadTraceIdEvent,
    ) -> Self {
        Self {
            pid,
            trace_id,
            is_sampled,
            check_times,
            event,
        }
    }

    pub fn is_expired(&mut self) -> bool {
        self.check_times -= 1;
        self.check_times < 0
    }

    pub fn update_trace_event(&mut self, enter_event: ThreadTraceIdEvent, is_sampled: bool) {
        if is_sampled && !self.is_sampled {
            // Mark Sampled
            self.is_sampled = true
        }
        // Update Url
        self.event.url = enter_event.url;
    }

    pub fn get_trace_id(&self) -> String {
        self.trace_id.to_string()
    }
}

#[derive(Debug, Clone, Default)]
pub struct ThreadTraceIdEvent {
    pub timestamp: u64,
    pub end_time: u64,
    pub apm_type: String,
    pub parent_span_id: String,
    pub span_id: String,
    pub url: String,
    pub thread_type: u64, // 0: businessThread 1: Async 2: I/O
    pub error: u64,
    pub tid: u32,
    pub container_id: String,
}

impl ThreadTraceIdEvent {
    pub fn merge_response(&mut self, new_evt: ThreadTraceIdEvent) {
        self.end_time = new_evt.timestamp;
        self.apm_type = new_evt.apm_type;
        self.thread_type = new_evt.thread_type;
        self.error = new_evt.error;
        self.span_id = new_evt.span_id;
        self.parent_span_id = new_evt.parent_span_id;
        self.tid = new_evt.tid;
        self.container_id = new_evt.container_id;
    }

    fn get_duration(&self) -> u64 {
        self.end_time - self.timestamp
    }

    pub fn is_business_thread(&self) -> bool {
        self.thread_type == 0
    }
}

#[derive(Default, Debug, Clone)]
pub struct SignalEvent {
    pub pid: u32,
    pub tid: u32,
    pub content_key: String,
    pub container_id: String,
    pub start_time: u64,
    pub end_time: u64,
}

impl SignalEvent {
    pub fn new(evt: &TraceEvent) -> Self {
        Self {
            pid: evt.pid,
            tid: evt.tid,
            content_key: evt.content_key.clone(),
            container_id: evt.container_id.clone(),
            start_time: evt.end_time - evt.duration,
            end_time: evt.end_time,
        }
    }
}

impl fmt::Display for SignalEvent {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(
            f,
            "SignalEvent: pid={}, tid={}, start_time={}, end_time={}",
            self.pid, self.tid, self.start_time, self.end_time,
        )
    }
}

#[derive(Default, Debug, Serialize, Clone)]
pub struct TraceEvent {
    pub pid: u32,
    pub tid: u32,
    pub protocol: String,
    pub content_key: String,
    pub http_url: String,
    pub is_slow: bool,
    pub is_server: bool,
    pub is_error: bool,
    pub is_profiled: bool,
    pub p90: f64,
    pub trace_id: String,
    pub apm_type: String,
    pub apm_span_id: String,
    pub container_id: String,
    pub container_name: String,
    pub workload_name: String,
    pub workload_kind: String,
    pub pod_ip: String,
    pub pod_name: String,
    pub namespace: String,
    pub duration: u64,
    pub end_time: u64,
    pub node_name: String,
    pub node_ip: String,
    pub offset_ts: i64,
}

impl fmt::Display for TraceEvent {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(
            f,
            "TraceEvent: pid={}, tid={}, protocol={}, content_key={}, http_url={}, \\
            is_slow={}, is_server={}, is_error={}, is_profiled={} \\
            p90={}, trace_id={}, apm_type={}, apm_span_id={}, duration={}, end_time={}, \\
            container_id={}, container_name={}, workload_name={}, workload_kind={}, pod_ip={}, pod_name={}, namespace={}, node_name: {}, node_ip: {}, offset_ts: {}",
            self.pid,
            self.tid,
            self.protocol,
            self.content_key,
            self.http_url,
            self.is_slow,
            self.is_server,
            self.is_error,
            self.is_profiled,
            self.p90,
            self.apm_type,
            self.trace_id,
            self.apm_span_id,
            self.duration,
            self.end_time,
            self.container_id,
            self.container_name,
            self.workload_name,
            self.workload_kind,
            self.pod_ip,
            self.pod_name,
            self.namespace,
            self.node_name,
            self.node_ip,
            self.offset_ts,
        )
    }
}
use opentelemetry_api::KeyValue;
impl TraceEvent {
    pub fn new(
        pid: u32,
        trace_id: String,
        event: ThreadTraceIdEvent,
        slow_threshold: u64,
        node_name: String,
        node_ip: String,
        k8s_pod_info: k8scache::K8sPodInfo,
    ) -> Self {
        Self {
            apm_span_id: event.span_id.clone(),
            container_id: event.container_id.clone(),
            content_key: event.url.clone(),
            http_url: event.url.clone(),
            duration: event.get_duration(),
            end_time: event.end_time,
            pid,
            protocol: "http".to_string(),
            is_error: event.error > 0,
            is_slow: event.get_duration() > slow_threshold,
            is_server: true,
            is_profiled: false,
            p90: 0.0,
            tid: event.tid,
            trace_id,
            apm_type: event.apm_type,
            container_name: k8s_pod_info.container_name,
            workload_name: k8s_pod_info.workload_name,
            workload_kind: k8s_pod_info.workload_kind,
            pod_ip: k8s_pod_info.pod_ip,
            pod_name: k8s_pod_info.pod_name,
            namespace: k8s_pod_info.namespace,
            node_name,
            node_ip,
            offset_ts: get_current_offset(),
        }
    }
    pub fn to_key_value(&self) -> Vec<KeyValue> {
        vec![
            KeyValue::new(consts::PID, self.pid.to_string()),
            KeyValue::new(consts::CONTENT_KEY, self.content_key.to_string()),
            KeyValue::new(consts::IS_ERROR, self.is_error.to_string()),
            KeyValue::new(consts::CONTAINER_ID, self.container_id.to_string()),
            KeyValue::new(consts::CONTAINER_NAME, self.container_name.to_string()),
            KeyValue::new(consts::WORKLOAD_NAME, self.workload_name.to_string()),
            KeyValue::new(consts::WORKLOAD_KIND, self.workload_kind.to_string()),
            KeyValue::new(consts::POD_IP, self.pod_ip.to_string()),
            KeyValue::new(consts::POD_NAME, self.pod_name.to_string()),
            KeyValue::new(consts::NODE_NAME, self.node_name.to_string()),
            KeyValue::new(consts::NAMESPACE, self.namespace.to_string()),
        ]
    }
}

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn add_trace_event() {
        let tid: u32 = 100;
        let enter_event: ThreadTraceIdEvent = ThreadTraceIdEvent {
            timestamp: 100000000,
            url: String::from("Undertow"),
            tid,
            ..ThreadTraceIdEvent::default()
        };
        let enter_url_event = ThreadTraceIdEvent {
            timestamp: 100000010,
            url: String::from("Get /http"),
            tid,
            ..ThreadTraceIdEvent::default()
        };
        let exit_event = ThreadTraceIdEvent {
            timestamp: 101000000,
            apm_type: "skywalking".to_owned(),
            thread_type: 1,
            parent_span_id: String::from(""),
            span_id: String::from(""),
            error: 0,
            tid,
            container_id: String::from(""),
            ..ThreadTraceIdEvent::default()
        };

        let mut trace_events = TraceIdEvents::new(
            1,
            "d3755d293ac24ad2899044853e521056.46.16793970149120001".to_string(),
            false,
            5,
            enter_event,
        );
        trace_events.update_trace_event(enter_url_event, false);
        trace_events.event.merge_response(exit_event);

        let request = trace_events.event;
        assert_eq!(request.get_duration(), 1000000);
    }
}
