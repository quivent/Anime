use serde::{Deserialize, Serialize};
use std::collections::HashMap;

// Region (must be defined before InstanceType since it's used there)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Region {
    pub name: String,
    pub description: String,
}

// Instance Types
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InstanceType {
    pub name: String,
    pub description: String,
    pub gpu_description: String,
    pub price_cents_per_hour: u64,
    pub specs: InstanceSpecs,
    pub regions_with_capacity_available: Vec<Region>,
}

// This is what comes from the API (nested inside InstanceTypeEntry)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InstanceTypeNested {
    pub name: String,
    pub description: String,
    pub gpu_description: String,
    pub price_cents_per_hour: u64,
    pub specs: InstanceSpecs,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InstanceSpecs {
    pub vcpus: u32,
    pub memory_gib: u32,
    pub storage_gib: u32,
    pub gpus: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InstanceTypeEntry {
    pub instance_type: InstanceTypeNested,
    pub regions_with_capacity_available: Vec<Region>,
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
    #[serde(default)]
    pub is_reserved: Option<bool>,
    #[serde(default)]
    pub actions: Option<serde_json::Value>,
    #[serde(default)]
    pub firewall_rulesets: Option<serde_json::Value>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum InstanceStatus {
    Active,
    Booting,
    Unhealthy,
    Terminated,
    Terminating,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InstanceTypeName {
    pub name: String,
    #[serde(default)]
    pub description: Option<String>,
    #[serde(default)]
    pub gpu_description: Option<String>,
    #[serde(default)]
    pub price_cents_per_hour: Option<u64>,
    #[serde(default)]
    pub specs: Option<InstanceSpecs>,
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
pub struct LaunchInstanceResponseData {
    pub instance_ids: Vec<String>,
}

#[derive(Debug, Deserialize)]
pub struct LaunchInstanceResponse {
    pub data: LaunchInstanceResponseData,
}

#[derive(Debug, Deserialize)]
pub struct ListInstancesResponse {
    pub data: Vec<Instance>,
}

#[derive(Debug, Deserialize)]
pub struct ListInstanceTypesResponse {
    pub data: HashMap<String, InstanceTypeEntry>,
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
pub struct TerminateInstanceResponseData {
    pub terminated_instances: Vec<Instance>,
}

#[derive(Debug, Deserialize)]
pub struct TerminateInstanceResponse {
    pub data: TerminateInstanceResponseData,
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
