// pub(super) const STORAGE_FILE: &str = "file";
// pub(super) const STORAGE_ELASTICSEARCH: &str = "elasticsearch";

// #[derive(Clone)]
// pub(crate) struct Config {
//     pub(crate) storage: String,
//     pub(crate) es_config: Option<EsConfig>,
//     pub(crate) file_config: Option<FileConfig>,
// }
// #[derive(Clone)]
// pub(crate) struct EsConfig {
//     pub(crate) es_host: String,
//     pub(crate) index_suffix: String,
// }
// #[derive(Clone)]
// pub(crate) struct FileConfig {
//     // StoragePath is the ABSOLUTE path of the directory where the profile file should be saved
//     pub(crate) storage_path: String,
//     // Storage constrains for each process
//     pub(crate) max_file_count_each_process: i32,
// }

// pub(crate) fn new_default_config() -> Config {
//     return Config {
//         storage: String::from(STORAGE_FILE),
//         file_config: Some(FileConfig {
//             storage_path: String::from("/tmp/kindling/"),
//             max_file_count_each_process: 50,
//         }),
//         es_config: None,
//     };
// }
