use tauri::State;
use std::sync::Mutex;
use super::{LambdaClient, Instance, InstanceType, SSHKey, FileSystem, LaunchInstanceRequest};

pub struct LambdaState {
    pub client: Mutex<Option<LambdaClient>>,
}

#[tauri::command]
pub async fn set_lambda_api_key(
    api_key: String,
    state: State<'_, LambdaState>,
) -> Result<String, String> {
    let client = LambdaClient::new(api_key).map_err(|e| e.to_string())?;

    // Test the API key by fetching instance types
    client.list_instance_types().await.map_err(|e| e.to_string())?;

    let mut client_guard = state.client.lock().unwrap();
    *client_guard = Some(client);

    Ok("API key set successfully".to_string())
}

#[tauri::command]
pub async fn check_lambda_connection(
    state: State<'_, LambdaState>,
) -> Result<bool, String> {
    let client_guard = state.client.lock().unwrap();
    Ok(client_guard.is_some())
}

// Instance operations
#[tauri::command]
pub async fn lambda_list_instances(
    state: State<'_, LambdaState>,
) -> Result<Vec<Instance>, String> {
    let client_guard = state.client.lock().unwrap();
    let client = client_guard.as_ref().ok_or("Lambda API key not set")?;

    client.list_instances().await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn lambda_list_instance_types(
    state: State<'_, LambdaState>,
) -> Result<Vec<InstanceType>, String> {
    let client_guard = state.client.lock().unwrap();
    let client = client_guard.as_ref().ok_or("Lambda API key not set")?;

    client.list_instance_types().await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn lambda_launch_instance(
    instance_type: String,
    region: String,
    ssh_keys: Vec<String>,
    name: Option<String>,
    quantity: Option<u32>,
    state: State<'_, LambdaState>,
) -> Result<Vec<String>, String> {
    let client_guard = state.client.lock().unwrap();
    let client = client_guard.as_ref().ok_or("Lambda API key not set")?;

    let request = LaunchInstanceRequest {
        instance_type_name: instance_type,
        region_name: region,
        ssh_key_names: ssh_keys,
        file_system_names: None,
        quantity,
        name,
    };

    client.launch_instance(request).await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn lambda_terminate_instances(
    instance_ids: Vec<String>,
    state: State<'_, LambdaState>,
) -> Result<Vec<String>, String> {
    let client_guard = state.client.lock().unwrap();
    let client = client_guard.as_ref().ok_or("Lambda API key not set")?;

    client.terminate_instances(instance_ids).await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn lambda_restart_instances(
    instance_ids: Vec<String>,
    state: State<'_, LambdaState>,
) -> Result<Vec<String>, String> {
    let client_guard = state.client.lock().unwrap();
    let client = client_guard.as_ref().ok_or("Lambda API key not set")?;

    client.restart_instances(instance_ids).await.map_err(|e| e.to_string())
}

// SSH Key operations
#[tauri::command]
pub async fn lambda_list_ssh_keys(
    state: State<'_, LambdaState>,
) -> Result<Vec<SSHKey>, String> {
    let client_guard = state.client.lock().unwrap();
    let client = client_guard.as_ref().ok_or("Lambda API key not set")?;

    client.list_ssh_keys().await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn lambda_add_ssh_key(
    name: String,
    public_key: String,
    state: State<'_, LambdaState>,
) -> Result<SSHKey, String> {
    let client_guard = state.client.lock().unwrap();
    let client = client_guard.as_ref().ok_or("Lambda API key not set")?;

    client.add_ssh_key(name, public_key).await.map_err(|e| e.to_string())
}

// File System operations
#[tauri::command]
pub async fn lambda_list_file_systems(
    state: State<'_, LambdaState>,
) -> Result<Vec<FileSystem>, String> {
    let client_guard = state.client.lock().unwrap();
    let client = client_guard.as_ref().ok_or("Lambda API key not set")?;

    client.list_file_systems().await.map_err(|e| e.to_string())
}
