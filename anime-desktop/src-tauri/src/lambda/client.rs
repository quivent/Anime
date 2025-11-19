use reqwest::{Client, header};
use anyhow::{Result, anyhow};
use super::types::*;

const LAMBDA_API_BASE: &str = "https://cloud.lambdalabs.com/api/v1";

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
        let response = self.client.get(&url).send().await?;

        if !response.status().is_success() {
            let error: ApiError = response.json().await?;
            return Err(anyhow!("API Error: {}", error.error.message));
        }

        let result: ListInstanceTypesResponse = response.json().await?;
        Ok(result.data)
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
        let response = self.client.post(&url).json(&request).send().await?;

        if !response.status().is_success() {
            let error: ApiError = response.json().await?;
            return Err(anyhow!("API Error: {}", error.error.message));
        }

        let result: LaunchInstanceResponse = response.json().await?;
        Ok(result.instance_ids)
    }

    pub async fn terminate_instances(&self, instance_ids: Vec<String>) -> Result<Vec<String>> {
        let url = format!("{}/instance-operations/terminate", LAMBDA_API_BASE);
        let request = TerminateInstanceRequest { instance_ids };
        let response = self.client.post(&url).json(&request).send().await?;

        if !response.status().is_success() {
            let error: ApiError = response.json().await?;
            return Err(anyhow!("API Error: {}", error.error.message));
        }

        let result: TerminateInstanceResponse = response.json().await?;
        Ok(result.terminated_instances)
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
