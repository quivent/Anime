# Lambda GH200 Setup Guide

This Ansible playbook sets up a complete AI/ML development environment on a fresh Lambda Ubuntu GH200 machine.

## What Gets Installed

### Core System
- NVIDIA Drivers (550) and CUDA Toolkit 12.4
- Python 3 with pip and venv
- Node.js 20.x and npm
- Docker and Docker Compose
- Build tools and dependencies

### AI/ML Libraries
- **PyTorch** (with CUDA support)
- **Transformers** (Hugging Face)
- **Diffusers** (Stable Diffusion pipeline)
- **Accelerate**, **PEFT**, **TRL**
- **xformers** (memory-efficient attention)
- **ONNX** and **ONNX Runtime GPU**
- Scientific libraries: NumPy, SciPy, scikit-learn, pandas
- Visualization: matplotlib, tensorboard
- Additional: opencv-python, pillow, einops, kornia

### LLM Tools
- **Ollama** with models:
  - Llama 3.3 70B
  - Qwen 2.2
  - Mistral
- **Claude Code** CLI

### Applications
- **ComfyUI** (Stable Diffusion UI)
  - ComfyUI Manager pre-installed
- Jupyter Notebook
- Hugging Face Hub CLI

### Development Tools
- Git, vim, tmux, htop
- TypeScript, ts-node
- Yarn, pnpm, pm2

## Prerequisites

1. **On your local machine**, install Ansible:
   ```bash
   # macOS
   brew install ansible

   # Ubuntu/Debian
   sudo apt install ansible

   # Or via pip
   pip install ansible
   ```

2. **SSH access** to your Lambda GH200 machine with sudo privileges

3. **SSH key** configured for passwordless authentication

## Setup Instructions

### 1. Configure Inventory

Edit `inventory.ini` and replace the placeholder values:

```ini
[gh200]
lambda-gh200 ansible_host=YOUR_LAMBDA_IP_HERE ansible_user=ubuntu ansible_ssh_private_key_file=~/.ssh/your_key.pem
```

Example:
```ini
[gh200]
lambda-gh200 ansible_host=192.168.1.100 ansible_user=ubuntu ansible_ssh_private_key_file=~/.ssh/lambda_key.pem
```

### 2. Test Connection

```bash
ansible -i inventory.ini gh200 -m ping
```

You should see a SUCCESS message.

### 3. Run the Playbook

```bash
ansible-playbook -i inventory.ini setup-gh200.yml
```

This will take 30-60 minutes depending on network speed (downloading large models).

### 4. Optional: Run with verbose output

```bash
ansible-playbook -i inventory.ini setup-gh200.yml -v
```

## Post-Installation

### SSH into your machine

```bash
ssh -i ~/.ssh/your_key.pem ubuntu@YOUR_LAMBDA_IP
```

### Start Services

```bash
~/start-services.sh
```

This starts:
- Ollama (port 11434)
- ComfyUI (port 8188)

### Access ComfyUI

Open your browser to: `http://YOUR_LAMBDA_IP:8188`

### Test Ollama

```bash
ollama list  # See installed models
ollama run llama3.3:70b "Hello, how are you?"
```

### Test Claude Code

```bash
claude-code --help
```

### Test Python/PyTorch

```bash
python3 -c "import torch; print(f'CUDA available: {torch.cuda.is_available()}'); print(f'GPU count: {torch.cuda.device_count()}')"
```

### Verify NVIDIA/CUDA

```bash
nvidia-smi
nvcc --version
```

## Customization

### Change Ollama Models

Edit `setup-gh200.yml` and modify the `ollama_models` variable:

```yaml
vars:
  ollama_models:
    - llama3.3:70b
    - qwen2.2:latest
    - mistral:latest
    - codellama:34b  # Add more models
```

### Change Node.js Version

The playbook uses Node.js 20.x. To use a different version, modify the NodeSource URL in the playbook.

### Install Additional Python Packages

Add to the Python AI/ML libraries task:

```yaml
- name: Install Python AI/ML libraries
  pip:
    name:
      # ... existing packages ...
      - your-package-here
```

## Troubleshooting

### NVIDIA drivers not loading
```bash
sudo reboot
# After reboot
nvidia-smi
```

### Ollama not starting
```bash
sudo systemctl status ollama
sudo systemctl restart ollama
journalctl -u ollama -f
```

### ComfyUI errors
```bash
cd ~/ComfyUI
python3 main.py --help
# Check logs
cat /tmp/comfyui.log
```

### Out of disk space
Large models require significant space. Check available space:
```bash
df -h
```

Consider cleaning up:
```bash
# Remove unused Docker images
docker system prune -a

# Remove unused pip cache
pip cache purge

# Remove unused Ollama models
ollama rm model-name
```

### Permission issues with Docker
Log out and back in for group changes to take effect:
```bash
exit
# SSH back in
ssh -i ~/.ssh/your_key.pem ubuntu@YOUR_LAMBDA_IP
```

## Additional Notes

- The GH200 has ARM64 architecture - all packages are installed for ARM64
- CUDA 12.4 is installed for maximum compatibility
- PyTorch is installed with CUDA 12.4 support
- Models are downloaded in parallel to save time
- The playbook is idempotent - safe to run multiple times

## Files Created

- `/usr/local/bin/ollama` - Ollama binary
- `~/ComfyUI/` - ComfyUI installation
- `~/start-services.sh` - Service startup script
- `/etc/systemd/system/ollama.service` - Ollama service

## Useful Commands

```bash
# Monitor GPU usage
watch -n 1 nvidia-smi

# Check running services
systemctl status ollama
docker ps

# View system resources
htop

# Check Python packages
pip list | grep torch

# Check Node.js version
node --version
npm --version

# Run Jupyter notebook
jupyter notebook --ip=0.0.0.0 --no-browser

# Update Ollama models
ollama pull llama3.3:70b
```

## Security Notes

- The playbook opens ports 8188 (ComfyUI) and 11434 (Ollama) on 0.0.0.0
- Consider using a firewall or VPN for production use
- Change default passwords if any services require authentication
- Keep SSH keys secure and use strong passphrases

## License

This setup script is provided as-is for use with Lambda Labs GH200 instances.
