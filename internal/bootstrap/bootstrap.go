package bootstrap

import (
	"log/slog"

	"github.com/alexperezortuno/portx/internal/cli"
	"github.com/alexperezortuno/portx/internal/config"
	"github.com/alexperezortuno/portx/internal/logger"
	"github.com/alexperezortuno/portx/internal/runtime"
)

type Bootstrap struct {
	root    *cli.RootCommand
	runtime *runtime.Runtime
	config  *config.Config
}

func New() (*Bootstrap, error) {
	root := cli.NewRoot()

	cfg, err := config.Load("")
	if err != nil {
		return nil, err
	}

	logLevel := parseLogLevel(cfg.LogLevel)
	log := logger.New(nil, logLevel)

	rt := runtime.New(
		runtime.WithLogger(log),
		runtime.WithConfig(cfg),
	)

	b := &Bootstrap{
		root:    root,
		runtime: rt,
		config:  cfg,
	}

	root.SetConfig(cfg)

	return b, nil
}

func (a *Bootstrap) Run(args []string) error {
	return a.root.Execute(args)
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
