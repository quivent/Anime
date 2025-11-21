use anyhow::Result;
use super::ssh::ServerConnection;
use super::status::*;
use chrono::Utc;

pub struct ServerMonitor {
    connection: ServerConnection,
    instance_id: String,
}

impl ServerMonitor {
    pub fn new(connection: ServerConnection, instance_id: String) -> Self {
        Self {
            connection,
            instance_id,
        }
    }

    /// Get comprehensive server status
    pub fn get_status(&self) -> Result<ServerStatus> {
        Ok(ServerStatus {
            instance_id: self.instance_id.clone(),
            hostname: self.connection.hostname().to_string(),
            uptime_seconds: self.get_uptime()?,
            cpu: self.get_cpu_status()?,
            memory: self.get_memory_status()?,
            gpu: self.get_gpu_status().ok(),
            disk: self.get_disk_status()?,
            network: self.get_network_status()?,
            processes: self.get_process_info()?,
            timestamp: Utc::now().to_rfc3339(),
        })
    }

    fn get_uptime(&self) -> Result<u64> {
        let output = self.connection.execute_command("cat /proc/uptime | awk '{print $1}'")?;
        let uptime: f64 = output.parse()?;
        Ok(uptime as u64)
    }

    fn get_cpu_status(&self) -> Result<CpuStatus> {
        // Get CPU usage
        let cpu_usage = self.connection.execute_command(
            "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1}'"
        )?;

        // Get number of cores
        let cores = self.connection.execute_command("nproc")?;

        // Get load average
        let load_avg = self.connection.execute_command("cat /proc/loadavg | awk '{print $1,$2,$3}'")?;
        let loads: Vec<f32> = load_avg.split_whitespace()
            .filter_map(|s| s.parse().ok())
            .collect();

        // Try to get temperature (might not be available)
        let temp = self.connection.execute_command(
            "sensors 2>/dev/null | grep 'Core 0' | awk '{print $3}' | tr -d '+°C'"
        ).ok().and_then(|t| t.parse().ok());

        Ok(CpuStatus {
            usage_percent: cpu_usage.parse().unwrap_or(0.0),
            cores: cores.parse().unwrap_or(1),
            load_average: (
                loads.get(0).copied().unwrap_or(0.0),
                loads.get(1).copied().unwrap_or(0.0),
                loads.get(2).copied().unwrap_or(0.0),
            ),
            temperature: temp,
        })
    }

    fn get_memory_status(&self) -> Result<MemoryStatus> {
        let output = self.connection.execute_command(
            "free -g | awk 'NR==2{print $2,$3,$7} NR==3{print $2,$3}'"
        )?;

        let values: Vec<f32> = output.split_whitespace()
            .filter_map(|s| s.parse().ok())
            .collect();

        let total_gb = values.get(0).copied().unwrap_or(0.0);
        let used_gb = values.get(1).copied().unwrap_or(0.0);

        Ok(MemoryStatus {
            total_gb,
            used_gb,
            available_gb: values.get(2).copied().unwrap_or(0.0),
            usage_percent: if total_gb > 0.0 { (used_gb / total_gb) * 100.0 } else { 0.0 },
            swap_total_gb: values.get(3).copied().unwrap_or(0.0),
            swap_used_gb: values.get(4).copied().unwrap_or(0.0),
        })
    }

    fn get_gpu_status(&self) -> Result<GpuStatus> {
        // Check if nvidia-smi is available
        let gpu_count = self.connection.execute_command(
            "nvidia-smi --query-gpu=count --format=csv,noheader | head -1"
        )?;

        let count: u32 = gpu_count.parse()?;
        let mut gpus = Vec::new();

        // Get detailed info for each GPU
        for gpu_id in 0..count {
            let query = format!(
                "nvidia-smi -i {} --query-gpu=gpu_name,memory.total,memory.used,utilization.gpu,temperature.gpu,power.draw,power.limit --format=csv,noheader,nounits",
                gpu_id
            );

            let output = self.connection.execute_command(&query)?;
            let parts: Vec<&str> = output.split(',').map(|s| s.trim()).collect();

            if parts.len() >= 7 {
                let memory_total = parts[1].parse::<f32>().unwrap_or(0.0) / 1024.0; // MB to GB
                let memory_used = parts[2].parse::<f32>().unwrap_or(0.0) / 1024.0;

                gpus.push(GpuInfo {
                    id: gpu_id,
                    name: parts[0].to_string(),
                    memory_total_gb: memory_total,
                    memory_used_gb: memory_used,
                    memory_usage_percent: if memory_total > 0.0 {
                        (memory_used / memory_total) * 100.0
                    } else {
                        0.0
                    },
                    utilization_percent: parts[3].parse().unwrap_or(0.0),
                    temperature: parts[4].parse().unwrap_or(0.0),
                    power_draw_watts: parts[5].parse().unwrap_or(0.0),
                    power_limit_watts: parts[6].parse().unwrap_or(0.0),
                });
            }
        }

        let model = gpus.first().map(|g| g.name.clone()).unwrap_or_default();

        Ok(GpuStatus {
            model,
            count,
            gpus,
        })
    }

    fn get_disk_status(&self) -> Result<DiskStatus> {
        let output = self.connection.execute_command(
            "df -BG / | awk 'NR==2{print $2,$3,$4}' | tr -d 'G'"
        )?;

        let values: Vec<f32> = output.split_whitespace()
            .filter_map(|s| s.parse().ok())
            .collect();

        let total_gb = values.get(0).copied().unwrap_or(0.0);
        let used_gb = values.get(1).copied().unwrap_or(0.0);

        Ok(DiskStatus {
            total_gb,
            used_gb,
            available_gb: values.get(2).copied().unwrap_or(0.0),
            usage_percent: if total_gb > 0.0 { (used_gb / total_gb) * 100.0 } else { 0.0 },
        })
    }

    fn get_network_status(&self) -> Result<NetworkStatus> {
        // Get network stats from /proc/net/dev
        let output = self.connection.execute_command(
            "cat /proc/net/dev | grep -E 'eth0|ens' | head -1 | awk '{print $2,$10}'"
        )?;

        let values: Vec<u64> = output.split_whitespace()
            .filter_map(|s| s.parse().ok())
            .collect();

        let total_rx = values.get(0).copied().unwrap_or(0) as f32 / 1_073_741_824.0; // bytes to GB
        let total_tx = values.get(1).copied().unwrap_or(0) as f32 / 1_073_741_824.0;

        Ok(NetworkStatus {
            rx_bytes_per_sec: 0, // Would need historical data
            tx_bytes_per_sec: 0,
            total_rx_gb: total_rx,
            total_tx_gb: total_tx,
        })
    }

    fn get_process_info(&self) -> Result<Vec<ProcessInfo>> {
        // Get info for important processes (ComfyUI, Python, etc.)
        let output = self.connection.execute_command(
            "ps aux | grep -E 'comfyui|python.*main\\.py|ollama' | grep -v grep | awk '{print $2,$11,$3,$4,$10}'"
        )?;

        let mut processes = Vec::new();

        for line in output.lines() {
            let parts: Vec<&str> = line.split_whitespace().collect();
            if parts.len() >= 5 {
                processes.push(ProcessInfo {
                    pid: parts[0].parse().unwrap_or(0),
                    name: parts[1].to_string(),
                    cpu_percent: parts[2].parse().unwrap_or(0.0),
                    memory_mb: parts[3].parse::<f32>().unwrap_or(0.0) * 10.0, // rough estimate
                    status: "running".to_string(),
                    uptime_seconds: Self::parse_uptime(parts[4]),
                });
            }
        }

        Ok(processes)
    }

    fn parse_uptime(time_str: &str) -> u64 {
        // Parse time format like "1:23" or "12:34:56"
        let parts: Vec<&str> = time_str.split(':').collect();
        match parts.len() {
            2 => {
                let mins: u64 = parts[0].parse().unwrap_or(0);
                let secs: u64 = parts[1].parse().unwrap_or(0);
                mins * 60 + secs
            }
            3 => {
                let hours: u64 = parts[0].parse().unwrap_or(0);
                let mins: u64 = parts[1].parse().unwrap_or(0);
                let secs: u64 = parts[2].parse().unwrap_or(0);
                hours * 3600 + mins * 60 + secs
            }
            _ => 0,
        }
    }
}
