package main

import (
	"log/slog"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/tasks"
)

func initRunner(opts *Options, enclaveDir string) (tasks.Runner, error) {
	real, err := tasks.NewAnsibleRunner(enclaveDir)
	if err != nil {
		slog.Warn("task runner unavailable, tasks API disabled", "error", err)
		return applyRunnerMode(opts, nil, enclaveDir)
	}
	if err := real.Recover(); err != nil {
		slog.Warn("task recovery failed", "error", err)
	}
	return applyRunnerMode(opts, real, enclaveDir)
}
