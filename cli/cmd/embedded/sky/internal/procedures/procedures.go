package procedures

import (
    "fmt"
    "github.com/sky-cli/sky/internal/config"
    "github.com/sky-cli/sky/ui"
)

// Procedure represents a setup procedure
type Procedure struct {
    ID          string
    Name        string
    Description string
    Steps       []Step
    Duration    string
    Difficulty  string
}

// Step represents a procedure step
type Step struct {
    Number      int
    Title       string
    Description string
    Commands    []string
    Verification string
}

// Procedures handles procedure listing and execution
type Procedures struct {
    Config *config.Config
}

// New creates a new Procedures instance
func New(cfg *config.Config) *Procedures {
    return &Procedures{Config: cfg}
}

// GetAll returns all procedures in order
func (p *Procedures) GetAll() []Procedure {
    return []Procedure{
        p.hardwareVerification(),
        p.environmentSetup(),
        p.modelDownload(),
        p.dependencyInstall(),
        p.parallelismConfig(),
        p.optimizationConfig(),
        p.validationTest(),
    }
}

// PrintAll prints all procedures
func (p *Procedures) PrintAll() {
    ui.PrintHeader("Setup Procedures")

    procedures := p.GetAll()

    headers := []string{"#", "Procedure", "Difficulty", "Est. Time"}
    rows := make([][]string, len(procedures))
    for i, proc := range procedures {
        rows[i] = []string{
            fmt.Sprintf("%d", i+1),
            proc.Name,
            proc.Difficulty,
            proc.Duration,
        }
    }
    ui.PrintTable(headers, rows)

    fmt.Println()
    ui.PrintStatus("info", "Run 'sky procedures show <number>' for detailed steps")
    ui.PrintStatus("info", "Run 'sky procedures run <number>' to execute a procedure")
}

// PrintProcedure prints a specific procedure
func (p *Procedures) PrintProcedure(id int) {
    procedures := p.GetAll()
    if id < 1 || id > len(procedures) {
        ui.PrintSuggestion(fmt.Sprintf("Procedure %d not found", id), []string{
            fmt.Sprintf("Valid procedure numbers: 1-%d", len(procedures)),
            "Run 'sky procedures' to see all procedures",
        })
        return
    }

    proc := procedures[id-1]
    ui.PrintHeader(fmt.Sprintf("Procedure %d: %s", id, proc.Name))

    ui.PrintKeyValue("Description", proc.Description)
    ui.PrintKeyValue("Difficulty", proc.Difficulty)
    ui.PrintKeyValue("Estimated Time", proc.Duration)

    ui.PrintSection("Steps")
    for _, step := range proc.Steps {
        fmt.Printf("\n  %s%d.%s %s\n", ui.BrightCyan, step.Number, ui.Reset, ui.Title(step.Title))
        fmt.Printf("     %s\n", ui.Muted(step.Description))

        if len(step.Commands) > 0 {
            fmt.Printf("\n     %s\n", ui.Key("Commands:"))
            for _, cmd := range step.Commands {
                fmt.Printf("     %s %s\n", ui.Muted("$"), ui.Value(cmd))
            }
        }

        if step.Verification != "" {
            fmt.Printf("\n     %s %s\n", ui.Key("Verify:"), step.Verification)
        }
    }
}

func (p *Procedures) hardwareVerification() Procedure {
    return Procedure{
        ID:          "hardware",
        Name:        "Hardware Verification",
        Description: "Verify GPU hardware and NVLink connectivity",
        Duration:    "5 min",
        Difficulty:  "Easy",
        Steps: []Step{
            {
                Number:      1,
                Title:       "Check GPU Detection",
                Description: "Verify all 4 H100 GPUs are detected",
                Commands: []string{
                    "nvidia-smi -L",
                },
                Verification: "Should list 4 H100 GPUs",
            },
            {
                Number:      2,
                Title:       "Verify NVLink Topology",
                Description: "Check NVLink connectivity between GPUs",
                Commands: []string{
                    "nvidia-smi topo -m",
                },
                Verification: "Should show NV18 connections between all GPU pairs",
            },
            {
                Number:      3,
                Title:       "Check Memory",
                Description: "Verify GPU memory availability",
                Commands: []string{
                    "nvidia-smi --query-gpu=memory.total,memory.free --format=csv",
                },
                Verification: "Each GPU should show ~80GB total",
            },
            {
                Number:      4,
                Title:       "Test NVLink Bandwidth",
                Description: "Run bandwidth test between GPUs",
                Commands: []string{
                    "# Optional: Run CUDA samples p2pBandwidthLatencyTest",
                    "/usr/local/cuda/samples/1_Utilities/p2pBandwidthLatencyTest/p2pBandwidthLatencyTest",
                },
                Verification: "Should show ~800+ GB/s bidirectional",
            },
        },
    }
}

func (p *Procedures) environmentSetup() Procedure {
    return Procedure{
        ID:          "environment",
        Name:        "Environment Setup",
        Description: "Set up Python environment and CUDA",
        Duration:    "15 min",
        Difficulty:  "Easy",
        Steps: []Step{
            {
                Number:      1,
                Title:       "Create Virtual Environment",
                Description: "Create isolated Python environment",
                Commands: []string{
                    "python3 -m venv ~/.venv/skyreels",
                    "source ~/.venv/skyreels/bin/activate",
                },
                Verification: "Prompt should show (skyreels)",
            },
            {
                Number:      2,
                Title:       "Verify CUDA Version",
                Description: "Ensure CUDA 12.x is installed",
                Commands: []string{
                    "nvcc --version",
                },
                Verification: "Should show CUDA 12.1 or higher",
            },
            {
                Number:      3,
                Title:       "Install PyTorch",
                Description: "Install PyTorch with CUDA support",
                Commands: []string{
                    "pip install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cu121",
                },
                Verification: "Run: python -c \"import torch; print(torch.cuda.is_available())\" # Should print True",
            },
            {
                Number:      4,
                Title:       "Set Environment Variables",
                Description: "Configure CUDA and NCCL settings",
                Commands: []string{
                    "export CUDA_VISIBLE_DEVICES=0,1,2,3",
                    "export NCCL_P2P_LEVEL=NVL",
                    "export NCCL_IB_DISABLE=1",
                },
                Verification: "Add to ~/.bashrc for persistence",
            },
        },
    }
}

func (p *Procedures) modelDownload() Procedure {
    return Procedure{
        ID:          "model",
        Name:        "Model Download",
        Description: "Download SkyReel model weights",
        Duration:    "30-60 min",
        Difficulty:  "Easy",
        Steps: []Step{
            {
                Number:      1,
                Title:       "Install Hugging Face CLI",
                Description: "Install HF tools for model download",
                Commands: []string{
                    "pip install huggingface_hub",
                },
                Verification: "huggingface-cli --help should work",
            },
            {
                Number:      2,
                Title:       "Login to Hugging Face",
                Description: "Authenticate with HF (if model requires)",
                Commands: []string{
                    "huggingface-cli login",
                },
                Verification: "Follow prompts to enter token",
            },
            {
                Number:      3,
                Title:       "Download Model",
                Description: "Download SkyReels-V2-DF-14B model",
                Commands: []string{
                    "huggingface-cli download Skywork/SkyReels-V2-DF-14B-540P --local-dir /models/skyreels",
                },
                Verification: "Check /models/skyreels contains model files",
            },
            {
                Number:      4,
                Title:       "Download T5-XXL",
                Description: "Download text encoder",
                Commands: []string{
                    "huggingface-cli download google/t5-xxl-lm-adapt --local-dir /models/t5-xxl",
                },
                Verification: "Check /models/t5-xxl contains encoder files",
            },
        },
    }
}

func (p *Procedures) dependencyInstall() Procedure {
    return Procedure{
        ID:          "dependencies",
        Name:        "Dependency Installation",
        Description: "Install SkyReel and required packages",
        Duration:    "20 min",
        Difficulty:  "Medium",
        Steps: []Step{
            {
                Number:      1,
                Title:       "Clone SkyReels-V2",
                Description: "Get the SkyReel repository",
                Commands: []string{
                    "git clone git@github.com:SkyworkAI/SkyReels-V2.git",
                    "cd SkyReels-V2",
                },
                Verification: "Directory should contain inference scripts",
            },
            {
                Number:      2,
                Title:       "Install Requirements",
                Description: "Install Python dependencies",
                Commands: []string{
                    "pip install -r requirements.txt",
                },
                Verification: "No errors during installation",
            },
            {
                Number:      3,
                Title:       "Install xDiT",
                Description: "Install distributed inference framework",
                Commands: []string{
                    "pip install xfuser",
                },
                Verification: "python -c \"import xfuser\" should work",
            },
            {
                Number:      4,
                Title:       "Install Flash Attention",
                Description: "Install optimized attention",
                Commands: []string{
                    "pip install flash-attn --no-build-isolation",
                },
                Verification: "May take 10+ minutes to compile",
            },
            {
                Number:      5,
                Title:       "Install FP8 Support",
                Description: "Install transformer engine for FP8",
                Commands: []string{
                    "pip install transformer-engine",
                },
                Verification: "python -c \"import transformer_engine\" should work",
            },
        },
    }
}

func (p *Procedures) parallelismConfig() Procedure {
    return Procedure{
        ID:          "parallelism",
        Name:        "Parallelism Configuration",
        Description: "Configure multi-GPU parallelism strategy",
        Duration:    "10 min",
        Difficulty:  "Medium",
        Steps: []Step{
            {
                Number:      1,
                Title:       "Configure Context Parallel",
                Description: "Set up 4-way context parallelism",
                Commands: []string{
                    "# In inference script or config:",
                    "export ULYSSES_DEGREE=4  # Context parallel degree",
                    "export RING_DEGREE=1     # Ring attention degree",
                },
                Verification: "Total GPUs = ULYSSES_DEGREE × RING_DEGREE",
            },
            {
                Number:      2,
                Title:       "Configure CFG Parallel",
                Description: "Set up CFG parallelism (optional)",
                Commands: []string{
                    "# For CP2 + CFG2 hybrid:",
                    "export ULYSSES_DEGREE=2",
                    "export CFG_PARALLEL=2",
                },
                Verification: "Use for shorter sequences",
            },
            {
                Number:      3,
                Title:       "Enable VAE Parallel",
                Description: "Distribute VAE across GPUs",
                Commands: []string{
                    "# In config or script:",
                    "vae_parallel=True",
                },
                Verification: "Reduces VAE memory per GPU",
            },
            {
                Number:      4,
                Title:       "Verify Parallelism",
                Description: "Run test inference to verify setup",
                Commands: []string{
                    "torchrun --nproc_per_node=4 test_inference.py",
                },
                Verification: "All 4 GPUs should show activity",
            },
        },
    }
}

func (p *Procedures) optimizationConfig() Procedure {
    return Procedure{
        ID:          "optimization",
        Name:        "Optimization Configuration",
        Description: "Configure TeaCache, FP8, and other optimizations",
        Duration:    "10 min",
        Difficulty:  "Medium",
        Steps: []Step{
            {
                Number:      1,
                Title:       "Enable FP8 Quantization",
                Description: "Use FP8 for model weights",
                Commands: []string{
                    "# In inference config:",
                    "--precision fp8",
                    "--quantization fp8_e4m3fn",
                },
                Verification: "Memory usage should drop ~50%",
            },
            {
                Number:      2,
                Title:       "Configure TeaCache",
                Description: "Enable token-level caching",
                Commands: []string{
                    "--teacache_enabled true",
                    "--teacache_thresh 0.3",
                },
                Verification: "Should see ~20-30% speedup",
            },
            {
                Number:      3,
                Title:       "Enable Model Compilation",
                Description: "Use torch.compile for optimization",
                Commands: []string{
                    "--compile_model true",
                    "# Or in Python:",
                    "model = torch.compile(model, mode='reduce-overhead')",
                },
                Verification: "First run slower, subsequent runs faster",
            },
            {
                Number:      4,
                Title:       "Configure Diffusion Forcing",
                Description: "Set ar_step for continuity",
                Commands: []string{
                    "--ar_step 5  # For async mode with continuity",
                    "--ar_step 0  # For sync mode (faster)",
                },
                Verification: "ar_step=5 gives better continuity",
            },
        },
    }
}

func (p *Procedures) validationTest() Procedure {
    return Procedure{
        ID:          "validation",
        Name:        "Validation Test",
        Description: "Run validation to confirm setup",
        Duration:    "5 min",
        Difficulty:  "Easy",
        Steps: []Step{
            {
                Number:      1,
                Title:       "Run Short Test",
                Description: "Generate short test video",
                Commands: []string{
                    "python inference.py \\",
                    "  --prompt \"A serene lake at sunset\" \\",
                    "  --num_frames 25 \\",
                    "  --width 544 --height 960 \\",
                    "  --ar_step 0",
                },
                Verification: "Should complete in ~30-60 seconds",
            },
            {
                Number:      2,
                Title:       "Monitor GPU Utilization",
                Description: "Check all GPUs are being used",
                Commands: []string{
                    "watch -n 1 nvidia-smi",
                },
                Verification: "All 4 GPUs should show high utilization",
            },
            {
                Number:      3,
                Title:       "Verify Output",
                Description: "Check generated video",
                Commands: []string{
                    "ffprobe output/generated.mp4",
                },
                Verification: "Video should match configured resolution/fps",
            },
            {
                Number:      4,
                Title:       "Run Continuity Test",
                Description: "Test sequential generation",
                Commands: []string{
                    "python inference.py \\",
                    "  --prompt \"A bird flying across a mountain range\" \\",
                    "  --num_frames 97 \\",
                    "  --ar_step 5 \\",
                    "  --extend_video true",
                },
                Verification: "Should maintain visual continuity",
            },
        },
    }
}
