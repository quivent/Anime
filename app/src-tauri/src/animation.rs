use anyhow::{anyhow, Result};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use tauri::{AppHandle, Emitter};

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "lowercase")]
pub enum AnimationModel {
    AnimateDiff,
    SVD,
    Deforum,
    Pika,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ModelCard {
    pub id: String,
    pub name: String,
    pub model_type: AnimationModel,
    pub description: String,
    pub capabilities: Vec<String>,
    pub max_duration: u32, // in seconds
    pub recommended_fps: Vec<u32>,
    pub recommended_resolutions: Vec<String>,
    pub installed: bool,
    pub size_gb: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "lowercase")]
pub enum GenerationStatus {
    Queued,
    Preparing,
    Generating,
    PostProcessing,
    Completed,
    Failed,
    Cancelled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GenerationJob {
    pub id: String,
    pub model: AnimationModel,
    pub prompt: String,
    pub negative_prompt: Option<String>,
    pub input_image: Option<String>,
    pub settings: VideoSettings,
    pub status: GenerationStatus,
    pub progress: f32, // 0.0 - 100.0
    pub current_frame: u32,
    pub total_frames: u32,
    pub output_path: Option<String>,
    pub error_message: Option<String>,
    pub created_at: String,
    pub completed_at: Option<String>,
    pub estimated_time_remaining: Option<u32>, // in seconds
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct VideoSettings {
    pub duration: u32, // in seconds
    pub fps: u32,
    pub resolution: String, // "512x512", "768x768", "1024x1024", etc.
    pub motion_scale: f32, // 0.0 - 2.0
    pub style_preset: Option<String>,
    pub seed: Option<i64>,
    pub camera_motion: CameraMotion,
    pub guidance_scale: f32, // 1.0 - 20.0
    pub num_inference_steps: u32, // 20-50
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CameraMotion {
    pub pan_x: f32,  // -1.0 to 1.0
    pub pan_y: f32,  // -1.0 to 1.0
    pub zoom: f32,   // 0.5 to 2.0
    pub rotate: f32, // -180.0 to 180.0 degrees
}

impl Default for CameraMotion {
    fn default() -> Self {
        Self {
            pan_x: 0.0,
            pan_y: 0.0,
            zoom: 1.0,
            rotate: 0.0,
        }
    }
}

impl Default for VideoSettings {
    fn default() -> Self {
        Self {
            duration: 2,
            fps: 8,
            resolution: "512x512".to_string(),
            motion_scale: 1.27,
            style_preset: None,
            seed: None,
            camera_motion: CameraMotion::default(),
            guidance_scale: 7.5,
            num_inference_steps: 25,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StylePreset {
    pub id: String,
    pub name: String,
    pub description: String,
    pub model_type: AnimationModel,
    pub positive_prompt_suffix: String,
    pub negative_prompt: String,
    pub recommended_settings: VideoSettings,
}

// Global state for animation generation
#[derive(Clone)]
pub struct AnimationState {
    pub jobs: Arc<Mutex<HashMap<String, GenerationJob>>>,
    pub models: Arc<Mutex<Vec<ModelCard>>>,
}

impl AnimationState {
    pub fn new() -> Self {
        Self {
            jobs: Arc::new(Mutex::new(HashMap::new())),
            models: Arc::new(Mutex::new(get_available_models())),
        }
    }
}

// Get available animation models
fn get_available_models() -> Vec<ModelCard> {
    vec![
        ModelCard {
            id: "animatediff".to_string(),
            name: "AnimateDiff".to_string(),
            model_type: AnimationModel::AnimateDiff,
            description: "Motion module for Stable Diffusion - converts any SD model to video".to_string(),
            capabilities: vec![
                "Text-to-video".to_string(),
                "Image-to-video".to_string(),
                "Motion control".to_string(),
                "Style transfer".to_string(),
            ],
            max_duration: 8,
            recommended_fps: vec![8, 12, 16],
            recommended_resolutions: vec![
                "512x512".to_string(),
                "768x768".to_string(),
                "512x768".to_string(),
            ],
            installed: false,
            size_gb: 1.7,
        },
        ModelCard {
            id: "svd".to_string(),
            name: "Stable Video Diffusion".to_string(),
            model_type: AnimationModel::SVD,
            description: "Stability AI's dedicated video generation model - high quality motion".to_string(),
            capabilities: vec![
                "Image-to-video".to_string(),
                "Smooth motion".to_string(),
                "14-25 frame generation".to_string(),
                "High quality output".to_string(),
            ],
            max_duration: 4,
            recommended_fps: vec![6, 12, 24],
            recommended_resolutions: vec![
                "576x1024".to_string(),
                "1024x576".to_string(),
            ],
            installed: false,
            size_gb: 3.5,
        },
        ModelCard {
            id: "deforum".to_string(),
            name: "Deforum".to_string(),
            model_type: AnimationModel::Deforum,
            description: "Advanced keyframe animation with 2D/3D camera movement".to_string(),
            capabilities: vec![
                "Keyframe animation".to_string(),
                "3D camera motion".to_string(),
                "Depth warping".to_string(),
                "Long videos".to_string(),
            ],
            max_duration: 30,
            recommended_fps: vec![12, 24, 30],
            recommended_resolutions: vec![
                "512x512".to_string(),
                "768x768".to_string(),
                "1024x1024".to_string(),
            ],
            installed: false,
            size_gb: 2.2,
        },
        ModelCard {
            id: "pika".to_string(),
            name: "Pika-style Motion".to_string(),
            model_type: AnimationModel::Pika,
            description: "Cinematic motion generation with camera controls".to_string(),
            capabilities: vec![
                "Text-to-video".to_string(),
                "Camera controls".to_string(),
                "Cinematic effects".to_string(),
                "Scene transitions".to_string(),
            ],
            max_duration: 3,
            recommended_fps: vec![24],
            recommended_resolutions: vec![
                "1024x576".to_string(),
                "576x1024".to_string(),
            ],
            installed: false,
            size_gb: 4.1,
        },
    ]
}

// Get available style presets
pub fn get_style_presets() -> Vec<StylePreset> {
    vec![
        StylePreset {
            id: "anime".to_string(),
            name: "Anime".to_string(),
            description: "Japanese animation style with vibrant colors".to_string(),
            model_type: AnimationModel::AnimateDiff,
            positive_prompt_suffix: ", anime style, cel shaded, vibrant colors, studio quality".to_string(),
            negative_prompt: "photorealistic, 3d render, low quality, blurry".to_string(),
            recommended_settings: VideoSettings {
                guidance_scale: 9.0,
                num_inference_steps: 30,
                ..Default::default()
            },
        },
        StylePreset {
            id: "realistic".to_string(),
            name: "Realistic".to_string(),
            description: "Photorealistic video generation".to_string(),
            model_type: AnimationModel::SVD,
            positive_prompt_suffix: ", photorealistic, 8k uhd, high quality, cinematic lighting".to_string(),
            negative_prompt: "cartoon, anime, illustration, painting, low quality".to_string(),
            recommended_settings: VideoSettings {
                guidance_scale: 7.0,
                num_inference_steps: 25,
                ..Default::default()
            },
        },
        StylePreset {
            id: "artistic".to_string(),
            name: "Artistic".to_string(),
            description: "Painterly and artistic style".to_string(),
            model_type: AnimationModel::AnimateDiff,
            positive_prompt_suffix: ", oil painting, artistic, brushstrokes, masterpiece".to_string(),
            negative_prompt: "photorealistic, 3d, low quality".to_string(),
            recommended_settings: VideoSettings {
                guidance_scale: 8.5,
                num_inference_steps: 35,
                ..Default::default()
            },
        },
        StylePreset {
            id: "cinematic".to_string(),
            name: "Cinematic".to_string(),
            description: "Movie-like cinematic quality".to_string(),
            model_type: AnimationModel::Pika,
            positive_prompt_suffix: ", cinematic, film grain, professional cinematography, 35mm".to_string(),
            negative_prompt: "amateur, low quality, cartoon".to_string(),
            recommended_settings: VideoSettings {
                guidance_scale: 7.5,
                num_inference_steps: 30,
                fps: 24,
                ..Default::default()
            },
        },
    ]
}

// Tauri commands

#[tauri::command]
pub fn get_animation_models(
    state: tauri::State<'_, AnimationState>,
) -> Result<Vec<ModelCard>, String> {
    let models = state.models.lock().map_err(|e| e.to_string())?;
    Ok(models.clone())
}

#[tauri::command]
pub fn get_animation_style_presets() -> Result<Vec<StylePreset>, String> {
    Ok(get_style_presets())
}

#[tauri::command]
pub async fn submit_generation_job<R: tauri::Runtime>(
    app: AppHandle<R>,
    state: tauri::State<'_, AnimationState>,
    model: String,
    prompt: String,
    negative_prompt: Option<String>,
    input_image: Option<String>,
    settings: VideoSettings,
) -> Result<String, String> {
    // Parse model type
    let model_type = match model.as_str() {
        "animatediff" => AnimationModel::AnimateDiff,
        "svd" => AnimationModel::SVD,
        "deforum" => AnimationModel::Deforum,
        "pika" => AnimationModel::Pika,
        _ => return Err("Invalid model type".to_string()),
    };

    // Generate unique job ID
    let job_id = format!("job_{}", chrono::Utc::now().timestamp_millis());

    // Calculate total frames
    let total_frames = settings.duration * settings.fps;

    // Create job
    let job = GenerationJob {
        id: job_id.clone(),
        model: model_type,
        prompt: prompt.clone(),
        negative_prompt,
        input_image,
        settings: settings.clone(),
        status: GenerationStatus::Queued,
        progress: 0.0,
        current_frame: 0,
        total_frames,
        output_path: None,
        error_message: None,
        created_at: chrono::Utc::now().to_rfc3339(),
        completed_at: None,
        estimated_time_remaining: Some(total_frames * 2), // Rough estimate: 2 seconds per frame
    };

    // Store job
    {
        let mut jobs = state.jobs.lock().map_err(|e| e.to_string())?;
        jobs.insert(job_id.clone(), job.clone());
    }

    // Emit job created event
    app.emit("generation_job_update", job.clone())
        .map_err(|e| e.to_string())?;

    // Clone job_id for return value
    let job_id_return = job_id.clone();

    // Start generation in background
    let state_clone = state.inner().clone();
    tokio::spawn(async move {
        if let Err(e) = process_generation_job(app.clone(), state_clone, job_id.clone()).await {

        }
    });

    Ok(job_id_return)
}

#[tauri::command]
pub fn get_generation_jobs(
    state: tauri::State<'_, AnimationState>,
) -> Result<Vec<GenerationJob>, String> {
    let jobs = state.jobs.lock().map_err(|e| e.to_string())?;
    let mut job_list: Vec<GenerationJob> = jobs.values().cloned().collect();

    // Sort by creation time (newest first)
    job_list.sort_by(|a, b| b.created_at.cmp(&a.created_at));

    Ok(job_list)
}

#[tauri::command]
pub fn get_generation_job(
    state: tauri::State<'_, AnimationState>,
    job_id: String,
) -> Result<GenerationJob, String> {
    let jobs = state.jobs.lock().map_err(|e| e.to_string())?;
    jobs.get(&job_id)
        .cloned()
        .ok_or_else(|| format!("Job {} not found", job_id))
}

#[tauri::command]
pub fn cancel_generation_job(
    state: tauri::State<'_, AnimationState>,
    job_id: String,
) -> Result<(), String> {
    let mut jobs = state.jobs.lock().map_err(|e| e.to_string())?;

    if let Some(job) = jobs.get_mut(&job_id) {
        if job.status == GenerationStatus::Queued
            || job.status == GenerationStatus::Preparing
            || job.status == GenerationStatus::Generating {
            job.status = GenerationStatus::Cancelled;
            job.completed_at = Some(chrono::Utc::now().to_rfc3339());
        }
    }

    Ok(())
}

#[tauri::command]
pub fn retry_generation_job<R: tauri::Runtime>(
    app: AppHandle<R>,
    state: tauri::State<'_, AnimationState>,
    job_id: String,
) -> Result<String, String> {
    let old_job = {
        let jobs = state.jobs.lock().map_err(|e| e.to_string())?;
        jobs.get(&job_id)
            .cloned()
            .ok_or_else(|| format!("Job {} not found", job_id))?
    };

    // Create new job with same parameters
    let new_job_id = format!("job_{}", chrono::Utc::now().timestamp_millis());

    let new_job = GenerationJob {
        id: new_job_id.clone(),
        model: old_job.model.clone(),
        prompt: old_job.prompt.clone(),
        negative_prompt: old_job.negative_prompt.clone(),
        input_image: old_job.input_image.clone(),
        settings: old_job.settings.clone(),
        status: GenerationStatus::Queued,
        progress: 0.0,
        current_frame: 0,
        total_frames: old_job.total_frames,
        output_path: None,
        error_message: None,
        created_at: chrono::Utc::now().to_rfc3339(),
        completed_at: None,
        estimated_time_remaining: Some(old_job.total_frames * 2),
    };

    // Store new job
    {
        let mut jobs = state.jobs.lock().map_err(|e| e.to_string())?;
        jobs.insert(new_job_id.clone(), new_job.clone());
    }

    // Emit job created event
    app.emit("generation_job_update", new_job)
        .map_err(|e| e.to_string())?;

    // Clone new_job_id for return value
    let new_job_id_return = new_job_id.clone();

    // Start generation in background
    let state_clone = state.inner().clone();
    tokio::spawn(async move {
        if let Err(e) = process_generation_job(app, state_clone, new_job_id.clone()).await {

        }
    });

    Ok(new_job_id_return)
}

#[tauri::command]
pub fn delete_generation_job(
    state: tauri::State<'_, AnimationState>,
    job_id: String,
) -> Result<(), String> {
    let mut jobs = state.jobs.lock().map_err(|e| e.to_string())?;
    jobs.remove(&job_id);
    Ok(())
}

// Process a generation job (simulated for now)
async fn process_generation_job<R: tauri::Runtime>(
    app: AppHandle<R>,
    state: AnimationState,
    job_id: String,
) -> Result<()> {
    // Update status to preparing
    update_job_status(&app, &state, &job_id, GenerationStatus::Preparing, 0.0, None)?;
    tokio::time::sleep(tokio::time::Duration::from_secs(2)).await;

    // Update status to generating
    update_job_status(&app, &state, &job_id, GenerationStatus::Generating, 5.0, None)?;

    // Get total frames
    let total_frames = {
        let jobs = state.jobs.lock().map_err(|e| anyhow!(e.to_string()))?;
        jobs.get(&job_id)
            .ok_or_else(|| anyhow!("Job not found"))?
            .total_frames
    };

    // Simulate frame-by-frame generation
    for frame in 1..=total_frames {
        // Check if cancelled
        {
            let jobs = state.jobs.lock().map_err(|e| anyhow!(e.to_string()))?;
            if let Some(job) = jobs.get(&job_id) {
                if job.status == GenerationStatus::Cancelled {
                    return Ok(());
                }
            }
        }

        let progress = (frame as f32 / total_frames as f32) * 90.0;

        // Update progress
        {
            let mut jobs = state.jobs.lock().map_err(|e| anyhow!(e.to_string()))?;
            if let Some(job) = jobs.get_mut(&job_id) {
                job.current_frame = frame;
                job.progress = progress;

                // Estimate time remaining
                let frames_remaining = total_frames - frame;
                job.estimated_time_remaining = Some(frames_remaining * 2);

                app.emit("generation_job_update", job.clone())
                    .map_err(|e| anyhow!(e.to_string()))?;
            }
        }

        // Simulate frame generation time (0.5-2 seconds per frame)
        tokio::time::sleep(tokio::time::Duration::from_millis(500)).await;
    }

    // Post-processing
    update_job_status(&app, &state, &job_id, GenerationStatus::PostProcessing, 95.0, None)?;
    tokio::time::sleep(tokio::time::Duration::from_secs(2)).await;

    // Complete
    let output_path = format!("/tmp/anime_output_{}.mp4", job_id);

    {
        let mut jobs = state.jobs.lock().map_err(|e| anyhow!(e.to_string()))?;
        if let Some(job) = jobs.get_mut(&job_id) {
            job.status = GenerationStatus::Completed;
            job.progress = 100.0;
            job.output_path = Some(output_path);
            job.completed_at = Some(chrono::Utc::now().to_rfc3339());
            job.estimated_time_remaining = Some(0);

            app.emit("generation_job_update", job.clone())
                .map_err(|e| anyhow!(e.to_string()))?;
        }
    }

    Ok(())
}

// Helper function to update job status
fn update_job_status<R: tauri::Runtime>(
    app: &AppHandle<R>,
    state: &AnimationState,
    job_id: &str,
    status: GenerationStatus,
    progress: f32,
    error_message: Option<String>,
) -> Result<()> {
    let mut jobs = state.jobs.lock().map_err(|e| anyhow!(e.to_string()))?;

    if let Some(job) = jobs.get_mut(job_id) {
        job.status = status;
        job.progress = progress;
        job.error_message = error_message;

        app.emit("generation_job_update", job.clone())
            .map_err(|e| anyhow!(e.to_string()))?;
    }

    Ok(())
}
