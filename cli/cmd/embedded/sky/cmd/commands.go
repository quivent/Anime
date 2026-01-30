package cmd

import (
    "strconv"

    "github.com/sky-cli/sky/internal/analysis"
    "github.com/sky-cli/sky/internal/architecture"
    "github.com/sky-cli/sky/internal/benchmark"
    "github.com/sky-cli/sky/internal/config"
    "github.com/sky-cli/sky/internal/doctor"
    "github.com/sky-cli/sky/internal/enhance"
    "github.com/sky-cli/sky/internal/load"
    "github.com/sky-cli/sky/internal/metrics"
    "github.com/sky-cli/sky/internal/procedures"
    "github.com/sky-cli/sky/internal/reel"
    "github.com/sky-cli/sky/internal/sequence"
    "github.com/sky-cli/sky/internal/status"
    "github.com/sky-cli/sky/ui"
    "github.com/sky-cli/sky/internal/variants"
)

func runAnalysis(args []string) {
    a := analysis.New(cfg)

    if len(args) == 0 {
        a.PrintFull()
        return
    }

    switch args[0] {
    case "overview":
        a.PrintOverview()
    case "hardware":
        a.PrintHardware()
    case "model":
        a.PrintModel()
    case "parallelism":
        a.PrintParallelism()
    case "memory":
        a.PrintMemory()
    case "continuity":
        a.PrintContinuity()
    case "throughput":
        a.PrintThroughput()
    case "explain":
        if len(args) > 1 {
            a.PrintExplanation(args[1])
        } else {
            ui.PrintSuggestion("Topic required", []string{
                "Usage: sky analysis explain <topic>",
                "Topics: parallelism, diffusion-forcing, memory, teacache, nvlink",
            })
        }
    default:
        ui.PrintSuggestion("Unknown subcommand: "+args[0], []string{
            "Subcommands: overview, hardware, model, parallelism, memory, continuity, throughput, explain",
        })
    }
}

func runArchitecture(args []string) {
    a := architecture.New(cfg)

    if len(args) == 0 {
        a.PrintCurrent()
        return
    }

    switch args[0] {
    case "linear":
        a.PrintLinear()
    case "diagram":
        a.PrintDiagram()
    case "current":
        a.PrintCurrent()
    default:
        ui.PrintSuggestion("Unknown subcommand: "+args[0], []string{
            "Subcommands: linear, diagram, current",
        })
    }
}

func runProcedures(args []string) {
    p := procedures.New(cfg)

    if len(args) == 0 {
        p.PrintAll()
        return
    }

    switch args[0] {
    case "show":
        if len(args) > 1 {
            id, err := strconv.Atoi(args[1])
            if err != nil {
                ui.PrintSuggestion("Invalid procedure number", []string{
                    "Usage: sky procedures show <number>",
                    "Example: sky procedures show 3",
                })
                return
            }
            p.PrintProcedure(id)
        } else {
            ui.PrintSuggestion("Procedure number required", []string{
                "Usage: sky procedures show <number>",
                "Run 'sky procedures' to see all procedures",
            })
        }
    case "run":
        ui.PrintSuggestion("Procedure execution not implemented", []string{
            "Use 'sky procedures show <number>' to see commands",
            "Execute commands manually in your terminal",
        })
    default:
        // Try to parse as number
        id, err := strconv.Atoi(args[0])
        if err == nil {
            p.PrintProcedure(id)
        } else {
            ui.PrintSuggestion("Unknown subcommand: "+args[0], []string{
                "Subcommands: show, run",
                "Or use number directly: sky procedures 3",
            })
        }
    }
}

func runStatus(args []string) {
    s := status.New(cfg)

    if len(args) == 0 {
        s.Print()
        return
    }

    switch args[0] {
    case "full", "--full", "-f":
        s.PrintFull()
    case "load":
        s.PrintLoad()
    case "gpus", "gpu":
        s.PrintGPUs()
    default:
        ui.PrintSuggestion("Unknown subcommand: "+args[0], []string{
            "Subcommands: full, load, gpus",
            "  full  - Full status check including slow Python dependency checks",
            "  load  - Show GPU memory and utilization (MB/MB)",
            "  gpus  - Show detected GPUs information",
        })
    }
}

func runBenchmark(args []string) {
    b := benchmark.New(cfg)

    if len(args) == 0 {
        b.PrintInfo()
        return
    }

    switch args[0] {
    case "info":
        b.PrintInfo()
    case "estimate":
        b.PrintEstimate()
    case "compare":
        b.PrintComparison()
    case "run":
        name := "standard"
        if len(args) > 1 {
            name = args[1]
        }
        b.PrintSimulated(name)
    default:
        ui.PrintSuggestion("Unknown subcommand: "+args[0], []string{
            "Subcommands: info, estimate, compare, run",
        })
    }
}

func runVariants(args []string) {
    v := variants.New(cfg)

    if len(args) == 0 {
        v.PrintList()
        return
    }

    switch args[0] {
    case "list":
        v.PrintList()
    case "show":
        if len(args) > 1 {
            v.PrintDetails(args[1])
        } else {
            ui.PrintSuggestion("Variant name required", []string{
                "Usage: sky variants show <name>",
                "Example: sky variants show SkyReels-V2-DF-14B",
            })
        }
    case "recommend":
        v.PrintRecommendation()
    default:
        // Try as variant name
        v.PrintDetails(args[0])
    }
}

func runMetrics(args []string) {
    m := metrics.New(cfg)

    if len(args) == 0 {
        m.Print()
        return
    }

    switch args[0] {
    case "gpu":
        m.Print()
    case "performance", "perf":
        m.PrintPerformance()
    default:
        ui.PrintSuggestion("Unknown subcommand: "+args[0], []string{
            "Subcommands: gpu, performance",
        })
    }
}

func runEnhance(args []string) {
    e := enhance.New(cfg)

    if len(args) == 0 {
        e.Print()
        return
    }

    switch args[0] {
    case "list":
        e.Print()
    case "apply":
        target := "all"
        if len(args) > 1 {
            target = args[1]
        }
        e.PrintApply(target)
    case "profile":
        profile := ""
        if len(args) > 1 {
            profile = args[1]
        }
        e.PrintProfile(profile)
    default:
        ui.PrintSuggestion("Unknown subcommand: "+args[0], []string{
            "Subcommands: list, apply, profile",
        })
    }
}

func runSequence(args []string) {
    s := sequence.New(cfg)

    if len(args) == 0 {
        s.Print()
        return
    }

    switch args[0] {
    case "list":
        s.Print()
    case "show":
        if len(args) > 1 {
            s.PrintProtocol(args[1])
        } else {
            ui.PrintSuggestion("Protocol ID required", []string{
                "Usage: sky sequence show <id>",
                "IDs: quickstart, production, continuity, optimization",
            })
        }
    case "run":
        if len(args) > 1 {
            s.PrintRun(args[1])
        } else {
            ui.PrintSuggestion("Protocol ID required", []string{
                "Usage: sky sequence run <id>",
            })
        }
    default:
        // Try as protocol ID
        s.PrintProtocol(args[0])
    }
}

func runConfig(args []string) {
    if len(args) == 0 {
        ui.PrintHeader("Configuration")
        ui.PrintKeyValue("Config File", cfg.Paths.ConfigPath)
        ui.PrintSection("Current Settings")
        ui.PrintKeyValue("Model", cfg.Model.Variant)
        ui.PrintKeyValue("Precision", cfg.Model.Precision)
        ui.PrintKeyValue("GPUs", string(rune('0'+cfg.Hardware.GPUCount)))
        ui.PrintKeyValue("Context Parallel", string(rune('0'+cfg.Parallelism.ContextParallel)))
        ui.PrintKeyValue("ar_step", string(rune('0'+cfg.Diffusion.ARStep)))

        ui.PrintSection("Commands")
        ui.PrintList("sky config show - Show full configuration")
        ui.PrintList("sky config set <key> <value> - Set a value")
        ui.PrintList("sky config reset - Reset to defaults")
        return
    }

    switch args[0] {
    case "show":
        // Would print full config as YAML/JSON
        ui.PrintSuggestion("Full config display not implemented", []string{
            "View config file directly: cat ~/.sky/config.json",
        })
    case "set":
        ui.PrintSuggestion("Config set not implemented", []string{
            "Edit config file directly: vim ~/.sky/config.json",
        })
    case "reset":
        ui.PrintSuggestion("Config reset not implemented", []string{
            "Delete config file: rm ~/.sky/config.json",
            "Run 'sky init' to create new defaults",
        })
    }
}

func runInit() {
    ui.PrintHeader("Initialize Configuration")

    // Save default config
    defaultCfg := config.DefaultConfig()
    if err := defaultCfg.Save(); err != nil {
        ui.PrintStatus("error", "Failed to save config: "+err.Error())
        return
    }

    ui.PrintStatus("success", "Configuration file created")
    ui.PrintKeyValue("Location", config.ConfigPath())

    ui.PrintSection("Default Settings")
    ui.PrintKeyValue("Model", defaultCfg.Model.Variant)
    ui.PrintKeyValue("GPUs", "4× H100-SXM5")
    ui.PrintKeyValue("Parallelism", "CP4 (4-way context parallel)")
    ui.PrintKeyValue("Precision", defaultCfg.Model.Precision)

    ui.PrintSection("Next Steps")
    ui.PrintList("Run 'sky status' to check system readiness")
    ui.PrintList("Run 'sky procedures' to see setup steps")
    ui.PrintList("Run 'sky config' to view/modify settings")
}

func runReel(args []string) {
    reel.Handle(args)
}

func runDoctor(args []string) {
    d := doctor.New(cfg)
    d.Run()
}

func runReload(args []string) {
    d := doctor.New(cfg)
    d.Reload()
}

func runLoad(args []string) {
    l := load.New(cfg)

    if len(args) == 0 {
        l.RunWizard()
        return
    }

    switch args[0] {
    case "quick", "-q":
        l.QuickLoad()
    case "wizard", "-w":
        l.RunWizard()
    default:
        ui.PrintSuggestion("Unknown subcommand: "+args[0], []string{
            "Subcommands:",
            "  (none)  - Run interactive wizard",
            "  quick   - Quick load with recommended defaults",
            "  wizard  - Run interactive wizard",
        })
    }
}
