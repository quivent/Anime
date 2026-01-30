package cmd

import (
    "fmt"
    "os"

    "github.com/sky-cli/sky/internal/config"
    "github.com/sky-cli/sky/ui"
)

var cfg *config.Config

// Execute runs the CLI
func Execute() error {
    // Load config
    var err error
    cfg, err = config.Load()
    if err != nil {
        cfg = config.DefaultConfig()
    }

    // Parse args
    if len(os.Args) < 2 {
        printMainHelp()
        return nil
    }

    command := os.Args[1]
    args := os.Args[2:]

    switch command {
    case "help", "-h", "--help":
        printMainHelp()
    case "version", "-v", "--version":
        printVersion()
    case "analysis":
        runAnalysis(args)
    case "architecture", "arch":
        runArchitecture(args)
    case "procedures", "proc":
        runProcedures(args)
    case "status":
        runStatus(args)
    case "benchmark", "bench":
        runBenchmark(args)
    case "variants":
        runVariants(args)
    case "metrics":
        runMetrics(args)
    case "enhance":
        runEnhance(args)
    case "sequence", "seq":
        runSequence(args)
    case "config":
        runConfig(args)
    case "init":
        runInit()
    case "reel":
        runReel(args)
    case "doctor":
        runDoctor(args)
    case "reload":
        runReload(args)
    case "load":
        runLoad(args)
    default:
        ui.PrintSuggestion("Unknown command: "+command, []string{
            "Run 'sky help' to see available commands",
            "Common commands: analysis, architecture, status, benchmark",
        })
    }

    return nil
}

func printMainHelp() {
    ui.PrintHeader("Sky CLI - SkyReel Video Generation Manager")

    fmt.Printf("  %s\n\n", ui.Muted("High-throughput sequential video generation for 4×H100"))

    ui.PrintSection("Usage")
    fmt.Printf("  %s <command> [subcommand] [options]\n\n", ui.Value("sky"))

    ui.PrintSection("Commands")

    commands := []struct {
        name string
        desc string
    }{
        {"load", "Interactive wizard to load models on GPUs"},
        {"doctor", "Check if models are loaded on GPUs"},
        {"reload", "Quick reload models with defaults"},
        {"status", "Check system status and readiness"},
        {"analysis", "View complete architecture analysis"},
        {"architecture", "Display architecture diagrams and config"},
        {"procedures", "Setup procedures and installation steps"},
        {"sequence", "Implementation sequence protocols"},
        {"benchmark", "Run performance benchmarks"},
        {"variants", "List and configure model variants"},
        {"metrics", "Display current system metrics"},
        {"enhance", "Get optimization suggestions"},
        {"config", "View and modify configuration"},
        {"init", "Initialize configuration file"},
        {"reel", "Run SkyReel video generation CLI"},
    }

    for _, cmd := range commands {
        fmt.Printf("    %s%-14s%s %s\n", ui.BrightCyan, cmd.name, ui.Reset, ui.Muted(cmd.desc))
    }

    ui.PrintSection("Examples")
    fmt.Printf("    %s sky analysis%s                 %s\n", ui.Muted("$"), ui.Reset, ui.Muted("# View full analysis"))
    fmt.Printf("    %s sky analysis explain memory%s  %s\n", ui.Muted("$"), ui.Reset, ui.Muted("# Deep dive on memory"))
    fmt.Printf("    %s sky architecture diagram%s     %s\n", ui.Muted("$"), ui.Reset, ui.Muted("# Show ASCII diagram"))
    fmt.Printf("    %s sky procedures show 3%s        %s\n", ui.Muted("$"), ui.Reset, ui.Muted("# Show procedure #3"))
    fmt.Printf("    %s sky status%s                   %s\n", ui.Muted("$"), ui.Reset, ui.Muted("# Check system status"))
    fmt.Printf("    %s sky benchmark estimate%s       %s\n", ui.Muted("$"), ui.Reset, ui.Muted("# Performance estimates"))

    ui.PrintSection("Quick Start")
    fmt.Printf("    1. Run %s to check your system\n", ui.Value("sky status"))
    fmt.Printf("    2. Run %s to see setup steps\n", ui.Value("sky procedures"))
    fmt.Printf("    3. Run %s for architecture overview\n", ui.Value("sky analysis"))

    fmt.Println()
}

func printVersion() {
    fmt.Printf("%s %s\n", ui.Title("Sky CLI"), ui.Value("v1.0.0"))
    fmt.Printf("%s\n", ui.Muted("SkyReel Video Generation Manager for 4×H100"))
    fmt.Printf("%s\n", ui.Muted("Architecture: xDiT Context Parallel + Diffusion Forcing"))
}
