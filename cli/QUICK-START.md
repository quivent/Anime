# Lambda GH200 - Quick Start Card

## Copy-Paste Method (Fastest!)

```bash
# 1. SSH into your Lambda machine
ssh ubuntu@YOUR_LAMBDA_IP

# 2. Copy and paste one of these workflows:
```

### Minimal Setup (~8 min, ~$3)
```bash
# Core + PyTorch + 1 small model
curl -sSL https://raw.githubusercontent.com/YOUR_REPO/setup-core.sh | bash && \
curl -sSL https://raw.githubusercontent.com/YOUR_REPO/setup-pytorch.sh | bash && \
curl -sSL https://raw.githubusercontent.com/YOUR_REPO/setup-ollama.sh | bash && \
ollama pull mistral:latest
```

### Standard Setup (~20 min, ~$10)
```bash
# Everything + recommended models
wget https://raw.githubusercontent.com/YOUR_REPO/setup-core.sh
bash setup-core.sh
bash setup-all-optional.sh
bash setup-models.sh  # Then choose "recommended"
```

### Local SCP Method
```bash
# From your Mac:
cd /Users/joshkornreich/lambda
scp setup-*.sh ubuntu@YOUR_LAMBDA_IP:~/

# Then SSH in:
ssh ubuntu@YOUR_LAMBDA_IP
bash setup-core.sh      # Required, 5 min
bash setup-pytorch.sh   # If you need AI, 2 min
bash setup-ollama.sh    # If you need LLMs, 1 min
bash setup-models.sh    # Interactive, choose what you want
```

## What Each Script Does

| Script | Time | What | Cost @ $20/hr |
|--------|------|------|---------------|
| `setup-core.sh` | 5 min | CUDA, Python, Node, Docker | $1.67 |
| `setup-pytorch.sh` | 2 min | PyTorch + AI libs | $0.67 |
| `setup-ollama.sh` | 1 min | Ollama server | $0.33 |
| `setup-models.sh` | 2-40 min | **Interactive** model picker | $0.67-$13 |
| `setup-comfyui.sh` | 2 min | ComfyUI | $0.67 |
| `setup-claude.sh` | 30 sec | Claude Code | $0.17 |
| `setup-all-optional.sh` | 7 min | All above except models | $2.33 |

## Model Download Times

| Model | Size | Time | Cost @ $20/hr |
|-------|------|------|---------------|
| mistral:latest | 4.1GB | 2 min | $0.67 |
| llama3.3:8b | 4.7GB | 3 min | $1.00 |
| qwen3:8b | 5.2GB | 3 min | $1.00 |
| qwen3:14b | 9.3GB | 7 min | $2.33 |
| qwen3:32b | 20GB | 12 min | $4.00 |
| mixtral:8x7b | 26GB | 12 min | $4.00 |
| deepseek-coder:33b | 19GB | 10 min | $3.33 |
| llama3.3:70b | 40GB | 20 min | $6.67 |
| qwen3:235b | 142GB | 60 min | $20.00 |

## Pre-Made Workflows

### "Just Testing" - $3, 7 minutes
```bash
bash setup-core.sh && bash setup-pytorch.sh
python3 -c "import torch; print(torch.cuda.is_available())"
```

### "Small LLMs" - $5, 12 minutes
```bash
bash setup-core.sh
bash setup-ollama.sh
bash setup-models.sh  # Choose "all-small"
```

### "Production Ready" - $12, 22 minutes
```bash
bash setup-core.sh
bash setup-all-optional.sh
bash setup-models.sh  # Choose "recommended"
```

### "Everything" - $30, 60 minutes
```bash
bash setup-core.sh
bash setup-all-optional.sh
bash setup-models.sh  # Choose "all-large"
```

## After Installation

```bash
# Verify GPU
nvidia-smi

# Test PyTorch
python3 -c "import torch; print(f'CUDA: {torch.cuda.is_available()}')"

# Test Ollama
ollama list
ollama run mistral:latest "Hello!"

# Start ComfyUI
cd ~/ComfyUI && python3 main.py --listen 0.0.0.0 --port 8188
# Access at: http://YOUR_IP:8188

# Test Claude Code
claude-code --version
```

## Money-Saving Tips

1. ✓ **Install core first**, test GPU, then decide what else you need
2. ✓ **Use small models** (7B) for testing - 2-3 min each
3. ✓ **Skip ComfyUI** if you don't need Stable Diffusion UI
4. ✓ **Download 1 model at a time** - test before getting more
5. ✓ **Snapshot** your instance after setup
6. ✓ **Terminate** and restore from snapshot when needed

## Emergency Stop

If script is taking too long:
```bash
Ctrl+C  # Stop current script
# Bills stop immediately
# You can resume or pick different components
```

## Files Created

All scripts are in: `/Users/joshkornreich/lambda/`

To copy just one:
```bash
cat /Users/joshkornreich/lambda/setup-core.sh | pbcopy
```

Or SCP everything:
```bash
scp /Users/joshkornreich/lambda/setup-*.sh ubuntu@YOUR_IP:~/
```
