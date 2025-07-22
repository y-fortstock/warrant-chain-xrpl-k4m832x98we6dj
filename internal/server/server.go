package server

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	logger     *slog.Logger
}

func NewServer(logger *slog.Logger) *Server {
	return &Server{
		grpcServer: grpc.NewServer(),
		logger:     logger,
	}
}

func NewServerWithGRPC(logger *slog.Logger, grpcServer *grpc.Server) *Server {
	return &Server{
		grpcServer: grpcServer,
		logger:     logger,
	}
}

func (s *Server) Run(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.logger.Info("gRPC server listening", "addr", addr)
	return s.grpcServer.Serve(lis)
}

// RunWithGracefulShutdown запускает gRPC сервер и обеспечивает graceful shutdown по завершению контекста или сигналу SIGINT/SIGTERM.
func (s *Server) RunWithGracefulShutdown(ctx context.Context, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.logger.Info("gRPC server listening", "addr", addr)

	g, gctx := errgroup.WithContext(ctx)

	// Серверная горутина
	g.Go(func() error {
		return s.grpcServer.Serve(lis)
	})

	// Горутина для graceful shutdown по сигналу или отмене контекста
	g.Go(func() error {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigCh)

		select {
		case sig := <-sigCh:
			s.logger.Info("Received signal, shutting down gracefully", "signal", sig.String())
		case <-gctx.Done():
			s.logger.Info("Context cancelled, shutting down gracefully")
		}
		// Graceful shutdown
		s.grpcServer.GracefulStop()
		return nil
	})

	return g.Wait()
}
