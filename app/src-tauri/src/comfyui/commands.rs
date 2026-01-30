use super::client::ComfyUIClient;
use super::types::*;
use super::workflows::get_builtin_workflows;
use serde_json::{json, Value};
use std::collections::HashMap;
use tokio::sync::Mutex;
use tauri::State;

pub struct ComfyUIState {
    pub client: Mutex<Option<ComfyUIClient>>,
    pub host: Mutex<String>,
    pub port: Mutex<u16>,
    pub executions: Mutex<HashMap<String, WorkflowExecution>>,
}

impl Default for ComfyUIState {
    fn default() -> Self {
        Self {
            client: Mutex::new(None),
            host: Mutex::new("localhost".to_string()),
            port: Mutex::new(8188),
            executions: Mutex::new(HashMap::new()),
        }
    }
}

/// Set ComfyUI connection settings
#[tauri::command]
pub async fn comfyui_set_connection(
    state: State<'_, ComfyUIState>,
    host: String,
    port: u16,
) -> Result<(), String> {
    let mut host_lock = state.host.lock().await;
    let mut port_lock = state.port.lock().await;
    let mut client_lock = state.client.lock().await;

    *host_lock = host.clone();
    *port_lock = port;
    *client_lock = Some(ComfyUIClient::new(&host, port));

    Ok(())
}

/// Check ComfyUI connection status
#[tauri::command]
pub async fn comfyui_check_connection(state: State<'_, ComfyUIState>) -> Result<bool, String> {
    let client_lock = state.client.lock().await;

    if let Some(client) = client_lock.as_ref() {
        client
            .check_connection()
            .await
            .map_err(|e| e.to_string())
    } else {
        // Try default connection
        let host_lock = state.host.lock().await;
        let port_lock = state.port.lock().await;
        drop(client_lock);

        let client = ComfyUIClient::new(&host_lock, *port_lock);
        let connected = client
            .check_connection()
            .await
            .map_err(|e| e.to_string())?;

        if connected {
            let mut client_lock = state.client.lock().await;
            *client_lock = Some(client);
        }

        Ok(connected)
    }
}

/// Get ComfyUI status
#[tauri::command]
pub async fn comfyui_get_status(state: State<'_, ComfyUIState>) -> Result<ComfyUIStatus, String> {
    let client = get_client(&state).await?;

    let connected = client
        .check_connection()
        .await
        .map_err(|e| e.to_string())?;

    if !connected {
        return Ok(ComfyUIStatus {
            connected: false,
            queue_remaining: 0,
            queue_running: 0,
            system_stats: None,
        });
    }

    let queue = client.get_queue().await.map_err(|e| e.to_string())?;
    let system_stats = client.get_system_stats().await.ok();

    Ok(ComfyUIStatus {
        connected: true,
        queue_remaining: queue.queue_pending.len() as u32,
        queue_running: queue.queue_running.len() as u32,
        system_stats,
    })
}

/// Get list of built-in workflows
#[tauri::command]
pub fn comfyui_list_workflows() -> Result<Vec<ComfyUIWorkflow>, String> {
    Ok(get_builtin_workflows())
}

/// Get a specific workflow by ID
#[tauri::command]
pub fn comfyui_get_workflow(workflow_id: String) -> Result<ComfyUIWorkflow, String> {
    get_builtin_workflows()
        .into_iter()
        .find(|w| w.id == workflow_id)
        .ok_or_else(|| format!("Workflow '{}' not found", workflow_id))
}

/// Execute a workflow with parameters
#[tauri::command]
pub async fn comfyui_execute_workflow(
    state: State<'_, ComfyUIState>,
    workflow_id: String,
    parameters: HashMap<String, Value>,
) -> Result<WorkflowExecution, String> {
    let client = get_client(&state).await?;

    // Get the workflow
    let workflow = comfyui_get_workflow(workflow_id.clone())?;

    // Parse the workflow JSON
    let mut workflow_json: Value =
        serde_json::from_str(&workflow.workflow_json).map_err(|e| e.to_string())?;

    // Apply parameters to the workflow
    for param in &workflow.parameters {
        if let Some(value) = parameters.get(&param.id) {
            if let (Some(node_id), Some(field_name)) = (&param.node_id, &param.field_name) {
                if let Some(node) = workflow_json.get_mut(node_id) {
                    if let Some(inputs) = node.get_mut("inputs") {
                        inputs[field_name] = value.clone();
                    }
                }
            }
        } else if param.required && param.default_value.is_none() {
            return Err(format!("Required parameter '{}' not provided", param.id));
        } else if let Some(default) = &param.default_value {
            if let (Some(node_id), Some(field_name)) = (&param.node_id, &param.field_name) {
                if let Some(node) = workflow_json.get_mut(node_id) {
                    if let Some(inputs) = node.get_mut("inputs") {
                        inputs[field_name] = default.clone();
                    }
                }
            }
        }
    }

    // Generate a random seed
    let seed = rand::random::<u32>();
    if let Some(sampler_node) = workflow_json.get_mut("3") {
        if let Some(inputs) = sampler_node.get_mut("inputs") {
            inputs["seed"] = serde_json::json!(seed);
        }
    }

    // Queue the prompt
    let response = client
        .queue_prompt(workflow_json, None)
        .await
        .map_err(|e| e.to_string())?;

    // Create execution record
    let execution_id = uuid::Uuid::new_v4().to_string();
    let execution = WorkflowExecution {
        id: execution_id.clone(),
        workflow_id: workflow_id.clone(),
        status: "queued".to_string(),
        progress: 0.0,
        prompt_id: Some(response.prompt_id.clone()),
        queue_position: Some(response.number),
        current_node: None,
        error: None,
        started_at: Some(chrono::Utc::now().to_rfc3339()),
        completed_at: None,
        outputs: None,
    };

    // Store execution
    let mut executions = state.executions.lock().await;
    executions.insert(execution_id.clone(), execution.clone());

    Ok(execution)
}

/// Get execution status
#[tauri::command]
pub async fn comfyui_get_execution(
    state: State<'_, ComfyUIState>,
    execution_id: String,
) -> Result<WorkflowExecution, String> {
    let client = get_client(&state).await?;
    let mut executions = state.executions.lock().await;

    let execution = executions
        .get_mut(&execution_id)
        .ok_or_else(|| format!("Execution '{}' not found", execution_id))?;

    // If execution is complete, return cached version
    if execution.status == "completed" || execution.status == "failed" {
        return Ok(execution.clone());
    }

    // Check prompt status
    if let Some(prompt_id) = &execution.prompt_id {
        let history = client
            .get_history(Some(prompt_id))
            .await
            .map_err(|e| e.to_string())?;

        if let Some(history_item) = history.get(prompt_id) {
            // Update status
            execution.status = if history_item.status.completed {
                "completed"
            } else {
                "running"
            }
            .to_string();

            execution.progress = if history_item.status.completed {
                100.0
            } else {
                50.0
            };

            if history_item.status.completed {
                execution.completed_at = Some(chrono::Utc::now().to_rfc3339());

                // Extract outputs
                let mut outputs = Vec::new();
                for (_node_id, output) in &history_item.outputs {
                    if let Some(images) = output.get("images").and_then(|v| v.as_array()) {
                        for img in images {
                            if let (Some(filename), Some(subfolder), Some(output_type)) = (
                                img.get("filename").and_then(|v| v.as_str()),
                                img.get("subfolder").and_then(|v| v.as_str()),
                                img.get("type").and_then(|v| v.as_str()),
                            ) {
                                let host_lock = state.host.lock().await;
                                let port_lock = state.port.lock().await;

                                outputs.push(WorkflowExecutionOutput {
                                    filename: filename.to_string(),
                                    subfolder: subfolder.to_string(),
                                    output_type: output_type.to_string(),
                                    url: format!(
                                        "http://{}:{}/view?filename={}&subfolder={}&type={}",
                                        *host_lock, *port_lock, filename, subfolder, output_type
                                    ),
                                });
                            }
                        }
                    }
                }
                execution.outputs = Some(outputs);
            }
        }
    }

    Ok(execution.clone())
}

/// List all executions
#[tauri::command]
pub async fn comfyui_list_executions(
    state: State<'_, ComfyUIState>,
) -> Result<Vec<WorkflowExecution>, String> {
    let executions = state.executions.lock().await;
    Ok(executions.values().cloned().collect())
}

/// Cancel an execution
#[tauri::command]
pub async fn comfyui_cancel_execution(
    state: State<'_, ComfyUIState>,
    execution_id: String,
) -> Result<(), String> {
    let client = get_client(&state).await?;
    let mut executions = state.executions.lock().await;

    let execution = executions
        .get_mut(&execution_id)
        .ok_or_else(|| format!("Execution '{}' not found", execution_id))?;

    if let Some(prompt_id) = &execution.prompt_id {
        client
            .cancel_prompt(prompt_id)
            .await
            .map_err(|e| e.to_string())?;
    }

    execution.status = "failed".to_string();
    execution.error = Some("Cancelled by user".to_string());

    Ok(())
}

/// Interrupt current execution
#[tauri::command]
pub async fn comfyui_interrupt(state: State<'_, ComfyUIState>) -> Result<(), String> {
    let client = get_client(&state).await?;
    client.interrupt().await.map_err(|e| e.to_string())
}

/// Clear the queue
#[tauri::command]
pub async fn comfyui_clear_queue(state: State<'_, ComfyUIState>) -> Result<(), String> {
    let client = get_client(&state).await?;
    client.clear_queue().await.map_err(|e| e.to_string())
}

/// Upload an image to ComfyUI
#[tauri::command]
pub async fn comfyui_upload_image(
    state: State<'_, ComfyUIState>,
    image_data: Vec<u8>,
    filename: String,
) -> Result<String, String> {
    let client = get_client(&state).await?;
    client
        .upload_image(image_data, &filename, true)
        .await
        .map_err(|e| e.to_string())
}

// Helper function to get client
async fn get_client(state: &State<'_, ComfyUIState>) -> Result<ComfyUIClient, String> {
    let host = state.host.lock().await;
    let port = state.port.lock().await;
    Ok(ComfyUIClient::new(&host, *port))
}
