// use std::fs::{self, DirEntry, OpenOptions};
// use std::io::{self, Write};
// use std::path::{Path, PathBuf};
// use std::time::{SystemTime, UNIX_EPOCH};

// use super::config::FileConfig;
// use crate::model::{consts, Trace, TraceProfiling};
// pub(crate) struct FileWriter {
//     config: FileConfig,
// }
// const DIVIDING_LINE: &str = "\n------\n";

// impl FileWriter {
//     fn new(config: FileConfig) -> io::Result<Self> {
//         Ok(FileWriter { config })
//     }

//     fn pid_file_path(
//         &self,
//         workload_name: &str,
//         pod_name: &str,
//         container_name: &str,
//         pid: i64,
//     ) -> PathBuf {
//         let pid_str = pid.to_string();
//         let mut path = PathBuf::new();
//         path.push(&self.config.storage_path);
//         path.push(workload_name);
//         path.push(pod_name);
//         path.push(container_name);
//         path.push(pid_str);
//         path
//     }

//     fn get_file_name(protocol: &str, content_key: &str, timestamp: u64, is_server: bool) -> String {
//         let mut file_name = String::new();
//         file_name.push_str(protocol);
//         file_name.push_str("_");
//         file_name.push_str(content_key);
//         file_name.push_str("_");
//         file_name.push_str(&timestamp.to_string());
//         if is_server {
//             file_name.push_str("_server");
//         } else {
//             file_name.push_str("_client");
//         }
//         file_name
//     }

//     fn write_trace(
//         &self,
//         group: model::Trace,
//     ) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
//         let trace_timestamp = group.labels.get_int_value(consts::TIMESTAMP);
//         let path_elements = get_file_path_elements(group, trace_timestamp as u64);
//         let base_dir = self.pid_file_path(
//             &path_elements.workload_name,
//             &path_elements.pod_name,
//             &path_elements.container_name,
//             path_elements.pid,
//         );
//         let file_name = Self::get_file_name(
//             &path_elements.protocol,
//             &path_elements.content_key,
//             path_elements.timestamp,
//             path_elements.is_server,
//         );
//         let file_path = base_dir.join(&file_name);
//         let mut file = OpenOptions::new()
//             .create(true)
//             .append(true)
//             .open(&file_path)?;
//         writeln!(file, "------")?;
//         let events_bytes = serde_json::to_vec(group)?;
//         file.write_all(&events_bytes)?;
//         Ok(())
//     }

//     fn write_profiling(
//         &self,
//         group: model::TraceProfiling,
//     ) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
//         let trace_timestamp = group.labels.get_int_value(consts::TIMESTAMP);
//         let path_elements = get_file_path_elements(group, trace_timestamp as u64);
//         let base_dir = self.pid_file_path(
//             &path_elements.workload_name,
//             &path_elements.pod_name,
//             &path_elements.container_name,
//             path_elements.pid,
//         );
//         let file_name = Self::get_file_name(
//             &path_elements.protocol,
//             &path_elements.content_key,
//             path_elements.timestamp,
//             path_elements.is_server,
//         );
//         let file_path = base_dir.join(&file_name);
//         let mut file = OpenOptions::new()
//             .create(true)
//             .append(true)
//             .open(&file_path)?;
//         writeln!(file, "{}", DIVIDING_LINE)?;
//         let events_bytes = serde_json::to_vec(group)?;
//         file.write_all(&events_bytes)?;
//         Ok(())
//     }

//     fn rotate_files(&self, base_dir: &Path) -> io::Result<()> {
//         let mut entries = fs::read_dir(base_dir)?
//             .map(|res| res.map(|e| e.path()))
//             .collect::<Result<Vec<_>, io::Error>>()?;
//         entries.sort_by_key(|entry| entry.metadata().unwrap().modified().unwrap());
//         let max_files = self.config.max_files;
//         if entries.len() > max_files {
//             let num_files_to_delete = entries.len() - max_files;
//             for entry in entries.iter().take(num_files_to_delete) {
//                 fs::remove_file(entry)?;
//             }
//         }
//         Ok(())
//     }

//     fn get_dir_entry_in_time_order(path: &Path) -> io::Result<Vec<DirEntry>> {
//         let mut entries = fs::read_dir(path)?.collect::<Result<Vec<_>, io::Error>>()?;
//         entries.sort_by_key(|entry| entry.metadata().unwrap().modified().unwrap());
//         Ok(entries)
//     }

//     fn name(&self) -> &str {
//         config::STORAGE_FILE
//     }
// }

// fn get_date_string(timestamp: i64) -> String {
//     let time_unix = UNIX_EPOCH + std::time::Duration::from_nanos(timestamp as u64);
//     let time = time_unix.duration_since(SystemTime::UNIX_EPOCH).unwrap();
//     let date = time.as_secs();
//     let nano = time.subsec_nanos();
//     let year = (date / 31536000) as i32;
//     let month = ((date - (year as u64 * 31536000)) / 2592000) as u32;
//     let day = ((date - (year as u64 * 31536000) - (month as u64 * 2592000)) / 86400) as u32;
//     let hour = ((date - (year as u64 * 31536000) - (month as u64 * 2592000) - (day as u64 * 86400))
//         / 3600) as u32;
//     let minute = ((date
//         - (year as u64 * 31536000)
//         - (month as u64 * 2592000)
//         - (day as u64 * 86400)
//         - (hour as u64 * 3600))
//         / 60) as u32;
//     let second = date
//         - (year as u64 * 31536000)
//         - (month as u64 * 2592000)
//         - (day as u64 * 86400)
//         - (hour as u64 * 3600)
//         - (minute as u64 * 60);
//     return format!(
//         "{}{}{}{}{}{}.{}",
//         date_to_string(year),
//         date_to_string(month),
//         date_to_string(day),
//         date_to_string(hour),
//         date_to_string(minute),
//         date_to_string(second),
//         nano
//     );
// }

// fn date_to_string(date: u32) -> String {
//     if date >= 0 && date <= 9 {
//         return format!("0{}", date);
//     } else {
//         return format!("{}", date);
//     }
// }

// pub struct FilePathElements {
//     pub workload_name: String,
//     pub pod_name: String,
//     pub container_name: String,
//     pub pid: i64,
//     pub is_server: bool,
//     pub protocol: String,
//     pub timestamp: u64,
//     pub content_key: String,
// }

// pub fn get_file_path_elements(group: &TraceProfiling, timestamp: u64) -> FilePathElements {
//     FilePathElements {
//         workload_name: group.workload_name,
//         pod_name: group.pod_name,
//         container_name: group.container_name,
//         pid: group.pid,
//         is_server: group.is_server,
//         protocol: group.protocol,
//         timestamp: group.timestamp,
//         content_key: group.content_key,
//     }
// }
