//go:build dev

package main

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/runner"
)

type DevOptions struct {
	Record         bool `help:"Record ansible runs to fixtures/recordings/" default:"false"`
	DemoDeploy     bool `help:"Demo mode: simulate deployment from recordings" default:"false"`
	DemoValidation bool `help:"Demo mode: simulate validation from recordings" default:"false"`
	Speed          int  `help:"Demo replay speed factor (0=instant, 1=real-time, 10=10x faster)" default:"10"`
}

func applyRunnerMode(opts *Options, r runner.Runner, enclaveDir string) (runner.Runner, error) {
	artifactsDir := filepath.Join(enclaveDir, "artifacts")
	recordingsDir, _ := filepath.Abs("fixtures/recordings")

	hasDemoTypes := opts.DemoDeploy || opts.DemoValidation

	if hasDemoTypes {
		replay, err := runner.NewReplayRunner(enclaveDir, recordingsDir, float64(opts.Speed))
		if err != nil {
			return nil, fmt.Errorf("replay runner init: %w", err)
		}
		slog.Info("demo mode enabled",
			"recordings", recordingsDir,
			"speed", opts.Speed,
			"deploy", opts.DemoDeploy,
			"validation", opts.DemoValidation,
		)
		return replay, nil
	}

	if r != nil && opts.Record {
		rec, err := runner.NewRecordingRunner(r, artifactsDir, recordingsDir)
		if err != nil {
			return nil, fmt.Errorf("recording runner init: %w", err)
		}
		slog.Info("record mode enabled", "recordings", recordingsDir)
		return rec, nil
	}

	return r, nil
}
