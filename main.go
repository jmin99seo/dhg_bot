package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jm199seo/dhg_bot/cmd"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	if err := cmd.Execute(ctx); err != nil {
		os.Exit(1)
	}
	<-ctx.Done()
	os.Exit(0)
}
