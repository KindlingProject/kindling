use crate::cpu_analyzer::circle_queue::CircleQueue;
use crate::cpu_analyzer::model::TimedEventKind::{
    TimedCpuEventKind, TimedJavaFutexEventKind, TimedTransactionIdEventKind,
};
use crate::model::TraceProfiling;
use crate::traceid_analyzer::TraceEvent;
use chrono::{DateTime, Local};
use serde_derive::Deserialize;
use serde_derive::Serialize;
use std::any::Any;
use std::fmt;
use std::fmt::Debug;

pub trait TimedEvent {
    fn start_timestamp(&self) -> u64;
    fn end_timestamp(&self) -> u64;
    fn kind(&self) -> TimedEventKind;
    fn as_any(&self) -> &dyn Any;
}
#[derive(Default, Debug, Clone, Serialize, Deserialize)]
pub struct CpuEvent {
    #[serde(rename = "startTime")]
    pub start_time: u64,
    #[serde(rename = "endTime")]
    pub end_time: u64,
    #[serde(rename = "typeSpecs")]
    pub type_specs: Vec<u64>,
    #[serde(rename = "runqLatency")]
    pub runq_latency: Vec<u64>,
    #[serde(rename = "timeType")]
    pub time_type: Vec<u8>,
    #[serde(rename = "onInfo")]
    pub on_info: String,
    #[serde(rename = "offInfo")]
    pub off_info: String,
    pub log: String,
    pub stack: String,
}

impl fmt::Display for CpuEvent {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "CpuEvent: start_time={}, end_time={}, type_specs={:?}, runq_latency={:?}, time_type={:?}, on_info={}, off_info={}, log={}, stack={}",
               self.start_time, self.end_time, self.type_specs,
               self.runq_latency, self.time_type, self.on_info, self.off_info, self.log, self.stack)
    }
}

impl TimedEvent for CpuEvent {
    fn start_timestamp(&self) -> u64 {
        self.start_time
    }

    fn end_timestamp(&self) -> u64 {
        self.end_time
    }

    fn kind(&self) -> TimedEventKind {
        TimedCpuEventKind
    }
    fn as_any(&self) -> &dyn Any {
        self
    }
}

#[derive(Debug)]
pub struct TimeSegments {
    pub pid: u32,
    pub tid: u32,
    pub thread_name: String,
    pub base_time: u64,
    pub segments: CircleQueue,
    pub latest_time: u64,
}

impl TimeSegments {
    pub fn new(
        pid: u32,
        tid: u32,
        thread_name: String,
        base_time: u64,
        segments: CircleQueue,
    ) -> Self {
        TimeSegments {
            pid,
            tid,
            thread_name,
            base_time,
            segments,
            latest_time: 0,
        }
    }
    pub fn update_thread_name(&mut self, thread_name: &str) {
        self.thread_name = thread_name.to_string();
    }
    pub fn update_time(&mut self, end_time: u64) {
        self.latest_time = end_time;
    }
}

#[derive(Default, Debug, Serialize)]
pub struct Segment {
    pub start_time: u64,
    pub end_time: u64,
    pub cpu_events: Vec<CpuEvent>,
    pub java_futex_event: Vec<JavaFutexEvent>,
    pub transaction_id_event: Vec<TransactionIdEvent>,
    pub is_send: i32,
    pub index_timestamp: String,
}

impl Clone for Segment {
    fn clone(&self) -> Self {
        Segment {
            start_time: self.start_time,
            end_time: self.end_time,
            cpu_events: self.cpu_events.iter().cloned().collect(),
            java_futex_event: self.java_futex_event.iter().cloned().collect(),
            transaction_id_event: self.transaction_id_event.iter().cloned().collect(),
            is_send: self.is_send,
            index_timestamp: self.index_timestamp.clone(),
        }
    }
}

impl Segment {
    pub fn new(start_time: u64, end_time: u64) -> Self {
        Segment {
            start_time,
            end_time,
            cpu_events: Vec::new(),
            java_futex_event: Vec::new(),
            transaction_id_event: Vec::new(),
            is_send: 0,
            index_timestamp: String::new(),
        }
    }
    pub fn put_cpu_event(&mut self, event: CpuEvent) {
        self.cpu_events.push(event);
    }
    pub fn put_java_futex_event(&mut self, event: JavaFutexEvent) {
        self.java_futex_event.push(event);
    }

    pub fn put_transaction_id_event(&mut self, event: TransactionIdEvent) {
        self.transaction_id_event.push(event);
    }
    pub fn is_not_empty(&self) -> bool {
        !self.cpu_events.is_empty()
    }

    pub fn update_index_timestamp(&mut self) {
        let local_time: DateTime<Local> = Local::now();
        self.index_timestamp = local_time.to_string();
    }

    pub fn to_data_group(
        &self,
        pid: u32,
        tid: u32,
        thread_name: &str,
        trace_event: &TraceEvent,
    ) -> TraceProfiling {
        let mut profiling = TraceProfiling::new(self.start_time, self.end_time, trace_event);
        let mut labels = &mut profiling.labels;
        labels.cpuEvents = serde_json::to_string(&self.cpu_events).unwrap();
        labels.isSent = self.is_send;
        labels.javaFutexEvents = serde_json::to_string(&self.java_futex_event).unwrap();
        labels.pid = pid;
        labels.threadName = thread_name.to_string();
        labels.tid = tid;
        labels.transactionIds = serde_json::to_string(&self.transaction_id_event).unwrap();

        profiling
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct JavaFutexEvent {
    #[serde(rename = "startTime")]
    pub start_time: u64,
    #[serde(rename = "endTime")]
    pub end_time: u64,
    #[serde(rename = "dataValue")]
    pub data_val: String,
}

impl TimedEvent for JavaFutexEvent {
    fn start_timestamp(&self) -> u64 {
        self.start_time
    }

    fn end_timestamp(&self) -> u64 {
        self.end_time
    }

    fn kind(&self) -> TimedEventKind {
        TimedJavaFutexEventKind
    }
    fn as_any(&self) -> &dyn Any {
        self
    }
}

impl fmt::Display for JavaFutexEvent {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "JavaFutexEvent(start_time: {}, end_time: {}, data_val: {})",
            self.start_time, self.end_time, self.data_val
        )
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TransactionIdEvent {
    pub timestamp: u64,
    #[serde(rename = "traceId")]
    pub trace_id: String,
    #[serde(rename = "isEntry")]
    pub is_entry: String,
}

impl TimedEvent for TransactionIdEvent {
    fn start_timestamp(&self) -> u64 {
        self.timestamp
    }

    fn end_timestamp(&self) -> u64 {
        self.timestamp
    }

    fn kind(&self) -> TimedEventKind {
        TimedTransactionIdEventKind
    }
    fn as_any(&self) -> &dyn Any {
        self
    }
}

impl fmt::Display for TransactionIdEvent {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(
            f,
            "TransactionIdEvent: timestamp={}, trace_id={}, is_entry={}",
            self.timestamp, self.trace_id, self.is_entry
        )
    }
}

#[derive(Debug, PartialEq)]
pub enum TimedEventKind {
    TimedCpuEventKind,
    TimedJavaFutexEventKind,
    TimedTransactionIdEventKind,
    TimedApmSpanEventKind,
    TimedInnerCallEventKind,
}
