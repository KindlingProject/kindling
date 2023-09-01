use log::info;

use crate::cpu_analyzer::{self, CpuAnalyzer};
use crate::probe_to_rust::kindling_event;
use crate::probe_to_rust::kindling_event::{
    catchSignalUp, event_params_for_subscribe, getCaptureStatistics, getEventsByInterval,
    initKindlingEventForGo, startProfile, stopProfile, subEventForGo, KindlingEventForGo, SubEvent,
};
use crate::traceid_analyzer::TraceIdAnalyzer;
use std::ffi::{c_void, CStr, CString};
use std::slice;
use std::sync::{Arc, Mutex};

use super::kindling_event::{startProfileDebug, stopProfileDebug};

pub fn sub_event() {
    let subscribe_info = vec![SubEvent {
        category: "".to_string(),
        name: "tracepoint-procexit".to_string(),
        params: Default::default(),
    }];

    if subscribe_info.is_empty() {
        info!("No events are subscribed by cgo receiver. Please check your configuration.");
    } else {
        info!("The subscribed events are: {:?}", subscribe_info);
    }

    for value in subscribe_info {
        //to do. analyze params filed in the value
        let params_list = vec![event_params_for_subscribe {
            name: CString::new("terminator")
                .expect("CString::new failed")
                .into_raw(),
            value: CString::new("").expect("CString::new failed").into_raw(),
        }];

        let name = CString::new(value.name.clone()).unwrap().into_raw();
        let category = CString::new(value.category.clone()).unwrap().into_raw();
        let params = params_list.as_ptr() as *mut c_void;

        unsafe {
            subEventForGo(name, category, params);
            drop(CString::from_raw(name));
            drop(CString::from_raw(category));
        }
    }
}

pub fn get_kindlin_events(ca: &Arc<Mutex<CpuAnalyzer>>, ta: &mut TraceIdAnalyzer) {
    let mut count = 0;

    const KEY_VALUE_ARRAY_SIZE: usize = 16;

    let mut np_kindling_event: Vec<KindlingEventForGo> = vec![KindlingEventForGo::default(); 100];
    let np_kindling_event_ptr: *mut KindlingEventForGo =
        np_kindling_event.as_mut_slice().as_mut_ptr();
    let np_kindling_event_void_ptr: *mut std::ffi::c_void =
        np_kindling_event_ptr as *mut std::ffi::c_void;

    // let mut npKindlingEvent: Vec<kindling::KindlingEventForGo> = vec![kindling::KindlingEventForGo::default(); 1000];
    // npKindlingEvent = npKindlingEvent.as_mut_ptr()as *mut std::ffi::c_void;
    unsafe {
        initKindlingEventForGo(100, np_kindling_event_void_ptr);
    }

    loop {
        let res = unsafe {
            getEventsByInterval(
                100000000,
                np_kindling_event_void_ptr,
                &mut count as *mut _ as *mut libc::c_void,
            )
        };
        if res == 0 {
            let events =
                unsafe { slice::from_raw_parts(np_kindling_event.as_ptr(), count as usize) };
            for i in 0..count {
                let event = &events[i];

                let ev_name = unsafe { CStr::from_ptr(event.name) };
                let ev_name_string = ev_name.to_str().expect("Invalid UTF-8");
                match ev_name_string {
                    kindling_event::CPU_EVENT => cpu_analyzer::consume_cpu_event(event, ca),
                    kindling_event::JAVA_FUTEX_INFO => {
                        // 处理 java futex 的逻辑
                        cpu_analyzer::consume_java_futex_event(event, ca)
                    }
                    kindling_event::TRANSACTION_ID_EVENT => {
                        // 处理 trace_id 的逻辑
                        cpu_analyzer::consume_transaction_id_event(event, ca);
                        ta.consume_traceid_event(event);
                    }
                    kindling_event::PROCESS_EXIT_EVENT => {
                        // 线程退出事件
                        cpu_analyzer::consume_procexit_event(event, ca);
                    }
                    _ => {
                        // 默认情况，处理其他所有情况的逻辑
                    }
                }
            }
        }
        count = 0;
    }
}

pub fn start_profile() {
    if unsafe { startProfile() } == 0 {
        info!("start profile success!");
    }
}

pub fn stop_profile() {
    if unsafe { stopProfile() } == 0 {
        info!("stop profile success!");
    }
}

pub fn start_profile_debug(pid: i32, tid: i32) {
    unsafe { startProfileDebug(pid, tid) }
}

pub fn stop_profile_debug() {
    unsafe { stopProfileDebug() }
}

pub fn get_capture_statistics() {
    unsafe {
        getCaptureStatistics();
    }
}

pub fn catch_signal_up() {
    unsafe {
        catchSignalUp();
    }
}
