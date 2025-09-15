# chain-xrpl

A comprehensive Go-based microservice that provides enterprise-grade gRPC APIs for XRPL (XRP Ledger) blockchain operations, specifically designed for warrant and asset-backed token management systems.

## Overview

`chain-xrpl` is a production-ready blockchain service that serves as a secure bridge between enterprise applications and the XRPL network. Built with modern Go practices and clean architecture principles, it provides robust APIs for:

- **Account Management**: Secure account creation, balance management, and XRP transfer operations
- **Multi-Purpose Token (MPT) Operations**: Advanced token creation, transfer, and redemption for asset-backed warrants
- **Blockchain Integration**: Direct XRPL network connectivity with transaction submission and real-time querying
- **Lending Operations**: Optional creditor-owner token transfer functionality for lending scenarios
- **Enterprise Features**: Comprehensive logging, configuration management, and graceful shutdown capabilities

## Architecture

The service is built using modern Go practices and follows a clean architecture pattern:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   gRPC Client  │───▶│   API Layer     │───▶│  XRPL Network   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │  Blockchain     │
                       │  Interface      │
                       └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │  Crypto Utils   │
                       │  & Wallet Mgmt  │
                       └─────────────────┘
```

### Core Components

- **API Layer**: Comprehensive gRPC service implementations for Account, Token, and Blockchain operations
- **Blockchain Interface**: Robust abstraction layer for XRPL network interactions with error handling and retry mechanisms
- **Crypto Utilities**: Secure BIP-44 wallet derivation and XRPL-specific key management with seed phrase support
- **Dependency Injection**: Google Wire for compile-time dependency resolution ensuring type safety
- **Configuration Management**: Viper-based configuration with environment variable support and feature flags
- **Logging System**: Structured logging with configurable levels and formats for production monitoring
- **Server Management**: Graceful shutdown capabilities and signal handling for production deployments

## Features

### Account Management
- **Secure Account Creation**: Generate XRPL accounts from BIP-44 compatible seed phrases with derivation index support
- **Balance Operations**: Deposit XRP from system account to user accounts with transaction validation
- **Account Cleanup**: Clear account balances and return funds to system account safely
- **Account Queries**: Real-time balance and account information retrieval with comprehensive error handling

### Token Operations
- **MPT Creation**: Create Multi-Purpose Tokens (MPTs) for asset-backed warrants with document hash validation
- **Token Transfers**: Secure token transfers between accounts with signature verification
- **Lending Support**: Optional creditor-owner token transfer functionality for lending scenarios
- **Token Redemption**: Warehouse operations and token redemption with proper authorization
- **Metadata Management**: Token metadata handling and document hash association

### Blockchain Integration
- **Network Connectivity**: Direct XRPL network connectivity with configurable endpoints
- **Transaction Management**: Secure transaction submission and signing with proper error handling
- **Real-time Queries**: Live balance and transaction status queries with caching support
- **Multi-network Support**: Seamless support for testnet, mainnet, and custom XRPL networks
- **Error Recovery**: Robust error handling with retry mechanisms and graceful degradation

## Prerequisites

- **Go 1.24+**: Latest Go version with full module support
- **XRPL Network Access**: Connection to XRPL testnet, mainnet, or custom network
- **System Account**: XRPL account with sufficient XRP for operations and transaction fees
- **Dependencies**: All required Go modules will be automatically downloaded

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd chain-xrpl
```

2. Install dependencies:
```bash
go mod download
```

3. Generate Wire dependency injection code:
```bash
go generate ./...
```

## Configuration

The service uses Viper for configuration management. Create a configuration file or set environment variables:

### Configuration File (config.yaml)
```yaml
log:
  level: "info"          # Log level: debug, info, warn, error
  format: "logfmt"       # Log format: json, logfmt

network:
  url: "https://s.altnet.rippletest.net:51234/"  # XRPL network endpoint
  timeout: 30            # Network request timeout in seconds
  system:
    account: "rYourSystemAccount"    # System XRPL account address
    secret: "sYourSystemSecret"      # System account secret key
    public: "YourSystemPublicKey"    # System account public key

server:
  listen: ":8099"        # gRPC server listen address

features:
  loan: false            # Enable lending functionality (optional)
```

### Environment Variables
```bash
# Logging configuration
export LOG_LEVEL=info
export LOG_FORMAT=logfmt

# Network configuration
export NETWORK_URL=https://s.altnet.rippletest.net:51234/
export NETWORK_TIMEOUT=30

# System account credentials (keep secure!)
export CHAIN_SYSTEM_ACCOUNT=rYourSystemAccount
export CHAIN_SYSTEM_SECRET=sYourSystemSecret
export CHAIN_SYSTEM_PUBLIC=YourSystemPublicKey

# Server configuration
export SERVER_LISTEN=:8099

# Feature flags
export FEATURES_LOAN=false
```

## Usage

### Running the Service

1. **Development Run**:
```bash
go run cmd/chain-xrpl/main.go
```

2. **With Custom Config**:
```bash
go run cmd/chain-xrpl/main.go --config=./config.production.yaml
```

3. **Production Build**:
```bash
go build -o chain-xrpl cmd/chain-xrpl/main.go
./chain-xrpl
```

4. **Docker Development**:
```bash
docker-compose up --build
```

### Docker

```bash
# Build the Docker image
docker build -t chain-xrpl .

# Run with environment variables
docker run -p 8099:8099 \
  -e CHAIN_SYSTEM_ACCOUNT=rYourSystemAccount \
  -e CHAIN_SYSTEM_SECRET=sYourSystemSecret \
  -e CHAIN_SYSTEM_PUBLIC=YourSystemPublicKey \
  chain-xrpl

# Run with config file
docker run -p 8099:8099 -v $(pwd)/config.yaml:/app/config.yaml chain-xrpl
```

### API Usage

The service exposes gRPC APIs that can be consumed by any gRPC client:

#### Account Operations
```go
// Create account from seed phrase
createReq := &accountv1.CreateRequest{
    Password: "hexSeed-derivationIndex", // Format: "64charHexSeed-derivationIndex"
}
resp, err := accountClient.Create(ctx, createReq)

// Get account balance
balanceReq := &accountv1.GetBalanceRequest{
    AccountId: "rAccountAddress",
}
balanceResp, err := accountClient.GetBalance(ctx, balanceReq)

// Deposit XRP to account
depositReq := &accountv1.DepositRequest{
    AccountId: "rAccountAddress",
    Amount:    "1000000", // Amount in drops (1 XRP = 1,000,000 drops)
}
depositResp, err := accountClient.Deposit(ctx, depositReq)
```

#### Token Operations
```go
// Create Multi-Purpose Token (MPT)
emissionReq := &tokenv1.EmissionRequest{
    DocumentHash:         "documentHash",
    WarehouseAddressId:   "rWarehouseAddress",
    OwnerAddressId:       "rOwnerAddress", 
    Signature:            "signature",
    WarehousePass:        "hexSeed-derivationIndex",
}
resp, err := tokenClient.Emission(ctx, emissionReq)

// Transfer token between accounts
transferReq := &tokenv1.TransferRequest{
    TokenId:     "tokenId",
    FromAddress: "rFromAddress",
    ToAddress:   "rToAddress",
    Amount:      "1000000",
    Signature:   "signature",
    FromPass:    "hexSeed-derivationIndex",
}
transferResp, err := tokenClient.Transfer(ctx, transferReq)
```

## Development

### Project Structure
```
chain-xrpl/
├── cmd/                    # Application entry points
│   └── chain-xrpl/        # Main application command
├── internal/              # Private application code
│   ├── api/              # gRPC API implementations
│   │   ├── account.go    # Account management API
│   │   ├── token.go      # Token operations API
│   │   └── blockchain.go # Blockchain interface
│   ├── config/           # Configuration management
│   ├── crypto/           # Cryptographic utilities & wallet management
│   ├── di/               # Dependency injection with Wire
│   ├── logger/           # Structured logging configuration
│   └── server/           # gRPC server implementation
├── proto/                # Protocol buffer definitions
├── Dockerfile            # Container configuration
├── config.yaml           # Default configuration
├── go.mod               # Go module dependencies
└── README.md            # This file
```

### Testing

Run the complete test suite:
```bash
go test ./...
```

Run tests with coverage analysis:
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

Run specific test packages:
```bash
go test ./internal/api/...      # API layer tests
go test ./internal/crypto/...   # Cryptographic utilities tests
go test ./internal/config/...   # Configuration tests
```

Run tests with verbose output:
```bash
go test -v ./...
```

### Code Generation

The project uses several code generation tools:

1. **Wire** (Dependency Injection):
```bash
go generate ./internal/di/...
```

2. **Protocol Buffers**:
```bash
# Generate Go code from proto files
protoc --go_out=. --go-grpc_out=. proto/**/*.proto

# Or use the project's protobuf generation script
cd ../protobuf && npm run generate
```

3. **Generate All**:
```bash
go generate ./...
```

## Security Considerations

- **Private Keys**: Never log or expose private keys or secrets in application logs
- **Network Security**: Use HTTPS/TLS for production deployments and secure network communication
- **Access Control**: Implement proper authentication and authorization for gRPC endpoints
- **Configuration**: Use environment variables for sensitive configuration in production environments
- **Seed Phrase Security**: Ensure secure storage and transmission of seed phrases and derivation indices
- **Transaction Signing**: All transactions are signed securely using XRPL-compatible cryptographic methods
- **Error Handling**: Sensitive information is redacted from error messages and logs

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

[Add your license information here]

## Support

For support and questions:
- Create an issue in the repository
- Contact the development team
- Check the documentation and examples

## Roadmap

- [x] **Core XRPL Integration**: Account management and token operations
- [x] **gRPC API**: Comprehensive API for blockchain operations
- [x] **Configuration Management**: Flexible configuration with environment variables
- [x] **Docker Support**: Containerized deployment options
- [x] **Structured Logging**: Production-ready logging system
- [ ] **Enhanced Error Handling**: Advanced retry mechanisms and circuit breakers
- [ ] **Metrics & Monitoring**: Prometheus metrics and health checks
- [ ] **Rate Limiting**: API rate limiting and throttling
- [ ] **WebSocket Support**: Real-time updates and notifications
- [ ] **Multi-signature Support**: Enhanced security for enterprise use cases
- [ ] **Performance Optimizations**: Connection pooling and caching
- [ ] **API Documentation**: OpenAPI/Swagger documentation generation
