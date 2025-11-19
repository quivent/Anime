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
    eprintln!("[set_lambda_api_key] Starting - API key length: {}", api_key.len());

    let client = LambdaClient::new(api_key.clone()).map_err(|e| {
        eprintln!("[set_lambda_api_key] Failed to create client: {}", e);
        format!("Failed to create client: {}", e)
    })?;

    eprintln!("[set_lambda_api_key] Client created, testing API key...");

    // Test the API key by fetching instance types
    let test_result = client.list_instance_types().await;
    match &test_result {
        Ok(types) => eprintln!("[set_lambda_api_key] API verification successful - found {} instance types", types.len()),
        Err(e) => eprintln!("[set_lambda_api_key] API verification failed: {}", e),
    }

    test_result.map_err(|e| {
        format!("Failed to verify API key: {}. Please check that your API key is valid and has the correct permissions.", e)
    })?;

    eprintln!("[set_lambda_api_key] Storing client in state");
    let mut client_guard = state.client.lock().unwrap();
    *client_guard = Some(client);

    eprintln!("[set_lambda_api_key] Success!");
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
    let client = {
        let client_guard = state.client.lock().unwrap();
        client_guard.as_ref().ok_or("Lambda API key not set")?.clone()
    };

    client.list_instances().await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn lambda_list_instance_types(
    state: State<'_, LambdaState>,
) -> Result<Vec<InstanceType>, String> {
    let client = {
        let client_guard = state.client.lock().unwrap();
        client_guard.as_ref().ok_or("Lambda API key not set")?.clone()
    };

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
    let client = {
        let client_guard = state.client.lock().unwrap();
        client_guard.as_ref().ok_or("Lambda API key not set")?.clone()
    };

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
    let client = {
        let client_guard = state.client.lock().unwrap();
        client_guard.as_ref().ok_or("Lambda API key not set")?.clone()
    };

    client.terminate_instances(instance_ids).await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn lambda_restart_instances(
    instance_ids: Vec<String>,
    state: State<'_, LambdaState>,
) -> Result<Vec<String>, String> {
    let client = {
        let client_guard = state.client.lock().unwrap();
        client_guard.as_ref().ok_or("Lambda API key not set")?.clone()
    };

    client.restart_instances(instance_ids).await.map_err(|e| e.to_string())
}

// SSH Key operations
#[tauri::command]
pub async fn lambda_list_ssh_keys(
    state: State<'_, LambdaState>,
) -> Result<Vec<SSHKey>, String> {
    let client = {
        let client_guard = state.client.lock().unwrap();
        client_guard.as_ref().ok_or("Lambda API key not set")?.clone()
    };

    client.list_ssh_keys().await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn lambda_add_ssh_key(
    name: String,
    public_key: String,
    state: State<'_, LambdaState>,
) -> Result<SSHKey, String> {
    let client = {
        let client_guard = state.client.lock().unwrap();
        client_guard.as_ref().ok_or("Lambda API key not set")?.clone()
    };

    client.add_ssh_key(name, public_key).await.map_err(|e| e.to_string())
}

// File System operations
#[tauri::command]
pub async fn lambda_list_file_systems(
    state: State<'_, LambdaState>,
) -> Result<Vec<FileSystem>, String> {
    let client = {
        let client_guard = state.client.lock().unwrap();
        client_guard.as_ref().ok_or("Lambda API key not set")?.clone()
    };

    client.list_file_systems().await.map_err(|e| e.to_string())
}
