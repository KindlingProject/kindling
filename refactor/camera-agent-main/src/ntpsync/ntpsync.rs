use std::process::Command;
use std::str;
use std::thread;
use std::time::Duration;

use log::info;

// 全局变量，存储当前的偏移量
static mut CURRENT_OFFSET: i64 = 0;

// Function to get the NTP offset
pub fn get_ntp_offset() -> Result<i64, Box<dyn std::error::Error>> {
    let output = Command::new("ntpdate")
        .arg("-q")
        .arg("-u")
        .arg("10.10.103.148") // Replace with your desired NTP server endpoint
        .output()?;

    if output.status.success() {
        let stdout = str::from_utf8(&output.stdout)?;
        if let Some(adjust_start) = stdout.find("adjust time server ") {
            let adjust_str = &stdout[adjust_start + 19..]; // Offset start after "adjust time server "
            if let Some(offset_start) = adjust_str.find("offset ") {
                let offset_str = &adjust_str[offset_start + 7..]; // Offset start after "offset "
                if let Some(offset_end) = offset_str.find(" sec") {
                    let offset_val = offset_str[..offset_end].trim().parse::<f64>()?;
                    let offset_ns = (offset_val * 1_000_000_000.0).round() as i64;
                    return Ok(offset_ns);
                }
            }
        }
    }

    Err("Failed to get NTP offset".into())
}

// Function to update the current offset every 10 seconds
pub fn update_current_offset() {
    loop {
        match get_ntp_offset() {
            Ok(offset) => {
                // Use unsafe block to update global variable CURRENT_OFFSET
                unsafe {
                    CURRENT_OFFSET = offset;
                }
                info!("update offset: {}", offset);
            }
            Err(e) => {
                eprintln!("Failed to get NTP offset: {}", e);
            }
        }

        // Sleep for 10 seconds
        thread::sleep(Duration::from_secs(10));
    }
}

// Function to get the current offset
pub fn get_current_offset() -> i64 {
    // Use unsafe block to access global variable CURRENT_OFFSET
    unsafe {
        CURRENT_OFFSET
    }
}