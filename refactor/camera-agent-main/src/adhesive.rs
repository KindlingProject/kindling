use std::thread;

use crate::logger::init_logger;
use crate::ntpsync::update_current_offset;
use crate::probe_to_rust::{start_http_server, start_probe_to_rust};
use crossbeam::channel;
use tokio::runtime::Runtime;

pub fn start() {
    init_logger();
    thread::spawn(update_current_offset);
    let rt = Runtime::new().unwrap();

    rt.spawn(async {
        start_http_server().await;
    });
    let profiling_signal_channel = channel::bounded(10000);
    let metric_signal_channel = channel::bounded(10000);

    start_probe_to_rust(profiling_signal_channel, metric_signal_channel);
}
