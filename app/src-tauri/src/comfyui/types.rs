use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ComfyUIWorkflow {
    pub id: String,
    pub name: String,
    pub description: String,
    pub category: String,
    pub icon: String,
    pub thumbnail: Option<String>,
    pub workflow_json: String,
    pub parameters: Vec<WorkflowParameter>,
    pub outputs: Vec<WorkflowOutput>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WorkflowParameter {
    pub id: String,
    pub name: String,
    #[serde(rename = "type")]
    pub param_type: String,
    pub description: String,
    pub required: bool,
    pub default_value: Option<serde_json::Value>,
    pub options: Option<Vec<String>>,
    pub min: Option<f64>,
    pub max: Option<f64>,
    pub node_id: Option<String>,
    pub field_name: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WorkflowOutput {
    #[serde(rename = "type")]
    pub output_type: String,
    pub name: String,
    pub format: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WorkflowExecution {
    pub id: String,
    pub workflow_id: String,
    pub status: String,
    pub progress: f64,
    pub prompt_id: Option<String>,
    pub queue_position: Option<u32>,
    pub current_node: Option<String>,
    pub error: Option<String>,
    pub started_at: Option<String>,
    pub completed_at: Option<String>,
    pub outputs: Option<Vec<WorkflowExecutionOutput>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WorkflowExecutionOutput {
    pub filename: String,
    pub subfolder: String,
    #[serde(rename = "type")]
    pub output_type: String,
    pub url: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ComfyUIStatus {
    pub connected: bool,
    pub queue_remaining: u32,
    pub queue_running: u32,
    pub system_stats: Option<SystemStats>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SystemStats {
    pub devices: Vec<DeviceStats>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DeviceStats {
    pub name: String,
    #[serde(rename = "type")]
    pub device_type: String,
    pub vram_total: u64,
    pub vram_free: u64,
    pub torch_vram_total: u64,
    pub torch_vram_free: u64,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct QueuePromptRequest {
    pub client_id: Option<String>,
    pub prompt: serde_json::Value,
    pub extra_data: Option<serde_json::Value>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct QueuePromptResponse {
    pub prompt_id: String,
    pub number: u32,
    pub node_errors: Option<HashMap<String, serde_json::Value>>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct HistoryResponse {
    #[serde(flatten)]
    pub history: HashMap<String, HistoryItem>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct HistoryItem {
    pub prompt: serde_json::Value,
    pub outputs: HashMap<String, serde_json::Value>,
    pub status: HistoryStatus,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct HistoryStatus {
    pub status_str: String,
    pub completed: bool,
    pub messages: Vec<serde_json::Value>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct QueueStatus {
    pub queue_running: Vec<QueueItem>,
    pub queue_pending: Vec<QueueItem>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct QueueItem {
    pub prompt_id: String,
    pub number: u32,
    pub prompt: serde_json::Value,
}
