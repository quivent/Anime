package config

// ModelVariant represents a SkyReel model variant
type ModelVariant struct {
    Name        string
    Parameters  string
    Resolution  string
    VRAM        string
    Speed       string
    Quality     string
    Description string
}

// GetVariants returns all available model variants
func GetVariants() []ModelVariant {
    return []ModelVariant{
        {
            Name:        "SkyReels-V2-DF-1.3B-540P",
            Parameters:  "1.3B",
            Resolution:  "540P (544×960)",
            VRAM:        "~15GB",
            Speed:       "Fast",
            Quality:     "Good",
            Description: "Lightweight model for resource-constrained environments",
        },
        {
            Name:        "SkyReels-V2-DF-5B-540P",
            Parameters:  "5B",
            Resolution:  "540P (544×960)",
            VRAM:        "~25GB",
            Speed:       "Medium",
            Quality:     "Better",
            Description: "Balanced model (coming soon)",
        },
        {
            Name:        "SkyReels-V2-DF-14B-540P",
            Parameters:  "14B",
            Resolution:  "540P (544×960)",
            VRAM:        "~51GB",
            Speed:       "Slower",
            Quality:     "Best",
            Description: "High-quality generation for production use",
        },
        {
            Name:        "SkyReels-V2-DF-14B-720P",
            Parameters:  "14B",
            Resolution:  "720P (720×1280)",
            VRAM:        "~65GB",
            Speed:       "Slowest",
            Quality:     "Best",
            Description: "Maximum quality at higher resolution",
        },
        {
            Name:        "SkyReels-V1-HunyuanVideo",
            Parameters:  "13B",
            Resolution:  "544×960",
            VRAM:        "~48GB",
            Speed:       "Medium",
            Quality:     "Good",
            Description: "V1 model based on HunyuanVideo foundation",
        },
    }
}

// ParallelismStrategy represents a parallelism configuration
type ParallelismStrategy struct {
    Name            string
    ContextParallel int
    CFGParallel     int
    Description     string
    BestFor         string
}

// GetParallelismStrategies returns available parallelism strategies
func GetParallelismStrategies() []ParallelismStrategy {
    return []ParallelismStrategy{
        {
            Name:            "CP4",
            ContextParallel: 4,
            CFGParallel:     1,
            Description:     "4-way context parallel, full frame distribution",
            BestFor:         "Maximum throughput, long sequences",
        },
        {
            Name:            "CP2+CFG2",
            ContextParallel: 2,
            CFGParallel:     2,
            Description:     "Hybrid: 2-way context + 2-way CFG parallel",
            BestFor:         "Balanced latency and quality",
        },
        {
            Name:            "CFG4",
            ContextParallel: 1,
            CFGParallel:     4,
            Description:     "4-way CFG parallel only",
            BestFor:         "Short sequences, maximum CFG flexibility",
        },
    }
}
