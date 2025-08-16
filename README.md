# Coolify CLI

A simple command-line interface for interacting with your Coolify instance.

## Features

- 🔐 Secure API key management through configuration files
- 📋 Fetch application logs
- 🔄 Follow logs in real-time
- 🧪 Test API connectivity
- ⚙️ Easy configuration management

## Installation

1. Clone this repository
2. Build the CLI:
   ```bash
   go build -o coolify-cli
   ```

## Configuration

1. Initialize the configuration:
   ```bash
   ./coolify-cli config init
   ```

2. Edit the configuration file at `~/.coolify-cli/config.yaml`:
   ```yaml
   api_key: "your-actual-api-key-here"
   host_url: "https://app.coolify.io"
   api_path: "/api/v1"
   ```

3. Test your configuration:
   ```bash
   ./coolify-cli config test
   ```

## Usage

### View Configuration
```bash
./coolify-cli config show
```

### Fetch Application Logs
```bash
./coolify-cli logs nk4kcskcsswg0wskk88skcsg
```

### Follow Logs (Real-time)
```bash
./coolify-cli logs -f nk4kcskcsswg0wskk88skcsg
```

### Show Help
```bash
./coolify-cli --help
./coolify-cli logs --help
```

## Configuration File

The CLI uses a YAML configuration file located at `~/.coolify-cli/config.yaml`:

```yaml
# Coolify CLI Configuration
# Get your API key from your Coolify instance
api_key: "your-api-key-here"

# Your Coolify instance URL (without /api/v1)
host_url: "https://app.coolify.io"

# API path (usually /api/v1)
api_path: "/api/v1"
```

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
