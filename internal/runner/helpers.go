package runner

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
)

func artifactGet(artifactsDir, id string) (*models.TaskRun, error) {
	runDir := filepath.Join(artifactsDir, id)
	run, err := readRunJSON(runDir)
	if err != nil {
		return nil, ErrNotFound
	}
	return run, nil
}

func artifactList(artifactsDir string) ([]models.TaskRun, error) {
	entries, err := os.ReadDir(artifactsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.TaskRun{}, nil
		}
		return nil, err
	}

	var runs []models.TaskRun
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		run, err := artifactGet(artifactsDir, entry.Name())
		if err != nil {
			continue
		}
		runs = append(runs, *run)
	}

	sort.Slice(runs, func(i, j int) bool {
		return runs[i].StartedAt.After(runs[j].StartedAt)
	})

	return runs, nil
}

func artifactLogs(artifactsDir, id string) ([]byte, error) {
	runDir := filepath.Join(artifactsDir, id)
	if _, err := readRunJSON(runDir); err != nil {
		return nil, ErrNotFound
	}
	data, err := os.ReadFile(filepath.Join(runDir, "stdout"))
	if err != nil {
		if os.IsNotExist(err) {
			return []byte{}, nil
		}
		return nil, err
	}
	return data, nil
}

func artifactEvents(artifactsDir, id string) ([]json.RawMessage, error) {
	runDir := filepath.Join(artifactsDir, id)
	if _, err := readRunJSON(runDir); err != nil {
		return nil, ErrNotFound
	}

	eventsDir := filepath.Join(runDir, "job_events")
	entries, err := os.ReadDir(eventsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []json.RawMessage{}, nil
		}
		return nil, err
	}

	sortEventEntries(entries)

	var events []json.RawMessage
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(eventsDir, entry.Name()))
		if err != nil {
			continue
		}
		events = append(events, json.RawMessage(data))
	}

	return events, nil
}

func readRunJSON(runDir string) (*models.TaskRun, error) {
	data, err := os.ReadFile(filepath.Join(runDir, "run.json"))
	if err != nil {
		return nil, err
	}
	var run models.TaskRun
	if err := json.Unmarshal(data, &run); err != nil {
		return nil, err
	}
	return &run, nil
}

func writeRunJSON(runDir string, run *models.TaskRun) error {
	data, err := json.MarshalIndent(run, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(runDir, "run.json")
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0640); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return nil
}

func readAnsibleRunnerStatus(runDir string) string {
	data, err := os.ReadFile(filepath.Join(runDir, "status"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func readAnsibleRunnerRC(runDir string) (int, error) {
	data, err := os.ReadFile(filepath.Join(runDir, "rc"))
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

func processAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

func sortEventEntries(entries []os.DirEntry) {
	sort.Slice(entries, func(i, j int) bool {
		ni, _ := strconv.Atoi(strings.SplitN(entries[i].Name(), "-", 2)[0])
		nj, _ := strconv.Atoi(strings.SplitN(entries[j].Name(), "-", 2)[0])
		return ni < nj
	})
}

func generateRunID() string {
	var uuid [16]byte
	rand.Read(uuid[:])
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}
