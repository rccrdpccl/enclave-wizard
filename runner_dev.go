//go:build dev

package main

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/tasks"
)

type DevOptions struct {
	Record         bool `help:"Record ansible runs to fixtures/recordings/" default:"false"`
	DemoDeploy     bool `help:"Demo mode: simulate deployment from recordings" default:"false"`
	DemoValidation bool `help:"Demo mode: simulate validation from recordings" default:"false"`
	Speed          int  `help:"Demo replay speed factor (0=instant, 1=real-time, 10=10x faster)" default:"10"`
}

func applyRunnerMode(opts *Options, runner tasks.Runner, enclaveDir string) (tasks.Runner, error) {
	artifactsDir := filepath.Join(enclaveDir, "artifacts")
	recordingsDir, _ := filepath.Abs("fixtures/recordings")

	demoTypes := map[models.TaskType]bool{}
	if opts.DemoDeploy {
		demoTypes[models.TaskTypeDeploy] = true
		demoTypes[models.TaskTypeDeployPhase] = true
		demoTypes[models.TaskTypeDeployPlugin] = true
	}
	if opts.DemoValidation {
		demoTypes[models.TaskTypeValidate] = true
	}

	if len(demoTypes) > 0 {
		r, err := tasks.NewDemoRunner(enclaveDir, recordingsDir, float64(opts.Speed), demoTypes)
		if err != nil {
			return nil, fmt.Errorf("demo runner init: %w", err)
		}
		slog.Info("demo mode enabled",
			"recordings", recordingsDir,
			"speed", opts.Speed,
			"deploy", opts.DemoDeploy,
			"validation", opts.DemoValidation,
		)
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
