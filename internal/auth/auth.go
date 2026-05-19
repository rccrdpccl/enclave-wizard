package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Store struct {
	passwordFile string
	changedFile  string

	mu     sync.RWMutex
	hash   []byte
	tokens map[string]time.Time
}

func NewStore(passwordFile string) *Store {
	return &Store{
		passwordFile: passwordFile,
		changedFile:  passwordFile + ".changed",
		tokens:       make(map[string]time.Time),
	}
}

// Init loads the password hash from disk, or generates a new password if none exists.
// Returns the generated plaintext password (empty string if loaded from existing file).
func (s *Store) Init() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.passwordFile)
	if err == nil && len(data) > 0 {
		s.hash = []byte(strings.TrimSpace(string(data)))
		return "", nil
	}

	plain, err := generatePassword(16)
	if err != nil {
		return "", fmt.Errorf("generate password: %w", err)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	if err := os.WriteFile(s.passwordFile, append(hashed, '\n'), 0600); err != nil {
		return "", fmt.Errorf("write password file: %w", err)
	}

	s.hash = hashed
	return plain, nil
}

func (s *Store) CheckPassword(password string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return bcrypt.CompareHashAndPassword(s.hash, []byte(password)) == nil
}

func (s *Store) ChangePassword(currentPassword, newPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if bcrypt.CompareHashAndPassword(s.hash, []byte(currentPassword)) != nil {
		return fmt.Errorf("current password is incorrect")
	}

	if len(newPassword) < 8 {
		return fmt.Errorf("new password must be at least 8 characters")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := os.WriteFile(s.passwordFile, append(hashed, '\n'), 0600); err != nil {
		return fmt.Errorf("write password file: %w", err)
	}

	s.hash = hashed

	// Mark password as changed
	os.WriteFile(s.changedFile, []byte("changed\n"), 0600)

	// Invalidate all tokens
	s.tokens = make(map[string]time.Time)

	return nil
}

func (s *Store) MustChangePassword() bool {
	_, err := os.Stat(s.changedFile)
	return err != nil
}

func (s *Store) CreateToken() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	token, err := generateToken(32)
	if err != nil {
		return "", err
	}

	s.tokens[token] = time.Now().Add(24 * time.Hour)

	// Prune expired tokens
	now := time.Now()
	for t, exp := range s.tokens {
		if now.After(exp) {
			delete(s.tokens, t)
		}
	}

	return token, nil
}

func (s *Store) ValidateToken(token string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	exp, ok := s.tokens[token]
	if !ok {
		return false
	}
	return time.Now().Before(exp)
}

// HashPassword hashes a plaintext password for the password file.
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func generatePassword(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	const charset = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}

func generateToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
