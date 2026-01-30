use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use std::io::{Read, Write};
use std::thread;
use tauri::Emitter;
use serde::{Serialize, Deserialize};
use uuid::Uuid;
use portable_pty::{CommandBuilder, NativePtySystem, PtySize, PtySystem, MasterPty};

use lazy_static::lazy_static;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TerminalOutput {
    pub data: String,
}

pub struct TerminalSession {
    #[allow(dead_code)]
    session_id: String,
}

pub struct TerminalState {
    pub sessions: Mutex<HashMap<String, TerminalSession>>,
}

// Global storage for PTY writers
lazy_static! {
    static ref GLOBAL_WRITERS: Mutex<HashMap<String, Arc<Mutex<Box<dyn Write + Send>>>>> =
        Mutex::new(HashMap::new());
}

// Global storage for PTY masters (for resize operations)
lazy_static! {
    static ref GLOBAL_PTY_MASTERS: Mutex<HashMap<String, Arc<Mutex<Box<dyn MasterPty + Send>>>>> =
        Mutex::new(HashMap::new());
}

#[tauri::command]
pub async fn terminal_connect<R: tauri::Runtime>(
    app: tauri::AppHandle<R>,
    state: tauri::State<'_, TerminalState>,
    host: String,
    username: String,
) -> Result<String, String> {
    let session_id = Uuid::new_v4().to_string();
    println!("[Terminal {}] Connecting to {}@{}", session_id, username, host);

    // Create PTY system
    let pty_system = NativePtySystem::default();

    // Create PTY pair with size
    let pty_pair = pty_system
        .openpty(PtySize {
            rows: 24,
            cols: 80,
            pixel_width: 0,
            pixel_height: 0,
        })
        .map_err(|e| format!("Failed to create PTY: {}", e))?;

    // Build SSH command
    let mut cmd = CommandBuilder::new("ssh");
    cmd.arg("-o");
    cmd.arg("StrictHostKeyChecking=no");
    cmd.arg("-o");
    cmd.arg("UserKnownHostsFile=/dev/null");
    cmd.arg(format!("{}@{}", username, host));

    // Spawn child process
    let mut child = pty_pair
        .slave
        .spawn_command(cmd)
        .map_err(|e| format!("Failed to spawn SSH: {}", e))?;

    // Get the PTY master
    let mut master = pty_pair.master;

    // Clone reader before taking writer
    let mut reader = master.try_clone_reader()
        .map_err(|e| format!("Failed to clone PTY reader: {}", e))?;

    // Take the writer - after this, master cannot be used for writing
    let writer = master.take_writer()
        .map_err(|e| format!("Failed to take PTY writer: {}", e))?;

    let writer = Arc::new(Mutex::new(writer));
    let writer_for_input = writer.clone();

    // Note: We skip storing the master for resize operations since it's consumed after take_writer()
    // Resize functionality will not be available, but terminal I/O will work

    // Spawn thread to read from PTY and emit to frontend
    let sid_for_thread = session_id.clone();
    let app_for_thread = app.clone();
    thread::spawn(move || {
        let mut buf = [0u8; 8192];
        loop {
            match reader.read(&mut buf) {
                Ok(0) => break, // EOF
                Ok(n) => {
                    let data = String::from_utf8_lossy(&buf[..n]).to_string();
                    let _ = app_for_thread.emit("terminal_output", TerminalOutput {
                        data: data.clone(),
                    });
                }
                Err(e) => {
                    eprintln!("[Terminal {}] Read error: {}", sid_for_thread, e);
                    break;
                }
            }
        }

        // Wait for child to exit
        let exit_status = child.wait();
        eprintln!("[Terminal {}] Session ended with status: {:?}", sid_for_thread, exit_status);
    });

    // Store session with writer reference
    // For now, we'll handle input through a separate mechanism
    let session = TerminalSession {
        session_id: session_id.clone(),
    };

    // Lock ordering: Always acquire locks in this order to prevent deadlocks:
    // 1. state.sessions 2. GLOBAL_WRITERS 3. GLOBAL_PTY_MASTERS
    let mut sessions = state.sessions.lock()
        .map_err(|e| format!("Failed to lock sessions: {}", e))?;
    sessions.insert(session_id.clone(), session);

    // Also store writer globally so terminal_input can access it
    GLOBAL_WRITERS.lock()
        .map_err(|e| format!("Failed to lock writers: {}", e))?
        .insert(session_id.clone(), writer_for_input);

    // Note: Not storing master for resize since it's consumed after take_writer()

    Ok(session_id)
}

#[tauri::command]
pub async fn terminal_connect_local<R: tauri::Runtime>(
    app: tauri::AppHandle<R>,
    state: tauri::State<'_, TerminalState>,
) -> Result<String, String> {
    let session_id = Uuid::new_v4().to_string();
    println!("[Terminal {}] Starting local shell", session_id);

    // Create PTY system
    let pty_system = NativePtySystem::default();

    // Create PTY pair with size
    let pty_pair = pty_system
        .openpty(PtySize {
            rows: 24,
            cols: 80,
            pixel_width: 0,
            pixel_height: 0,
        })
        .map_err(|e| format!("Failed to create PTY: {}", e))?;

    // Build shell command - use user's default shell
    let shell = std::env::var("SHELL").unwrap_or_else(|_| "/bin/bash".to_string());
    let mut cmd = CommandBuilder::new(shell);
    cmd.arg("-l"); // Login shell to load user's environment

    // Spawn child process
    let mut child = pty_pair
        .slave
        .spawn_command(cmd)
        .map_err(|e| format!("Failed to spawn shell: {}", e))?;

    // Get the PTY master
    let mut master = pty_pair.master;

    // Clone reader before taking writer
    let mut reader = master.try_clone_reader()
        .map_err(|e| format!("Failed to clone PTY reader: {}", e))?;

    // Take the writer - after this, master cannot be used for writing
    let writer = master.take_writer()
        .map_err(|e| format!("Failed to take PTY writer: {}", e))?;

    let writer = Arc::new(Mutex::new(writer));
    let writer_for_input = writer.clone();

    // Note: We skip storing the master for resize operations since it's consumed after take_writer()
    // Resize functionality will not be available, but terminal I/O will work

    // Spawn thread to read from PTY and emit to frontend
    let sid_for_thread = session_id.clone();
    let app_for_thread = app.clone();
    thread::spawn(move || {
        let mut buf = [0u8; 8192];
        loop {
            match reader.read(&mut buf) {
                Ok(0) => break, // EOF
                Ok(n) => {
                    let data = String::from_utf8_lossy(&buf[..n]).to_string();
                    let _ = app_for_thread.emit("terminal_output", TerminalOutput {
                        data: data.clone(),
                    });
                }
                Err(e) => {
                    eprintln!("[Terminal {}] Read error: {}", sid_for_thread, e);
                    break;
                }
            }
        }

        // Wait for child to exit
        let exit_status = child.wait();
        eprintln!("[Terminal {}] Session ended with status: {:?}", sid_for_thread, exit_status);
    });

    // Store session
    let session = TerminalSession {
        session_id: session_id.clone(),
    };

    // Lock ordering: Always acquire locks in this order to prevent deadlocks:
    // 1. state.sessions 2. GLOBAL_WRITERS 3. GLOBAL_PTY_MASTERS
    let mut sessions = state.sessions.lock()
        .map_err(|e| format!("Failed to lock sessions: {}", e))?;
    sessions.insert(session_id.clone(), session);

    // Also store writer globally so terminal_input can access it
    GLOBAL_WRITERS.lock()
        .map_err(|e| format!("Failed to lock writers: {}", e))?
        .insert(session_id.clone(), writer_for_input);

    // Note: Not storing master for resize since it's consumed after take_writer()

    Ok(session_id)
}

#[tauri::command]
pub async fn terminal_input(
    session_id: String,
    data: String,
) -> Result<(), String> {
    let writers = GLOBAL_WRITERS.lock()
        .map_err(|e| format!("Failed to lock writers: {}", e))?;

    if let Some(writer_arc) = writers.get(&session_id) {
        let mut writer = writer_arc.lock()
            .map_err(|e| format!("Failed to lock writer for session: {}", e))?;
        writer.write_all(data.as_bytes())
            .map_err(|e| format!("Failed to write to terminal: {}", e))?;
        writer.flush()
            .map_err(|e| format!("Failed to flush terminal: {}", e))?;
        Ok(())
    } else {
        Err("Session not found".to_string())
    }
}

#[tauri::command]
pub async fn terminal_disconnect(
    state: tauri::State<'_, TerminalState>,
    session_id: String,
) -> Result<(), String> {
    // Lock ordering: Always acquire locks in this order to prevent deadlocks:
    // 1. state.sessions 2. GLOBAL_WRITERS 3. GLOBAL_PTY_MASTERS
    let mut sessions = state.sessions.lock()
        .map_err(|e| format!("Failed to lock sessions: {}", e))?;
    sessions.remove(&session_id);

    let mut writers = GLOBAL_WRITERS.lock()
        .map_err(|e| format!("Failed to lock writers: {}", e))?;
    writers.remove(&session_id);

    let mut pty_masters = GLOBAL_PTY_MASTERS.lock()
        .map_err(|e| format!("Failed to lock PTY masters: {}", e))?;
    pty_masters.remove(&session_id);

    Ok(())
}

#[tauri::command]
pub async fn terminal_resize(
    session_id: String,
    rows: u16,
    cols: u16,
) -> Result<(), String> {
    let pty_masters = GLOBAL_PTY_MASTERS.lock()
        .map_err(|e| format!("Failed to lock PTY masters: {}", e))?;

    if let Some(master_arc) = pty_masters.get(&session_id) {
        let master = master_arc.lock()
            .map_err(|e| format!("Failed to lock PTY master for session: {}", e))?;
        let size = PtySize {
            rows,
            cols,
            pixel_width: 0,
            pixel_height: 0,
        };
        master.resize(size)
            .map_err(|e| format!("Failed to resize terminal: {}", e))?;
        Ok(())
    } else {
        Err("Session not found".to_string())
    }
}

