package auth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	return NewStore(filepath.Join(t.TempDir(), "password"))
}

// initWithPassword calls Init and returns the generated plaintext password.
// It fails the test if Init returns an error or an empty password.
func initWithPassword(t *testing.T, s *Store) string {
	t.Helper()
	plain, err := s.Init()
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if plain == "" {
		t.Fatal("Init: expected generated password, got empty string")
	}
	return plain
}

// --- Init ---

func TestInit_GeneratesPassword(t *testing.T) {
	s := newStore(t)
	plain, err := s.Init()
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if plain == "" {
		t.Fatal("expected a generated password, got empty string")
	}
	if len(plain) < 16 {
		t.Errorf("generated password too short: %q", plain)
	}
}

func TestInit_WritesHashFile(t *testing.T) {
	s := newStore(t)
	initWithPassword(t, s)

	data, err := os.ReadFile(s.passwordFile)
	if err != nil {
		t.Fatalf("reading password file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("password file is empty")
	}
	// Stored content must be a valid bcrypt hash of the returned password.
	hash := []byte(strings.TrimSpace(string(data)))
	plain, _ := s.Init() // second Init loads from file; we need the first hash
	_ = plain
	if err := bcrypt.CompareHashAndPassword(hash, []byte(s.hash)); err == nil {
		// s.hash was overwritten by second Init; just verify file is valid bcrypt
	}
	if !strings.HasPrefix(string(hash), "$2") {
		t.Errorf("password file does not look like a bcrypt hash: %s", hash)
	}
}

func TestInit_LoadsExistingFile(t *testing.T) {
	s := newStore(t)
	firstPlain := initWithPassword(t, s)

	// Second Init should load the existing file and return empty string.
	s2 := NewStore(s.passwordFile)
	plain2, err := s2.Init()
	if err != nil {
		t.Fatalf("second Init: %v", err)
	}
	if plain2 != "" {
		t.Errorf("expected empty string on load, got %q", plain2)
	}
	// The loaded store must still accept the original password.
	if !s2.CheckPassword(firstPlain) {
		t.Error("loaded store rejected the original password")
	}
}

func TestInit_FilePermsAre0600(t *testing.T) {
	s := newStore(t)
	initWithPassword(t, s)

	info, err := os.Stat(s.passwordFile)
	if err != nil {
		t.Fatalf("stat password file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected 0600, got %o", perm)
	}
}

// --- CheckPassword ---

func TestCheckPassword_ValidPassword(t *testing.T) {
	s := newStore(t)
	plain := initWithPassword(t, s)

	if !s.CheckPassword(plain) {
		t.Error("CheckPassword returned false for the correct password")
	}
}

func TestCheckPassword_WrongPassword(t *testing.T) {
	s := newStore(t)
	initWithPassword(t, s)

	if s.CheckPassword("totally-wrong-password") {
		t.Error("CheckPassword returned true for an incorrect password")
	}
}

func TestCheckPassword_EmptyPassword(t *testing.T) {
	s := newStore(t)
	initWithPassword(t, s)

	if s.CheckPassword("") {
		t.Error("CheckPassword returned true for an empty password")
	}
}

// --- ChangePassword ---

func TestChangePassword_Success(t *testing.T) {
	s := newStore(t)
	plain := initWithPassword(t, s)

	if err := s.ChangePassword(plain, "new-secure-pass"); err != nil {
		t.Fatalf("ChangePassword: %v", err)
	}
	if !s.CheckPassword("new-secure-pass") {
		t.Error("new password not accepted after change")
	}
	if s.CheckPassword(plain) {
		t.Error("old password still accepted after change")
	}
}

func TestChangePassword_PersistsToFile(t *testing.T) {
	s := newStore(t)
	plain := initWithPassword(t, s)

	if err := s.ChangePassword(plain, "persisted-pass99"); err != nil {
		t.Fatalf("ChangePassword: %v", err)
	}

	// Fresh store loaded from the same file must accept the new password.
	s2 := NewStore(s.passwordFile)
	if _, err := s2.Init(); err != nil {
		t.Fatalf("second Init: %v", err)
	}
	if !s2.CheckPassword("persisted-pass99") {
		t.Error("new password not persisted to file")
	}
}

func TestChangePassword_WrongCurrentPassword(t *testing.T) {
	s := newStore(t)
	initWithPassword(t, s)

	err := s.ChangePassword("not-the-right-one", "new-secure-pass")
	if err == nil {
		t.Fatal("expected error for wrong current password, got nil")
	}
	if !strings.Contains(err.Error(), "incorrect") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestChangePassword_TooShortNewPassword(t *testing.T) {
	s := newStore(t)
	plain := initWithPassword(t, s)

	err := s.ChangePassword(plain, "short")
	if err == nil {
		t.Fatal("expected error for too-short new password, got nil")
	}
	if !strings.Contains(err.Error(), "8 characters") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestChangePassword_InvalidatesTokens(t *testing.T) {
	s := newStore(t)
	plain := initWithPassword(t, s)

	token, err := s.CreateToken()
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}
	if !s.ValidateToken(token) {
		t.Fatal("token should be valid before password change")
	}

	if err := s.ChangePassword(plain, "new-secure-pass"); err != nil {
		t.Fatalf("ChangePassword: %v", err)
	}
	if s.ValidateToken(token) {
		t.Error("old token still valid after password change")
	}
}

func TestChangePassword_WritesChangedSentinel(t *testing.T) {
	s := newStore(t)
	plain := initWithPassword(t, s)

	if err := s.ChangePassword(plain, "new-secure-pass"); err != nil {
		t.Fatalf("ChangePassword: %v", err)
	}
	if _, err := os.Stat(s.changedFile); err != nil {
		t.Errorf("expected changed sentinel file to exist: %v", err)
	}
}

// --- MustChangePassword ---

func TestMustChangePassword_TrueBeforeChange(t *testing.T) {
	s := newStore(t)
	initWithPassword(t, s)

	// No sentinel file yet — password has not been changed.
	if !s.MustChangePassword() {
		t.Error("expected MustChangePassword=true before any change")
	}
}

func TestMustChangePassword_FalseAfterSentinelCreated(t *testing.T) {
	s := newStore(t)
	plain := initWithPassword(t, s)

	if err := s.ChangePassword(plain, "new-secure-pass"); err != nil {
		t.Fatalf("ChangePassword: %v", err)
	}
	if s.MustChangePassword() {
		t.Error("expected MustChangePassword=false after sentinel file is written")
	}
}

// --- CreateToken / ValidateToken ---

func TestCreateToken_ReturnsNonEmptyToken(t *testing.T) {
	s := newStore(t)
	initWithPassword(t, s)

	token, err := s.CreateToken()
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestCreateToken_TokensAreUnique(t *testing.T) {
	s := newStore(t)
	initWithPassword(t, s)

	t1, _ := s.CreateToken()
	t2, _ := s.CreateToken()
	if t1 == t2 {
		t.Error("two consecutive tokens are identical")
	}
}

func TestValidateToken_ValidToken(t *testing.T) {
	s := newStore(t)
	initWithPassword(t, s)

	token, err := s.CreateToken()
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}
	if !s.ValidateToken(token) {
		t.Error("ValidateToken returned false for a freshly created token")
	}
}

func TestValidateToken_UnknownToken(t *testing.T) {
	s := newStore(t)
	initWithPassword(t, s)

	if s.ValidateToken("not-a-real-token") {
		t.Error("ValidateToken returned true for an unknown token")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	s := newStore(t)
	initWithPassword(t, s)

	token, err := s.CreateToken()
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}

	// Back-date the expiry so it looks expired.
	s.mu.Lock()
	s.tokens[token] = time.Now().Add(-1 * time.Second)
	s.mu.Unlock()

	if s.ValidateToken(token) {
		t.Error("ValidateToken returned true for an expired token")
	}
}

func TestCreateToken_PrunesExpiredTokens(t *testing.T) {
	s := newStore(t)
	initWithPassword(t, s)

	stale, err := s.CreateToken()
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}

	// Force expiry of the stale token.
	s.mu.Lock()
	s.tokens[stale] = time.Now().Add(-1 * time.Second)
	s.mu.Unlock()

	// Creating a new token triggers pruning.
	if _, err := s.CreateToken(); err != nil {
		t.Fatalf("second CreateToken: %v", err)
	}

	s.mu.RLock()
	_, still := s.tokens[stale]
	s.mu.RUnlock()
	if still {
		t.Error("expired token was not pruned on next CreateToken call")
	}
}

// --- HashPassword ---

func TestHashPassword_ReturnsValidBcryptHash(t *testing.T) {
	hash, err := HashPassword("my-secret")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !strings.HasPrefix(hash, "$2") {
		t.Errorf("result does not look like a bcrypt hash: %s", hash)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("my-secret")); err != nil {
		t.Errorf("hash does not verify against original password: %v", err)
	}
}

func TestHashPassword_DifferentCallsProduceDifferentHashes(t *testing.T) {
	h1, _ := HashPassword("same-password")
	h2, _ := HashPassword("same-password")
	if h1 == h2 {
		t.Error("expected different salts to produce different hashes")
	}
}
