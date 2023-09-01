mod kindling_event;
mod rust_receiver;

use crate::camera_exporter::camera_exporter::CameraExporter;
use crate::config;
use crate::controller::start_serve;
use crate::cpu_analyzer::{clear_tid, delay_send_metrics, CpuAnalyzer};
use crate::kubernetes::k8scache;
use crate::probe_to_rust::kindling_event::{runForGo, startProfile};
use crate::probe_to_rust::rust_receiver::{
    catch_signal_up, get_capture_statistics, get_kindlin_events, start_profile_debug,
    stop_profile_debug,sub_event,
};
use crate::sample_processor::SampleProcessor;
use crate::traceid_analyzer::{SignalEvent, TraceEvent, TraceIdAnalyzer};
use log::{debug, info};
use std::sync::{Arc, Mutex};
use std::thread;
use std::time::{Duration, Instant};
use warp::{Filter, Rejection, Reply};

use crate::metric_exporter;
use crossbeam::channel::{Receiver, Sender};
pub use kindling_event::sampleUrl;
pub use kindling_event::KindlingEventForGo;
pub use kindling_event::{startProfileDebug, stopProfileDebug};

pub fn start_probe_to_rust(
    profiling_signal_chan: (Sender<TraceEvent>, Receiver<TraceEvent>),
    metric_signal_chan: (Sender<Vec<SignalEvent>>, Receiver<Vec<SignalEvent>>),
) {
    // 初始化probe
    unsafe { runForGo() };
    unsafe { startProfile() };

    // 读取配置
    let conf = config::load_conf();

    // 订阅事件
    sub_event();

    let metric_exporter = get_metric_exporter(&conf);
    // 创建prometheus-exporter并启动webserver

    // 创建 K8S缓存
    let k8s_cache = Arc::new(k8scache::K8sMetaDataCache::new());

    // 创建Exproter
    let storage_exporter = Arc::new(CameraExporter::new(&conf.exporters.camera_exporter).unwrap());

    // 初始化on-off cpu分析器
    let cpu_analyzer = Arc::new(Mutex::new(CpuAnalyzer::new(
        metric_exporter.clone(),
        storage_exporter.clone(),
    )));

    // Start receive metrics signal
    let metric_signal_receiver = metric_signal_chan.1;
    let cpu_analyzer_for_metric = cpu_analyzer.clone();
    thread::spawn(move || loop {
        let signals = metric_signal_receiver.recv().unwrap();
        info!("signal {} metrics", signals.len());
        cpu_analyzer_for_metric
            .lock()
            .unwrap()
            .handle_metric_signal(signals);
    });

    // 初始化 trace-id 分析器
    let sample_processor = SampleProcessor::new(
        &conf.processors.sample_processor,
        profiling_signal_chan.0,
        storage_exporter,
    );
    let mut trace_id_analyzer = TraceIdAnalyzer::new(
        &conf.analyzers.trace_id_analyzer,
        sample_processor,
        metric_signal_chan.0,
        metric_exporter,
        k8s_cache,
    );

    // 启动内核事件统计
    thread::spawn(move || {
        get_capture_statistics();
    });

    let cpu_analyzer_clone_clear_tid = Arc::clone(&cpu_analyzer);
    thread::spawn(move || {
        let mut start_time = Instant::now(); // 记录起始时间
        let interval = Duration::from_secs(30); // 指定间隔为0秒

        loop {
            // 等待间隔时间
            thread::sleep(interval);

            // 检查是否已经过了指定的间隔时间
            if start_time.elapsed() >= interval {
                // 执行要定期执行的代码
                clear_tid(&cpu_analyzer_clone_clear_tid, Duration::from_secs(10));

                // 重置起始时间
                start_time = Instant::now();
            }
        }
    });

    let cpu_analyzer_clone_delay_send_metrics = Arc::clone(&cpu_analyzer);
    thread::spawn(move || loop {
        // 等待间隔时间
        thread::sleep(Duration::from_secs(1));
        delay_send_metrics(&cpu_analyzer_clone_delay_send_metrics);
    });

    // 启动异常退出打印gdb日志
    thread::spawn(move || {
        catch_signal_up();
    });

    // 启动组件
    trace_id_analyzer.start();

    // 开始获取事件
    let cpu_analyzer_clone = Arc::clone(&cpu_analyzer);
    thread::spawn(move || {
        get_kindlin_events(&cpu_analyzer_clone, &mut trace_id_analyzer);
    });
    // Start receive profiling signal
    let profiling_signal_receiver = profiling_signal_chan.1;
    let cpu_analyzer_for_profiling = cpu_analyzer.clone();
    loop {
        let signal = profiling_signal_receiver.recv().unwrap();
        debug!("signal profile: {:?}", &signal);
        let cpu_analyzer_for_profiling = cpu_analyzer_for_profiling.clone();
        thread::spawn(move || {
            thread::sleep(Duration::from_secs(2));
            let mut cpu_analyzer_for_profiling = cpu_analyzer_for_profiling.lock().unwrap();
            cpu_analyzer_for_profiling.handle_profiling_signal(&signal);
        });
    }
}

struct DebugInfo {
    debug_pid: i32,
    debug_tid: i32,
    is_open: bool,
}

fn get_metric_exporter(conf: &config::GlobalConfig) -> Arc<dyn metric_exporter::MetricExporter> {
    let metric_exporter: Arc<dyn metric_exporter::MetricExporter> = match conf
        .exporters
        .metric_exporter
        .exporter_type
        .as_ref()
        .unwrap()
        .as_str()
    {
        config::PROMETHEUS_EXPORTER => {
            let prometheus_exporter = Arc::new(metric_exporter::PromExporter::new());
            prometheus_exporter.runtime.spawn(start_serve(
                prometheus_exporter.clone(),
                (
                    [0, 0, 0, 0],
                    conf.exporters.metric_exporter.prometheus.port.unwrap(),
                ),
            ));
            prometheus_exporter
        }
        config::LOG_EXPORTER => {
            let log_exporter = Arc::new(metric_exporter::log_exporter::LogExporter::new());
            log_exporter
        }
        &_ => {
            let log_exporter = Arc::new(metric_exporter::log_exporter::LogExporter::new());
            log_exporter
        }
    };
    metric_exporter
}

pub async fn start_http_server() {
    info!("start http server 19877!");
    // 定义路由过滤器
    let update_cpu_debug =
        warp::path!("updateCpuDebug" / "debug_pid" / i32 / "debug_tid" / i32 / "is_open" / bool)
            .map(|debug_pid: i32, debug_tid: i32, is_open: bool| {
                // 创建DebugInfo结构体并打印
                let debug_info = DebugInfo {
                    debug_pid,
                    debug_tid,
                    is_open,
                };
                println!(
                    "Received debug pid: {}   tid: {}    is_open: {}",
                    debug_info.debug_pid, debug_info.debug_tid, debug_info.is_open
                );
                if debug_info.is_open {
                    start_profile_debug(debug_info.debug_pid, debug_info.debug_tid);
                } else {
                    stop_profile_debug();
                }

                // 返回响应
                warp::reply::html("Received and processed the debug info")
            })
            .recover(handle_rejection);

    // 启动服务器并绑定到指定端口
    warp::serve(update_cpu_debug)
        .run(([0, 0, 0, 0], 19877))
        .await;
}

async fn handle_rejection(err: Rejection) -> Result<impl Reply, Rejection> {
    Ok(warp::reply::with_status(
        warp::reply::html("404 Not Found"),
        warp::http::StatusCode::NOT_FOUND,
    ))
}

#[cfg(test)]
mod tests {
    use crate::config;
    use crate::metric_exporter;
    use crate::probe_to_rust::get_metric_exporter;
    use std::any::type_name;
    #[test]
    fn test_exporter() {
        let conf = config::load_conf();
        let exporter = get_metric_exporter(&conf);
        println!("exporter type: {:?}", test_type(exporter));
    }
}
