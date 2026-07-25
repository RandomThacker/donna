package main

import (
	"context"
	"fmt"
	"os"

	"github.com/RandomThacker/donna/services/api/internal/app"
	"github.com/RandomThacker/donna/services/api/internal/config"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		// Bootstrap only: Logger Factory is not available until config loads.
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	logFactory := logger.NewFactory(logger.Options{
		Service:     constant.ServiceAPI,
		Environment: cfg.App.Environment,
		Level:       cfg.App.LogLevel,
	})
	appLog := logFactory.Module(constant.ModuleApp)

	if err := app.Run(context.Background(), cfg, logFactory); err != nil {
		appLog.Error(context.Background(), "application stopped with error", constant.LogAttrError, err)
		os.Exit(1)
	}
}
