package sequence

import (
    "fmt"

    "github.com/sky-cli/sky/internal/config"
    "github.com/sky-cli/sky/ui"
)

// Protocol represents an implementation sequence protocol
type Protocol struct {
    ID          string
    Name        string
    Description string
    Phases      []Phase
    Duration    string
}

// Phase represents a protocol phase
type Phase struct {
    Number      int
    Name        string
    Description string
    Steps       []string
    Checkpoint  string
}

// Sequence handles sequence protocol management
type Sequence struct {
    Config *config.Config
}

// New creates a new Sequence instance
func New(cfg *config.Config) *Sequence {
    return &Sequence{Config: cfg}
}

// GetProtocols returns all available protocols
func (s *Sequence) GetProtocols() []Protocol {
    return []Protocol{
        s.quickStartProtocol(),
        s.productionProtocol(),
        s.continuityProtocol(),
        s.optimizationProtocol(),
    }
}

// Print prints all protocols
func (s *Sequence) Print() {
    ui.PrintHeader("Implementation Sequence Protocols")

    protocols := s.GetProtocols()

    ui.PrintSection("Available Protocols")
    headers := []string{"ID", "Name", "Phases", "Est. Time"}
    rows := make([][]string, len(protocols))
    for i, p := range protocols {
        rows[i] = []string{p.ID, p.Name, fmt.Sprintf("%d", len(p.Phases)), p.Duration}
    }
    ui.PrintTable(headers, rows)

    fmt.Println()
    ui.PrintStatus("info", "Run 'sky sequence show <id>' to see protocol details")
    ui.PrintStatus("info", "Run 'sky sequence run <id>' to execute a protocol")
}

// PrintProtocol prints a specific protocol
func (s *Sequence) PrintProtocol(id string) {
    protocols := s.GetProtocols()

    var found *Protocol
    for _, p := range protocols {
        if p.ID == id {
            found = &p
            break
        }
    }

    if found == nil {
        ui.PrintSuggestion("Protocol not found: "+id, []string{
            "Run 'sky sequence' to see all protocols",
            "Valid IDs: quickstart, production, continuity, optimization",
        })
        return
    }

    ui.PrintHeader("Protocol: " + found.Name)

    ui.PrintKeyValue("ID", found.ID)
    ui.PrintKeyValue("Description", found.Description)
    ui.PrintKeyValue("Total Phases", fmt.Sprintf("%d", len(found.Phases)))
    ui.PrintKeyValue("Estimated Time", found.Duration)

    for _, phase := range found.Phases {
        ui.PrintSection(fmt.Sprintf("Phase %d: %s", phase.Number, phase.Name))
        fmt.Printf("  %s\n\n", ui.Muted(phase.Description))

        for i, step := range phase.Steps {
            fmt.Printf("  %s%d.%d%s %s\n", ui.BrightCyan, phase.Number, i+1, ui.Reset, step)
        }

        if phase.Checkpoint != "" {
            fmt.Printf("\n  %s %s\n", ui.Key("Checkpoint:"), ui.Success(phase.Checkpoint))
        }
    }
}

// PrintRun simulates running a protocol
func (s *Sequence) PrintRun(id string) {
    protocols := s.GetProtocols()

    var found *Protocol
    for _, p := range protocols {
        if p.ID == id {
            found = &p
            break
        }
    }

    if found == nil {
        ui.PrintSuggestion("Protocol not found: "+id, []string{
            "Run 'sky sequence' to see all protocols",
        })
        return
    }

    ui.PrintHeader("Executing Protocol: " + found.Name)

    totalSteps := 0
    for _, phase := range found.Phases {
        totalSteps += len(phase.Steps)
    }

    currentStep := 0
    for _, phase := range found.Phases {
        ui.PrintSection(fmt.Sprintf("Phase %d: %s", phase.Number, phase.Name))

        for _, step := range phase.Steps {
            currentStep++
            ui.PrintStatus("running", fmt.Sprintf("[%d/%d] %s", currentStep, totalSteps, step))
        }

        if phase.Checkpoint != "" {
            ui.PrintStatus("success", "Checkpoint: "+phase.Checkpoint)
        }
        fmt.Println()
    }

    ui.PrintSection("Execution Summary")
    ui.PrintProgressBar(totalSteps, totalSteps, 50)
    ui.PrintStatus("success", fmt.Sprintf("Protocol '%s' execution simulated", found.Name))

    fmt.Println()
    ui.PrintSuggestion("Simulation mode - no actual commands run", []string{
        "To execute, run the commands in each phase manually",
        "Run 'sky procedures' for detailed command instructions",
        "Run 'sky status' to verify current state",
    })
}

func (s *Sequence) quickStartProtocol() Protocol {
    return Protocol{
        ID:          "quickstart",
        Name:        "Quick Start",
        Description: "Fastest path to running video generation",
        Duration:    "30-45 min",
        Phases: []Phase{
            {
                Number:      1,
                Name:        "Environment Setup",
                Description: "Set up Python and CUDA environment",
                Steps: []string{
                    "Verify NVIDIA drivers and CUDA installation",
                    "Create Python virtual environment",
                    "Install PyTorch with CUDA support",
                    "Set CUDA_VISIBLE_DEVICES environment variable",
                },
                Checkpoint: "python -c 'import torch; print(torch.cuda.is_available())' returns True",
            },
            {
                Number:      2,
                Name:        "Model Setup",
                Description: "Download and configure the model",
                Steps: []string{
                    "Install huggingface_hub CLI",
                    "Download SkyReels-V2-DF-1.3B model (smaller for quick start)",
                    "Verify model files are complete",
                },
                Checkpoint: "Model directory contains config.json and safetensors files",
            },
            {
                Number:      3,
                Name:        "Dependencies",
                Description: "Install required Python packages",
                Steps: []string{
                    "Clone SkyReels-V2 repository",
                    "Install requirements.txt",
                    "Install xfuser for distributed inference",
                },
                Checkpoint: "All imports succeed without errors",
            },
            {
                Number:      4,
                Name:        "First Generation",
                Description: "Run your first video generation",
                Steps: []string{
                    "Run test inference with short prompt",
                    "Verify output video is created",
                    "Check video plays correctly",
                },
                Checkpoint: "Generated video file exists and plays",
            },
        },
    }
}

func (s *Sequence) productionProtocol() Protocol {
    return Protocol{
        ID:          "production",
        Name:        "Production Setup",
        Description: "Full production-ready configuration for 4×H100",
        Duration:    "2-3 hours",
        Phases: []Phase{
            {
                Number:      1,
                Name:        "Hardware Verification",
                Description: "Verify all hardware is correctly configured",
                Steps: []string{
                    "Run nvidia-smi to verify all 4 H100 GPUs detected",
                    "Check NVLink topology with nvidia-smi topo -m",
                    "Verify NV18 connections (full NVLink mesh)",
                    "Test inter-GPU bandwidth with p2pBandwidthLatencyTest",
                    "Verify system RAM (512GB+ recommended)",
                },
                Checkpoint: "All GPUs show NV18 connections, 800+ GB/s bandwidth",
            },
            {
                Number:      2,
                Name:        "Environment Configuration",
                Description: "Set up optimized environment",
                Steps: []string{
                    "Create dedicated virtual environment",
                    "Install CUDA 12.1+ toolkit",
                    "Configure NCCL for NVLink (NCCL_P2P_LEVEL=NVL)",
                    "Set up persistent environment variables",
                    "Configure system limits (ulimit, shared memory)",
                },
                Checkpoint: "Environment variables persist across sessions",
            },
            {
                Number:      3,
                Name:        "Model Installation",
                Description: "Install production model",
                Steps: []string{
                    "Download SkyReels-V2-DF-14B-540P model",
                    "Download T5-XXL text encoder",
                    "Verify model checksums",
                    "Configure model paths in sky config",
                    "Pre-load models to verify memory fit",
                },
                Checkpoint: "Models load without OOM errors",
            },
            {
                Number:      4,
                Name:        "Optimization Setup",
                Description: "Configure all optimizations",
                Steps: []string{
                    "Install Flash Attention 2",
                    "Install transformer-engine for FP8",
                    "Enable FP8 quantization",
                    "Configure TeaCache (thresh=0.3)",
                    "Enable torch.compile",
                    "Set up 4-way context parallelism",
                },
                Checkpoint: "Memory usage ~39GB/GPU with 41GB headroom",
            },
            {
                Number:      5,
                Name:        "Validation",
                Description: "Validate production configuration",
                Steps: []string{
                    "Run full benchmark suite",
                    "Generate test videos at all resolutions",
                    "Test continuity with sequential generation",
                    "Verify throughput meets targets (~45s for 97 frames)",
                    "Test error recovery and stability",
                },
                Checkpoint: "All benchmarks pass, stable over 10+ generations",
            },
        },
    }
}

func (s *Sequence) continuityProtocol() Protocol {
    return Protocol{
        ID:          "continuity",
        Name:        "Sequential Continuity",
        Description: "Set up video continuity for long-form generation",
        Duration:    "45-60 min",
        Phases: []Phase{
            {
                Number:      1,
                Name:        "Diffusion Forcing Setup",
                Description: "Configure autoregressive diffusion forcing",
                Steps: []string{
                    "Set ar_step=5 for async mode",
                    "Configure block stagger timing",
                    "Test basic continuity with 2-segment video",
                    "Verify no visual discontinuities",
                },
                Checkpoint: "Two segments blend smoothly",
            },
            {
                Number:      2,
                Name:        "Infinite Length Config",
                Description: "Enable infinite-length generation",
                Steps: []string{
                    "Configure segment overlap settings",
                    "Set up prompt continuation logic",
                    "Configure memory management for long sequences",
                    "Test 3+ segment generation",
                },
                Checkpoint: "Can generate 3+ segments without memory issues",
            },
            {
                Number:      3,
                Name:        "Quality Tuning",
                Description: "Optimize continuity quality",
                Steps: []string{
                    "Experiment with ar_step values (3, 5, 10)",
                    "Tune guidance scale for consistency",
                    "Test different prompt structures",
                    "Document optimal settings for use case",
                },
                Checkpoint: "Consistent style across all segments",
            },
            {
                Number:      4,
                Name:        "Pipeline Integration",
                Description: "Integrate into production pipeline",
                Steps: []string{
                    "Create segment queuing system",
                    "Implement prompt chaining logic",
                    "Set up output concatenation",
                    "Add quality monitoring",
                },
                Checkpoint: "End-to-end pipeline generates continuous video",
            },
        },
    }
}

func (s *Sequence) optimizationProtocol() Protocol {
    return Protocol{
        ID:          "optimization",
        Name:        "Performance Optimization",
        Description: "Maximize throughput and efficiency",
        Duration:    "1-2 hours",
        Phases: []Phase{
            {
                Number:      1,
                Name:        "Baseline Measurement",
                Description: "Establish performance baseline",
                Steps: []string{
                    "Run sky benchmark standard",
                    "Record baseline throughput (frames/sec)",
                    "Record baseline memory usage",
                    "Document current configuration",
                },
                Checkpoint: "Baseline metrics documented",
            },
            {
                Number:      2,
                Name:        "Memory Optimization",
                Description: "Reduce memory footprint",
                Steps: []string{
                    "Enable FP8 quantization if not already",
                    "Verify memory reduction (~50%)",
                    "Tune batch sizes for memory efficiency",
                    "Configure gradient checkpointing if needed",
                },
                Checkpoint: "Memory usage reduced, headroom increased",
            },
            {
                Number:      3,
                Name:        "Speed Optimization",
                Description: "Maximize generation speed",
                Steps: []string{
                    "Enable TeaCache with thresh=0.3",
                    "Enable torch.compile",
                    "Run warmup generation (first run slower)",
                    "Verify speedup in subsequent runs",
                    "Tune inference steps (try 25, 30, 35)",
                },
                Checkpoint: "20-30% speedup from baseline",
            },
            {
                Number:      4,
                Name:        "Parallelism Tuning",
                Description: "Optimize multi-GPU parallelism",
                Steps: []string{
                    "Test CP4 vs CP2+CFG2 configurations",
                    "Measure throughput for each",
                    "Identify optimal for your workload",
                    "Configure final parallelism strategy",
                },
                Checkpoint: "Best parallelism strategy identified",
            },
            {
                Number:      5,
                Name:        "Final Benchmark",
                Description: "Measure optimized performance",
                Steps: []string{
                    "Run full benchmark suite",
                    "Compare to baseline",
                    "Document improvements",
                    "Run stability test (10+ generations)",
                },
                Checkpoint: "Throughput improved, system stable",
            },
        },
    }
}
