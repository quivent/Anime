use super::types::*;
use serde_json::json;

/// Get built-in workflow templates
pub fn get_builtin_workflows() -> Vec<ComfyUIWorkflow> {
    vec![
        // Text-to-Image workflow
        ComfyUIWorkflow {
            id: "txt2img_flux".to_string(),
            name: "Text to Image (FLUX)".to_string(),
            description: "Generate high-quality images from text prompts using FLUX model".to_string(),
            category: "image".to_string(),
            icon: "🎨".to_string(),
            thumbnail: None,
            workflow_json: create_flux_txt2img_workflow(),
            parameters: vec![
                WorkflowParameter {
                    id: "prompt".to_string(),
                    name: "Prompt".to_string(),
                    param_type: "text".to_string(),
                    description: "Describe the image you want to generate".to_string(),
                    required: true,
                    default_value: Some(json!("A beautiful landscape with mountains and a lake")),
                    options: None,
                    min: None,
                    max: None,
                    node_id: Some("6".to_string()),
                    field_name: Some("text".to_string()),
                },
                WorkflowParameter {
                    id: "negative_prompt".to_string(),
                    name: "Negative Prompt".to_string(),
                    param_type: "text".to_string(),
                    description: "What to avoid in the image".to_string(),
                    required: false,
                    default_value: Some(json!("blurry, low quality, distorted")),
                    options: None,
                    min: None,
                    max: None,
                    node_id: Some("7".to_string()),
                    field_name: Some("text".to_string()),
                },
                WorkflowParameter {
                    id: "width".to_string(),
                    name: "Width".to_string(),
                    param_type: "number".to_string(),
                    description: "Image width in pixels".to_string(),
                    required: true,
                    default_value: Some(json!(1024)),
                    options: None,
                    min: Some(512.0),
                    max: Some(2048.0),
                    node_id: Some("5".to_string()),
                    field_name: Some("width".to_string()),
                },
                WorkflowParameter {
                    id: "height".to_string(),
                    name: "Height".to_string(),
                    param_type: "number".to_string(),
                    description: "Image height in pixels".to_string(),
                    required: true,
                    default_value: Some(json!(1024)),
                    options: None,
                    min: Some(512.0),
                    max: Some(2048.0),
                    node_id: Some("5".to_string()),
                    field_name: Some("height".to_string()),
                },
                WorkflowParameter {
                    id: "steps".to_string(),
                    name: "Steps".to_string(),
                    param_type: "number".to_string(),
                    description: "Number of denoising steps".to_string(),
                    required: true,
                    default_value: Some(json!(20)),
                    options: None,
                    min: Some(1.0),
                    max: Some(150.0),
                    node_id: Some("3".to_string()),
                    field_name: Some("steps".to_string()),
                },
                WorkflowParameter {
                    id: "cfg".to_string(),
                    name: "CFG Scale".to_string(),
                    param_type: "number".to_string(),
                    description: "Classifier-free guidance scale".to_string(),
                    required: true,
                    default_value: Some(json!(7.0)),
                    options: None,
                    min: Some(1.0),
                    max: Some(20.0),
                    node_id: Some("3".to_string()),
                    field_name: Some("cfg".to_string()),
                },
            ],
            outputs: vec![
                WorkflowOutput {
                    output_type: "image".to_string(),
                    name: "Generated Image".to_string(),
                    format: "png".to_string(),
                },
            ],
        },
        // Image-to-Image workflow
        ComfyUIWorkflow {
            id: "img2img".to_string(),
            name: "Image to Image".to_string(),
            description: "Transform existing images with AI guidance".to_string(),
            category: "image".to_string(),
            icon: "🖼️".to_string(),
            thumbnail: None,
            workflow_json: create_img2img_workflow(),
            parameters: vec![
                WorkflowParameter {
                    id: "input_image".to_string(),
                    name: "Input Image".to_string(),
                    param_type: "image".to_string(),
                    description: "The image to transform".to_string(),
                    required: true,
                    default_value: None,
                    options: None,
                    min: None,
                    max: None,
                    node_id: Some("1".to_string()),
                    field_name: Some("image".to_string()),
                },
                WorkflowParameter {
                    id: "prompt".to_string(),
                    name: "Prompt".to_string(),
                    param_type: "text".to_string(),
                    description: "Describe the transformation".to_string(),
                    required: true,
                    default_value: Some(json!("enhance details, vibrant colors")),
                    options: None,
                    min: None,
                    max: None,
                    node_id: Some("6".to_string()),
                    field_name: Some("text".to_string()),
                },
                WorkflowParameter {
                    id: "denoise".to_string(),
                    name: "Denoise Strength".to_string(),
                    param_type: "number".to_string(),
                    description: "How much to change the image (0-1)".to_string(),
                    required: true,
                    default_value: Some(json!(0.75)),
                    options: None,
                    min: Some(0.0),
                    max: Some(1.0),
                    node_id: Some("3".to_string()),
                    field_name: Some("denoise".to_string()),
                },
            ],
            outputs: vec![
                WorkflowOutput {
                    output_type: "image".to_string(),
                    name: "Transformed Image".to_string(),
                    format: "png".to_string(),
                },
            ],
        },
        // Upscaling workflow
        ComfyUIWorkflow {
            id: "upscale".to_string(),
            name: "Image Upscaling".to_string(),
            description: "Upscale images to higher resolution with AI enhancement".to_string(),
            category: "upscaling".to_string(),
            icon: "⬆️".to_string(),
            thumbnail: None,
            workflow_json: create_upscale_workflow(),
            parameters: vec![
                WorkflowParameter {
                    id: "input_image".to_string(),
                    name: "Input Image".to_string(),
                    param_type: "image".to_string(),
                    description: "The image to upscale".to_string(),
                    required: true,
                    default_value: None,
                    options: None,
                    min: None,
                    max: None,
                    node_id: Some("1".to_string()),
                    field_name: Some("image".to_string()),
                },
                WorkflowParameter {
                    id: "upscale_factor".to_string(),
                    name: "Upscale Factor".to_string(),
                    param_type: "select".to_string(),
                    description: "How much to upscale".to_string(),
                    required: true,
                    default_value: Some(json!("2x")),
                    options: Some(vec!["2x".to_string(), "4x".to_string()]),
                    min: None,
                    max: None,
                    node_id: Some("2".to_string()),
                    field_name: Some("upscale_method".to_string()),
                },
            ],
            outputs: vec![
                WorkflowOutput {
                    output_type: "image".to_string(),
                    name: "Upscaled Image".to_string(),
                    format: "png".to_string(),
                },
            ],
        },
        // Video Generation workflow
        ComfyUIWorkflow {
            id: "video_gen".to_string(),
            name: "Text to Video".to_string(),
            description: "Generate videos from text prompts using video models".to_string(),
            category: "video".to_string(),
            icon: "🎬".to_string(),
            thumbnail: None,
            workflow_json: create_video_workflow(),
            parameters: vec![
                WorkflowParameter {
                    id: "prompt".to_string(),
                    name: "Prompt".to_string(),
                    param_type: "text".to_string(),
                    description: "Describe the video you want to generate".to_string(),
                    required: true,
                    default_value: Some(json!("A beautiful sunset over the ocean")),
                    options: None,
                    min: None,
                    max: None,
                    node_id: Some("6".to_string()),
                    field_name: Some("text".to_string()),
                },
                WorkflowParameter {
                    id: "frames".to_string(),
                    name: "Number of Frames".to_string(),
                    param_type: "number".to_string(),
                    description: "Video length in frames".to_string(),
                    required: true,
                    default_value: Some(json!(24)),
                    options: None,
                    min: Some(8.0),
                    max: Some(120.0),
                    node_id: Some("3".to_string()),
                    field_name: Some("frames".to_string()),
                },
            ],
            outputs: vec![
                WorkflowOutput {
                    output_type: "video".to_string(),
                    name: "Generated Video".to_string(),
                    format: "mp4".to_string(),
                },
            ],
        },
    ]
}

/// Create a basic FLUX text-to-image workflow
fn create_flux_txt2img_workflow() -> String {
    let workflow = json!({
        "3": {
            "inputs": {
                "seed": 42,
                "steps": 20,
                "cfg": 7.0,
                "sampler_name": "euler",
                "scheduler": "normal",
                "denoise": 1,
                "model": ["4", 0],
                "positive": ["6", 0],
                "negative": ["7", 0],
                "latent_image": ["5", 0]
            },
            "class_type": "KSampler"
        },
        "4": {
            "inputs": {
                "ckpt_name": "flux1-dev.safetensors"
            },
            "class_type": "CheckpointLoaderSimple"
        },
        "5": {
            "inputs": {
                "width": 1024,
                "height": 1024,
                "batch_size": 1
            },
            "class_type": "EmptyLatentImage"
        },
        "6": {
            "inputs": {
                "text": "A beautiful landscape",
                "clip": ["4", 1]
            },
            "class_type": "CLIPTextEncode"
        },
        "7": {
            "inputs": {
                "text": "blurry, low quality",
                "clip": ["4", 1]
            },
            "class_type": "CLIPTextEncode"
        },
        "8": {
            "inputs": {
                "samples": ["3", 0],
                "vae": ["4", 2]
            },
            "class_type": "VAEDecode"
        },
        "9": {
            "inputs": {
                "filename_prefix": "ComfyUI",
                "images": ["8", 0]
            },
            "class_type": "SaveImage"
        }
    });

    serde_json::to_string(&workflow).unwrap()
}

/// Create an image-to-image workflow
fn create_img2img_workflow() -> String {
    let workflow = json!({
        "1": {
            "inputs": {
                "image": "input.png",
                "upload": "image"
            },
            "class_type": "LoadImage"
        },
        "2": {
            "inputs": {
                "samples": ["1", 0],
                "vae": ["4", 2]
            },
            "class_type": "VAEEncode"
        },
        "3": {
            "inputs": {
                "seed": 42,
                "steps": 20,
                "cfg": 7.0,
                "sampler_name": "euler",
                "scheduler": "normal",
                "denoise": 0.75,
                "model": ["4", 0],
                "positive": ["6", 0],
                "negative": ["7", 0],
                "latent_image": ["2", 0]
            },
            "class_type": "KSampler"
        },
        "4": {
            "inputs": {
                "ckpt_name": "flux1-dev.safetensors"
            },
            "class_type": "CheckpointLoaderSimple"
        },
        "6": {
            "inputs": {
                "text": "enhance details",
                "clip": ["4", 1]
            },
            "class_type": "CLIPTextEncode"
        },
        "7": {
            "inputs": {
                "text": "blurry, low quality",
                "clip": ["4", 1]
            },
            "class_type": "CLIPTextEncode"
        },
        "8": {
            "inputs": {
                "samples": ["3", 0],
                "vae": ["4", 2]
            },
            "class_type": "VAEDecode"
        },
        "9": {
            "inputs": {
                "filename_prefix": "img2img",
                "images": ["8", 0]
            },
            "class_type": "SaveImage"
        }
    });

    serde_json::to_string(&workflow).unwrap()
}

/// Create an upscaling workflow
fn create_upscale_workflow() -> String {
    let workflow = json!({
        "1": {
            "inputs": {
                "image": "input.png",
                "upload": "image"
            },
            "class_type": "LoadImage"
        },
        "2": {
            "inputs": {
                "upscale_method": "nearest-exact",
                "scale_by": 2.0,
                "image": ["1", 0]
            },
            "class_type": "ImageScaleBy"
        },
        "3": {
            "inputs": {
                "filename_prefix": "upscaled",
                "images": ["2", 0]
            },
            "class_type": "SaveImage"
        }
    });

    serde_json::to_string(&workflow).unwrap()
}

/// Create a video generation workflow
fn create_video_workflow() -> String {
    let workflow = json!({
        "3": {
            "inputs": {
                "frames": 24,
                "width": 512,
                "height": 512,
                "batch_size": 1
            },
            "class_type": "EmptyVideoLatent"
        },
        "4": {
            "inputs": {
                "ckpt_name": "mochi_preview_dit_fp8_e4m3fn.safetensors"
            },
            "class_type": "CheckpointLoaderSimple"
        },
        "6": {
            "inputs": {
                "text": "A beautiful sunset",
                "clip": ["4", 1]
            },
            "class_type": "CLIPTextEncode"
        },
        "7": {
            "inputs": {
                "seed": 42,
                "steps": 25,
                "cfg": 7.0,
                "sampler_name": "euler",
                "scheduler": "normal",
                "denoise": 1,
                "model": ["4", 0],
                "positive": ["6", 0],
                "latent_image": ["3", 0]
            },
            "class_type": "KSampler"
        },
        "8": {
            "inputs": {
                "samples": ["7", 0],
                "vae": ["4", 2]
            },
            "class_type": "VAEDecode"
        },
        "9": {
            "inputs": {
                "filename_prefix": "video",
                "images": ["8", 0]
            },
            "class_type": "SaveVideo"
        }
    });

    serde_json::to_string(&workflow).unwrap()
}
