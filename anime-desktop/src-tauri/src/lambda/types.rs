use serde::{Deserialize, Serialize};

// Instance Types
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InstanceType {
    pub name: String,
    pub description: String,
    pub price_cents_per_hour: u64,
    pub specs: InstanceSpecs,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InstanceSpecs {
    pub vcpus: u32,
    pub memory_gib: u32,
    pub storage_gib: u32,
    pub gpus: Vec<GpuSpec>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GpuSpec {
    pub gpu_type: String,
    pub count: u32,
    pub memory_gib: u32,
}

// Instance
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Instance {
    pub id: String,
    pub name: Option<String>,
    pub ip: Option<String>,
    pub private_ip: Option<String>,
    pub status: InstanceStatus,
    pub ssh_key_names: Vec<String>,
    pub file_system_names: Vec<String>,
    pub region: Region,
    pub instance_type: InstanceTypeName,
    pub hostname: Option<String>,
    pub jupyter_token: Option<String>,
    pub jupyter_url: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum InstanceStatus {
    Active,
    Booting,
    Unhealthy,
    Terminated,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Region {
    pub name: String,
    pub description: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InstanceTypeName {
    pub name: String,
}

// SSH Keys
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SSHKey {
    pub id: String,
    pub name: String,
    pub public_key: String,
    pub private_key: Option<String>,
}

// File Systems
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FileSystem {
    pub id: String,
    pub name: String,
    #[serde(rename = "created")]
    pub created_at: String,
    pub mount_point: Option<String>,
    pub is_in_use: bool,
    pub bytes_used: Option<u64>,
}

// API Request/Response Types
#[derive(Debug, Serialize)]
pub struct LaunchInstanceRequest {
    pub instance_type_name: String,
    pub region_name: String,
    pub ssh_key_names: Vec<String>,
    pub file_system_names: Option<Vec<String>>,
    pub quantity: Option<u32>,
    pub name: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct LaunchInstanceResponse {
    pub instance_ids: Vec<String>,
}

#[derive(Debug, Deserialize)]
pub struct ListInstancesResponse {
    pub data: Vec<Instance>,
}

#[derive(Debug, Deserialize)]
pub struct ListInstanceTypesResponse {
    pub data: Vec<InstanceType>,
}

#[derive(Debug, Deserialize)]
pub struct ListSSHKeysResponse {
    pub data: Vec<SSHKey>,
}

#[derive(Debug, Deserialize)]
pub struct ListFileSystemsResponse {
    pub data: Vec<FileSystem>,
}

#[derive(Debug, Serialize)]
pub struct AddSSHKeyRequest {
    pub name: String,
    pub public_key: String,
}

#[derive(Debug, Deserialize)]
pub struct AddSSHKeyResponse {
    pub data: SSHKey,
}

#[derive(Debug, Serialize)]
pub struct TerminateInstanceRequest {
    pub instance_ids: Vec<String>,
}

#[derive(Debug, Deserialize)]
pub struct TerminateInstanceResponse {
    pub terminated_instances: Vec<String>,
}

#[derive(Debug, Serialize)]
pub struct RestartInstanceRequest {
    pub instance_ids: Vec<String>,
}

#[derive(Debug, Deserialize)]
pub struct RestartInstanceResponse {
    pub restarted_instances: Vec<String>,
}

// Error types
#[derive(Debug, Deserialize)]
pub struct ApiError {
    pub error: ErrorDetail,
}

#[derive(Debug, Deserialize)]
pub struct ErrorDetail {
    pub code: String,
    pub message: String,
}
