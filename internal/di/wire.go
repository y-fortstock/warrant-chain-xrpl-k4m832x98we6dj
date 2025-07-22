//go:build wireinject
// +build wireinject

// Package di provides dependency injection providers for the application using Google Wire.
package di

import (
	"log/slog"

	"github.com/google/wire"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/api"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/logger"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/server"
	accountv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/account/v1"
	tokenv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/token/v1"
	"google.golang.org/grpc"
)

// ProvideLogger returns a new slog.Logger instance using the logger package.
func ProvideLogger() *slog.Logger {
	return logger.NewLogger()
}

// ProvideAccountAPI returns an implementation of the AccountAPIServer.
func ProvideAccountAPI() accountv1.AccountAPIServer {
	return api.NewAccount()
}

// ProvideTokenAPI returns an implementation of the TokenAPIServer.
func ProvideTokenAPI() tokenv1.TokenAPIServer {
	return api.NewToken()
}

// ProvideGRPCServer returns a new gRPC server with registered Account and Token APIs.
func ProvideGRPCServer(accountAPI accountv1.AccountAPIServer, tokenAPI tokenv1.TokenAPIServer) *grpc.Server {
	s := grpc.NewServer()
	accountv1.RegisterAccountAPIServer(s, accountAPI)
	tokenv1.RegisterTokenAPIServer(s, tokenAPI)
	return s
}

// ProvideAppServer returns a new application Server using the provided logger and gRPC server.
func ProvideAppServer(logger *slog.Logger, grpcServer *grpc.Server) *server.Server {
	return server.NewServerWithGRPC(logger, grpcServer)
}

// InitializeServer creates and initializes a new application server using dependency injection.
func InitializeServer() *server.Server {
	wire.Build(
		ProvideLogger,
		ProvideAccountAPI,
		ProvideTokenAPI,
		ProvideGRPCServer,
		ProvideAppServer,
	)
	return &server.Server{}
}
