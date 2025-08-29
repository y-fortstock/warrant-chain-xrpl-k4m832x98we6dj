// Package server provides the gRPC server implementation and related utilities.
// It handles server lifecycle management, graceful shutdown, and signal handling
// for the XRPL blockchain service.
package server

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	accountv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/account/v1"
	tokenv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/token/v1"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

// Server represents the gRPC server and its associated components.
// It manages the server lifecycle, including startup, shutdown, and signal handling.
//
// The Server struct encapsulates both the gRPC server instance and a logger
// for operational logging and debugging.
type Server struct {
	// grpcServer is the underlying gRPC server instance.
	// It handles all gRPC communication and request processing.
	grpcServer *grpc.Server

	// logger is used for operational logging and debugging.
	// It provides structured logging capabilities throughout the server lifecycle.
	logger *slog.Logger
}

// NewServer creates a new Server with its own gRPC server instance.
// This constructor is useful when you need a server with default gRPC configuration.
//
// Parameters:
// - logger: A configured logger instance for server operations
//
// Returns a new Server instance with a default gRPC server.
// The gRPC server will need to have services registered before use.
func NewServer(logger *slog.Logger) *Server {
	return &Server{
		grpcServer: grpc.NewServer(),
		logger:     logger,
	}
}

// NewServerWithGRPC creates a new Server using the provided gRPC server instance.
// This constructor is useful when you have a pre-configured gRPC server
// with services already registered.
//
// Parameters:
// - logger: A configured logger instance for server operations
// - grpcServer: A pre-configured gRPC server with services registered
//
// Returns a new Server instance using the provided gRPC server.
// This is typically used with dependency injection systems.
func NewServerWithGRPC(logger *slog.Logger, grpcServer *grpc.Server) *Server {
	return &Server{
		grpcServer: grpcServer,
		logger:     logger,
	}
}

// NewServerWithAPIs creates a new Server using the provided API server implementations.
// This constructor creates a gRPC server internally and registers the provided APIs.
// This is useful when you want to create a server directly from API implementations
// without going through the dependency injection system.
//
// Parameters:
// - logger: A configured logger instance for server operations
// - accountAPI: The account management API implementation
// - tokenAPI: The token management API implementation
//
// Returns a new Server instance with the APIs registered on an internal gRPC server.
func NewServerWithAPIs(logger *slog.Logger, accountAPI accountv1.AccountAPIServer, tokenAPI tokenv1.TokenAPIServer) *Server {
	grpcServer := grpc.NewServer()
	accountv1.RegisterAccountAPIServer(grpcServer, accountAPI)
	tokenv1.RegisterTokenAPIServer(grpcServer, tokenAPI)

	return &Server{
		grpcServer: grpcServer,
		logger:     logger,
	}
}

// Run starts the gRPC server on the specified address.
// This is a simple blocking call that starts the server and waits for it to stop.
//
// The server will listen for incoming connections on the specified address.
// This method blocks until the server stops or encounters an error.
//
// Parameters:
// - addr: The network address to listen on (e.g., ":8080", "localhost:9090")
//
// Returns an error if the server fails to start or encounters a fatal error.
// The server will continue running until manually stopped or an error occurs.
func (s *Server) Run(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.logger.Info("gRPC server listening", "addr", addr)
	return s.grpcServer.Serve(lis)
}

// RunWithGracefulShutdown starts the gRPC server and performs graceful shutdown
// on context cancellation or SIGINT/SIGTERM signal.
//
// This method provides production-ready server management with proper signal handling
// and graceful shutdown capabilities. It ensures that in-flight requests are completed
// before the server stops.
//
// The server listens for the following signals:
// - SIGINT: Interrupt signal (Ctrl+C)
// - SIGTERM: Termination signal (system shutdown)
//
// Graceful shutdown ensures that:
// - New connections are rejected
// - Existing connections are allowed to complete
// - The server stops cleanly after all requests finish
//
// Parameters:
// - ctx: Context for cancellation and timeout control
// - addr: The network address to listen on (e.g., ":8080", "localhost:9090")
//
// Returns an error if the server fails to start or encounters a fatal error.
// The server will automatically shut down when the context is cancelled or signals are received.
func (s *Server) RunWithGracefulShutdown(ctx context.Context, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.logger.Info("gRPC server listening", "addr", addr)

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.grpcServer.Serve(lis)
	})

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
