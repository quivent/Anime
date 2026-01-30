package workflow

import (
	"fmt"
	"time"
)

// Workflow represents an AI workflow that can be applied to a collection
type Workflow struct {
	ID          string
	Name        string
	Description string
	Type        string   // image, video, mixed
	Models      []string // Required models (wan2, mochi, comfyui, etc.)
	Parameters  []Parameter
	EstimatedTime time.Duration
}

// Parameter represents a configurable parameter for a workflow
type Parameter struct {
	Name        string
	Description string
	Type        string // string, int, float, bool, choice
	Default     interface{}
	Choices     []string // For choice type
	Required    bool
}

// WorkflowExecution represents a running or completed workflow
type WorkflowExecution struct {
	ID           string
	WorkflowID   string
	Collection   string
	Status       string // pending, running, completed, failed
	Progress     int    // 0-100
	CurrentFile  string
	TotalFiles   int
	ProcessedFiles int
	StartTime    time.Time
	EndTime      *time.Time
	Error        string
	Parameters   map[string]interface{}
}

// Available workflows
var Workflows = map[string]Workflow{
	"animate": {
		ID:          "animate",
		Name:        "Animate Images",
		Description: "Transform static images into dynamic videos with AI-powered motion using Wan2.2 or Mochi-1",
		Type:        "image",
		Models:      []string{"wan2", "mochi"},
		Parameters: []Parameter{
			{
				Name:        "model",
				Description: "Video generation model to use",
				Type:        "choice",
				Choices:     []string{"wan2", "mochi"},
				Default:     "wan2",
				Required:    true,
			},
			{
				Name:        "duration",
				Description: "Video duration in seconds",
				Type:        "int",
				Default:     3,
				Required:    false,
			},
			{
				Name:        "fps",
				Description: "Frames per second",
				Type:        "int",
				Default:     24,
				Required:    false,
			},
			{
				Name:        "output_dir",
				Description: "Output directory for videos",
				Type:        "string",
				Default:     "./output",
				Required:    false,
			},
		},
		EstimatedTime: 2 * time.Minute, // per image
	},
	"upscale": {
		ID:          "upscale",
		Name:        "AI Upscale 4K/8K",
		Description: "Upscale images to 4K or 8K resolution using ESRGAN/RealESRGAN AI models",
		Type:        "image",
		Models:      []string{"comfyui"},
		Parameters: []Parameter{
			{
				Name:        "scale",
				Description: "Upscale factor",
				Type:        "choice",
				Choices:     []string{"2x", "4x"},
				Default:     "2x",
				Required:    true,
			},
			{
				Name:        "model",
				Description: "Upscaling model",
				Type:        "choice",
				Choices:     []string{"esrgan", "realesrgan"},
				Default:     "realesrgan",
				Required:    true,
			},
			{
				Name:        "output_dir",
				Description: "Output directory",
				Type:        "string",
				Default:     "./upscaled",
				Required:    false,
			},
		},
		EstimatedTime: 30 * time.Second, // per image
	},
	"style-transfer": {
		ID:          "style-transfer",
		Name:        "Style Transfer",
		Description: "Apply artistic style to images using ComfyUI",
		Type:        "image",
		Models:      []string{"comfyui"},
		Parameters: []Parameter{
			{
				Name:        "style_image",
				Description: "Path to style reference image",
				Type:        "string",
				Required:    true,
			},
			{
				Name:        "strength",
				Description: "Style transfer strength (0.0-1.0)",
				Type:        "float",
				Default:     0.7,
				Required:    false,
			},
			{
				Name:        "output_dir",
				Description: "Output directory",
				Type:        "string",
				Default:     "./styled",
				Required:    false,
			},
		},
		EstimatedTime: 45 * time.Second, // per image
	},
	"batch-inference": {
		ID:          "batch-inference",
		Name:        "Batch Inference",
		Description: "Run LLM vision inference on images using Ollama",
		Type:        "image",
		Models:      []string{"ollama"},
		Parameters: []Parameter{
			{
				Name:        "model",
				Description: "Vision model to use",
				Type:        "choice",
				Choices:     []string{"llava", "bakllava"},
				Default:     "llava",
				Required:    true,
			},
			{
				Name:        "prompt",
				Description: "Prompt for the model",
				Type:        "string",
				Default:     "Describe this image in detail",
				Required:    true,
			},
			{
				Name:        "output_file",
				Description: "Output JSON file",
				Type:        "string",
				Default:     "./inference_results.json",
				Required:    false,
			},
		},
		EstimatedTime: 10 * time.Second, // per image
	},
	"video-enhance": {
		ID:          "video-enhance",
		Name:        "Video Enhancement",
		Description: "Enhance video quality and upscale",
		Type:        "video",
		Models:      []string{"comfyui"},
		Parameters: []Parameter{
			{
				Name:        "upscale",
				Description: "Upscale factor",
				Type:        "choice",
				Choices:     []string{"none", "2x", "4x"},
				Default:     "2x",
				Required:    false,
			},
			{
				Name:        "denoise",
				Description: "Apply denoising",
				Type:        "bool",
				Default:     true,
				Required:    false,
			},
			{
				Name:        "output_dir",
				Description: "Output directory",
				Type:        "string",
				Default:     "./enhanced",
				Required:    false,
			},
		},
		EstimatedTime: 5 * time.Minute, // per video
	},
	"text2img": {
		ID:          "text2img",
		Name:        "Generate from Prompts",
		Description: "Generate high-quality images from text prompts using Stable Diffusion XL",
		Type:        "mixed",
		Models:      []string{"comfyui"},
		Parameters: []Parameter{
			{
				Name:        "prompts_file",
				Description: "File containing prompts (one per line)",
				Type:        "string",
				Required:    true,
			},
			{
				Name:        "model",
				Description: "Stable Diffusion model checkpoint",
				Type:        "string",
				Default:     "sd_xl_base_1.0",
				Required:    false,
			},
			{
				Name:        "width",
				Description: "Image width",
				Type:        "int",
				Default:     1024,
				Required:    false,
			},
			{
				Name:        "height",
				Description: "Image height",
				Type:        "int",
				Default:     1024,
				Required:    false,
			},
			{
				Name:        "steps",
				Description: "Number of sampling steps",
				Type:        "int",
				Default:     30,
				Required:    false,
			},
			{
				Name:        "output_dir",
				Description: "Output directory",
				Type:        "string",
				Default:     "./generated",
				Required:    false,
			},
		},
		EstimatedTime: 30 * time.Second, // per image
	},
	"depth-3d": {
		ID:          "depth-3d",
		Name:        "3D Depth Map Generation",
		Description: "Generate 3D depth maps from 2D images for 3D reconstruction and rendering",
		Type:        "image",
		Models:      []string{"comfyui"},
		Parameters: []Parameter{
			{
				Name:        "model",
				Description: "Depth estimation model",
				Type:        "choice",
				Choices:     []string{"midas", "depth-anything"},
				Default:     "depth-anything",
				Required:    true,
			},
			{
				Name:        "output_format",
				Description: "Output format",
				Type:        "choice",
				Choices:     []string{"png", "exr", "obj"},
				Default:     "png",
				Required:    false,
			},
			{
				Name:        "output_dir",
				Description: "Output directory",
				Type:        "string",
				Default:     "./depth-maps",
				Required:    false,
			},
		},
		EstimatedTime: 20 * time.Second, // per image
	},
	"3d-render": {
		ID:          "3d-render",
		Name:        "3D Model Rendering",
		Description: "Render 3D models from multiple angles with photorealistic lighting",
		Type:        "mixed",
		Models:      []string{"comfyui"},
		Parameters: []Parameter{
			{
				Name:        "angles",
				Description: "Number of camera angles",
				Type:        "int",
				Default:     8,
				Required:    false,
			},
			{
				Name:        "resolution",
				Description: "Output resolution",
				Type:        "choice",
				Choices:     []string{"1K", "2K", "4K", "8K"},
				Default:     "4K",
				Required:    false,
			},
			{
				Name:        "lighting",
				Description: "Lighting preset",
				Type:        "choice",
				Choices:     []string{"studio", "outdoor", "dramatic", "natural"},
				Default:     "studio",
				Required:    false,
			},
			{
				Name:        "output_dir",
				Description: "Output directory",
				Type:        "string",
				Default:     "./renders",
				Required:    false,
			},
		},
		EstimatedTime: 45 * time.Second, // per render
	},
	"motion-track": {
		ID:          "motion-track",
		Name:        "Motion Tracking & Stabilization",
		Description: "Track motion in videos and apply professional stabilization",
		Type:        "video",
		Models:      []string{"comfyui"},
		Parameters: []Parameter{
			{
				Name:        "stabilize",
				Description: "Apply stabilization",
				Type:        "bool",
				Default:     true,
				Required:    false,
			},
			{
				Name:        "track_points",
				Description: "Number of tracking points",
				Type:        "int",
				Default:     100,
				Required:    false,
			},
			{
				Name:        "output_dir",
				Description: "Output directory",
				Type:        "string",
				Default:     "./tracked",
				Required:    false,
			},
		},
		EstimatedTime: 3 * time.Minute, // per video
	},
	"interpolate": {
		ID:          "interpolate",
		Name:        "Frame Interpolation (Slowmo)",
		Description: "Create smooth slow-motion by interpolating frames to 60/120fps",
		Type:        "video",
		Models:      []string{"comfyui"},
		Parameters: []Parameter{
			{
				Name:        "target_fps",
				Description: "Target frame rate",
				Type:        "choice",
				Choices:     []string{"60", "120", "240"},
				Default:     "60",
				Required:    true,
			},
			{
				Name:        "quality",
				Description: "Interpolation quality",
				Type:        "choice",
				Choices:     []string{"fast", "balanced", "high"},
				Default:     "balanced",
				Required:    false,
			},
			{
				Name:        "output_dir",
				Description: "Output directory",
				Type:        "string",
				Default:     "./interpolated",
				Required:    false,
			},
		},
		EstimatedTime: 4 * time.Minute, // per video
	},
	"segment": {
		ID:          "segment",
		Name:        "AI Background Removal",
		Description: "Segment and remove backgrounds with AI precision (SAM/RMBG)",
		Type:        "image",
		Models:      []string{"comfyui"},
		Parameters: []Parameter{
			{
				Name:        "model",
				Description: "Segmentation model",
				Type:        "choice",
				Choices:     []string{"sam", "rembg", "u2net"},
				Default:     "rembg",
				Required:    true,
			},
			{
				Name:        "output_format",
				Description: "Output format",
				Type:        "choice",
				Choices:     []string{"png", "webp"},
				Default:     "png",
				Required:    false,
			},
			{
				Name:        "output_dir",
				Description: "Output directory",
				Type:        "string",
				Default:     "./segmented",
				Required:    false,
			},
		},
		EstimatedTime: 15 * time.Second, // per image
	},
	"colorize": {
		ID:          "colorize",
		Name:        "AI Colorization",
		Description: "Colorize black & white images/videos using deep learning",
		Type:        "mixed",
		Models:      []string{"comfyui"},
		Parameters: []Parameter{
			{
				Name:        "model",
				Description: "Colorization model",
				Type:        "choice",
				Choices:     []string{"deoldify", "colorful"},
				Default:     "deoldify",
				Required:    true,
			},
			{
				Name:        "render_factor",
				Description: "Quality factor (higher = better)",
				Type:        "int",
				Default:     35,
				Required:    false,
			},
			{
				Name:        "output_dir",
				Description: "Output directory",
				Type:        "string",
				Default:     "./colorized",
				Required:    false,
			},
		},
		EstimatedTime: 40 * time.Second, // per image
	},
	"face-restore": {
		ID:          "face-restore",
		Name:        "Face Restoration",
		Description: "Restore and enhance faces in old or low-quality photos",
		Type:        "image",
		Models:      []string{"comfyui"},
		Parameters: []Parameter{
			{
				Name:        "model",
				Description: "Face restoration model",
				Type:        "choice",
				Choices:     []string{"gfpgan", "codeformer"},
				Default:     "codeformer",
				Required:    true,
			},
			{
				Name:        "fidelity",
				Description: "Fidelity weight (0.0-1.0)",
				Type:        "float",
				Default:     0.7,
				Required:    false,
			},
			{
				Name:        "output_dir",
				Description: "Output directory",
				Type:        "string",
				Default:     "./restored",
				Required:    false,
			},
		},
		EstimatedTime: 25 * time.Second, // per image
	},
}

// GetWorkflowsForType returns workflows applicable to a collection type
func GetWorkflowsForType(collectionType string) []Workflow {
	var applicable []Workflow
	for _, workflow := range Workflows {
		if workflow.Type == collectionType || workflow.Type == "mixed" || collectionType == "mixed" {
			applicable = append(applicable, workflow)
		}
	}
	return applicable
}

// GetWorkflow returns a workflow by ID
func GetWorkflow(id string) (*Workflow, error) {
	workflow, exists := Workflows[id]
	if !exists {
		return nil, fmt.Errorf("workflow %s not found", id)
	}
	return &workflow, nil
}

// ValidateParameters validates workflow parameters
func ValidateParameters(workflow *Workflow, params map[string]interface{}) error {
	for _, param := range workflow.Parameters {
		value, provided := params[param.Name]

		// Check required parameters
		if param.Required && !provided {
			return fmt.Errorf("required parameter %s not provided", param.Name)
		}

		// Type validation if value provided
		if provided {
			switch param.Type {
			case "choice":
				strVal, ok := value.(string)
				if !ok {
					return fmt.Errorf("parameter %s must be a string", param.Name)
				}
				validChoice := false
				for _, choice := range param.Choices {
					if strVal == choice {
						validChoice = true
						break
					}
				}
				if !validChoice {
					return fmt.Errorf("parameter %s must be one of: %v", param.Name, param.Choices)
				}
			case "int":
				if _, ok := value.(int); !ok {
					return fmt.Errorf("parameter %s must be an integer", param.Name)
				}
			case "float":
				if _, ok := value.(float64); !ok {
					return fmt.Errorf("parameter %s must be a float", param.Name)
				}
			case "bool":
				if _, ok := value.(bool); !ok {
					return fmt.Errorf("parameter %s must be a boolean", param.Name)
				}
			}
		}
	}
	return nil
}
