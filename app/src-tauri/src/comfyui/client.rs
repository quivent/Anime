use super::types::*;
use reqwest::Client;
use serde_json::Value;
use std::collections::HashMap;
use std::time::Duration;

pub struct ComfyUIClient {
    base_url: String,
    client: Client,
}

impl ComfyUIClient {
    pub fn new(host: &str, port: u16) -> Self {
        let base_url = format!("http://{}:{}", host, port);
        let client = Client::builder()
            .timeout(Duration::from_secs(30))
            .build()
            .unwrap_or_else(|_| Client::new());

        Self { base_url, client }
    }

    /// Check if ComfyUI is running and accessible
    pub async fn check_connection(&self) -> Result<bool, Box<dyn std::error::Error>> {
        let url = format!("{}/system_stats", self.base_url);
        match self.client.get(&url).send().await {
            Ok(response) => Ok(response.status().is_success()),
            Err(_) => Ok(false),
        }
    }

    /// Get ComfyUI system stats
    pub async fn get_system_stats(&self) -> Result<SystemStats, Box<dyn std::error::Error>> {
        let url = format!("{}/system_stats", self.base_url);
        let response = self.client.get(&url).send().await?;
        let stats = response.json::<SystemStats>().await?;
        Ok(stats)
    }

    /// Get queue status
    pub async fn get_queue(&self) -> Result<QueueStatus, Box<dyn std::error::Error>> {
        let url = format!("{}/queue", self.base_url);
        let response = self.client.get(&url).send().await?;
        let queue = response.json::<QueueStatus>().await?;
        Ok(queue)
    }

    /// Queue a prompt for execution
    pub async fn queue_prompt(
        &self,
        workflow: Value,
        client_id: Option<String>,
    ) -> Result<QueuePromptResponse, Box<dyn std::error::Error>> {
        let url = format!("{}/prompt", self.base_url);

        let request = QueuePromptRequest {
            client_id,
            prompt: workflow,
            extra_data: None,
        };

        let response = self.client.post(&url).json(&request).send().await?;

        if !response.status().is_success() {
            let error_text = response.text().await?;
            return Err(format!("Failed to queue prompt: {}", error_text).into());
        }

        let result = response.json::<QueuePromptResponse>().await?;
        Ok(result)
    }

    /// Get execution history
    pub async fn get_history(
        &self,
        prompt_id: Option<&str>,
    ) -> Result<HashMap<String, HistoryItem>, Box<dyn std::error::Error>> {
        let url = if let Some(id) = prompt_id {
            format!("{}/history/{}", self.base_url, id)
        } else {
            format!("{}/history", self.base_url)
        };

        let response = self.client.get(&url).send().await?;
        let history = response.json::<HistoryResponse>().await?;
        Ok(history.history)
    }

    /// Cancel a prompt in the queue
    pub async fn cancel_prompt(&self, prompt_id: &str) -> Result<(), Box<dyn std::error::Error>> {
        let url = format!("{}/queue", self.base_url);
        let mut delete_data = HashMap::new();
        delete_data.insert("delete", vec![prompt_id]);

        self.client.post(&url).json(&delete_data).send().await?;
        Ok(())
    }

    /// Get an output image
    pub async fn get_image(
        &self,
        filename: &str,
        subfolder: &str,
        folder_type: &str,
    ) -> Result<Vec<u8>, Box<dyn std::error::Error>> {
        let url = format!(
            "{}/view?filename={}&subfolder={}&type={}",
            self.base_url, filename, subfolder, folder_type
        );

        let response = self.client.get(&url).send().await?;
        let bytes = response.bytes().await?;
        Ok(bytes.to_vec())
    }

    /// Upload an image to ComfyUI
    pub async fn upload_image(
        &self,
        image_data: Vec<u8>,
        filename: &str,
        overwrite: bool,
    ) -> Result<String, Box<dyn std::error::Error>> {
        let url = format!("{}/upload/image", self.base_url);

        let form = reqwest::multipart::Form::new()
            .part(
                "image",
                reqwest::multipart::Part::bytes(image_data)
                    .file_name(filename.to_string()),
            )
            .text("overwrite", overwrite.to_string());

        let response = self.client.post(&url).multipart(form).send().await?;

        if !response.status().is_success() {
            let error_text = response.text().await?;
            return Err(format!("Failed to upload image: {}", error_text).into());
        }

        let result: Value = response.json().await?;
        let uploaded_name = result["name"]
            .as_str()
            .ok_or("No filename in response")?
            .to_string();

        Ok(uploaded_name)
    }

    /// Interrupt current execution
    pub async fn interrupt(&self) -> Result<(), Box<dyn std::error::Error>> {
        let url = format!("{}/interrupt", self.base_url);
        self.client.post(&url).send().await?;
        Ok(())
    }

    /// Clear the queue
    pub async fn clear_queue(&self) -> Result<(), Box<dyn std::error::Error>> {
        let url = format!("{}/queue", self.base_url);
        let mut clear_data = HashMap::new();
        clear_data.insert("clear", true);

        self.client.post(&url).json(&clear_data).send().await?;
        Ok(())
    }
}
