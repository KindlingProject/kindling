use crate::camera_exporter::camera_exporter::CameraExporter;
use crate::config::SampleProcessorCfg;
use crate::sample_processor::sampler::{SampleCache, SampleTrace};
use crate::traceid_analyzer::TraceEvent;
use crossbeam::channel::Sender;
use log::info;
use std::sync::{Arc, Condvar, Mutex};
use std::thread;
use std::time::Duration;

const CHECK_INTERVAL: Duration = Duration::from_secs(1);

pub struct SampleProcessor {
    sample_cache: Arc<Mutex<SampleCache>>,
    trace_retry_num: i32,
    ignore_threshold: u64,
    query_grpc_periods: i32,
    running: Arc<(Mutex<bool>, Condvar)>,
    thread: Mutex<Option<thread::JoinHandle<()>>>,
}

impl SampleProcessor {
    pub fn new(
        cfg: &SampleProcessorCfg,
        profiling_signal_sender: Sender<TraceEvent>,
        trace_exporter: Arc<CameraExporter>,
    ) -> Self {
        Self {
            sample_cache: Arc::new(Mutex::new(SampleCache::new(
                cfg,
                profiling_signal_sender,
                trace_exporter,
            ))),
            trace_retry_num: cfg.sample_trace_repeat_num.unwrap(),
            ignore_threshold: cfg.sample_trace_ignore_threshold.unwrap() * 1000000,
            query_grpc_periods: cfg.promethues_query_interval.unwrap(),
            running: Arc::new((Mutex::new(false), Condvar::new())),
            thread: Mutex::new(None),
        }
    }

    pub fn start(&self) {
        {
            let (started, _) = &*self.running;
            let mut started = started.lock().unwrap();
            if *started {
                return;
            }
            *started = true;
        }

        let sample_cache = Arc::clone(&self.sample_cache);
        let running = self.running.clone();
        let p9x_interval = self.query_grpc_periods;
        let thread = thread::Builder::new()
            .name("sample-processor-check".to_owned())
            .spawn(move || {
                let mut count = p9x_interval;
                loop {
                    do_work(&sample_cache, count);
                    if count == 0 {
                        count = p9x_interval;
                    } else {
                        count -= 1;
                    }

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
                info!("SampleProcessor exited");
            })
            .unwrap();
        self.thread.lock().unwrap().replace(thread);
        info!("SampleProcessor started");
    }

    pub fn stop(&self) {
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
        info!("SampleProcessor stopped");
    }

    pub fn consume(&self, event: TraceEvent, is_sampled: bool) {
        if event.duration < self.ignore_threshold {
            // Ignore Low Valued Datas
            return;
        }
        let trace = SampleTrace::new(event, is_sampled, self.trace_retry_num);
        process_event(&self.sample_cache, trace, is_sampled);
    }
}

fn do_work(sample_cache: &Arc<Mutex<SampleCache>>, count: i32) {
    let mut cache_guard = sample_cache.lock().unwrap();
    // Check tailBase Traces and clean expired traceIds per second.
    cache_guard.check_tail_base_traces();
    // Send local sampled traceIds to receiver
    // Get tailbase sampled traceIds from receiver
    cache_guard.send_and_recv_sampled_traces();

    if count <= 0 {
        cache_guard.update_p9x_by_grpc();
    }
}

fn process_event(sample_cache: &Arc<Mutex<SampleCache>>, mut trace: SampleTrace, is_sampled: bool) {
    let mut cache_guard = sample_cache.lock().unwrap();
    if !is_sampled && cache_guard.is_sampled(&mut trace) {
        // Store Trace and Profiling
        //info!("Start Profile");
        cache_guard.store_profiling(&mut trace, false);
        cache_guard.store_trace(&trace);
    } else if cache_guard.is_tail_base_sampled(&trace) {
        cache_guard.tail_base_profiling(&mut trace);
    } else {
        // Store datas into SampleCache for none-error, none slow or hit datas in N seconds.
        cache_guard.cache_sample_trace(trace)
    }
}
