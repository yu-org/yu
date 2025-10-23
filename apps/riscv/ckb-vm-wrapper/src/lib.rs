use std::ffi::CString;
use std::os::raw::{c_char, c_int};
use std::ptr;
use std::slice;

// 引入 CKB-VM 相关类型
use ckb_vm::{DefaultCoreMachine, DefaultMachine, DefaultMachineBuilder, SparseMemory, ISA_IMC};

#[repr(C)]
pub struct ckb_vm_cell_data_t {
    pub data: *const u8,
    pub length: usize,
}

#[repr(C)]
pub struct ckb_vm_result_t {
    pub exit_code: c_int,
    pub error_message: *const c_char,
}

// Global state for VM instance
static mut VM_STATE: Option<VMState> = None;

struct VMState {
    program_loaded: bool,
    machine: Option<DefaultMachine<DefaultCoreMachine<u64, SparseMemory<u64>>>>,
    last_result: ckb_vm_result_t,
}

impl VMState {
    fn new() -> Self {
        Self {
            program_loaded: false,
            machine: None,
            last_result: ckb_vm_result_t {
                exit_code: 0,
                error_message: ptr::null(),
            },
        }
    }

    fn set_error(&mut self, exit_code: c_int, message: String) {
        // Free previous error message if exists
        if !self.last_result.error_message.is_null() {
            unsafe {
                let _ = CString::from_raw(self.last_result.error_message as *mut c_char);
            }
        }

        let c_string =
            CString::new(message).unwrap_or_else(|_| CString::new("Unknown error").unwrap());
        let ptr = c_string.into_raw();

        self.last_result = ckb_vm_result_t {
            exit_code,
            error_message: ptr,
        };
    }

    fn set_success(&mut self, exit_code: c_int) {
        // Free previous error message if exists
        if !self.last_result.error_message.is_null() {
            unsafe {
                let _ = CString::from_raw(self.last_result.error_message as *mut c_char);
            }
        }

        self.last_result = ckb_vm_result_t {
            exit_code,
            error_message: ptr::null(),
        };
    }
}

#[no_mangle]
pub extern "C" fn ckb_vm_init(
    _cell_data: *const ckb_vm_cell_data_t,
    _cell_data_length: usize,
) -> c_int {
    unsafe {
        VM_STATE = Some(VMState::new());
    }
    0
}

#[no_mangle]
pub extern "C" fn ckb_vm_load_program(program: *const u8, program_length: usize) -> c_int {
    unsafe {
        if VM_STATE.is_none() {
            return -1;
        }

        if program.is_null() || program_length == 0 {
            VM_STATE
                .as_mut()
                .unwrap()
                .set_error(-1, "Invalid program data".to_string());
            return -1;
        }

        let vm_state = VM_STATE.as_mut().unwrap();

        // 创建 CKB-VM 实例
        let core_machine = DefaultCoreMachine::<u64, SparseMemory<u64>>::new(ISA_IMC, 0, u64::MAX);
        let mut machine = DefaultMachineBuilder::new(core_machine).build();

        // 将程序数据转换为字节切片
        let program_slice = slice::from_raw_parts(program, program_length);
        let program_bytes = ckb_vm::Bytes::from(program_slice);

        // 加载程序到虚拟机
        if let Err(e) = machine.load_program(&program_bytes, &[]) {
            vm_state.set_error(-1, format!("Failed to load program: {:?}", e));
            return -1;
        }

        vm_state.machine = Some(machine);
        vm_state.program_loaded = true;
    }
    0
}

#[no_mangle]
pub extern "C" fn ckb_vm_run() -> c_int {
    unsafe {
        if VM_STATE.is_none() {
            return -1;
        }

        let vm_state = VM_STATE.as_mut().unwrap();
        if !vm_state.program_loaded {
            vm_state.set_error(-1, "No program loaded".to_string());
            return -1;
        }

        if vm_state.machine.is_none() {
            vm_state.set_error(-1, "No machine instance".to_string());
            return -1;
        }

        // 执行 RISC-V 程序
        let mut machine = vm_state.machine.take().unwrap();

        match machine.run() {
            Ok(exit_code) => {
                vm_state.set_success(exit_code as c_int);
                vm_state.machine = Some(machine);
                0
            }
            Err(e) => {
                vm_state.set_error(-1, format!("Execution failed: {:?}", e));
                vm_state.machine = Some(machine);
                -1
            }
        }
    }
}

#[no_mangle]
pub extern "C" fn ckb_vm_get_result() -> ckb_vm_result_t {
    unsafe {
        if let Some(ref vm_state) = VM_STATE {
            ckb_vm_result_t {
                exit_code: vm_state.last_result.exit_code,
                error_message: vm_state.last_result.error_message,
            }
        } else {
            ckb_vm_result_t {
                exit_code: -1,
                error_message: ptr::null(),
            }
        }
    }
}

#[no_mangle]
pub extern "C" fn ckb_vm_cleanup() {
    unsafe {
        if let Some(vm_state) = VM_STATE.take() {
            // Free error message if exists
            if !vm_state.last_result.error_message.is_null() {
                let _ = CString::from_raw(vm_state.last_result.error_message as *mut c_char);
            }
        }
    }
}
