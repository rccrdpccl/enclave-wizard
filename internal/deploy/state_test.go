package deploy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
)

func TestState_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	s, err := NewState(dir)
	if err != nil {
		t.Fatalf("NewState: %v", err)
	}

	now := time.Now().Truncate(time.Second)
	dep := &models.Deployment{
		ID:         "dep-1",
		Status:     models.TaskStatusRunning,
		TotalTasks: 350,
		StartedAt:  now,
		Phases: []models.DeploymentPhase{
			{Name: "main", TaskID: "task-1", Status: models.TaskStatusRunning},
			{Name: "osac", TaskID: "task-1", Status: models.TaskStatusPending},
		},
	}

	if err := s.Save(dep); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := s.Load("dep-1")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.ID != dep.ID {
		t.Errorf("ID: want %s, got %s", dep.ID, loaded.ID)
	}
	if loaded.Status != dep.Status {
		t.Errorf("Status: want %s, got %s", dep.Status, loaded.Status)
	}
	if loaded.TotalTasks != dep.TotalTasks {
		t.Errorf("TotalTasks: want %d, got %d", dep.TotalTasks, loaded.TotalTasks)
	}
	if len(loaded.Phases) != 2 {
		t.Fatalf("Phases: want 2, got %d", len(loaded.Phases))
	}
	if loaded.Phases[0].Name != "main" {
		t.Errorf("Phases[0].Name: want main, got %s", loaded.Phases[0].Name)
	}
	if loaded.Phases[1].Status != models.TaskStatusPending {
		t.Errorf("Phases[1].Status: want pending, got %s", loaded.Phases[1].Status)
	}
}

func TestState_Load_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, err := NewState(dir)
	if err != nil {
		t.Fatalf("NewState: %v", err)
	}

	_, err = s.Load("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing deployment")
	}
}

func TestState_SaveOverwrite(t *testing.T) {
	dir := t.TempDir()
	s, err := NewState(dir)
	if err != nil {
		t.Fatalf("NewState: %v", err)
	}

	dep := &models.Deployment{
		ID:     "dep-1",
		Status: models.TaskStatusRunning,
	}
	s.Save(dep)

	dep.Status = models.TaskStatusSuccessful
	s.Save(dep)

	loaded, err := s.Load("dep-1")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Status != models.TaskStatusSuccessful {
		t.Errorf("Status: want successful, got %s", loaded.Status)
	}
}

func TestState_LoadLatest(t *testing.T) {
	dir := t.TempDir()
	s, err := NewState(dir)
	if err != nil {
		t.Fatalf("NewState: %v", err)
	}

	now := time.Now()
	deps := []*models.Deployment{
		{ID: "oldest", Status: models.TaskStatusSuccessful, StartedAt: now.Add(-2 * time.Hour)},
		{ID: "newest", Status: models.TaskStatusRunning, StartedAt: now},
		{ID: "middle", Status: models.TaskStatusFailed, StartedAt: now.Add(-1 * time.Hour)},
	}
	for _, d := range deps {
		if err := s.Save(d); err != nil {
			t.Fatalf("Save(%s): %v", d.ID, err)
		}
	}

	latest, err := s.LoadLatest()
	if err != nil {
		t.Fatalf("LoadLatest: %v", err)
	}
	if latest == nil {
		t.Fatal("LoadLatest returned nil")
	}
	if latest.ID != "newest" {
		t.Errorf("LoadLatest: want newest, got %s", latest.ID)
	}
}

func TestState_LoadLatest_Empty(t *testing.T) {
	dir := t.TempDir()
	s, err := NewState(dir)
	if err != nil {
		t.Fatalf("NewState: %v", err)
	}

	latest, err := s.LoadLatest()
	if err != nil {
		t.Fatalf("LoadLatest: %v", err)
	}
	if latest != nil {
		t.Errorf("expected nil for empty state, got %+v", latest)
	}
}

func TestState_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	s, err := NewState(dir)
	if err != nil {
		t.Fatalf("NewState: %v", err)
	}

	dep := &models.Deployment{
		ID:         "atomic-test",
		Status:     models.TaskStatusRunning,
		TotalTasks: 100,
	}
	if err := s.Save(dep); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify the file is valid JSON.
	data, err := os.ReadFile(filepath.Join(dir, "atomic-test.json"))
	if err != nil {
		t.Fatalf("reading state file: %v", err)
	}
	var loaded models.Deployment
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("invalid JSON in state file: %v", err)
	}

	// Verify no tmp files are left behind.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("tmp file left behind: %s", e.Name())
		}
	}
}

func TestState_ConcurrentAccess(t *testing.T) {
	dir := t.TempDir()
	s, err := NewState(dir)
	if err != nil {
		t.Fatalf("NewState: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			dep := &models.Deployment{
				ID:         "concurrent",
				Status:     models.TaskStatusRunning,
				TotalTasks: n,
				StartedAt:  time.Now(),
			}
			s.Save(dep)
		}(i)
	}
	wg.Wait()

	// Should be able to load without corruption.
	loaded, err := s.Load("concurrent")
	if err != nil {
		t.Fatalf("Load after concurrent writes: %v", err)
	}
	if loaded == nil {
		t.Fatal("Load returned nil after concurrent writes")
	}
}

func TestState_LoadAll(t *testing.T) {
	dir := t.TempDir()
	s, err := NewState(dir)
	if err != nil {
		t.Fatalf("NewState: %v", err)
	}

	for _, id := range []string{"a", "b", "c"} {
		s.Save(&models.Deployment{ID: id, Status: models.TaskStatusSuccessful})
	}

	all, err := s.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 deployments, got %d", len(all))
	}
}

func TestState_SkipsTmpFiles(t *testing.T) {
	dir := t.TempDir()
	s, err := NewState(dir)
	if err != nil {
		t.Fatalf("NewState: %v", err)
	}

	s.Save(&models.Deployment{ID: "real", Status: models.TaskStatusSuccessful})

	// Write a stale tmp file.
	os.WriteFile(filepath.Join(dir, "stale.json.tmp"), []byte("{}"), 0640)

	all, err := s.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("expected 1 deployment (tmp file ignored), got %d", len(all))
	}
}
