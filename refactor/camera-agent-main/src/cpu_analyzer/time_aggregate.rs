use super::cpu_analyzer;
use super::model::{CpuEvent, TimeSegments, TimedEvent};
use super::CpuAnalyzer;
use crate::model::consts;
use log::info;

const CPUTYPE_MAX: usize = 7;
const NANO_TO_SECONDS: u64 = 1_000_000_000;

pub enum CpuTimeType {
    OnCpu = 0,
    File = 1,
    Net = 2,
    Futex = 3,
    Idle = 4,
    Other = 5,
    Epoll = 6,

    Max,
}

impl CpuTimeType {
    pub fn from_usize(value: usize) -> Option<Self> {
        match value {
            0 => Some(CpuTimeType::OnCpu),
            1 => Some(CpuTimeType::File),
            2 => Some(CpuTimeType::Net),
            3 => Some(CpuTimeType::Futex),
            4 => Some(CpuTimeType::Idle),
            5 => Some(CpuTimeType::Other),
            6 => Some(CpuTimeType::Epoll),
            _ => None,
        }
    }
}

pub fn get_metric_name(time_type: &CpuTimeType, metric_name: &str) -> &'static str {
    match time_type {
        CpuTimeType::OnCpu => match metric_name {
            consts::PROFILING_DURATION_SUFFIX => consts::PROFILING_CPU_DURATION_METRIC_NAME,
            consts::PROFILING_COUNT_SUFFIX => consts::PROFILING_CPU_COUNT_METRIC_NAME,
            _ => "undefined",
        },
        CpuTimeType::File => match metric_name {
            consts::PROFILING_DURATION_SUFFIX => consts::PROFILING_FILE_DURATION_METRIC_NAME,
            consts::PROFILING_COUNT_SUFFIX => consts::PROFILING_FILE_COUNT_METRIC_NAME,
            _ => "undefined",
        },
        CpuTimeType::Net => match metric_name {
            consts::PROFILING_DURATION_SUFFIX => consts::PROFILING_NET_DURATION_METRIC_NAME,
            consts::PROFILING_COUNT_SUFFIX => consts::PROFILING_NET_COUNT_METRIC_NAME,
            _ => "undefined",
        },
        CpuTimeType::Futex => match metric_name {
            consts::PROFILING_DURATION_SUFFIX => consts::PROFILING_FUTEX_DURATION_METRIC_NAME,
            consts::PROFILING_COUNT_SUFFIX => consts::PROFILING_FUTEX_COUNT_METRIC_NAME,
            _ => "undefined",
        },
        CpuTimeType::Idle => match metric_name {
            consts::PROFILING_DURATION_SUFFIX => consts::PROFILING_IDLE_DURATION_METRIC_NAME,
            consts::PROFILING_COUNT_SUFFIX => consts::PROFILING_IDLE_COUNT_METRIC_NAME,
            _ => "undefined",
        },
        CpuTimeType::Other => match metric_name {
            consts::PROFILING_DURATION_SUFFIX => consts::PROFILING_OTHER_DURATION_METRIC_NAME,
            consts::PROFILING_COUNT_SUFFIX => consts::PROFILING_OTHER_COUNT_METRIC_NAME,
            _ => "undefined",
        },
        CpuTimeType::Epoll => match metric_name {
            consts::PROFILING_DURATION_SUFFIX => consts::PROFILING_EPOLL_DURATION_METRIC_NAME,
            consts::PROFILING_COUNT_SUFFIX => consts::PROFILING_EPOLL_COUNT_METRIC_NAME,
            _ => "undefined",
        },
        CpuTimeType::Max => "MaxNumOfElements",
    }
}

impl ToString for CpuTimeType {
    fn to_string(&self) -> String {
        match self {
            CpuTimeType::OnCpu => consts::PROFILING_CPU_PREFIX.to_string(),
            CpuTimeType::File => consts::PROFILING_FILE_PREFIX.to_string(),
            CpuTimeType::Net => consts::PROFILING_NET_PREFIX.to_string(),
            CpuTimeType::Futex => consts::PROFILING_FUTEX_PREFIX.to_string(),
            CpuTimeType::Idle => consts::PROFILING_IDLE_PREFIX.to_string(),
            CpuTimeType::Other => consts::PROFILING_OTHER_PREFIX.to_string(),
            CpuTimeType::Epoll => consts::PROFILING_EPOLL_PREFIX.to_string(),
            CpuTimeType::Max => "MaxNumOfElements".to_string(),
        }
    }
}

#[derive(Default, Debug)]
pub struct AggregatedTime {
    // The index indicates the CpuTimeType
    // on, file, net, futex, idle, other, epoll
    // "times" contains the duration each type of time spent
    pub times: [u64; CpuTimeType::Max as usize],
    // "counts" contains the number of times each type of time spent
    pub counts: [u64; CpuTimeType::Max as usize],
    // Runq Time
    pub runq_duration: u64,
}

impl AggregatedTime {
    pub fn len(&self) -> usize {
        self.times.len()
    }
}

impl CpuAnalyzer {
    pub fn get_times(
        &self,
        pid: u32,
        tid: u32,
        start_time: u64,
        end_time: u64,
    ) -> Option<AggregatedTime> {
        let tid_map = self.cpu_pid_events.get(&pid)?;
        let segments = tid_map.get(&tid)?;
        if segments.latest_time < end_time {
            return None;
        }
        let cpu_events = get_cpu_events(segments, start_time, end_time);
        Some(aggregate_time(cpu_events, start_time, end_time))
    }
}

fn get_cpu_events(segments: &TimeSegments, start_time: u64, end_time: u64) -> Vec<&CpuEvent> {
    let max_segment_size = cpu_analyzer::MAX_SEGMENT_SIZE;
    let start_time_second = start_time / NANO_TO_SECONDS;
    let end_time_second = end_time / NANO_TO_SECONDS;
    if end_time_second < segments.base_time
        || start_time_second > segments.base_time + max_segment_size as u64
    {
        return vec![];
    }

    let mut start_index: usize = 0;
    if start_time_second > segments.base_time {
        start_index = (start_time_second - segments.base_time) as usize;
    }
    let mut end_index = end_time_second - segments.base_time;
    if end_index > segments.base_time + max_segment_size as u64 {
        end_index = segments.base_time + max_segment_size as u64;
    }

    let mut events = vec![];
    let mut last_cpu_end_time: u64 = 0;
    for i in start_index..=end_index.min(max_segment_size as u64) as usize {
        let val = segments.segments.get_by_index(i);
        if let Some(segment) = val {
            for e in &segment.cpu_events {
                if e.end_time > last_cpu_end_time {
                    // 基于事件有序性，过滤跨时段的相同cpuEvent数据.
                    events.push(e);
                    last_cpu_end_time = e.end_time;
                }
            }
        }
    }

    let mut start = lower_bound_event(&events, start_time);
    if start > 0 {
        // EqualTo start -= 1;
        start = start.saturating_sub(1);
    }
    let end = lower_bound_event(&events, end_time);

    events[start..end].to_vec()
}

fn aggregate_time(cpu_events: Vec<&CpuEvent>, start_time: u64, end_time: u64) -> AggregatedTime {
    let mut aggregated_time = AggregatedTime::default();
    for event in cpu_events {
        let mut current_time = event.start_timestamp();
        if current_time > end_time || event.end_timestamp() < start_time {
            // Ignore cpuEvents outof Trace TimeRange
            continue;
        }
        for i in 0..event.type_specs.len() {
            let (on_off_time, runq_time) =
                get_on_off_runq_time(event, i, start_time, end_time, current_time);
            if on_off_time > 0 || runq_time > 0 {
                aggregated_time.times[event.time_type[i] as usize] += on_off_time;
                aggregated_time.counts[event.time_type[i] as usize] += 1;
                aggregated_time.runq_duration += runq_time;
            }
            current_time += event.type_specs[i];
        }
    }

    aggregated_time
}

fn get_on_off_runq_time(
    event: &CpuEvent,
    i: usize,
    start_time: u64,
    end_time: u64,
    current_time: u64,
) -> (u64, u64) {
    let mut on_off_time = event.type_specs[i];
    if current_time + on_off_time < start_time || current_time > end_time {
        // 不计算Trace时间段外的数据
        return (0, 0);
    }

    if current_time < start_time {
        // 采用Trace开始时间
        on_off_time = on_off_time + current_time - start_time
    } else if current_time + on_off_time > end_time {
        // 采用Trace结束时间
        on_off_time = end_time - current_time
    }

    let mut runq_time: u64 = 0;
    if event.time_type[i] > 0 {
        runq_time = event.runq_latency[i / 2] * 1000; // us -> ns
        if runq_time > 0 {
            if on_off_time < runq_time {
                info!("[Waring] Runq is larger than OffTime {}", event);
                runq_time = on_off_time;
                on_off_time = 0;
            } else {
                // OffData Minus Runq
                on_off_time -= runq_time;
            }
        }
    }
    (on_off_time, runq_time)
}

fn lower_bound_event(events: &Vec<&CpuEvent>, target_time: u64) -> usize {
    let end = events.len();
    if end == 0 {
        return 0;
    }
    let (mut l, mut r) = (0, end - 1);
    while l < r {
        let mid = (l + r) / 2;
        if events[mid].start_timestamp() >= target_time {
            r = mid;
        } else {
            l = mid + 1;
        }
    }
    if r == end {
        end
    } else {
        r
    }
}
