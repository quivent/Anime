export interface CpuStatus {
  usage_percent: number
  cores: number
  load_average_1m: number
  load_average_5m: number
  load_average_15m: number
  temperature: number | null
}

export interface MemoryStatus {
  total_gb: number
  used_gb: number
  available_gb: number
  usage_percent: number
  swap_total_gb: number
  swap_used_gb: number
}

export interface GpuInfo {
  id: number
  name: string
  memory_total_gb: number
  memory_used_gb: number
  memory_usage_percent: number
  utilization_percent: number
  temperature: number
  power_draw_watts: number
  power_limit_watts: number
}

export interface GpuStatus {
  gpus: GpuInfo[]
  driver_version: string
}

export interface DiskStatus {
  total_gb: number
  used_gb: number
  available_gb: number
  usage_percent: number
}

export interface NetworkStatus {
  bytes_sent: number
  bytes_received: number
  packets_sent: number
  packets_received: number
}

export interface ProcessInfo {
  pid: number
  name: string
  cpu_percent: number
  memory_mb: number
  status: string
}

export interface ServerStatus {
  instance_id: string
  hostname: string
  uptime_seconds: number
  cpu: CpuStatus
  memory: MemoryStatus
  gpu: GpuStatus | null
  disk: DiskStatus
  network: NetworkStatus
  processes: ProcessInfo[]
  timestamp: string
}
