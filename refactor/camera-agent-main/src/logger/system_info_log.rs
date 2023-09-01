use libc::{uname, utsname};
use log::{info, error};
use sysinfo::{System, SystemExt};

pub fn print_hardware_info() {
    let mut system = System::new_all();

    system.refresh_memory();
    system.refresh_system();


    info!("total memory:{} MB", system.total_memory()/1024/1024);
    info!("free memory:{} MB", system.free_memory()/1024/1024);
    info!("total swap:{} MB", system.total_swap()/1024/1024);
    info!("free swap:{} MB", system.free_swap()/1024/1024);
    info!("available memory:{} MB", system.available_memory()/1024/1024);
    info!("cpu number:{} ", system.cpus().len());

    info!("kernel version:{} ", system.kernel_version().unwrap());
}



pub fn print_system_info() {
    // Create a `utsname` structure to store the system information
    let mut sys_info: utsname = unsafe { std::mem::zeroed() };

    // Call the `uname` function to retrieve system information
    let result = unsafe { uname(&mut sys_info as *mut utsname) };
    if result == 0 {
        // Extract system information from the `utsname` structure
        let sys_name = unsafe { std::ffi::CStr::from_ptr(sys_info.sysname.as_ptr()) };
        let node_name = unsafe { std::ffi::CStr::from_ptr(sys_info.nodename.as_ptr()) };
        let release = unsafe { std::ffi::CStr::from_ptr(sys_info.release.as_ptr()) };
        let version = unsafe { std::ffi::CStr::from_ptr(sys_info.version.as_ptr()) };
        let machine = unsafe { std::ffi::CStr::from_ptr(sys_info.machine.as_ptr()) };

        // Print the system information
        info!("System information:");
        info!("System Name: {}", sys_name.to_string_lossy());
        info!("Node Name: {}", node_name.to_string_lossy());
        info!("Release: {}", release.to_string_lossy());
        info!("Version: {}", version.to_string_lossy());
        info!("Machine: {}", machine.to_string_lossy());
    } else {
        let errno = std::io::Error::last_os_error();
        error!("Failed to retrieve system information: {}", errno);
    }
}



