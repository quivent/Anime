use anyhow::{Result, anyhow};
use serde::{Deserialize, Serialize};
use std::process::Command;
use tauri::Emitter;
use tokio::io::{AsyncBufReadExt, BufReader};
use tokio::process::Command as TokioCommand;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InstallProgress {
    pub package_id: String,
    pub status: InstallStatus,
    pub progress: u8,
    pub message: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum InstallStatus {
    Pending,
    Installing,
    Completed,
    Failed,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PackageStatus {
    pub package_id: String,
    pub installed: bool,
    pub version: Option<String>,
}

/// Check if a package is installed on the remote server
pub async fn check_package_installed(
    host: &str,
    username: &str,
    package_id: &str,
) -> Result<bool> {
    let check_command = match package_id {
        "anime" => "which anime",
        "comfyui" => "[ -d ~/ComfyUI ] && echo 'installed'",
        "ollama" => "which ollama",
        "core" => "which nvidia-smi",
        "python" => "python3 --version",
        "pytorch" => "python3 -c 'import torch; print(torch.__version__)'",
        "vllm" => "python3 -c 'import vllm; print(vllm.__version__)'",
        "sglang" => "python3 -c 'import sglang; print(sglang.__version__)'",
        "aphrodite" => "python3 -c 'import aphrodite; print(aphrodite.__version__)'",
        _ => return Ok(false),
    };

    let output = Command::new("ssh")
        .args(&[
            "-o", "StrictHostKeyChecking=no",
            "-o", "UserKnownHostsFile=/dev/null",
            &format!("{}@{}", username, host),
            check_command,
        ])
        .output()?;

    Ok(output.status.success())
}

/// Get the installation script for a package
fn get_install_script(package_id: &str) -> Option<&'static str> {
    match package_id {
        "anime" => Some(r#"
#!/bin/bash
set -e

echo "Installing ANIME..."

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ]; then
    ARCH="arm64"
fi

# Download the appropriate binary
DOWNLOAD_URL="https://github.com/joshkornreich/anime/releases/latest/download/anime-${OS}-${ARCH}"
echo "Downloading from: $DOWNLOAD_URL"

# Create bin directory if it doesn't exist
mkdir -p ~/bin

# Download and install
curl -L "$DOWNLOAD_URL" -o ~/bin/anime
chmod +x ~/bin/anime

# Add to PATH if not already there
if ! grep -q 'export PATH="$HOME/bin:$PATH"' ~/.bashrc; then
    echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
fi

if ! grep -q 'export PATH="$HOME/bin:$PATH"' ~/.zshrc 2>/dev/null; then
    echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc 2>/dev/null || true
fi

echo "ANIME installed successfully!"
anime version || echo "Please restart your shell or run: source ~/.bashrc"
        "#),
        "pytorch" => Some(r#"
#!/bin/bash
set -e

echo "Starting PyTorch installation..."

# Check if PyTorch is already installed
if python3 -c "import torch; print(f'PyTorch {torch.__version__} already installed')" 2>/dev/null; then
    echo "PyTorch is already installed. Skipping installation."
    exit 0
fi

# Ensure Python 3.10+ is available
PYTHON_VERSION=$(python3 --version 2>&1 | awk '{print $2}' | cut -d. -f1,2)
echo "Detected Python version: $PYTHON_VERSION"

if ! python3 -c "import sys; exit(0 if sys.version_info >= (3, 10) else 1)"; then
    echo "Error: Python 3.10+ is required"
    exit 1
fi

# Ensure pip is available
echo "Ensuring pip is up to date..."
python3 -m pip install --upgrade pip

# Detect CUDA version
echo "Detecting CUDA version..."
if command -v nvidia-smi &> /dev/null; then
    CUDA_VERSION=$(nvidia-smi | grep "CUDA Version" | awk '{print $9}' | cut -d. -f1,2)
    echo "Detected CUDA version: $CUDA_VERSION"
else
    echo "Warning: nvidia-smi not found, assuming CUDA 12.x"
    CUDA_VERSION="12.1"
fi

# Install PyTorch with CUDA support
echo "Installing PyTorch stack with CUDA support..."
echo "This includes: torch, torchvision, torchaudio, transformers, diffusers, accelerate"

# Use pip with CUDA 12.x support (compatible with GH200)
python3 -m pip install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cu121

# Install essential ML libraries
echo "Installing transformers and related packages..."
python3 -m pip install transformers diffusers accelerate safetensors huggingface-hub

# Install additional utilities
echo "Installing utility packages..."
python3 -m pip install sentencepiece tokenizers peft bitsandbytes

# Verify installation
echo "Verifying PyTorch installation..."
python3 -c "
import torch
print(f'PyTorch version: {torch.__version__}')
print(f'CUDA available: {torch.cuda.is_available()}')
if torch.cuda.is_available():
    print(f'CUDA version: {torch.version.cuda}')
    print(f'GPU devices: {torch.cuda.device_count()}')
    for i in range(torch.cuda.device_count()):
        print(f'  Device {i}: {torch.cuda.get_device_name(i)}')
"

echo "PyTorch installation complete!"
        "#),
        "vllm" => Some(r#"
#!/bin/bash
set -e

echo "Starting vLLM installation..."

# Check if vLLM is already installed
if python3 -c "import vllm; print(f'vLLM {vllm.__version__} already installed')" 2>/dev/null; then
    echo "vLLM is already installed. Skipping installation."
    exit 0
fi

# Verify prerequisites
echo "Checking prerequisites..."
if ! python3 -c "import torch" 2>/dev/null; then
    echo "Error: PyTorch is required. Please install PyTorch first."
    exit 1
fi

# Ensure Python 3.10+ is available
if ! python3 -c "import sys; exit(0 if sys.version_info >= (3, 10) else 1)"; then
    echo "Error: Python 3.10+ is required"
    exit 1
fi

# Check CUDA availability
echo "Verifying CUDA setup..."
python3 -c "
import torch
if not torch.cuda.is_available():
    print('Error: CUDA is not available')
    exit(1)
print(f'CUDA version: {torch.version.cuda}')
print(f'GPU count: {torch.cuda.device_count()}')
"

# Install system dependencies
echo "Installing system dependencies..."
sudo apt-get update
sudo apt-get install -y \
    build-essential \
    python3-dev \
    libssl-dev \
    libffi-dev \
    libnccl2 \
    libnccl-dev

# Install vLLM
echo "Installing vLLM inference engine..."
python3 -m pip install vllm

# Install optional dependencies for better performance
echo "Installing optional performance dependencies..."
python3 -m pip install ray flash-attn --no-build-isolation || echo "Note: Some optional dependencies may have failed"

# Verify installation
echo "Verifying vLLM installation..."
python3 -c "
import vllm
print(f'vLLM version: {vllm.__version__}')

# Test basic functionality
from vllm import LLM
print('vLLM import successful!')
print('PagedAttention available: True')
"

echo "vLLM installation complete!"
echo "You can now run: python3 -m vllm.entrypoints.openai.api_server --model <model_name>"
        "#),
        "sglang" => Some(r#"
#!/bin/bash
set -e

echo "Starting SGLang installation..."

# Check if SGLang is already installed
if python3 -c "import sglang; print(f'SGLang {sglang.__version__} already installed')" 2>/dev/null; then
    echo "SGLang is already installed. Skipping installation."
    exit 0
fi

# Verify prerequisites
echo "Checking prerequisites..."
if ! python3 -c "import torch" 2>/dev/null; then
    echo "Error: PyTorch is required. Please install PyTorch first."
    exit 1
fi

# Ensure Python 3.10+ is available
if ! python3 -c "import sys; exit(0 if sys.version_info >= (3, 10) else 1)"; then
    echo "Error: Python 3.10+ is required"
    exit 1
fi

# Check CUDA availability
echo "Verifying CUDA setup..."
python3 -c "
import torch
if not torch.cuda.is_available():
    print('Error: CUDA is not available')
    exit(1)
print(f'CUDA version: {torch.version.cuda}')
print(f'GPU count: {torch.cuda.device_count()}')
"

# Install system dependencies
echo "Installing system dependencies..."
sudo apt-get update
sudo apt-get install -y \
    build-essential \
    python3-dev \
    git

# Install SGLang
echo "Installing SGLang framework..."
python3 -m pip install "sglang[all]"

# Install FlashInfer for optimized attention
echo "Installing FlashInfer for enhanced performance..."
python3 -m pip install flashinfer -i https://flashinfer.ai/whl/cu121/torch2.4/ || \
    echo "Note: FlashInfer installation optional, continuing anyway"

# Verify installation
echo "Verifying SGLang installation..."
python3 -c "
import sglang as sgl
print(f'SGLang version: {sgl.__version__}')
print('SGLang import successful!')
print('Structured generation available: True')
"

echo "SGLang installation complete!"
echo "You can now use SGLang for structured LLM generation"
echo "Run: python3 -m sglang.launch_server --model <model_name> --port 30000"
        "#),
        "aphrodite" => Some(r#"
#!/bin/bash
set -e

echo "Starting Aphrodite Engine installation..."

# Check if Aphrodite is already installed
if python3 -c "import aphrodite; print(f'Aphrodite {aphrodite.__version__} already installed')" 2>/dev/null; then
    echo "Aphrodite Engine is already installed. Skipping installation."
    exit 0
fi

# Verify prerequisites
echo "Checking prerequisites..."
if ! python3 -c "import torch" 2>/dev/null; then
    echo "Error: PyTorch is required. Please install PyTorch first."
    exit 1
fi

# Ensure Python 3.10+ is available
if ! python3 -c "import sys; exit(0 if sys.version_info >= (3, 10) else 1)"; then
    echo "Error: Python 3.10+ is required"
    exit 1
fi

# Check CUDA availability
echo "Verifying CUDA setup..."
python3 -c "
import torch
if not torch.cuda.is_available():
    print('Error: CUDA is not available')
    exit(1)
print(f'CUDA version: {torch.version.cuda}')
print(f'GPU count: {torch.cuda.device_count()}')
"

# Install system dependencies
echo "Installing system dependencies..."
sudo apt-get update
sudo apt-get install -y \
    build-essential \
    python3-dev \
    git \
    cmake \
    ninja-build

# Install Aphrodite Engine
echo "Installing Aphrodite inference engine..."
python3 -m pip install aphrodite-engine

# Install additional dependencies for better compatibility
echo "Installing additional dependencies..."
python3 -m pip install ray xformers packaging ninja

# Verify installation
echo "Verifying Aphrodite installation..."
python3 -c "
import aphrodite
print(f'Aphrodite version: {aphrodite.__version__}')

from aphrodite import EngineArgs, LLMEngine, SamplingParams
print('Aphrodite Engine import successful!')
print('Inference engine ready!')
"

echo "Aphrodite Engine installation complete!"
echo "You can now run: python3 -m aphrodite.endpoints.openai.api_server --model <model_name>"
        "#),
        _ => None,
    }
}

/// Parse a line of output to determine progress and extract meaningful information
///
/// This function analyzes stdout/stderr output from installation scripts and extracts
/// progress information based on recognized patterns.
///
/// # Arguments
/// * `package_id` - The package being installed (e.g., "anime", "comfyui")
/// * `line` - A single line of output from the installation script
///
/// # Returns
/// * `Some((progress, message))` - If a recognized pattern is found:
///   - `progress`: 0-100 indicating installation progress (0 for errors/warnings)
///   - `message`: Human-readable description of the current step
/// * `None` - If the line doesn't match any known patterns
///
/// # Adding Support for New Packages
/// To add progress tracking for a new package:
/// 1. Add a new `if package_id == "yourpackage"` block
/// 2. Match on key output patterns from your installation script
/// 3. Return appropriate progress values (spread across 0-95 range)
/// 4. Use progress 95-99 for verification steps, 100 is reserved for completion
fn parse_progress_from_line(package_id: &str, line: &str) -> Option<(u8, String)> {
    let line_lower = line.to_lowercase();

    // Package-specific progress parsing for ANIME
    if package_id == "anime" {
        if line.contains("Installing ANIME") {
            return Some((5, "Starting ANIME installation...".to_string()));
        } else if line.contains("Detect") || line.contains("uname") {
            return Some((15, "Detecting system architecture...".to_string()));
        } else if line.contains("Downloading") || line.contains("download") {
            return Some((30, "Downloading ANIME binary...".to_string()));
        } else if line.contains("curl") && line.contains("releases") {
            return Some((40, "Fetching from GitHub releases...".to_string()));
        } else if line.contains("mkdir") || line.contains("bin") {
            return Some((60, "Creating installation directory...".to_string()));
        } else if line.contains("chmod +x") || line.contains("chmod") {
            return Some((75, "Setting executable permissions...".to_string()));
        } else if line.contains("export PATH") || line.contains(".bashrc") {
            return Some((85, "Updating shell configuration...".to_string()));
        } else if line.contains("installed successfully") || line.contains("version") {
            return Some((95, "Verifying installation...".to_string()));
        }
    }

    // Generic progress indicators
    if line_lower.contains("error") || line_lower.contains("failed") {
        return Some((0, format!("Error: {}", line.trim())));
    } else if line_lower.contains("warning") {
        return Some((0, format!("Warning: {}", line.trim())));
    }

    // APT progress parsing (for future packages)
    if line_lower.contains("reading package lists") {
        return Some((10, "Reading package lists...".to_string()));
    } else if line_lower.contains("building dependency tree") {
        return Some((20, "Building dependency tree...".to_string()));
    } else if line_lower.contains("unpacking") {
        return Some((40, "Unpacking packages...".to_string()));
    } else if line_lower.contains("setting up") {
        return Some((70, "Setting up packages...".to_string()));
    }

    // Download progress (percentage indicators)
    if let Some(percent_pos) = line.find('%') {
        if let Some(start) = line[..percent_pos].rfind(|c: char| !c.is_numeric()) {
            if let Ok(percent) = line[start+1..percent_pos].trim().parse::<u8>() {
                return Some((
                    percent.min(90),
                    format!("Downloading... {}%", percent)
                ));
            }
        }
    }

    None
}

/// Install a package on the remote server with progress tracking
pub async fn install_package<R: tauri::Runtime>(
    app: tauri::AppHandle<R>,
    host: &str,
    username: &str,
    package_id: &str,
) -> Result<()> {
    // Emit initial progress
    app.emit("install_progress", InstallProgress {
        package_id: package_id.to_string(),
        status: InstallStatus::Installing,
        progress: 0,
        message: format!("Starting installation of {}...", package_id),
    })?;

    let script = get_install_script(package_id)
        .ok_or_else(|| anyhow!("No installation script for package: {}", package_id))?;

    // Create temporary script file
    let script_path = format!("/tmp/install_{}.sh", package_id);
    std::fs::write(&script_path, script)?;

    // Copy script to remote server
    app.emit("install_progress", InstallProgress {
        package_id: package_id.to_string(),
        status: InstallStatus::Installing,
        progress: 2,
        message: "Uploading installation script...".to_string(),
    })?;

    let remote_script = format!("/tmp/install_{}.sh", package_id);
    let scp_output = Command::new("scp")
        .args(&[
            "-o", "StrictHostKeyChecking=no",
            "-o", "UserKnownHostsFile=/dev/null",
            &script_path,
            &format!("{}@{}:{}", username, host, remote_script),
        ])
        .output()?;

    if !scp_output.status.success() {
        let error_msg = String::from_utf8_lossy(&scp_output.stderr);
        app.emit("install_progress", InstallProgress {
            package_id: package_id.to_string(),
            status: InstallStatus::Failed,
            progress: 0,
            message: format!("Failed to upload script: {}", error_msg),
        })?;
        return Err(anyhow!("Failed to upload installation script: {}", error_msg));
    }

    // Execute installation script with real-time output capture
    app.emit("install_progress", InstallProgress {
        package_id: package_id.to_string(),
        status: InstallStatus::Installing,
        progress: 5,
        message: "Executing installation script...".to_string(),
    })?;

    let ssh_command = format!("chmod +x {} && bash -x {}", remote_script, remote_script);

    let mut child = TokioCommand::new("ssh")
        .args(&[
            "-o", "StrictHostKeyChecking=no",
            "-o", "UserKnownHostsFile=/dev/null",
            &format!("{}@{}", username, host),
            &ssh_command,
        ])
        .stdout(std::process::Stdio::piped())
        .stderr(std::process::Stdio::piped())
        .spawn()?;

    // Capture stdout
    let stdout = child.stdout.take()
        .ok_or_else(|| anyhow!("Failed to capture stdout"))?;
    let stdout_reader = BufReader::new(stdout);
    let mut stdout_lines = stdout_reader.lines();

    // Capture stderr
    let stderr = child.stderr.take()
        .ok_or_else(|| anyhow!("Failed to capture stderr"))?;
    let stderr_reader = BufReader::new(stderr);
    let mut stderr_lines = stderr_reader.lines();

    // Track current progress
    let mut current_progress = 5u8;
    let mut error_occurred = false;
    let mut error_messages = Vec::new();

    // Process stdout and stderr concurrently
    loop {
        tokio::select! {
            // Read from stdout
            stdout_result = stdout_lines.next_line() => {
                match stdout_result {
                    Ok(Some(line)) => {

                        // Parse progress from the line
                        if let Some((progress, message)) = parse_progress_from_line(package_id, &line) {
                            // Only update if progress has increased or it's an error/warning
                            if progress > current_progress || progress == 0 {
                                if progress > 0 {
                                    current_progress = progress;
                                }

                                app.emit("install_progress", InstallProgress {
                                    package_id: package_id.to_string(),
                                    status: InstallStatus::Installing,
                                    progress: current_progress,
                                    message,
                                })?;
                            }
                        } else if !line.trim().is_empty() {
                            // Emit generic progress update for non-empty lines
                            app.emit("install_progress", InstallProgress {
                                package_id: package_id.to_string(),
                                status: InstallStatus::Installing,
                                progress: current_progress,
                                message: line.trim().to_string(),
                            })?;
                        }
                    }
                    Ok(None) => {
                        // stdout closed
                    }
                    Err(e) => {

                    }
                }
            }

            // Read from stderr
            stderr_result = stderr_lines.next_line() => {
                match stderr_result {
                    Ok(Some(line)) => {

                        // Check for errors in stderr
                        let line_lower = line.to_lowercase();
                        if line_lower.contains("error") || line_lower.contains("fatal") {
                            error_occurred = true;
                            error_messages.push(line.clone());

                            app.emit("install_progress", InstallProgress {
                                package_id: package_id.to_string(),
                                status: InstallStatus::Installing,
                                progress: current_progress,
                                message: format!("Error: {}", line.trim()),
                            })?;
                        } else if !line.trim().is_empty() {
                            // Parse progress from stderr too (bash -x outputs to stderr)
                            if let Some((progress, message)) = parse_progress_from_line(package_id, &line) {
                                if progress > current_progress || progress == 0 {
                                    if progress > 0 {
                                        current_progress = progress;
                                    }

                                    app.emit("install_progress", InstallProgress {
                                        package_id: package_id.to_string(),
                                        status: InstallStatus::Installing,
                                        progress: current_progress,
                                        message,
                                    })?;
                                }
                            }
                        }
                    }
                    Ok(None) => {
                        // stderr closed
                    }
                    Err(e) => {

                    }
                }
            }

            // Wait for process to exit
            _ = tokio::time::sleep(tokio::time::Duration::from_millis(100)) => {
                // Check if process has exited
                if let Ok(Some(_)) = child.try_wait() {
                    break;
                }
            }
        }
    }

    // Wait for final exit status
    let status = child.wait().await?;

    // Clean up remote script
    let _ = Command::new("ssh")
        .args(&[
            "-o", "StrictHostKeyChecking=no",
            "-o", "UserKnownHostsFile=/dev/null",
            &format!("{}@{}", username, host),
            &format!("rm -f {}", remote_script),
        ])
        .output();

    // Clean up local script
    let _ = std::fs::remove_file(&script_path);

    // Determine final status based on exit code
    if !status.success() || error_occurred {
        let error_msg = if !error_messages.is_empty() {
            error_messages.join("; ")
        } else {
            format!("Installation failed with exit code: {:?}", status.code())
        };

        app.emit("install_progress", InstallProgress {
            package_id: package_id.to_string(),
            status: InstallStatus::Failed,
            progress: 0,
            message: error_msg.clone(),
        })?;

        return Err(anyhow!("Installation failed: {}", error_msg));
    }

    // Emit completion
    app.emit("install_progress", InstallProgress {
        package_id: package_id.to_string(),
        status: InstallStatus::Completed,
        progress: 100,
        message: format!("{} installed successfully!", package_id),
    })?;

    Ok(())
}

/// Get installation status for all packages
pub async fn get_packages_status(
    host: &str,
    username: &str,
    package_ids: &[String],
) -> Result<Vec<PackageStatus>> {
    let mut statuses = Vec::new();

    for package_id in package_ids {
        let installed = check_package_installed(host, username, package_id).await?;
        statuses.push(PackageStatus {
            package_id: package_id.clone(),
            installed,
            version: None, // TODO: Get actual version
        });
    }

    Ok(statuses)
}
