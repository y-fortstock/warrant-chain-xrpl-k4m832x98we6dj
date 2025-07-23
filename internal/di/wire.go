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

// ProvideAccountAPI returns an implementation of the AccountAPIServer.
func ProvideAccountAPI() accountv1.AccountAPIServer {
	return api.NewAccount()
}

// ProvideTokenAPI returns an implementation of the TokenAPIServer.
func ProvideTokenAPI(logger *slog.Logger) tokenv1.TokenAPIServer {
	return api.NewToken(logger)
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
func InitializeServer(cfg config.LogConfig) *server.Server {
	wire.Build(
		ProvideLogger,
		ProvideAccountAPI,
		ProvideTokenAPI,
		ProvideGRPCServer,
		ProvideAppServer,
	)
	return &server.Server{}
}
