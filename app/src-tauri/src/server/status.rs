use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ServerStatus {
    pub instance_id: String,
    pub hostname: String,
    pub uptime_seconds: u64,
    pub cpu: CpuStatus,
    pub memory: MemoryStatus,
    pub gpu: Option<GpuStatus>,
    pub disk: DiskStatus,
    pub network: NetworkStatus,
    pub processes: Vec<ProcessInfo>,
    pub timestamp: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CpuStatus {
    pub usage_percent: f32,
    pub cores: u32,
    pub load_average: (f32, f32, f32), // 1min, 5min, 15min
    pub temperature: Option<f32>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryStatus {
    pub total_gb: f32,
    pub used_gb: f32,
    pub available_gb: f32,
    pub usage_percent: f32,
    pub swap_total_gb: f32,
    pub swap_used_gb: f32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GpuStatus {
    pub model: String,
    pub count: u32,
    pub gpus: Vec<GpuInfo>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GpuInfo {
    pub id: u32,
    pub name: String,
    pub memory_total_gb: f32,
    pub memory_used_gb: f32,
    pub memory_usage_percent: f32,
    pub utilization_percent: f32,
    pub temperature: f32,
    pub power_draw_watts: f32,
    pub power_limit_watts: f32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DiskStatus {
    pub total_gb: f32,
    pub used_gb: f32,
    pub available_gb: f32,
    pub usage_percent: f32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NetworkStatus {
    pub rx_bytes_per_sec: u64,
    pub tx_bytes_per_sec: u64,
    pub total_rx_gb: f32,
    pub total_tx_gb: f32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProcessInfo {
    pub pid: u32,
    pub name: String,
    pub cpu_percent: f32,
    pub memory_mb: f32,
    pub status: String,
    pub uptime_seconds: u64,
}
