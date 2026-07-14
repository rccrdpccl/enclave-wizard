package deploy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
)

// State provides persistent storage for deployment records as JSON files.
// Each deployment is saved as {id}.json in the configured directory.
// All operations are protected by a mutex for concurrent access safety.
type State struct {
	dir string
	mu  sync.Mutex
}

// NewState creates a State that persists deployments in dir.
func NewState(dir string) (*State, error) {
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("creating state directory: %w", err)
	}
	return &State{dir: dir}, nil
}

// Save persists a deployment using atomic write (tmp file + rename).
func (s *State) Save(dep *models.Deployment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(dep, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling deployment: %w", err)
	}

	path := filepath.Join(s.dir, dep.ID+".json")
	tmpPath := path + ".tmp"

	if err := os.WriteFile(tmpPath, data, 0640); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}

// Load reads a deployment by ID.
func (s *State) Load(id string) (*models.Deployment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.loadLocked(id)
}

func (s *State) loadLocked(id string) (*models.Deployment, error) {
	path := filepath.Join(s.dir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("deployment not found: %s", id)
		}
		return nil, fmt.Errorf("reading deployment: %w", err)
	}

	var dep models.Deployment
	if err := json.Unmarshal(data, &dep); err != nil {
		return nil, fmt.Errorf("parsing deployment: %w", err)
	}
	return &dep, nil
}

// LoadLatest returns the most recent deployment by StartedAt.
// Returns nil if no deployments exist.
func (s *State) LoadLatest() (*models.Deployment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading state directory: %w", err)
	}

	var latest *models.Deployment
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) < 6 || name[len(name)-5:] != ".json" {
			continue
		}
		// Skip tmp files from atomic writes.
		if len(name) > 9 && name[len(name)-9:] == ".json.tmp" {
			continue
		}
		id := name[:len(name)-5]
		dep, err := s.loadLocked(id)
		if err != nil {
			continue
		}
		if latest == nil || dep.StartedAt.After(latest.StartedAt) {
			latest = dep
		}
	}

	return latest, nil
}

// LoadAll returns all stored deployments, most recent first.
func (s *State) LoadAll() ([]*models.Deployment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading state directory: %w", err)
	}

	var deps []*models.Deployment
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) < 6 || name[len(name)-5:] != ".json" {
			continue
		}
		if len(name) > 9 && name[len(name)-9:] == ".json.tmp" {
			continue
		}
		id := name[:len(name)-5]
		dep, err := s.loadLocked(id)
		if err != nil {
			continue
		}
		deps = append(deps, dep)
	}

	return deps, nil
}
