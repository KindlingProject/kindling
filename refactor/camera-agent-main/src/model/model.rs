use std::collections::HashMap;

use crate::traceid_analyzer::TraceEvent;
use serde::Serialize;

#[derive(Debug, Clone, Serialize)]
pub struct TraceProfiling {
    pub name: String,
    pub timestamp: u64,
    pub labels: ProfilingEvent,
}

impl TraceProfiling {
    pub fn new(start_time: u64, end_time: u64, event: &TraceEvent) -> Self {
        Self {
            name: "camera_event_group".to_string(),
            timestamp: start_time,
            labels: ProfilingEvent {
                content_key: event.content_key.to_string(),
                container_id: event.container_id.to_string(),
                container_name: event.container_name.to_string(),
                pod_ip: event.pod_ip.to_string(),
                pod_name: event.pod_name.to_string(),
                workload_name: event.workload_name.to_string(),
                workload_kind: event.workload_kind.to_string(),
                namespace: event.namespace.to_string(),
                node_name: event.node_name.to_string(),
                node_ip: event.node_ip.to_string(),
                endTime: end_time,
                innerCalls: "[]".to_string(),
                is_server: event.is_server,
                protocol: event.protocol.to_string(),
                spans: "[]".to_string(),
                startTime: start_time,
                offset_ts: event.offset_ts,
                ..ProfilingEvent::default()
            },
        }
    }
}

#[allow(non_snake_case)]
#[derive(Debug, Default, Clone, Serialize)]
pub struct ProfilingEvent {
    pub content_key: String,
    pub cpuEvents: String,
    pub container_id: String,
    pub container_name: String,
    pub pod_ip: String,
    pub pod_name: String,
    pub workload_name: String,
    pub workload_kind: String,
    pub namespace: String,
    pub node_name: String,
    pub node_ip: String,
    pub endTime: u64,
    pub innerCalls: String,
    pub isSent: i32,
    pub is_server: bool,
    pub javaFutexEvents: String,
    pub pid: u32,
    pub protocol: String,
    pub spans: String,
    pub startTime: u64,
    pub threadName: String,
    pub tid: u32,
    pub transactionIds: String,
    pub offset_ts: i64,
}

#[derive(Debug, Clone, Serialize)]
pub(crate) struct Trace {
    pub name: String,
    pub timestamp: u64,
    pub metrics: Vec<Metric>,
    pub labels: TraceEvent,
}

impl Trace {
    pub fn new(event: TraceEvent) -> Self {
        Self {
            name: "span_trace_group".to_string(),
            timestamp: event.end_time - event.duration,
            metrics: vec![Metric::new(
                "request_total_time".to_string(),
                event.duration,
            )],
            labels: event,
        }
    }
}

#[allow(non_snake_case)]
#[derive(Debug, Clone, Serialize)]
pub(crate) struct Metric {
    Name: String,
    Data: HashMap<String, u64>,
}

impl Metric {
    fn new(name: String, value: u64) -> Self {
        let mut map: HashMap<String, u64> = HashMap::new();
        map.insert("Value".to_string(), value);

        Self {
            Name: name,
            Data: map,
        }
    }
}
