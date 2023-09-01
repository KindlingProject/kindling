static mut ENABLE_PROFILE: bool = true;

pub fn is_profiled_enabled() -> bool {
    unsafe { ENABLE_PROFILE }
}
