package main

import (
	"context"
	"fmt"
	"os"

	"github.com/RandomThacker/donna/services/api/internal/app"
	"github.com/RandomThacker/donna/services/api/internal/config"
	"github.com/RandomThacker/donna/services/api/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.App.Environment, cfg.App.LogLevel)
	if err := app.Run(context.Background(), cfg, log); err != nil {
		log.Error("application stopped with error", "error", err)
		os.Exit(1)
	}
}
