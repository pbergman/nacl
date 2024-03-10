package main

import (
	"github.com/pbergman/logger"
	"github.com/pbergman/nacl/plugin"
)

type App struct {
	Config  *Config
	Plugins map[string]plugin.Plugin
	Logger  *logger.Logger
	LinkId  int
	Worker  *Worker
}

func (a *App) Close() error {
	return a.Worker.Close()
}
