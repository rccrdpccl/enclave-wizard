//go:build !dev

package main

import "github.com/rh-ecosystem-edge/enclave-wizard/internal/runner"

type DevOptions struct{}

func applyRunnerMode(_ *Options, r runner.Runner, _ string) (runner.Runner, error) {
	return r, nil
}
