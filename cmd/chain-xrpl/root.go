package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/di"
)

var rootCmd = &cobra.Command{
	Use:   "chain-xrpl",
	Short: "XRPL blockchain service",
	RunE: func(cmd *cobra.Command, args []string) error {
		server := di.InitializeServer()
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if err := server.RunWithGracefulShutdown(ctx, ":50051"); err != nil {
			return err
		}

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to run gRPC server: %v\n", err)
		os.Exit(1)
	}
}
