# pacproxy

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

A PAC-aware HTTP/HTTPS proxy with HTTP and SOCKS5 upstream support.

## Overview

**pacproxy** is a lightweight local HTTP/HTTPS proxy that applies PAC (Proxy Auto-Config) rules to dynamically route traffic either **directly** or through **upstream HTTP / SOCKS5 proxies**.

It is designed for tools that only support static `HTTP_PROXY / HTTPS_PROXY` environment variables but do not understand PAC files (e.g., AI CLI tools, SDKs, WebFetch implementations).

## Features

- **PAC Support**: Automatically fetches and parses PAC files
- **Multiple Upstream Types**: Supports DIRECT, HTTP, SOCKS5 proxies
- **HTTP/HTTPS Support**: Handles both HTTP requests and HTTPS CONNECT tunnels
- **Concurrent Connections**: Efficiently handles multiple simultaneous connections
- **Graceful Shutdown**: Clean shutdown on interrupt signals
- **Configurable Logging**: Multiple log levels (debug, info, warn, error)

## Installation

### From Source

```bash
go install github.com/mengbo/pacproxy@latest
```

### Build Locally

```bash
git clone https://github.com/mengbo/pacproxy.git
cd pacproxy
go build -o pacproxy .
```

### Using Make

```bash
make build        # Build binary
make test         # Run tests
make clean        # Clean build artifacts
```

## Usage

### Quick Start

```bash
# Start pacproxy with default settings
# Default: listens on 127.0.0.1:1080, uses http://clash-server:9090/ui/proxy.pac
./pacproxy

# Or specify custom PAC URL and listen address
./pacproxy --pac-url http://your-pac-server/proxy.pac --listen 127.0.0.1:8080
```

### Command Line Options

```
Flags:
  -h, --help               help for pacproxy
  -l, --listen string      Listen address (default "127.0.0.1:1080")
      --log-level string   Log level (debug, info, warn, error) (default "info")
  -p, --pac-url string     PAC file URL (default "http://clash-server:9090/ui/proxy.pac")
  -v, --version            version for pacproxy
```

### Configure Client Tools

Once pacproxy is running, configure your tools to use it:

```bash
# Set environment variables
export HTTP_PROXY=http://127.0.0.1:1080
export HTTPS_PROXY=http://127.0.0.1:1080
export ALL_PROXY=http://127.0.0.1:1080
```

### Usage Examples

#### curl
```bash
# Using environment variable
export HTTP_PROXY=http://127.0.0.1:1080
curl https://example.com

# Using command-line option
curl -x http://127.0.0.1:1080 https://example.com
```

#### wget
```bash
wget -e use_proxy=yes -e http_proxy=http://127.0.0.1:1080 https://example.com
```

#### Python requests
```python
import os
os.environ['HTTP_PROXY'] = 'http://127.0.0.1:1080'
os.environ['HTTPS_PROXY'] = 'http://127.0.0.1:1080'

import requests
response = requests.get('https://example.com')
```

#### Node.js / npm
```bash
export HTTP_PROXY=http://127.0.0.1:1080
export HTTPS_PROXY=http://127.0.0.1:1080
npm install
```

#### Git
```bash
git config --global http.proxy http://127.0.0.1:1080
git config --global https.proxy http://127.0.0.1:1080
```

#### Docker
```bash
# Build with proxy
docker build --build-arg HTTP_PROXY=http://127.0.0.1:1080 .

# Or configure Docker daemon
# ~/.docker/config.json
{
  "proxies": {
    "default": {
      "httpProxy": "http://127.0.0.1:1080",
      "httpsProxy": "http://127.0.0.1:1080"
    }
  }
}
```

## How It Works

1. **Client Request**: Client sends HTTP request or HTTPS CONNECT to pacproxy
2. **PAC Evaluation**: pacproxy evaluates the URL against the PAC file rules
3. **Routing Decision**: Based on PAC result (DIRECT, PROXY, SOCKS5), selects appropriate upstream
4. **Connection**: Establishes connection to target (directly or via upstream proxy)
5. **Data Transfer**: Proxies data between client and target

## Architecture

```
┌─────────────┐     HTTP/HTTPS      ┌─────────────┐
│   Client    │ ──────────────────▶ │  pacproxy   │
│  (curl/git) │                     │  (127.0.0.1)│
└─────────────┘                     └──────┬──────┘
                                           │
                    ┌──────────────────────┼──────────────────────┐
                    │                      │                      │
                    ▼                      ▼                      ▼
              ┌─────────┐           ┌──────────┐           ┌──────────┐
              │ DIRECT  │           │  HTTP    │           │  SOCKS5  │
              │         │           │  Proxy   │           │  Proxy   │
              └─────────┘           └──────────┘           └──────────┘
```

## Project Structure

```
pacproxy/
├── cmd/                    # CLI implementation
│   └── root.go
├── internal/
│   └── config/            # Configuration
│       ├── config.go
│       └── config_test.go
├── pac/                   # PAC file handling
│   ├── engine.go         # JavaScript execution
│   ├── fetcher.go        # PAC file fetching
│   ├── parser.go         # Proxy directive parsing
│   ├── engine_test.go
│   └── parser_test.go
├── proxy/                 # HTTP proxy server
│   ├── handler.go        # Request handlers
│   ├── router.go         # PAC-based routing
│   └── server.go         # HTTP server
├── upstream/              # Upstream connectors
│   ├── connector.go      # Connector interface
│   ├── direct.go         # Direct connection
│   ├── http.go           # HTTP proxy connector
│   └── socks5.go         # SOCKS5 connector
├── main.go               # Entry point
├── go.mod                # Go module
├── go.sum                # Dependencies
├── Makefile              # Build automation
├── LICENSE               # MIT License
└── README.md             # This file
```

## Development

### Running Tests

```bash
go test ./...
go test -v ./...        # Verbose output
go test -race ./...     # Race condition detection
```

### Building

```bash
# Build for current platform
go build -o pacproxy .

# Build for all platforms
make build-all

# Cross-compile examples
GOOS=linux GOARCH=amd64 go build -o pacproxy-linux-amd64 .
GOOS=darwin GOARCH=arm64 go build -o pacproxy-darwin-arm64 .
GOOS=windows GOARCH=amd64 go build -o pacproxy-windows-amd64.exe .
```

## Requirements

- Go 1.21 or later
- PAC file accessible via HTTP/HTTPS
- Upstream proxy (if PAC rules specify proxy usage)

## Limitations

- PAC file must be accessible via HTTP/HTTPS URL
- JavaScript PAC functions are executed in a sandbox with limited built-in functions
- No authentication support for PAC file fetching

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [goja](https://github.com/dop251/goja) - JavaScript engine for Go
- [cobra](https://github.com/spf13/cobra) - CLI framework
- [golang.org/x/net/proxy](https://golang.org/x/net/proxy) - SOCKS5 support

## Related Projects

- [genpac](https://github.com/JinnLynn/genpac) - PAC file generator
- [SwitchyOmega](https://github.com/FelisCatus/SwitchyOmega) - Browser extension for proxy switching
