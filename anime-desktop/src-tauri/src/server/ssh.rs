use anyhow::{Result, anyhow};
use ssh2::Session;
use std::net::TcpStream;
use std::path::Path;

pub struct ServerConnection {
    session: Session,
    host: String,
    username: String,
}

impl ServerConnection {
    /// Connect to a server using SSH key authentication
    pub fn connect(host: &str, username: &str, private_key_path: &str) -> Result<Self> {
        eprintln!("[SSH] Connecting to {}@{}...", username, host);

        // Establish TCP connection
        let tcp = TcpStream::connect(format!("{}:22", host))
            .map_err(|e| anyhow!("Failed to connect to {}: {}", host, e))?;

        // Create SSH session
        let mut session = Session::new()?;
        session.set_tcp_stream(tcp);
        session.handshake()?;

        // Authenticate with private key
        let key_path = Path::new(private_key_path);
        session.userauth_pubkey_file(username, None, key_path, None)
            .map_err(|e| anyhow!("SSH authentication failed: {}", e))?;

        if !session.authenticated() {
            return Err(anyhow!("SSH authentication failed"));
        }

        eprintln!("[SSH] Successfully connected to {}", host);

        Ok(Self {
            session,
            host: host.to_string(),
            username: username.to_string(),
        })
    }

    /// Execute a command on the remote server and return the output
    pub fn execute_command(&self, command: &str) -> Result<String> {
        eprintln!("[SSH] Executing: {}", command);

        let mut channel = self.session.channel_session()?;
        channel.exec(command)?;

        let mut output = String::new();
        channel.read_to_string(&mut output)?;

        channel.wait_close()?;
        let exit_status = channel.exit_status()?;

        if exit_status != 0 {
            let mut stderr = String::new();
            channel.stderr().read_to_string(&mut stderr)?;
            eprintln!("[SSH] Command failed with exit code {}: {}", exit_status, stderr);
            return Err(anyhow!("Command failed with exit code {}: {}", exit_status, stderr));
        }

        Ok(output.trim().to_string())
    }

    /// Get the hostname of the connected server
    pub fn hostname(&self) -> &str {
        &self.host
    }

    /// Get the username used for the connection
    pub fn username(&self) -> &str {
        &self.username
    }
}

impl Drop for ServerConnection {
    fn drop(&mut self) {
        let _ = self.session.disconnect(None, "Closing connection", None);
        eprintln!("[SSH] Disconnected from {}", self.host);
    }
}
