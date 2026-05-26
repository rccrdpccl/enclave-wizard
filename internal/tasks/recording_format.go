package tasks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
)

type Recording struct {
	Run     models.TaskRun    `json:"run"`
	Command json.RawMessage   `json:"command,omitempty"`
	RC      int               `json:"rc"`
	Status  string            `json:"status"`
	Stdout  string            `json:"stdout"`
	Stderr  string            `json:"stderr"`
	Events  []json.RawMessage `json:"events"`
}

func SaveRecording(srcDir, destFile string) error {
	run, err := readRunJSON(srcDir)
	if err != nil {
		return fmt.Errorf("reading run.json: %w", err)
	}

	rec := Recording{Run: *run}

	if data, err := os.ReadFile(filepath.Join(srcDir, "command")); err == nil {
		rec.Command = json.RawMessage(data)
	}
	if data, err := os.ReadFile(filepath.Join(srcDir, "rc")); err == nil {
		rc, _ := strconv.Atoi(strings.TrimSpace(string(data)))
		rec.RC = rc
	}
	if data, err := os.ReadFile(filepath.Join(srcDir, "status")); err == nil {
		rec.Status = strings.TrimSpace(string(data))
	}
	if data, err := os.ReadFile(filepath.Join(srcDir, "stdout")); err == nil {
		rec.Stdout = string(data)
	}
	if data, err := os.ReadFile(filepath.Join(srcDir, "stderr")); err == nil {
		rec.Stderr = string(data)
	}

	eventsDir := filepath.Join(srcDir, "job_events")
	entries, err := os.ReadDir(eventsDir)
	if err == nil {
		sortEventEntries(entries)
		for _, entry := range entries {
			if !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(eventsDir, entry.Name()))
			if err != nil {
				continue
			}
			rec.Events = append(rec.Events, json.RawMessage(data))
		}
	}

	out, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling recording: %w", err)
	}
	return os.WriteFile(destFile, out, 0640)
}

func LoadRecording(srcFile, destDir string) error {
	data, err := os.ReadFile(srcFile)
	if err != nil {
		return fmt.Errorf("reading recording file: %w", err)
	}

	var rec Recording
	if err := json.Unmarshal(data, &rec); err != nil {
		return fmt.Errorf("parsing recording: %w", err)
	}

	if err := os.MkdirAll(destDir, 0750); err != nil {
		return err
	}

	if err := writeRunJSON(destDir, &rec.Run); err != nil {
		return fmt.Errorf("writing run.json: %w", err)
	}
	if rec.Command != nil {
		os.WriteFile(filepath.Join(destDir, "command"), []byte(rec.Command), 0640)
	}
	os.WriteFile(filepath.Join(destDir, "rc"), []byte(strconv.Itoa(rec.RC)), 0640)
	os.WriteFile(filepath.Join(destDir, "status"), []byte(rec.Status), 0640)
	os.WriteFile(filepath.Join(destDir, "stdout"), []byte(rec.Stdout), 0640)
	os.WriteFile(filepath.Join(destDir, "stderr"), []byte(rec.Stderr), 0640)

	if len(rec.Events) > 0 {
		eventsDir := filepath.Join(destDir, "job_events")
		if err := os.MkdirAll(eventsDir, 0750); err != nil {
			return err
		}
		for _, event := range rec.Events {
			var meta struct {
				UUID    string `json:"uuid"`
				Counter int    `json:"counter"`
			}
			json.Unmarshal(event, &meta)
			filename := fmt.Sprintf("%03d-%s.json", meta.Counter, meta.UUID)
			os.WriteFile(filepath.Join(eventsDir, filename), []byte(event), 0640)
		}
	}

	return nil
}
