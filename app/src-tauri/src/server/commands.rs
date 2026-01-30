use tauri::State;
use std::sync::Mutex;
use std::collections::HashMap;
use std::path::PathBuf;
use std::fs;
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

    // Create SSH connection
    let connection = ServerConnection::connect(&host, &username, &private_key_path)
        .map_err(|e| format!("Failed to connect: {}", e))?;

    // Create monitor
    let monitor = ServerMonitor::new(connection, instance_id.clone());

    // Store in state
    let mut connections = state.connections.lock().unwrap();
    connections.insert(instance_id.clone(), monitor);

    Ok(format!("Connected to {}", instance_id))
}

#[tauri::command]
pub async fn disconnect_from_server(
    instance_id: String,
    state: State<'_, ServerState>,
) -> Result<String, String> {

    let mut connections = state.connections.lock().unwrap();
    connections.remove(&instance_id);

    Ok(format!("Disconnected from {}", instance_id))
}

#[tauri::command]
pub async fn get_server_status(
    instance_id: String,
    state: State<'_, ServerState>,
) -> Result<ServerStatus, String> {

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

    let connections = state.connections.lock().unwrap();
    let _monitor = connections.get(&instance_id)
        .ok_or_else(|| format!("Not connected to instance {}", instance_id))?;

    // Access the connection through the monitor (we'll need to add a method for this)
    Err("Command execution not yet implemented".to_string())
}

#[derive(serde::Serialize)]
pub struct SshKeyInfo {
    pub path: String,
    pub name: String,
    pub key_type: String,
    pub is_valid: bool,
}

#[tauri::command]
pub async fn find_ssh_keys() -> Result<Vec<SshKeyInfo>, String> {

    let mut keys = Vec::new();

    // Get home directory
    let home_dir = dirs::home_dir()
        .ok_or_else(|| "Could not determine home directory".to_string())?;

    let ssh_dir = home_dir.join(".ssh");

    // Check if .ssh directory exists
    if !ssh_dir.exists() {

        return Ok(keys);
    }

    // Common SSH key filenames
    let key_names = vec![
        "id_rsa",
        "id_ed25519",
        "id_ecdsa",
        "id_dsa",
    ];

    for key_name in key_names {
        let key_path = ssh_dir.join(key_name);

        if key_path.exists() {

            // Check file permissions (should be 600 or 400)
            let metadata = fs::metadata(&key_path)
                .map_err(|e| format!("Failed to read metadata for {}: {}", key_path.display(), e))?;

            #[cfg(unix)]
            let is_valid = {
                use std::os::unix::fs::PermissionsExt;
                let mode = metadata.permissions().mode();
                let permissions = mode & 0o777;
                permissions == 0o600 || permissions == 0o400
            };

            #[cfg(not(unix))]
            let is_valid = true; // On non-Unix systems, skip permission check

            // Determine key type from filename
            let key_type = if key_name.contains("rsa") {
                "RSA".to_string()
            } else if key_name.contains("ed25519") {
                "Ed25519".to_string()
            } else if key_name.contains("ecdsa") {
                "ECDSA".to_string()
            } else if key_name.contains("dsa") {
                "DSA".to_string()
            } else {
                "Unknown".to_string()
            };

            keys.push(SshKeyInfo {
                path: key_path.to_string_lossy().to_string(),
                name: key_name.to_string(),
                key_type,
                is_valid,
            });
        }
    }

    Ok(keys)
}

#[tauri::command]
pub async fn validate_ssh_key(key_path: String) -> Result<bool, String> {

    let path = PathBuf::from(&key_path);

    // Check if file exists
    if !path.exists() {
        return Err(format!("Key file does not exist: {}", key_path));
    }

    // Check if it's a file (not a directory)
    let metadata = fs::metadata(&path)
        .map_err(|e| format!("Failed to read file metadata: {}", e))?;

    if !metadata.is_file() {
        return Err(format!("Path is not a file: {}", key_path));
    }

    // Check file permissions on Unix systems
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        let mode = metadata.permissions().mode();
        let permissions = mode & 0o777;

        if permissions != 0o600 && permissions != 0o400 {
            return Err(format!(
                "Invalid key permissions: {:o} (should be 600 or 400)",
                permissions
            ));
        }
    }

    // Read first few bytes to check if it looks like a private key
    let content = fs::read_to_string(&path)
        .map_err(|e| format!("Failed to read key file: {}", e))?;

    let is_valid = content.contains("-----BEGIN") &&
                   (content.contains("PRIVATE KEY") || content.contains("RSA PRIVATE KEY") ||
                    content.contains("OPENSSH PRIVATE KEY"));

    if !is_valid {
        return Err("File does not appear to be a valid SSH private key".to_string());
    }

    Ok(true)
}
