//go:build wireinject
// +build wireinject

// Package di provides dependency injection providers for the application using Google Wire.
// It defines the dependency graph and provides functions for creating and wiring
// application components together.
//
// This package uses Google Wire for compile-time dependency injection, ensuring
// that all dependencies are properly resolved at build time rather than runtime.
package di

import (
	"log/slog"

	"github.com/google/wire"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/api"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/config"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/logger"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/server"
	accountv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/account/v1"
	tokenv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/token/v1"
	"google.golang.org/grpc"
)

// ProvideLogger returns a new slog.Logger instance using the logger package and the provided LogConfig.
// This provider creates a configured logger that can be used throughout the application.
//
// Parameters:
// - cfg: Logging configuration including level and format settings
//
// Returns a configured slog.Logger instance.
func ProvideLogger(cfg config.LogConfig) *slog.Logger {
	return logger.NewLogger(cfg)
}

// ProvideBlockchainOrPanic returns a new Blockchain instance using the provided NetworkConfig.
// It panics if blockchain creation fails, which is appropriate for application startup
// where a blockchain connection is essential.
//
// This provider creates the main blockchain interface that handles all XRPL network interactions.
// It's marked as "OrPanic" because the application cannot function without blockchain connectivity.
//
// Parameters:
// - cfg: Network configuration including RPC URL, timeout, and system account details
//
// Returns a configured Blockchain instance or panics if creation fails.
func ProvideBlockchainOrPanic(cfg config.NetworkConfig) *api.Blockchain {
	bc, err := api.NewBlockchain(cfg)
	if err != nil {
		slog.Error("failed to create blockchain", "error", err)
		panic(err)
	}
	return bc
}

// ProvideAccountAPI returns an implementation of the AccountAPIServer.
// This provider creates the account management API that handles account creation,
// balance queries, and XRP transfers.
//
// Parameters:
// - l: A configured logger instance
// - bc: The blockchain interface for XRPL network operations
//
// Returns an AccountAPIServer implementation.
func ProvideAccountAPI(l *slog.Logger, bc *api.Blockchain) accountv1.AccountAPIServer {
	return api.NewAccount(l, bc)
}

// ProvideTokenAPI returns an implementation of the TokenAPIServer.
// This provider creates the token management API that handles MPT creation,
// transfers, and token lifecycle operations.
//
// Parameters:
// - l: A configured logger instance
// - bc: The blockchain interface for XRPL network operations
//
// Returns a TokenAPIServer implementation.
func ProvideTokenAPI(l *slog.Logger, bc *api.Blockchain) tokenv1.TokenAPIServer {
	return api.NewToken(l, bc)
}

// ProvideGRPCServer returns a new gRPC server with registered Account and Token APIs.
// This provider creates and configures the gRPC server, registering all available
// API services for external communication.
//
// Parameters:
// - accountAPI: The account management API implementation
// - tokenAPI: The token management API implementation
//
// Returns a configured gRPC server with all APIs registered.
func ProvideGRPCServer(accountAPI accountv1.AccountAPIServer, tokenAPI tokenv1.TokenAPIServer) *grpc.Server {
	s := grpc.NewServer()
	accountv1.RegisterAccountAPIServer(s, accountAPI)
	tokenv1.RegisterTokenAPIServer(s, tokenAPI)
	return s
}

// ProvideAppServer returns a new application Server using the provided logger and gRPC server.
// This provider creates the main application server that manages the gRPC server lifecycle
// and provides graceful shutdown capabilities.
//
// Parameters:
// - l: A configured logger instance
// - grpcServer: The configured gRPC server with registered APIs
//
// Returns an application Server instance.
func ProvideAppServer(l *slog.Logger, grpcServer *grpc.Server) *server.Server {
	return server.NewServerWithGRPC(l, grpcServer)
}

// InitializeServer creates and initializes a new application server using dependency injection
// and the provided configuration.
//
// This is the main entry point for the Wire dependency injection system.
// It defines the complete dependency graph and ensures all components are properly wired.
//
// The function uses Wire's Build function to create the dependency graph:
// - Logger → Blockchain → APIs → gRPC Server → Application Server
//
// Parameters:
// - cfg: Logging configuration for the application
// - netCfg: Network configuration for XRPL connectivity
//
// Returns a fully configured and wired application server.
func InitializeServer(cfg config.LogConfig, netCfg config.NetworkConfig) *server.Server {
	wire.Build(
		ProvideLogger,
		ProvideBlockchainOrPanic,
		ProvideAccountAPI,
		ProvideTokenAPI,
		ProvideGRPCServer,
		ProvideAppServer,
	)
	return &server.Server{}
}