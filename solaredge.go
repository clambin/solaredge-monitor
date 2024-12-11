package main

import (
	"context"
	"github.com/clambin/solaredge-monitor/internal/cmd"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

var (
	version = "change_me"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cmd.RootCmd.SetContext(ctx)
	cmd.RootCmd.Version = version

	if err := cmd.RootCmd.Execute(); err != nil {
		slog.Error("failed to start", "err", err)
		os.Exit(1)
	}
}
