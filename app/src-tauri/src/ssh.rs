// SSH connection management
// TODO: Implement full SSH functionality using ssh2 crate

use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Server {
    pub id: String,
    pub name: String,
    pub host: String,
    pub port: u16,
    pub username: String,
    pub status: ServerStatus,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum ServerStatus {
    Connected,
    Disconnected,
    Error,
}

pub fn test_connection(_server: &Server) -> Result<(), String> {
    // TODO: Implement SSH connection test
    Ok(())
}

pub fn execute_remote_command(_server: &Server, _command: &str) -> Result<String, String> {
    // TODO: Implement remote command execution
    Ok("Command executed".to_string())
}
