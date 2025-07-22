//go:build wireinject
// +build wireinject

package di

import (
	"log/slog"
	"os"

	"github.com/google/wire"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/server"
	accountv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/account/v1"
	tokenv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/token/v1"
	"google.golang.org/grpc"
)

func ProvideLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

// Провайдеры-заглушки для gRPC сервисов (реализации должны быть определены отдельно)
func ProvideAccountAPI() accountv1.AccountAPIServer {
	panic("ProvideAccountAPI must be implemented")
}

func ProvideTokenAPI() tokenv1.TokenAPIServer {
	panic("ProvideTokenAPI must be implemented")
}

// Провайдер grpc.Server с регистрацией сервисов
func ProvideGRPCServer(accountAPI accountv1.AccountAPIServer, tokenAPI tokenv1.TokenAPIServer) *grpc.Server {
	s := grpc.NewServer()
	accountv1.RegisterAccountAPIServer(s, accountAPI)
	tokenv1.RegisterTokenAPIServer(s, tokenAPI)
	return s
}

// Провайдер для Server, принимающий готовый grpc.Server
func ProvideAppServer(logger *slog.Logger, grpcServer *grpc.Server) *server.Server {
	return server.NewServerWithGRPC(logger, grpcServer)
}

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
