# Hyperledger Fabric Chaincode Template (Go)

This repository serves as a template for developing Hyperledger Fabric chaincodes using Go. It provides a basic structure and configuration for building smart contracts that can be deployed to a Hyperledger Fabric network.

## Project Structure

```
chaincode-fabric-go-tmpl/
├── chaincode/
│   └── contract.go      # Main chaincode contract implementation
├── Dockerfile          # Container definition for chaincode deployment
├── go.mod             # Go module dependencies
├── go.sum             # Go module checksums
└── main.go            # Entry point and server configuration
```

## Prerequisites

- Go 1.23 or later
- [Air](https://github.com/cosmtrek/air) for live reloading during development
- Docker (for building deployment images)
- Access to a Hyperledger Fabric network

## Getting Started

1. Clone this template:
   ```bash
   git clone <your-repo-url>
   cd chaincode-fabric-go-tmpl
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Install Air for development:
   ```bash
   go install github.com/cosmtrek/air@latest
   ```

## Development with Air

Air enables automatic rebuilding of your chaincode as you make changes. To use Air:

1. Create a `.air.toml` configuration file in your project root:
   ```toml
   root = "."
   tmp_dir = "tmp"

   [build]
   cmd = "go build -o ./tmp/main ."
   bin = "./tmp/main"
   include_ext = ["go"]
   exclude_dir = ["tmp"]
   delay = 1000 # ms

   [log]
   time = true
   ```

2. Run Air:
   ```bash
   air
   ```

Air will now watch your Go files and automatically rebuild the project when changes are detected.

## Environment Variables

The chaincode server requires several environment variables to be set:

```bash
CORE_CHAINCODE_ID=your-chaincode-id
CORE_CHAINCODE_ADDRESS=:7052
CHAINCODE_TLS_DISABLED=true  # Set to false in production
```

For TLS configuration (when enabled):
```bash
CHAINCODE_TLS_KEY=path/to/key
CHAINCODE_TLS_CERT=path/to/cert
CHAINCODE_CLIENT_CA_CERT=path/to/ca-cert
```

## Building for Production

Build the Docker image:
```bash
docker build -t your-org/chaincode-name:version .
```


## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

