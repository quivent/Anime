use reqwest::{Client, header};
use anyhow::{Result, anyhow};
use super::types::*;

const LAMBDA_API_BASE: &str = "https://cloud.lambdalabs.com/api/v1";

#[derive(Clone)]
pub struct LambdaClient {
    client: Client,
    api_key: String,
}

impl LambdaClient {
    pub fn new(api_key: String) -> Result<Self> {
        let mut headers = header::HeaderMap::new();
        headers.insert(
            header::AUTHORIZATION,
            header::HeaderValue::from_str(&format!("Bearer {}", api_key))?,
        );
        headers.insert(
            header::CONTENT_TYPE,
            header::HeaderValue::from_static("application/json"),
        );

        let client = Client::builder()
            .default_headers(headers)
            .build()?;

        Ok(Self { client, api_key })
    }

    // Instance Types
    pub async fn list_instance_types(&self) -> Result<Vec<InstanceType>> {
        let url = format!("{}/instance-types", LAMBDA_API_BASE);
        eprintln!("[list_instance_types] Making request to: {}", url);

        let response = self.client.get(&url).send().await
            .map_err(|e| {
                eprintln!("[list_instance_types] Network error: {}", e);
                anyhow!("Network error: {}. Please check your internet connection.", e)
            })?;

        let status = response.status();
        eprintln!("[list_instance_types] Response status: {}", status);

        if !status.is_success() {
            let body = response.text().await.unwrap_or_else(|_| "Unable to read response".to_string());
            eprintln!("[list_instance_types] Error response body: {}", body);
            return Err(anyhow!("Lambda API request failed with status {}. Response: {}", status, body));
        }

        let body_text = response.text().await
            .map_err(|e| {
                eprintln!("[list_instance_types] Failed to read response body: {}", e);
                anyhow!("Failed to read response body: {}", e)
            })?;

        eprintln!("[list_instance_types] Raw response body (first 500 chars): {}",
            if body_text.len() > 500 { &body_text[..500] } else { &body_text });

        let result: ListInstanceTypesResponse = serde_json::from_str(&body_text)
            .map_err(|e| {
                eprintln!("[list_instance_types] JSON parse error: {}", e);
                eprintln!("[list_instance_types] Full response body: {}", body_text);
                anyhow!("Failed to parse API response: {}", e)
            })?;

        eprintln!("[list_instance_types] Successfully parsed {} instance types", result.data.len());

        // Convert HashMap to Vec of InstanceType with regions
        Ok(result.data.into_iter().map(|(_, entry)| {
            InstanceType {
                name: entry.instance_type.name,
                description: entry.instance_type.description,
                gpu_description: entry.instance_type.gpu_description,
                price_cents_per_hour: entry.instance_type.price_cents_per_hour,
                specs: entry.instance_type.specs,
                regions_with_capacity_available: entry.regions_with_capacity_available,
            }
        }).collect())
    }

    // Instances
    pub async fn list_instances(&self) -> Result<Vec<Instance>> {
        let url = format!("{}/instances", LAMBDA_API_BASE);
        let response = self.client.get(&url).send().await?;

        if !response.status().is_success() {
            let error: ApiError = response.json().await?;
            return Err(anyhow!("API Error: {}", error.error.message));
        }

        let result: ListInstancesResponse = response.json().await?;
        Ok(result.data)
    }

    pub async fn launch_instance(&self, request: LaunchInstanceRequest) -> Result<Vec<String>> {
        let url = format!("{}/instance-operations/launch", LAMBDA_API_BASE);
        eprintln!("[launch_instance] Making request to: {}", url);
        eprintln!("[launch_instance] Request payload: {:?}", request);

        let response = self.client.post(&url).json(&request).send().await
            .map_err(|e| {
                eprintln!("[launch_instance] Network error: {}", e);
                anyhow!("Network error: {}", e)
            })?;

        let status = response.status();
        eprintln!("[launch_instance] Response status: {}", status);

        let body_text = response.text().await
            .map_err(|e| {
                eprintln!("[launch_instance] Failed to read response body: {}", e);
                anyhow!("Failed to read response body: {}", e)
            })?;

        eprintln!("[launch_instance] Raw response body: {}", body_text);

        if !status.is_success() {
            eprintln!("[launch_instance] Error response");
            return Err(anyhow!("Lambda API request failed with status {}. Response: {}", status, body_text));
        }

        let result: LaunchInstanceResponse = serde_json::from_str(&body_text)
            .map_err(|e| {
                eprintln!("[launch_instance] Failed to parse response: {}", e);
                eprintln!("[launch_instance] Response was: {}", body_text);
                anyhow!("Failed to parse response: {}", e)
            })?;

        eprintln!("[launch_instance] Successfully launched: {:?}", result.data.instance_ids);
        Ok(result.data.instance_ids)
    }

    pub async fn terminate_instances(&self, instance_ids: Vec<String>) -> Result<Vec<String>> {
        let url = format!("{}/instance-operations/terminate", LAMBDA_API_BASE);
        eprintln!("[terminate_instances] Terminating instances: {:?}", instance_ids);

        let request = TerminateInstanceRequest { instance_ids };
        let response = self.client.post(&url).json(&request).send().await
            .map_err(|e| {
                eprintln!("[terminate_instances] Network error: {}", e);
                anyhow!("Network error: {}", e)
            })?;

        let status = response.status();
        eprintln!("[terminate_instances] Response status: {}", status);

        let body_text = response.text().await
            .map_err(|e| {
                eprintln!("[terminate_instances] Failed to read response body: {}", e);
                anyhow!("Failed to read response body: {}", e)
            })?;

        eprintln!("[terminate_instances] Raw response body: {}", body_text);

        if !status.is_success() {
            eprintln!("[terminate_instances] Error response");
            return Err(anyhow!("Lambda API request failed with status {}. Response: {}", status, body_text));
        }

        let result: TerminateInstanceResponse = serde_json::from_str(&body_text)
            .map_err(|e| {
                eprintln!("[terminate_instances] Failed to parse response: {}", e);
                eprintln!("[terminate_instances] Response was: {}", body_text);
                anyhow!("Failed to parse response: {}", e)
            })?;

        let terminated_ids: Vec<String> = result.data.terminated_instances.iter().map(|inst| inst.id.clone()).collect();
        eprintln!("[terminate_instances] Successfully terminated: {:?}", terminated_ids);
        Ok(terminated_ids)
    }

    pub async fn restart_instances(&self, instance_ids: Vec<String>) -> Result<Vec<String>> {
        let url = format!("{}/instance-operations/restart", LAMBDA_API_BASE);
        let request = RestartInstanceRequest { instance_ids };
        let response = self.client.post(&url).json(&request).send().await?;

        if !response.status().is_success() {
            let error: ApiError = response.json().await?;
            return Err(anyhow!("API Error: {}", error.error.message));
        }

        let result: RestartInstanceResponse = response.json().await?;
        Ok(result.restarted_instances)
    }

    // SSH Keys
    pub async fn list_ssh_keys(&self) -> Result<Vec<SSHKey>> {
        let url = format!("{}/ssh-keys", LAMBDA_API_BASE);
        let response = self.client.get(&url).send().await?;

        if !response.status().is_success() {
            let error: ApiError = response.json().await?;
            return Err(anyhow!("API Error: {}", error.error.message));
        }

        let result: ListSSHKeysResponse = response.json().await?;
        Ok(result.data)
    }

    pub async fn add_ssh_key(&self, name: String, public_key: String) -> Result<SSHKey> {
        let url = format!("{}/ssh-keys", LAMBDA_API_BASE);
        let request = AddSSHKeyRequest { name, public_key };
        let response = self.client.post(&url).json(&request).send().await?;

        if !response.status().is_success() {
            let error: ApiError = response.json().await?;
            return Err(anyhow!("API Error: {}", error.error.message));
        }

        let result: AddSSHKeyResponse = response.json().await?;
        Ok(result.data)
    }

    // File Systems
    pub async fn list_file_systems(&self) -> Result<Vec<FileSystem>> {
        let url = format!("{}/file-systems", LAMBDA_API_BASE);
        let response = self.client.get(&url).send().await?;

        if !response.status().is_success() {
            let error: ApiError = response.json().await?;
            return Err(anyhow!("API Error: {}", error.error.message));
        }

        let result: ListFileSystemsResponse = response.json().await?;
        Ok(result.data)
    }
}
