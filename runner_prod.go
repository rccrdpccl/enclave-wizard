//go:build !dev

package main

import "github.com/rh-ecosystem-edge/enclave-wizard/internal/tasks"

type DevOptions struct{}

func applyRunnerMode(_ *Options, runner tasks.Runner, _ string) (tasks.Runner, error) {
	return runner, nil
}
