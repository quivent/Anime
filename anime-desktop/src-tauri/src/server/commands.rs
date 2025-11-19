use tauri::State;
use std::sync::Mutex;
use std::collections::HashMap;
use super::{ServerConnection, ServerMonitor, ServerStatus};

pub struct ServerState {
    pub connections: Mutex<HashMap<String, ServerMonitor>>,
}

#[tauri::command]
pub async fn connect_to_server(
    instance_id: String,
    host: String,
    username: String,
    private_key_path: String,
    state: State<'_, ServerState>,
) -> Result<String, String> {
    eprintln!("[connect_to_server] Connecting to instance {} at {}", instance_id, host);

    // Create SSH connection
    let connection = ServerConnection::connect(&host, &username, &private_key_path)
        .map_err(|e| format!("Failed to connect: {}", e))?;

    // Create monitor
    let monitor = ServerMonitor::new(connection, instance_id.clone());

    // Store in state
    let mut connections = state.connections.lock().unwrap();
    connections.insert(instance_id.clone(), monitor);

    eprintln!("[connect_to_server] Successfully connected to {}", instance_id);
    Ok(format!("Connected to {}", instance_id))
}

#[tauri::command]
pub async fn disconnect_from_server(
    instance_id: String,
    state: State<'_, ServerState>,
) -> Result<String, String> {
    eprintln!("[disconnect_from_server] Disconnecting from {}", instance_id);

    let mut connections = state.connections.lock().unwrap();
    connections.remove(&instance_id);

    Ok(format!("Disconnected from {}", instance_id))
}

#[tauri::command]
pub async fn get_server_status(
    instance_id: String,
    state: State<'_, ServerState>,
) -> Result<ServerStatus, String> {
    eprintln!("[get_server_status] Getting status for {}", instance_id);

    let connections = state.connections.lock().unwrap();
    let monitor = connections.get(&instance_id)
        .ok_or_else(|| format!("Not connected to instance {}", instance_id))?;

    monitor.get_status().map_err(|e| format!("Failed to get status: {}", e))
}

#[tauri::command]
pub async fn is_server_connected(
    instance_id: String,
    state: State<'_, ServerState>,
) -> Result<bool, String> {
    let connections = state.connections.lock().unwrap();
    Ok(connections.contains_key(&instance_id))
}

#[tauri::command]
pub async fn list_connected_servers(
    state: State<'_, ServerState>,
) -> Result<Vec<String>, String> {
    let connections = state.connections.lock().unwrap();
    Ok(connections.keys().cloned().collect())
}

#[tauri::command]
pub async fn execute_server_command(
    instance_id: String,
    command: String,
    state: State<'_, ServerState>,
) -> Result<String, String> {
    eprintln!("[execute_server_command] Executing '{}' on {}", command, instance_id);

    let connections = state.connections.lock().unwrap();
    let monitor = connections.get(&instance_id)
        .ok_or_else(|| format!("Not connected to instance {}", instance_id))?;

    // Access the connection through the monitor (we'll need to add a method for this)
    Err("Command execution not yet implemented".to_string())
}
