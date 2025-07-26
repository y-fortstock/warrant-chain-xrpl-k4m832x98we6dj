//go:build wireinject
// +build wireinject

// Package di provides dependency injection providers for the application using Google Wire.
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
func ProvideLogger(cfg config.LogConfig) *slog.Logger {
	return logger.NewLogger(cfg)
}

// ProvideBlockchainOrPanic returns a new Blockchain instance using the provided NetworkConfig.
// It panics if blockchain creation fails.
func ProvideBlockchainOrPanic(cfg config.NetworkConfig) *api.Blockchain {
	bc, err := api.NewBlockchain(cfg)
	if err != nil {
		slog.Error("failed to create blockchain", "error", err)
		panic(err)
	}
	return bc
}

// ProvideAccountAPI returns an implementation of the AccountAPIServer.
func ProvideAccountAPI(l *slog.Logger, bc *api.Blockchain) accountv1.AccountAPIServer {
	return api.NewAccount(l, bc)
}

// ProvideTokenAPI returns an implementation of the TokenAPIServer.
func ProvideTokenAPI(l *slog.Logger) tokenv1.TokenAPIServer {
	return api.NewToken(l)
}

// ProvideGRPCServer returns a new gRPC server with registered Account and Token APIs.
func ProvideGRPCServer(accountAPI accountv1.AccountAPIServer, tokenAPI tokenv1.TokenAPIServer) *grpc.Server {
	s := grpc.NewServer()
	accountv1.RegisterAccountAPIServer(s, accountAPI)
	tokenv1.RegisterTokenAPIServer(s, tokenAPI)
	return s
}

// ProvideAppServer returns a new application Server using the provided logger and gRPC server.
func ProvideAppServer(l *slog.Logger, grpcServer *grpc.Server) *server.Server {
	return server.NewServerWithGRPC(l, grpcServer)
}

// InitializeServer creates and initializes a new application server using dependency injection and the provided LogConfig.
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
