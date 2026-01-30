use anyhow::{anyhow, Result};
use futures::StreamExt;
use reqwest::Client;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::fs::{self, File};
use std::io::Write;
use std::path::{Path, PathBuf};
use std::sync::Arc;
use tauri::{AppHandle, Emitter};
use tokio::sync::Mutex;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ModelDownloadProgress {
    pub model_id: String,
    pub status: DownloadStatus,
    pub bytes_downloaded: u64,
    pub total_bytes: u64,
    pub progress: f32, // 0.0 - 100.0
    pub message: String,
    pub download_speed: Option<f64>, // bytes/sec
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "lowercase")]
pub enum DownloadStatus {
    Pending,
    Downloading,
    Completed,
    Failed,
    Cancelled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InstalledModel {
    pub model_id: String,
    pub name: String,
    pub size_bytes: u64,
    pub install_path: String,
    pub installed_at: String,
    pub model_type: ModelType,
    pub loaded: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "lowercase")]
pub enum ModelType {
    Ollama,
    HuggingFace,
    ComfyUI,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ModelMetadata {
    pub model_id: String,
    pub name: String,
    pub model_type: ModelType,
    pub download_url: Option<String>,
    pub ollama_name: Option<String>,
    pub huggingface_repo: Option<String>,
    pub file_name: Option<String>,
    pub estimated_size: u64,
}

// Ollama API types
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OllamaPullRequest {
    pub name: String,
    pub stream: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OllamaPullResponse {
    pub status: String,
    #[serde(default)]
    pub digest: String,
    #[serde(default)]
    pub total: u64,
    #[serde(default)]
    pub completed: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OllamaListResponse {
    pub models: Vec<OllamaModel>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OllamaModel {
    pub name: String,
    pub modified_at: String,
    pub size: u64,
    pub digest: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OllamaLoadRequest {
    pub model: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OllamaGenerateRequest {
    pub model: String,
    pub prompt: String,
    pub stream: bool,
}

// Model download state manager
pub struct ModelDownloadManager {
    active_downloads: Arc<Mutex<HashMap<String, DownloadStatus>>>,
    http_client: Client,
    models_dir: PathBuf,
}

impl ModelDownloadManager {
    pub fn new() -> Result<Self> {
        let home = std::env::var("HOME").or_else(|_| std::env::var("USERPROFILE"))?;
        let models_dir = PathBuf::from(home).join(".anime-desktop").join("models");

        // Create models directory if it doesn't exist
        fs::create_dir_all(&models_dir)?;

        Ok(Self {
            active_downloads: Arc::new(Mutex::new(HashMap::new())),
            http_client: Client::builder()
                .user_agent("ANIME-Desktop/0.1.0")
                .timeout(std::time::Duration::from_secs(300))
                .build()?,
            models_dir,
        })
    }

    /// Download a model from Ollama
    pub async fn download_ollama_model<R: tauri::Runtime>(
        &self,
        app: &AppHandle<R>,
        model_id: String,
        ollama_name: String,
        ollama_url: Option<String>,
    ) -> Result<()> {
        let ollama_base = ollama_url.unwrap_or_else(|| "http://localhost:11434".to_string());

        // Mark download as in progress
        {
            let mut downloads = self.active_downloads.lock().await;
            downloads.insert(model_id.clone(), DownloadStatus::Downloading);
        }

        // Emit initial progress
        app.emit(
            "model_download_progress",
            ModelDownloadProgress {
                model_id: model_id.clone(),
                status: DownloadStatus::Downloading,
                bytes_downloaded: 0,
                total_bytes: 0,
                progress: 0.0,
                message: format!("Starting download of {}...", ollama_name),
                download_speed: None,
            },
        )?;

        // Pull the model using Ollama API
        let pull_url = format!("{}/api/pull", ollama_base);
        let request_body = OllamaPullRequest {
            name: ollama_name.clone(),
            stream: true,
        };

        let response = self
            .http_client
            .post(&pull_url)
            .json(&request_body)
            .send()
            .await
            .map_err(|e| anyhow!("Failed to connect to Ollama: {}", e))?;

        if !response.status().is_success() {
            return Err(anyhow!("Ollama pull request failed: {}", response.status()));
        }

        // Stream the progress
        let mut stream = response.bytes_stream();
        use futures::stream::StreamExt;

        let mut buffer = String::new();
        while let Some(chunk) = stream.next().await {
            let chunk = chunk.map_err(|e| anyhow!("Stream error: {}", e))?;
            buffer.push_str(&String::from_utf8_lossy(&chunk));

            // Process complete JSON lines
            while let Some(newline_pos) = buffer.find('\n') {
                let line = buffer[..newline_pos].to_string();
                buffer = buffer[newline_pos + 1..].to_string();

                if let Ok(pull_response) = serde_json::from_str::<OllamaPullResponse>(&line) {
                    let progress = if pull_response.total > 0 {
                        (pull_response.completed as f32 / pull_response.total as f32) * 100.0
                    } else {
                        0.0
                    };

                    app.emit(
                        "model_download_progress",
                        ModelDownloadProgress {
                            model_id: model_id.clone(),
                            status: DownloadStatus::Downloading,
                            bytes_downloaded: pull_response.completed,
                            total_bytes: pull_response.total,
                            progress,
                            message: pull_response.status.clone(),
                            download_speed: None,
                        },
                    )?;
                }
            }
        }

        // Mark as completed
        {
            let mut downloads = self.active_downloads.lock().await;
            downloads.insert(model_id.clone(), DownloadStatus::Completed);
        }

        app.emit(
            "model_download_progress",
            ModelDownloadProgress {
                model_id: model_id.clone(),
                status: DownloadStatus::Completed,
                bytes_downloaded: 0,
                total_bytes: 0,
                progress: 100.0,
                message: format!("{} downloaded successfully!", ollama_name),
                download_speed: None,
            },
        )?;

        Ok(())
    }

    /// Download a model from HuggingFace Hub
    pub async fn download_huggingface_model<R: tauri::Runtime>(
        &self,
        app: &AppHandle<R>,
        model_id: String,
        repo_id: String,
        file_name: String,
    ) -> Result<()> {
        // Mark download as in progress
        {
            let mut downloads = self.active_downloads.lock().await;
            downloads.insert(model_id.clone(), DownloadStatus::Downloading);
        }

        // Construct HuggingFace download URL
        let download_url = format!(
            "https://huggingface.co/{}/resolve/main/{}",
            repo_id, file_name
        );

        // Get file info to determine size
        let head_response = self.http_client.head(&download_url).send().await?;
        let total_bytes = head_response
            .headers()
            .get("content-length")
            .and_then(|v| v.to_str().ok())
            .and_then(|v| v.parse::<u64>().ok())
            .unwrap_or(0);

        app.emit(
            "model_download_progress",
            ModelDownloadProgress {
                model_id: model_id.clone(),
                status: DownloadStatus::Downloading,
                bytes_downloaded: 0,
                total_bytes,
                progress: 0.0,
                message: format!("Downloading {} ({:.2} GB)...", file_name, total_bytes as f64 / 1_073_741_824.0),
                download_speed: None,
            },
        )?;

        // Create destination path
        let dest_path = self.models_dir.join(&model_id).join(&file_name);
        let parent_dir = dest_path.parent()
            .ok_or_else(|| anyhow!("Invalid destination path: cannot determine parent directory"))?;
        fs::create_dir_all(parent_dir)?;

        // Download with progress tracking
        let response = self.http_client.get(&download_url).send().await?;

        if !response.status().is_success() {
            return Err(anyhow!("Download failed: {}", response.status()));
        }

        let mut file = File::create(&dest_path)?;
        let mut downloaded: u64 = 0;
        let mut stream = response.bytes_stream();
        let start_time = std::time::Instant::now();

        while let Some(chunk) = stream.next().await {
            let chunk = chunk.map_err(|e| anyhow!("Download error: {}", e))?;
            file.write_all(&chunk)?;
            downloaded += chunk.len() as u64;

            let progress = if total_bytes > 0 {
                (downloaded as f32 / total_bytes as f32) * 100.0
            } else {
                0.0
            };

            let elapsed = start_time.elapsed().as_secs_f64();
            let download_speed = if elapsed > 0.0 {
                Some(downloaded as f64 / elapsed)
            } else {
                None
            };

            // Emit progress every 1MB
            if downloaded % (1024 * 1024) == 0 || downloaded == total_bytes {
                app.emit(
                    "model_download_progress",
                    ModelDownloadProgress {
                        model_id: model_id.clone(),
                        status: DownloadStatus::Downloading,
                        bytes_downloaded: downloaded,
                        total_bytes,
                        progress,
                        message: format!(
                            "Downloading... {:.2} MB / {:.2} MB",
                            downloaded as f64 / 1_048_576.0,
                            total_bytes as f64 / 1_048_576.0
                        ),
                        download_speed,
                    },
                )?;
            }
        }

        file.flush()?;

        // Mark as completed
        {
            let mut downloads = self.active_downloads.lock().await;
            downloads.insert(model_id.clone(), DownloadStatus::Completed);
        }

        app.emit(
            "model_download_progress",
            ModelDownloadProgress {
                model_id: model_id.clone(),
                status: DownloadStatus::Completed,
                bytes_downloaded: total_bytes,
                total_bytes,
                progress: 100.0,
                message: format!("{} downloaded successfully!", file_name),
                download_speed: None,
            },
        )?;

        Ok(())
    }

    /// List installed models from Ollama
    pub async fn list_ollama_models(&self, ollama_url: Option<String>) -> Result<Vec<OllamaModel>> {
        let ollama_base = ollama_url.unwrap_or_else(|| "http://localhost:11434".to_string());
        let list_url = format!("{}/api/tags", ollama_base);

        let response = self.http_client.get(&list_url).send().await?;

        if !response.status().is_success() {
            return Err(anyhow!("Failed to list Ollama models: {}", response.status()));
        }

        let list_response: OllamaListResponse = response.json().await?;
        Ok(list_response.models)
    }

    /// Delete a model from Ollama
    pub async fn delete_ollama_model(
        &self,
        ollama_name: String,
        ollama_url: Option<String>,
    ) -> Result<()> {
        let ollama_base = ollama_url.unwrap_or_else(|| "http://localhost:11434".to_string());
        let delete_url = format!("{}/api/delete", ollama_base);

        let request_body = serde_json::json!({
            "name": ollama_name
        });

        let response = self
            .http_client
            .delete(&delete_url)
            .json(&request_body)
            .send()
            .await?;

        if !response.status().is_success() {
            return Err(anyhow!("Failed to delete Ollama model: {}", response.status()));
        }

        Ok(())
    }

    /// Load a model into Ollama memory
    pub async fn load_ollama_model(
        &self,
        ollama_name: String,
        ollama_url: Option<String>,
    ) -> Result<()> {
        let ollama_base = ollama_url.unwrap_or_else(|| "http://localhost:11434".to_string());
        let generate_url = format!("{}/api/generate", ollama_base);

        // Send a minimal generate request to load the model
        let request_body = OllamaGenerateRequest {
            model: ollama_name.clone(),
            prompt: "Hello".to_string(),
            stream: false,
        };

        let response = self
            .http_client
            .post(&generate_url)
            .json(&request_body)
            .send()
            .await?;

        if !response.status().is_success() {
            return Err(anyhow!("Failed to load Ollama model: {}", response.status()));
        }

        Ok(())
    }

    /// Delete a HuggingFace model from local storage
    pub async fn delete_huggingface_model(&self, model_id: String) -> Result<()> {
        let model_dir = self.models_dir.join(&model_id);

        if model_dir.exists() {
            fs::remove_dir_all(&model_dir)?;
        }

        Ok(())
    }

    /// List all locally downloaded HuggingFace models
    pub async fn list_local_models(&self) -> Result<Vec<InstalledModel>> {
        let mut models = Vec::new();

        if !self.models_dir.exists() {
            return Ok(models);
        }

        for entry in fs::read_dir(&self.models_dir)? {
            let entry = entry?;
            let path = entry.path();

            if path.is_dir() {
                let model_id = path.file_name().unwrap().to_string_lossy().to_string();
                let size_bytes = get_dir_size(&path)?;

                models.push(InstalledModel {
                    model_id: model_id.clone(),
                    name: model_id.clone(),
                    size_bytes,
                    install_path: path.to_string_lossy().to_string(),
                    installed_at: chrono::Utc::now().to_rfc3339(),
                    model_type: ModelType::HuggingFace,
                    loaded: false,
                });
            }
        }

        Ok(models)
    }

    /// Cancel an ongoing download
    pub async fn cancel_download(&self, model_id: String) -> Result<()> {
        let mut downloads = self.active_downloads.lock().await;
        downloads.insert(model_id, DownloadStatus::Cancelled);
        Ok(())
    }
}

/// Get total size of a directory
fn get_dir_size(path: &Path) -> Result<u64> {
    let mut total_size = 0;

    if path.is_file() {
        return Ok(path.metadata()?.len());
    }

    for entry in fs::read_dir(path)? {
        let entry = entry?;
        let metadata = entry.metadata()?;

        if metadata.is_file() {
            total_size += metadata.len();
        } else if metadata.is_dir() {
            total_size += get_dir_size(&entry.path())?;
        }
    }

    Ok(total_size)
}

/// Get model metadata based on package ID
pub fn get_model_metadata(package_id: &str) -> Option<ModelMetadata> {
    match package_id {
        // Small LLM models (7-8B)
        "llama-3_3-8b" => Some(ModelMetadata {
            model_id: package_id.to_string(),
            name: "Llama 3.3 8B".to_string(),
            model_type: ModelType::Ollama,
            download_url: None,
            ollama_name: Some("llama3.3".to_string()),
            huggingface_repo: None,
            file_name: None,
            estimated_size: 5_368_709_120,
        }),
        "mistral-7b" => Some(ModelMetadata {
            model_id: package_id.to_string(),
            name: "Mistral 7B".to_string(),
            model_type: ModelType::Ollama,
            download_url: None,
            ollama_name: Some("mistral".to_string()),
            huggingface_repo: None,
            file_name: None,
            estimated_size: 4_294_967_296,
        }),
        "qwen-2_5-7b" => Some(ModelMetadata {
            model_id: package_id.to_string(),
            name: "Qwen 2.5 7B".to_string(),
            model_type: ModelType::Ollama,
            download_url: None,
            ollama_name: Some("qwen2.5:7b".to_string()),
            huggingface_repo: None,
            file_name: None,
            estimated_size: 4_831_838_208,
        }),
        "phi-4" => Some(ModelMetadata {
            model_id: package_id.to_string(),
            name: "Phi-4".to_string(),
            model_type: ModelType::Ollama,
            download_url: None,
            ollama_name: Some("phi4".to_string()),
            huggingface_repo: None,
            file_name: None,
            estimated_size: 4_080_218_931,
        }),
        // Medium LLM models (14-34B)
        "qwen-2_5-14b" => Some(ModelMetadata {
            model_id: package_id.to_string(),
            name: "Qwen 2.5 14B".to_string(),
            model_type: ModelType::Ollama,
            download_url: None,
            ollama_name: Some("qwen2.5:14b".to_string()),
            huggingface_repo: None,
            file_name: None,
            estimated_size: 9_663_676_416,
        }),
        "mixtral-8x7b" => Some(ModelMetadata {
            model_id: package_id.to_string(),
            name: "Mixtral 8x7B".to_string(),
            model_type: ModelType::Ollama,
            download_url: None,
            ollama_name: Some("mixtral".to_string()),
            huggingface_repo: None,
            file_name: None,
            estimated_size: 27_917_287_424,
        }),
        "deepseek-coder-33b" => Some(ModelMetadata {
            model_id: package_id.to_string(),
            name: "DeepSeek Coder 33B".to_string(),
            model_type: ModelType::Ollama,
            download_url: None,
            ollama_name: Some("deepseek-coder:33b".to_string()),
            huggingface_repo: None,
            file_name: None,
            estimated_size: 20_401_094_656,
        }),
        "yi-34b" => Some(ModelMetadata {
            model_id: package_id.to_string(),
            name: "Yi 34B".to_string(),
            model_type: ModelType::Ollama,
            download_url: None,
            ollama_name: Some("yi:34b".to_string()),
            huggingface_repo: None,
            file_name: None,
            estimated_size: 21_474_836_480,
        }),
        // Large LLM models (70B+)
        "llama-3_3-70b" => Some(ModelMetadata {
            model_id: package_id.to_string(),
            name: "Llama 3.3 70B".to_string(),
            model_type: ModelType::Ollama,
            download_url: None,
            ollama_name: Some("llama3.3:70b".to_string()),
            huggingface_repo: None,
            file_name: None,
            estimated_size: 42_949_672_960,
        }),
        "qwen-2_5-72b" => Some(ModelMetadata {
            model_id: package_id.to_string(),
            name: "Qwen 2.5 72B".to_string(),
            model_type: ModelType::Ollama,
            download_url: None,
            ollama_name: Some("qwen2.5:72b".to_string()),
            huggingface_repo: None,
            file_name: None,
            estimated_size: 45_097_156_608,
        }),
        "deepseek-v3" => Some(ModelMetadata {
            model_id: package_id.to_string(),
            name: "DeepSeek V3".to_string(),
            model_type: ModelType::Ollama,
            download_url: None,
            ollama_name: Some("deepseek-v3".to_string()),
            huggingface_repo: None,
            file_name: None,
            estimated_size: 96_636_764_160,
        }),
        _ => None,
    }
}

// Global state for model manager
pub struct ModelManagerState {
    pub manager: Arc<Mutex<ModelDownloadManager>>,
}

// Tauri commands
#[tauri::command]
pub async fn download_model<R: tauri::Runtime>(
    app: AppHandle<R>,
    state: tauri::State<'_, ModelManagerState>,
    model_id: String,
    ollama_url: Option<String>,
) -> Result<(), String> {
    let metadata = get_model_metadata(&model_id)
        .ok_or_else(|| format!("Unknown model: {}", model_id))?;

    let manager = state.manager.lock().await;

    match metadata.model_type {
        ModelType::Ollama => {
            let ollama_name = metadata
                .ollama_name
                .ok_or_else(|| "No Ollama name configured".to_string())?;

            manager
                .download_ollama_model(&app, model_id, ollama_name, ollama_url)
                .await
                .map_err(|e| e.to_string())?;
        }
        ModelType::HuggingFace => {
            let repo = metadata
                .huggingface_repo
                .ok_or_else(|| "No HuggingFace repo configured".to_string())?;
            let file_name = metadata
                .file_name
                .ok_or_else(|| "No file name configured".to_string())?;

            manager
                .download_huggingface_model(&app, model_id, repo, file_name)
                .await
                .map_err(|e| e.to_string())?;
        }
        _ => {
            return Err("Unsupported model type".to_string());
        }
    }

    Ok(())
}

#[tauri::command]
pub async fn list_installed_models(
    state: tauri::State<'_, ModelManagerState>,
    ollama_url: Option<String>,
) -> Result<Vec<InstalledModel>, String> {
    let manager = state.manager.lock().await;

    // Get Ollama models
    let mut installed = Vec::new();

    match manager.list_ollama_models(ollama_url).await {
        Ok(ollama_models) => {
            for model in ollama_models {
                installed.push(InstalledModel {
                    model_id: model.name.clone(),
                    name: model.name.clone(),
                    size_bytes: model.size,
                    install_path: "ollama".to_string(),
                    installed_at: model.modified_at,
                    model_type: ModelType::Ollama,
                    loaded: false,
                });
            }
        }
        Err(e) => {

        }
    }

    // Get local HuggingFace models
    match manager.list_local_models().await {
        Ok(local_models) => {
            installed.extend(local_models);
        }
        Err(e) => {

        }
    }

    Ok(installed)
}

#[tauri::command]
pub async fn delete_model(
    state: tauri::State<'_, ModelManagerState>,
    model_id: String,
    model_type: Option<String>,
    ollama_url: Option<String>,
) -> Result<(), String> {
    let metadata = get_model_metadata(&model_id);
    let manager = state.manager.lock().await;

    let actual_model_type = if let Some(ref mt) = model_type {
        match mt.as_str() {
            "ollama" => ModelType::Ollama,
            "huggingface" => ModelType::HuggingFace,
            _ => {
                metadata
                    .as_ref()
                    .map(|m| m.model_type.clone())
                    .unwrap_or(ModelType::Ollama)
            }
        }
    } else {
        metadata
            .as_ref()
            .map(|m| m.model_type.clone())
            .unwrap_or(ModelType::Ollama)
    };

    match actual_model_type {
        ModelType::Ollama => {
            let ollama_name = if let Some(meta) = metadata {
                meta.ollama_name.unwrap_or(model_id)
            } else {
                model_id
            };

            manager
                .delete_ollama_model(ollama_name, ollama_url)
                .await
                .map_err(|e| e.to_string())?;
        }
        ModelType::HuggingFace => {
            manager
                .delete_huggingface_model(model_id)
                .await
                .map_err(|e| e.to_string())?;
        }
        _ => {
            return Err("Unsupported model type".to_string());
        }
    }

    Ok(())
}

#[tauri::command]
pub async fn load_model(
    state: tauri::State<'_, ModelManagerState>,
    model_id: String,
    ollama_url: Option<String>,
) -> Result<(), String> {
    let metadata = get_model_metadata(&model_id)
        .ok_or_else(|| format!("Unknown model: {}", model_id))?;

    if metadata.model_type != ModelType::Ollama {
        return Err("Only Ollama models can be loaded into memory".to_string());
    }

    let ollama_name = metadata
        .ollama_name
        .ok_or_else(|| "No Ollama name configured".to_string())?;

    let manager = state.manager.lock().await;

    manager
        .load_ollama_model(ollama_name, ollama_url)
        .await
        .map_err(|e| e.to_string())?;

    Ok(())
}

#[tauri::command]
pub async fn cancel_model_download(
    state: tauri::State<'_, ModelManagerState>,
    model_id: String,
) -> Result<(), String> {
    let manager = state.manager.lock().await;
    manager
        .cancel_download(model_id)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
pub fn get_model_info(model_id: String) -> Result<ModelMetadata, String> {
    get_model_metadata(&model_id).ok_or_else(|| format!("Unknown model: {}", model_id))
}
