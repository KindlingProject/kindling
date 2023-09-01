use std::time::{Duration, SystemTime};

#[derive(Debug)]
pub struct DeleteTid {
    pub pid: u32,
    pub tid: u32,
    pub exit_time: SystemTime,
}

impl Default for DeleteTid {
    fn default() -> Self {
        DeleteTid {
            pid: 0,
            tid: 0,
            exit_time: SystemTime::UNIX_EPOCH + Duration::from_secs(0),
        }
    }
}
