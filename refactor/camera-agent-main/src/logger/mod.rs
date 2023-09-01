mod system_info_log;
use log::info;

use crate::logger::system_info_log::{print_hardware_info};

use self::system_info_log::print_system_info;

pub fn init_logger() {
    log4rs::init_file("config/log4rs.yml", Default::default()).unwrap();
    info!("Log initialization completed.");
    //print_init_info();  
}

fn print_init_info() {
    print_hardware_info();
    print_system_info();
}
