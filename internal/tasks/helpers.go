package tasks

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
)

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
	return os.WriteFile(filepath.Join(runDir, "run.json"), data, 0640)
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

func generateRunID() string {
	var uuid [16]byte
	rand.Read(uuid[:])
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}
