use crate::camera_exporter::camera_exporter::CameraExporter;
use crate::cpu_analyzer::circle_queue::CircleQueue;
use crate::cpu_analyzer::model::TimedEventKind::TimedCpuEventKind;
use crate::cpu_analyzer::model::{CpuEvent, JavaFutexEvent, TimeSegments, TransactionIdEvent};
use crate::cpu_analyzer::model::{Segment, TimedEvent};
use crate::cpu_analyzer::time_aggregate::{get_metric_name, CpuTimeType};
use crate::probe_to_rust::KindlingEventForGo;
use byteorder::{LittleEndian, ReadBytesExt};
use log::{debug, info};
use std::collections::HashMap;
use std::io::Cursor;
use std::sync::{Arc, Mutex};
use std::time::{Duration, SystemTime};

use super::tid_exit_clear::DeleteTid;
use super::time_aggregate::AggregatedTime;
use crate::metric_exporter::{self};
use crate::model::consts;
use crate::traceid_analyzer::{SignalEvent, TraceEvent};
use opentelemetry_api::KeyValue;
const NANO_TO_SECONDS: u64 = 1_000_000_000;
const CHECK_INTERVAL: Duration = Duration::from_secs(1);

pub(crate) const MAX_SEGMENT_SIZE: usize = 40;

pub(crate) const UPPER_LIMITE_TS: u64 = 2000000000000000000;

pub(crate) const LOWER_LIMITE_TS: u64 = 1600000000000000000;

pub struct CpuAnalyzer {
    pub cpu_pid_events: HashMap<u32, HashMap<u32, TimeSegments>>,
    pub tid_exit_queue: Vec<DeleteTid>,
    pub metric_signals: Vec<SignalEvent>,
    pub metric_exporter: Arc<dyn metric_exporter::MetricExporter>,
    pub profile_exporter: Arc<CameraExporter>,
}

pub fn delay_send_metrics(cca: &Arc<Mutex<CpuAnalyzer>>) {
    let mut cpu_analyzer = cca.lock().unwrap();
    cpu_analyzer.send_remove_metrics();
}

pub fn clear_tid(cca: &Arc<Mutex<CpuAnalyzer>>, expired_duration: Duration) {
    let mut ca_guard = cca.lock().unwrap();
    ca_guard.tid_delete(expired_duration);
}

pub fn consume_cpu_event(event: &KindlingEventForGo, cca: &Arc<Mutex<CpuAnalyzer>>) {
    let mut ev = CpuEvent::default();
    for i in 0..event.params_number as usize {
        let user_attributes = event.user_attributes[i];
        match i {
            0 => ev.start_time = user_attributes.get_uint_value(),
            1 => ev.end_time = user_attributes.get_uint_value(),
            2 => {
                let val = user_attributes.get_value().unwrap();
                ev.type_specs = read_u64_values(val);
            }
            3 => {
                let val = user_attributes.get_value().unwrap();
                ev.runq_latency = read_u64_values(val);
            }
            4 => {
                let val = user_attributes.get_value().unwrap();
                ev.time_type = read_u8_values(val);
            }
            5 => ev.stack = read_string_value(user_attributes.get_value()),
            6 => ev.log = read_string_value(user_attributes.get_value()),
            7 => ev.on_info = read_string_value(user_attributes.get_value()),
            8 => ev.off_info = read_string_value(user_attributes.get_value()),
            _ => (),
        }
    }

    if ev.start_time < LOWER_LIMITE_TS {
        return;
    }

    // if event.get_pid() != 9689 as u32  && event.get_pid() != 4509 as u32 && event.get_pid() != 26600 as u32{
    //     info!("{}", &ev);
    //     info!("tid: {}, thread name: {}", event.get_tid(),&event.get_comm())
    // }

    let mut ca_guard = cca.lock().unwrap();
    if let Some((start, end, segments)) =
        ca_guard.put_event_to_segments(event.get_pid(), event.get_tid(), &event.get_comm(), &ev)
    {
        let mut corrected_start = start;
        if corrected_start < 0 {
            corrected_start = 0;
        }
        for i in corrected_start..end {
            if let Some(segment) = segments.get_by_index_mut(i as usize) {
                segment.put_cpu_event(ev.clone());
                segment.is_send = 0;
            }
        }
        if let Some(segment) = segments.get_by_index_mut(end as usize) {
            segment.put_cpu_event(ev);
            segment.is_send = 0;
        }
    }
}

pub fn consume_java_futex_event(event: &KindlingEventForGo, cca: &Arc<Mutex<CpuAnalyzer>>) {
    let ev = JavaFutexEvent {
        start_time: event.timestamp,
        end_time: event.user_attributes[0].get_uint_value(),
        data_val: read_string_value(event.user_attributes[1].get_value()),
    };
    //debug!("{}", &ev);

    let mut ca_guard = cca.lock().unwrap();
    if let Some((start, end, segments)) =
        ca_guard.put_event_to_segments(event.get_pid(), event.get_tid(), &event.get_comm(), &ev)
    {
        for i in start..end {
            if let Some(segment) = segments.get_by_index_mut(i as usize) {
                segment.put_java_futex_event(ev.clone());
                segment.is_send = 0;
            }
        }
        if let Some(segment) = segments.get_by_index_mut(end as usize) {
            segment.put_java_futex_event(ev);
            segment.is_send = 0;
        }
    }
}

pub fn consume_transaction_id_event(event: &KindlingEventForGo, cca: &Arc<Mutex<CpuAnalyzer>>) {
    let ev = TransactionIdEvent {
        timestamp: event.timestamp,
        trace_id: read_string_value(event.user_attributes[0].get_value()),
        is_entry: read_string_value(event.user_attributes[1].get_value()),
    };
    info!("{}", &ev);

    let mut ca_guard = cca.lock().unwrap();
    if let Some((start, end, segments)) =
        ca_guard.put_event_to_segments(event.get_pid(), event.get_tid(), &event.get_comm(), &ev)
    {
        for i in start..end {
            if let Some(segment) = segments.get_by_index_mut(i as usize) {
                segment.put_transaction_id_event(ev.clone());
                segment.is_send = 0;
            }
        }
        if let Some(segment) = segments.get_by_index_mut(end as usize) {
            segment.put_transaction_id_event(ev);
            segment.is_send = 0;
        }
    }
}

pub fn consume_procexit_event(event: &KindlingEventForGo, ca: &Arc<Mutex<CpuAnalyzer>>) {
    let pid = event.get_pid();
    let tid = event.get_tid();
    debug!("Receive a procexit event: pid={}, tid={}", pid, tid);
    let mut ca_guard = ca.lock().unwrap();
    ca_guard.trim_exited_thread(pid, tid);
}

impl CpuAnalyzer {
    pub fn new(
        metric_exporter: Arc<dyn metric_exporter::MetricExporter>,
        profile_exporter: Arc<CameraExporter>,
    ) -> Self {
        CpuAnalyzer {
            cpu_pid_events: HashMap::new(),
            tid_exit_queue: Vec::new(),
            metric_signals: Vec::new(),
            metric_exporter,
            profile_exporter,
        }
    }

    pub fn handle_profiling_signal(&mut self, event: &TraceEvent) {
        if event.is_normal {
            self.send_event(
                event.pid,
                event.tid,
                event.end_time - event.duration,
                event.end_time,
                event,
            );
        } else {
            self.send_events(
                event.pid,
                event.end_time - event.duration,
                event.end_time,
                event,
            );
        }
    }

    pub fn handle_metric_signal(&mut self, mut events: Vec<SignalEvent>) {
        events.drain(..).for_each(|event| {
            debug!("signal metric: {:?}", &event);
            if self.send_metric(&event).is_none() {
                // Cache SignalEvent to DelayQueue.
                self.cache_metric_signal(event);
            }
        });
    }

    pub fn send_remove_metrics(&mut self) {
        if self.metric_signals.is_empty() {
            return;
        }

        let mut removed_indices = vec![];
        // Send Metrics
        for (i, value) in self.metric_signals.iter().enumerate() {
            if self.send_metric(value).is_some() {
                removed_indices.push(i);
            }
        }

        let mut offset = 0;
        // Remove Sent Metrics
        self.metric_signals.retain(|_| {
            let current_index = offset;
            offset += 1;
            !removed_indices.contains(&current_index)
        });
    }

    fn cache_metric_signal(&mut self, event: SignalEvent) {
        self.metric_signals.push(event);
    }

    fn send_metric(&self, event: &SignalEvent) -> Option<AggregatedTime> {
        let agg = self.get_times(event.pid, event.tid, event.start_time, event.end_time)?;
        debug!("Metric Aggregate: {:?}", agg);
        let attributes = vec![
            KeyValue::new(consts::PID, event.pid.to_string()),
            KeyValue::new(consts::CONTENT_KEY, event.content_key.clone()),
            KeyValue::new(consts::CONTAINER_ID, event.container_id.clone()),
        ];
        for i in 0..agg.len() {
            let time_type = CpuTimeType::from_usize(i).unwrap();
            let duration = agg.times[i];
            let duration_name = get_metric_name(&time_type, consts::PROFILING_DURATION_SUFFIX);
            self.metric_exporter
                .record_metric(duration_name, duration, &attributes);
            let count = agg.counts[i];
            let count_name = get_metric_name(&time_type, consts::PROFILING_COUNT_SUFFIX);
            self.metric_exporter
                .record_metric(count_name, count, &attributes);
        }
        if agg.runq_duration > 0 {
            self.metric_exporter.record_metric(
                consts::PROFILING_RUNQ_DURATION_METRIC_NAME,
                agg.runq_duration,
                &attributes,
            );
        }
        Some(agg)
    }

    pub fn put_event_to_segments(
        &mut self,
        pid: u32,
        tid: u32,
        thread_name: &str,
        event: &dyn TimedEvent,
    ) -> Option<(i32, i32, &mut CircleQueue)> {
        if event.end_timestamp() > UPPER_LIMITE_TS || event.start_timestamp() > UPPER_LIMITE_TS {
            info!(
                "filiter err time event!  start: {:?}   end: {:?}    kind: {:?}!",
                event.start_timestamp(),
                event.end_timestamp(),
                event.kind()
            );
            return None;
        }
        let tid_cpu_events = self.cpu_pid_events.entry(pid).or_insert_with(HashMap::new);
        let time_segments = tid_cpu_events.entry(tid).or_insert_with(|| {
            let base_time = event.start_timestamp() / NANO_TO_SECONDS;
            let segments = create_initial_segments(base_time);
            TimeSegments::new(pid, tid, thread_name.to_string(), base_time, segments)
        });
        if event.kind() == TimedCpuEventKind {
            // Set LatestTime.
            time_segments.update_time(event.end_timestamp());
        }

        if event.end_timestamp() / NANO_TO_SECONDS < time_segments.base_time {
            return None;
        }
        let mut end_offset =
            (event.end_timestamp() / NANO_TO_SECONDS - time_segments.base_time) as i32;

        let mut start_offset =
            (event.start_timestamp() / NANO_TO_SECONDS - time_segments.base_time) as i32;
        let should_clear_segments =
            start_offset >= MAX_SEGMENT_SIZE as i32 || end_offset > MAX_SEGMENT_SIZE as i32;

        if should_clear_segments {
            if start_offset * 2 >= 3 * MAX_SEGMENT_SIZE as i32 {
                time_segments.segments.clear();
                time_segments.base_time = event.start_timestamp() / NANO_TO_SECONDS;
                end_offset -= start_offset;
                start_offset = 0;
                time_segments.segments = create_initial_segments(time_segments.base_time);
            } else if end_offset > MAX_SEGMENT_SIZE as i32 {
                time_segments.segments.clear();
                time_segments.base_time =
                    (event.end_timestamp() / NANO_TO_SECONDS) - ((MAX_SEGMENT_SIZE / 2) as u64);
                start_offset = ((event.start_timestamp() / NANO_TO_SECONDS)
                    - time_segments.base_time as u64) as i32;
                end_offset = (MAX_SEGMENT_SIZE / 2) as i32;
                time_segments.segments = create_initial_segments(time_segments.base_time);
            } else {
                let clear_size = MAX_SEGMENT_SIZE / 2;
                time_segments.base_time += clear_size as u64;
                if start_offset < clear_size as i32 {
                    start_offset = 0;
                } else {
                    start_offset -= clear_size as i32;
                }
                end_offset -= clear_size as i32;
                for i in 0..clear_size {
                    let moved_index = i + clear_size;
                    if let Some(_segment) = time_segments.segments.get_by_index(moved_index) {
                        time_segments.segments.update_by_moved_index(i, moved_index);
                    }
                    let segment_tmp = Segment::new(
                        (time_segments.base_time + (moved_index as u64)) * NANO_TO_SECONDS,
                        (time_segments.base_time + ((moved_index + 1) as u64)) * NANO_TO_SECONDS,
                    );
                    time_segments
                        .segments
                        .update_by_index(moved_index, segment_tmp);
                }
            }
        }
        time_segments.update_thread_name(thread_name);

        end_offset = end_offset.min(MAX_SEGMENT_SIZE as i32 - 1);
        Some((start_offset, end_offset, &mut time_segments.segments))
    }

    pub fn send_event(
        &mut self,
        pid: u32,
        tid: u32,
        start_time: u64,
        end_time: u64,
        trace_event: &TraceEvent,
    ) {
        let tid_cpu_events = match self.cpu_pid_events.get_mut(&pid) {
            Some(tid_cpu_events) => tid_cpu_events,
            None => {
                return;
            }
        };
        let time_segments = match tid_cpu_events.get_mut(&tid) {
            Some(time_segments) => time_segments,
            None => {
                return;
            }
        };

        let start_time_second = start_time / NANO_TO_SECONDS;
        let end_time_second = end_time / NANO_TO_SECONDS;

        if end_time_second < time_segments.base_time
            || start_time_second > time_segments.base_time + (MAX_SEGMENT_SIZE as u64)
        {
            return;
        }

        let start_index = (start_time_second - time_segments.base_time) as i32;
        let end_index =
            (end_time_second - time_segments.base_time).min(MAX_SEGMENT_SIZE as u64) as i32;

        let thread_name = time_segments.thread_name.as_str();
        for i in start_index..=end_index {
            if let Some(segment) = time_segments.segments.get_by_index_mut(i as usize) {
                if segment.is_not_empty() {
                    segment.update_index_timestamp();
                    // Send To Es.
                    let profiling_event = segment.to_data_group(pid, tid, thread_name, trace_event);
                    self.profile_exporter.consume_profiling(profiling_event);
                    debug!("{:?}", segment);
                    segment.is_send = 1;
                }
            }
        }
    }

    pub fn send_events(
        &mut self,
        pid: u32,
        start_time: u64,
        end_time: u64,
        trace_event: &TraceEvent,
    ) {
        let tid_cpu_events = match self.cpu_pid_events.get_mut(&pid) {
            Some(tid_cpu_events) => tid_cpu_events,
            None => {
                return;
            }
        };

        let start_time_second = start_time / NANO_TO_SECONDS;
        let end_time_second = end_time / NANO_TO_SECONDS;

        for time_segments in tid_cpu_events.values_mut() {
            if end_time_second < time_segments.base_time
                || start_time_second > time_segments.base_time + (MAX_SEGMENT_SIZE as u64)
            {
                continue;
            }

            let start_index = (start_time_second - time_segments.base_time) as i32;
            let end_index =
                (end_time_second - time_segments.base_time).min(MAX_SEGMENT_SIZE as u64) as i32;

            let pid = time_segments.pid;
            let tid = time_segments.tid;
            let thread_name = time_segments.thread_name.as_str();
            for i in start_index..=end_index {
                if let Some(segment) = time_segments.segments.get_by_index_mut(i as usize) {
                    if segment.is_not_empty() {
                        segment.update_index_timestamp();
                        // Send To Es.
                        let profiling_event =
                            segment.to_data_group(pid, tid, thread_name, trace_event);
                        self.profile_exporter.consume_profiling(profiling_event);
                        debug!("{:?}", segment);
                        segment.is_send = 1;
                    }
                }
            }
        }
    }

    pub fn trim_exited_thread(&mut self, pid: u32, tid: u32) {
        let cache_elem = DeleteTid {
            pid,
            tid,
            exit_time: SystemTime::now(),
        };

        self.tid_exit_queue.push(cache_elem);
    }

    pub fn tid_delete(&mut self, expired_duration: Duration) {
        let now = SystemTime::now();

        loop {
            if let Some(elem) = self.tid_exit_queue.first() {
                if elem.exit_time + expired_duration > now {
                    break;
                }

                if let Some(tid_events_map) = self.cpu_pid_events.get_mut(&elem.pid) {
                    tid_events_map.remove(&elem.tid);
                }

                self.tid_exit_queue.remove(0);
            } else {
                break;
            }
        }
    }
}

fn read_u64_values(val: &[u8]) -> Vec<u64> {
    let mut cursor = Cursor::new(val);
    let count = val.len() / 8;
    let mut values = Vec::with_capacity(count);
    for _ in 0..count {
        values.push(cursor.read_u64::<LittleEndian>().unwrap());
    }
    values
}

fn read_u8_values(val: &[u8]) -> Vec<u8> {
    val.to_vec()
}

fn read_string_value(val: Option<&[u8]>) -> String {
    match val {
        Some(bytes) => {
            let mut string = String::with_capacity(bytes.len());
            for &byte in bytes {
                string.push(byte as char);
            }

            string
        }
        None => String::new(),
    }
}

fn create_initial_segments(base_time: u64) -> CircleQueue {
    let mut segments = CircleQueue::new(MAX_SEGMENT_SIZE);
    for i in 0..MAX_SEGMENT_SIZE {
        let segment = Segment::new(
            (base_time + (i as u64)) * NANO_TO_SECONDS,
            (base_time + (i as u64) + 1) * NANO_TO_SECONDS,
        );
        segments.update_by_index(i, segment);
    }
    segments
}
