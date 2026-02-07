package launch

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// SIWEConfig holds configuration for Sign-In With Ethereum
type SIWEConfig struct {
	AppName          string
	Domain           string
	ChainID          int
	AllowedAddresses []string
	ServicePort      int
	NonceExpiry      string
	SessionTTL       string
	User             string
}

// DefaultSIWEConfig returns a default SIWE configuration
func DefaultSIWEConfig(appName, domain string) *SIWEConfig {
	return &SIWEConfig{
		AppName:     appName,
		Domain:      domain,
		ChainID:     1, // Ethereum mainnet
		ServicePort: 4181,
		NonceExpiry: "5m",
		SessionTTL:  "24h",
	}
}

// GenerateSIWENonce generates a random nonce for SIWE
func GenerateSIWENonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// SIWEMessage represents the components of a SIWE message
type SIWEMessage struct {
	Domain         string
	Address        string
	Statement      string
	URI            string
	Version        string
	ChainID        int
	Nonce          string
	IssuedAt       time.Time
	ExpirationTime *time.Time
	NotBefore      *time.Time
	RequestID      string
	Resources      []string
}

// FormatSIWEMessage formats a SIWE message according to EIP-4361
func FormatSIWEMessage(msg *SIWEMessage) string {
	var buf strings.Builder

	// First line: domain wants you to sign in
	buf.WriteString(fmt.Sprintf("%s wants you to sign in with your Ethereum account:\n", msg.Domain))
	buf.WriteString(fmt.Sprintf("%s\n", msg.Address))

	// Statement (optional)
	if msg.Statement != "" {
		buf.WriteString(fmt.Sprintf("\n%s\n", msg.Statement))
	}

	// Required fields
	buf.WriteString(fmt.Sprintf("\nURI: %s\n", msg.URI))
	buf.WriteString(fmt.Sprintf("Version: %s\n", msg.Version))
	buf.WriteString(fmt.Sprintf("Chain ID: %d\n", msg.ChainID))
	buf.WriteString(fmt.Sprintf("Nonce: %s\n", msg.Nonce))
	buf.WriteString(fmt.Sprintf("Issued At: %s", msg.IssuedAt.UTC().Format(time.RFC3339)))

	// Optional fields
	if msg.ExpirationTime != nil {
		buf.WriteString(fmt.Sprintf("\nExpiration Time: %s", msg.ExpirationTime.UTC().Format(time.RFC3339)))
	}
	if msg.NotBefore != nil {
		buf.WriteString(fmt.Sprintf("\nNot Before: %s", msg.NotBefore.UTC().Format(time.RFC3339)))
	}
	if msg.RequestID != "" {
		buf.WriteString(fmt.Sprintf("\nRequest ID: %s", msg.RequestID))
	}
	if len(msg.Resources) > 0 {
		buf.WriteString("\nResources:")
		for _, r := range msg.Resources {
			buf.WriteString(fmt.Sprintf("\n- %s", r))
		}
	}

	return buf.String()
}

// GenerateSIWEServiceSystemdUnit generates a systemd unit for the SIWE service
func GenerateSIWEServiceSystemdUnit(cfg *SIWEConfig) string {
	allowedAddrs := ""
	if len(cfg.AllowedAddresses) > 0 {
		allowedAddrs = fmt.Sprintf("--allowed-addresses=%s", strings.Join(cfg.AllowedAddresses, ","))
	}

	return fmt.Sprintf(`[Unit]
Description=SIWE Auth Service for %s
After=network-online.target

[Service]
ExecStart=/usr/local/bin/anime siwe-service \
    --domain=%s \
    --chain-id=%d \
    --port=%d \
    --nonce-expiry=%s \
    --session-ttl=%s %s
User=%s
Restart=always
RestartSec=3
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
`, cfg.AppName, cfg.Domain, cfg.ChainID, cfg.ServicePort, cfg.NonceExpiry, cfg.SessionTTL, allowedAddrs, cfg.User)
}

// GenerateSIWENginxBlocks generates nginx configuration for SIWE
type SIWENginxBlocks struct {
	// ServiceLocation - location block for /siwe/ endpoints
	ServiceLocation string
	// AuthRequest - auth_request directive for protected locations
	AuthRequest string
	// LoginPage - optional static login page location
	LoginPage string
}

// GenerateSIWENginxBlocks creates nginx blocks for SIWE integration
func GenerateSIWENginxBlocks(port int) *SIWENginxBlocks {
	if port == 0 {
		port = 4181
	}

	serviceLocation := fmt.Sprintf(`
    # SIWE Authentication Service
    location /siwe/ {
        proxy_pass http://127.0.0.1:%d/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location = /siwe/auth {
        internal;
        proxy_pass http://127.0.0.1:%d/session;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header Cookie $http_cookie;
        proxy_pass_request_body off;
        proxy_set_header Content-Length "";
    }
`, port, port)

	authRequest := `        auth_request /siwe/auth;
        error_page 401 = /siwe/login;
        auth_request_set $wallet_address $upstream_http_x_wallet_address;
        proxy_set_header X-Wallet-Address $wallet_address;
`

	loginPage := `
    # SIWE Login Page
    location = /siwe/login {
        default_type text/html;
        alias /var/www/siwe-login.html;
    }
`

	return &SIWENginxBlocks{
		ServiceLocation: serviceLocation,
		AuthRequest:     authRequest,
		LoginPage:       loginPage,
	}
}

// GenerateSIWELoginPage generates a simple HTML login page for SIWE
func GenerateSIWELoginPage(domain string, chainID int) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Sign In with Ethereum</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
        }
        .container {
            background: white;
            padding: 2rem;
            border-radius: 12px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            text-align: center;
            max-width: 400px;
        }
        h1 { margin-bottom: 1rem; color: #333; }
        p { color: #666; margin-bottom: 1.5rem; }
        button {
            background: #3b82f6;
            color: white;
            border: none;
            padding: 12px 24px;
            font-size: 16px;
            border-radius: 8px;
            cursor: pointer;
            width: 100%%;
        }
        button:hover { background: #2563eb; }
        button:disabled { background: #94a3b8; cursor: not-allowed; }
        .error { color: #ef4444; margin-top: 1rem; }
        .address { font-family: monospace; font-size: 14px; color: #666; margin-top: 1rem; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔐 Sign In</h1>
        <p>Connect your Ethereum wallet to continue</p>
        <button id="connect" onclick="connectWallet()">Connect Wallet</button>
        <button id="sign" onclick="signMessage()" style="display:none">Sign Message</button>
        <div id="address" class="address"></div>
        <div id="error" class="error"></div>
    </div>
    <script>
        const domain = '%s';
        const chainId = %d;
        let address = '';
        let nonce = '';

        async function connectWallet() {
            try {
                if (!window.ethereum) {
                    throw new Error('Please install MetaMask or another Web3 wallet');
                }
                const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
                address = accounts[0];
                document.getElementById('address').textContent = address;
                document.getElementById('connect').style.display = 'none';
                document.getElementById('sign').style.display = 'block';

                // Fetch nonce
                const res = await fetch('/siwe/nonce');
                const data = await res.json();
                nonce = data.nonce;
            } catch (err) {
                document.getElementById('error').textContent = err.message;
            }
        }

        async function signMessage() {
            try {
                const message = createSIWEMessage();
                const signature = await window.ethereum.request({
                    method: 'personal_sign',
                    params: [message, address]
                });

                const res = await fetch('/siwe/verify', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ message, signature })
                });

                if (res.ok) {
                    window.location.href = '/';
                } else {
                    const data = await res.json();
                    throw new Error(data.error || 'Verification failed');
                }
            } catch (err) {
                document.getElementById('error').textContent = err.message;
            }
        }

        function createSIWEMessage() {
            const now = new Date().toISOString();
            return domain + ' wants you to sign in with your Ethereum account:\n' +
                address + '\n\n' +
                'Sign in to ' + domain + '\n\n' +
                'URI: https://' + domain + '\n' +
                'Version: 1\n' +
                'Chain ID: ' + chainId + '\n' +
                'Nonce: ' + nonce + '\n' +
                'Issued At: ' + now;
        }
    </script>
</body>
</html>
`, domain, chainID)
}

// ValidateEthereumAddress checks if an address is valid
func ValidateEthereumAddress(address string) bool {
	if len(address) != 42 {
		return false
	}
	if !strings.HasPrefix(address, "0x") {
		return false
	}
	// Check if rest is hex
	_, err := hex.DecodeString(address[2:])
	return err == nil
}

// NormalizeEthereumAddress normalizes an Ethereum address to checksum format
func NormalizeEthereumAddress(address string) string {
	// Simple lowercase normalization
	// Full EIP-55 checksum would require keccak256
	return strings.ToLower(address)
}

// IsAddressAllowed checks if an address is in the allowed list
func IsAddressAllowed(address string, allowedAddresses []string) bool {
	if len(allowedAddresses) == 0 {
		return true // No restrictions
	}

	normalized := NormalizeEthereumAddress(address)
	for _, allowed := range allowedAddresses {
		if NormalizeEthereumAddress(allowed) == normalized {
			return true
		}
	}
	return false
}

// ChainNames maps chain IDs to human-readable names
var ChainNames = map[int]string{
	1:        "Ethereum Mainnet",
	5:        "Goerli Testnet",
	11155111: "Sepolia Testnet",
	137:      "Polygon Mainnet",
	80001:    "Polygon Mumbai",
	42161:    "Arbitrum One",
	10:       "Optimism",
	56:       "BNB Smart Chain",
	43114:    "Avalanche C-Chain",
	250:      "Fantom Opera",
	8453:     "Base",
}

// GetChainName returns the name for a chain ID
func GetChainName(chainID int) string {
	if name, ok := ChainNames[chainID]; ok {
		return name
	}
	return fmt.Sprintf("Chain %d", chainID)
}
