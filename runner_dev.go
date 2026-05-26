//go:build dev

package main

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/tasks"
)

type DevOptions struct {
	Record bool `help:"Record ansible runs to fixtures/recordings/" default:"false"`
	Demo   bool `help:"Demo mode: use fake-runner to simulate ansible from recordings" default:"false"`
	Speed  int  `help:"Demo/replay speed factor (0=instant, 1=real-time, 10=10x faster)" default:"10"`
}

func applyRunnerMode(opts *Options, runner tasks.Runner, enclaveDir string) (tasks.Runner, error) {
	artifactsDir := filepath.Join(enclaveDir, "artifacts")
	recordingsDir, _ := filepath.Abs("fixtures/recordings")

	if opts.Demo {
		r, err := tasks.NewDemoRunner(enclaveDir, recordingsDir, float64(opts.Speed))
		if err != nil {
			return nil, fmt.Errorf("demo runner init: %w", err)
		}
		slog.Info("demo mode enabled", "recordings", recordingsDir, "speed", opts.Speed)
		return r, nil
	}

	if runner != nil && opts.Record {
		rec, err := tasks.NewRecordingRunner(runner, artifactsDir, recordingsDir)
		if err != nil {
			return nil, fmt.Errorf("recording runner init: %w", err)
		}
		slog.Info("record mode enabled", "recordings", recordingsDir)
		return rec, nil
	}

	return runner, nil
}
