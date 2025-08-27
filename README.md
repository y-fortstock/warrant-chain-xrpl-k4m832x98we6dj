# chain-xrpl

A Go-based XRPL (XRP Ledger) blockchain service that provides gRPC APIs for account management, token operations, and blockchain interactions.

## Overview

`chain-xrpl` is a microservice that acts as a bridge between applications and the XRPL network. It provides a clean, gRPC-based interface for common blockchain operations including:

- **Account Management**: Create accounts, check balances, and manage XRP transfers
- **Token Operations**: Create and manage Multi-Purpose Tokens (MPTs) for asset-backed warrants
- **Blockchain Integration**: Direct interaction with XRPL network for transaction submission and querying

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

- **API Layer**: gRPC service implementations for Account and Token operations
- **Blockchain Interface**: Abstraction layer for XRPL network interactions
- **Crypto Utilities**: BIP-44 wallet derivation and XRPL-specific key management
- **Dependency Injection**: Google Wire for compile-time dependency resolution
- **Configuration Management**: Viper-based configuration with environment variable support

## Features

### Account Management
- Create XRPL accounts from seed phrases
- Deposit XRP from system account to user accounts
- Clear account balances (return funds to system)
- Query account balances and information

### Token Operations
- Create Multi-Purpose Tokens (MPTs) for asset-backed warrants
- Transfer tokens between accounts
- Support for creditor-owner token transfers (lending scenarios)
- Token redemption and warehouse operations

### Blockchain Integration
- Direct XRPL network connectivity
- Transaction submission and signing
- Real-time balance and transaction queries
- Support for both testnet and mainnet

## Prerequisites

- Go 1.21 or later
- Access to XRPL network (testnet or mainnet)
- XRPL account with sufficient XRP for operations

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
  level: "info"
  format: "logfmt"

network:
  url: "https://s.altnet.rippletest.net:51234"
  timeout: 30
  system:
    account: "rYourSystemAccount"
    secret: "sYourSystemSecret"
    public: "YourSystemPublicKey"

server:
  listen: ":8080"
```

### Environment Variables
```bash
export LOG_LEVEL=info
export LOG_FORMAT=logfmt
export NETWORK_URL=https://s.altnet.rippletest.net:51234
export NETWORK_TIMEOUT=30
export NETWORK_SYSTEM_ACCOUNT=rYourSystemAccount
export NETWORK_SYSTEM_SECRET=sYourSystemSecret
export NETWORK_SYSTEM_PUBLIC=YourSystemPublicKey
export SERVER_LISTEN=:8080
```

## Usage

### Running the Service

1. **Simple Run**:
```bash
go run cmd/chain-xrpl/main.go
```

2. **With Graceful Shutdown**:
```bash
go run cmd/chain-xrpl/main.go --graceful
```

3. **Build and Run**:
```bash
go build -o chain-xrpl cmd/chain-xrpl/main.go
./chain-xrpl
```

### Docker

```bash
docker build -t chain-xrpl .
docker run -p 8080:8080 chain-xrpl
```

### API Usage

The service exposes gRPC APIs that can be consumed by any gRPC client:

#### Account Operations
```go
// Create account
createReq := &accountv1.CreateRequest{
    Password: "hexSeed-derivationIndex",
}
resp, err := accountClient.Create(ctx, createReq)

// Get balance
balanceReq := &accountv1.GetBalanceRequest{
    AccountId: "rAccountAddress",
}
balanceResp, err := accountClient.GetBalance(ctx, balanceReq)
```

#### Token Operations
```go
// Create token
emissionReq := &tokenv1.EmissionRequest{
    DocumentHash: "documentHash",
    WarehouseAddressId: "rWarehouseAddress",
    OwnerAddressId: "rOwnerAddress",
    Signature: "signature",
    WarehousePass: "hexSeed-derivationIndex",
}
resp, err := tokenClient.Emission(ctx, emissionReq)
```

## Development

### Project Structure
```
chain-xrpl/
├── cmd/                    # Application entry points
├── internal/              # Private application code
│   ├── api/              # gRPC API implementations
│   ├── config/           # Configuration management
│   ├── crypto/           # Cryptographic utilities
│   ├── di/               # Dependency injection
│   ├── logger/           # Logging configuration
│   └── server/           # Server implementation
├── protobuf/             # Protocol buffer definitions
├── Dockerfile            # Container configuration
├── go.mod               # Go module dependencies
└── README.md            # This file
```

### Testing

Run the test suite:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test -cover ./...
```

Run specific test packages:
```bash
go test ./internal/api/...
go test ./internal/crypto/...
```

### Code Generation

The project uses several code generation tools:

1. **Wire** (Dependency Injection):
```bash
go generate ./internal/di/...
```

2. **Protocol Buffers**:
```bash
protoc --go_out=. --go-grpc_out=. protobuf/**/*.proto
```

## Security Considerations

- **Private Keys**: Never log or expose private keys or secrets
- **Network Security**: Use HTTPS for production deployments
- **Access Control**: Implement proper authentication and authorization
- **Configuration**: Use environment variables for sensitive configuration in production

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

- [ ] Enhanced error handling and retry mechanisms
- [ ] Metrics and monitoring integration
- [ ] Rate limiting and throttling
- [ ] Multi-network support (testnet, mainnet, devnet)
- [ ] WebSocket support for real-time updates
- [ ] Enhanced security features
- [ ] Performance optimizations
