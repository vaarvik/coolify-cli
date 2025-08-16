# Coolify CLI

A simple command-line interface for interacting with your Coolify instance.

## Features

- 🔐 Secure API key management through configuration files
- 📋 Fetch application logs
- 🔄 Follow logs in real-time
- 🧪 Test API connectivity
- ⚙️ Easy configuration management

## Installation

### Quick Install (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/vaarvik/coolify-cli/main/install.sh | bash
```

This will:
- Install the CLI to `/usr/local/bin/coolify-cli`
- Create configuration at `~/.coolify-cli/config.json`

### Manual Installation

1. Download the latest release for your platform from [GitHub Releases](https://github.com/vaarvik/coolify-cli/releases)
2. Install (replace platform with your system):
   ```bash
   # For macOS Intel (darwin-amd64)
   curl -L -o coolify-cli.tar.gz https://github.com/vaarvik/coolify-cli/releases/download/v0.0.1/coolify-cli-darwin-amd64.tar.gz

   # For macOS Apple Silicon (darwin-arm64)
   curl -L -o coolify-cli.tar.gz https://github.com/vaarvik/coolify-cli/releases/download/v0.0.1/coolify-cli-darwin-arm64.tar.gz

   # For Linux (linux-amd64)
   curl -L -o coolify-cli.tar.gz https://github.com/vaarvik/coolify-cli/releases/download/v0.0.1/coolify-cli-linux-amd64.tar.gz

   # For Windows (windows-amd64)
   curl -L -o coolify-cli.tar.gz https://github.com/vaarvik/coolify-cli/releases/download/v0.0.1/coolify-cli-windows-amd64.exe.tar.gz

   # Then for all platforms except Windows:
   tar xzf coolify-cli.tar.gz
   sudo mv coolify-cli-* /usr/local/bin/coolify-cli
   ```

### Build from Source

1. Clone this repository
2. Install system-wide:
   ```bash
   make install
   ```

   Or for development:
   ```bash
   make install-local
   ```

## Configuration

Get a **token** from your Coolify dashboard (Cloud or self-hosted) at `/security/api-tokens`

### Cloud

1. Initialize the configuration:
   ```bash
   ./coolify-cli config init
   ```

2. Add the token:
   ```bash
   ./coolify-cli instances set token cloud <token>
   ```

### Self-hosted

Add your self-hosted instance:
```bash
./coolify-cli instances add -d <name> <fqdn> <token>
```

Replace `<name>` with the name you want to give to the instance.
Replace `<fqdn>` with the fully qualified domain name of your Coolify instance.

### Change default instance

You can change the default instance with:
```bash
./coolify-cli instances set default <name>
```

### Test your configuration

```bash
./coolify-cli config test
```

## Usage

### Instance Management
```bash
# List all configured instances
./coolify-cli instances list

# Add a new self-hosted instance
./coolify-cli instances add myserver https://coolify.mycompany.com my-token-123

# Add and set as default
./coolify-cli instances add -d myserver https://coolify.mycompany.com my-token-123

# Set token for existing instance
./coolify-cli instances set token cloud my-cloud-token

# Change default instance
./coolify-cli instances set default myserver

# Remove an instance
./coolify-cli instances remove myserver
```

### View Configuration
```bash
./coolify-cli config show
```

### Fetch Application Logs
```bash
# Use default instance
./coolify-cli logs nk4kcskcsswg0wskk88skcsg

# Use specific instance
./coolify-cli logs -i myserver nk4kcskcsswg0wskk88skcsg
```

### Follow Logs (Real-time)
```bash
./coolify-cli logs -f nk4kcskcsswg0wskk88skcsg
```

### Show Help
```bash
./coolify-cli --help
./coolify-cli logs --help
./coolify-cli instances --help
```

## Configuration File

The CLI uses a JSON configuration file located at `~/.coolify-cli/config.json`:

```json
{
  "instances": [
    {
      "fqdn": "https://app.coolify.io",
      "name": "cloud",
      "token": ""
    },
    {
      "fqdn": "http://localhost:8000",
      "name": "localhost",
      "token": ""
    },
    {
      "default": true,
      "fqdn": "https://coolify.yourdomain.com",
      "name": "yourdomain",
      "token": "your-token-here"
    }
  ],
  "lastupdatechecktime": "2025-08-16T09:38:20.429802+02:00"
}
```

### Multiple Instance Support

- **instances**: Array of Coolify instances you can connect to
- **default**: Mark one instance as default (used when no instance is specified)
- **fqdn**: Full URL of your Coolify instance
- **name**: Friendly name for the instance
- **token**: API token for authentication

## API Key Security

- The configuration file is created with restricted permissions (0600)
- API keys are masked when displayed in `config show`
- Never commit your configuration file to version control

## Development

This CLI is built with:
- [Cobra](https://github.com/spf13/cobra) for CLI structure
- [Viper](https://github.com/spf13/viper) for configuration management
- Go's standard `net/http` package for API calls

## License

MIT License
