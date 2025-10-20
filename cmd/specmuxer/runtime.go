package main

import (
	"fmt"

	"github.com/bagaking/specmuxer/pkg/adapters"
	"github.com/bagaking/specmuxer/pkg/orchestrator/config"
	"github.com/bagaking/specmuxer/pkg/orchestrator/resume"
	"github.com/bagaking/specmuxer/pkg/orchestrator/run"
	"github.com/bagaking/specmuxer/pkg/runtime/storage"
	"github.com/bagaking/specmuxer/pkg/runtime/tmux"
	"github.com/bagaking/specmuxer/pkg/telemetry/logs"
	"github.com/bagaking/specmuxer/pkg/telemetry/stats"
)

type runtimeDeps struct {
	config    *config.Config
	store     *storage.YAMLStore
	logs      *logs.Manager
	registry  *adapters.Registry
	tmux      *tmux.Client
	runSvc    *run.Service
	resumeSvc *resume.Service
	collector *stats.Collector
}

func loadRuntime() (*runtimeDeps, error) {
	cfg, err := config.Load(globalOpts.Workspace, globalOpts.Config)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	store, err := storage.NewYAMLStore(cfg.Paths().SessionsDir, cfg.Paths().StatsPath)
	if err != nil {
		return nil, fmt.Errorf("init store: %w", err)
	}

	logManager, err := logs.NewManager(cfg.Paths().LogsDir)
	if err != nil {
		return nil, fmt.Errorf("init logs: %w", err)
	}

	registry := adapters.NewRegistry()
	tmuxClient := tmux.New()
	runSvc := run.NewService(run.ServiceConfig{
		Config:     cfg,
		Store:      store,
		LogManager: logManager,
		Registry:   registry,
		Tmux:       tmuxClient,
	})
	resumeSvc := resume.NewService(resume.ServiceConfig{
		Config:     cfg,
		Store:      store,
		LogManager: logManager,
		Registry:   registry,
		Tmux:       tmuxClient,
	})

	projectPaths := map[string]string{}
	if cfg != nil {
		projectPaths[cfg.ProjectID] = cfg.WorkspaceRoot
	}

	return &runtimeDeps{
		config:    cfg,
		store:     store,
		logs:      logManager,
		registry:  registry,
		tmux:      tmuxClient,
		runSvc:    runSvc,
		resumeSvc: resumeSvc,
		collector: stats.NewCollector(stats.WithProjectPaths(projectPaths)),
	}, nil
}
