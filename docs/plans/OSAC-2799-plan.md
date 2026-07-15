STATUS: READY

# OSAC-2799: Clean up golang code — Remove `else` blocks

## 1. Goal

Remove all `else` blocks from the Go backend source code, replacing them with early returns, guard clauses, or restructured logic per the project's coding standard (CLAUDE.md: "`else` is never necessary").

## 2. Context

A full-codebase grep (`grep -rn '\belse\b' --include='*.go' --exclude='*_test.go' --exclude='mock_*.go'`) reveals exactly **2 `else` blocks** in the entire non-test Go codebase, both in the same file:

**`internal/config/writer.go:57-70`** — inside `WriteAll()`:
```go
if osacCfg != nil {
    if err := writeYAMLFile(filepath.Join(pluginsDir, "osac.yaml"), osacCfg); err != nil {
        return fmt.Errorf("writing osac.yaml: %w", err)
    }
} else {
    removeIfExists(filepath.Join(pluginsDir, "osac.yaml"))
}
if rhbkCfg != nil {
    if err := writeYAMLFile(filepath.Join(pluginsDir, "rhbk.yaml"), rhbkCfg); err != nil {
        return fmt.Errorf("writing rhbk.yaml: %w", err)
    }
} else {
    removeIfExists(filepath.Join(pluginsDir, "rhbk.yaml"))
}
```

Both follow the same pattern: write the plugin config file if the config is non-nil, otherwise remove any stale file. This can be restructured using early-action-then-continue (handle the nil case first with a guard, then proceed to the write).

**Existing test coverage:** `internal/config/config_test.go` has comprehensive tests covering both branches:
- `TestWriteAllThenReadAll_OsacPluginRoundTrips` — osac config present
- `TestWriteAllThenReadAll_RhbkPluginRoundTrips` — rhbk config present
- `TestWriteAll_NoPluginFiles_WhenFieldsEmpty` — no plugin configs (nil case)
- `TestWriteAll_RemovesPluginFilesWhenFieldsCleared` — stale file removal (else branch)

## 3. Files to Modify

- **`internal/config/writer.go`** — Restructure the two `if/else` blocks in `WriteAll()` (lines 57-70) to eliminate `else`.

## 4. New Files to Create

None.

## 5. Implementation Steps

1. **Refactor the osac config block (lines 57-63):** Replace the `if/else` with a guard clause. Handle the `nil` case first (call `removeIfExists` and skip the write), then handle the non-nil case with the write logic. The refactored code:

```go
if osacCfg == nil {
    removeIfExists(filepath.Join(pluginsDir, "osac.yaml"))
} // remove else — fall through handled by next block
```

Wait — this still uses `else` semantics implicitly. The correct no-else pattern for a two-branch if/else that doesn't return early requires restructuring as sequential guards:

```go
osacPath := filepath.Join(pluginsDir, "osac.yaml")
if osacCfg == nil {
    removeIfExists(osacPath)
}
if osacCfg != nil {
    if err := writeYAMLFile(osacPath, osacCfg); err != nil {
        return fmt.Errorf("writing osac.yaml: %w", err)
    }
}
```

This is correct but redundant. A cleaner approach: extract a helper function that allows early return:

**Best approach — inline with early-continue logic using a helper:**

Extract a `writeOrRemovePluginFile` helper that uses early return:

```go
func writeOrRemovePluginFile[T any](path string, cfg *T) error {
    if cfg == nil {
        removeIfExists(path)
        return nil
    }
    return writeYAMLFile(path, cfg)
}
```

Then the call sites become:
```go
if err := writeOrRemovePluginFile(filepath.Join(pluginsDir, "osac.yaml"), osacCfg); err != nil {
    return fmt.Errorf("writing osac.yaml: %w", err)
}
if err := writeOrRemovePluginFile(filepath.Join(pluginsDir, "rhbk.yaml"), rhbkCfg); err != nil {
    return fmt.Errorf("writing rhbk.yaml: %w", err)
}
```

However — this adds a new abstraction for just 2 call sites, which goes against the project guideline of avoiding premature abstractions. The simplest no-else approach without adding a helper is:

**Recommended approach — invert the condition and use early-return-style guard clauses:**

```go
if osacCfg == nil {
    removeIfExists(filepath.Join(pluginsDir, "osac.yaml"))
}
if osacCfg != nil {
    if err := writeYAMLFile(filepath.Join(pluginsDir, "osac.yaml"), osacCfg); err != nil {
        return fmt.Errorf("writing osac.yaml: %w", err)
    }
}
if rhbkCfg == nil {
    removeIfExists(filepath.Join(pluginsDir, "rhbk.yaml"))
}
if rhbkCfg != nil {
    if err := writeYAMLFile(filepath.Join(pluginsDir, "rhbk.yaml"), rhbkCfg); err != nil {
        return fmt.Errorf("writing rhbk.yaml: %w", err)
    }
}
```

This is correct, eliminates all `else` blocks, preserves identical behavior, and adds no abstractions.

2. **Run `go vet ./...`** to confirm no static analysis issues.

3. **Run `go test ./internal/config/...`** to confirm all existing tests pass, especially `TestWriteAll_RemovesPluginFilesWhenFieldsCleared` which exercises the nil/removal path.

4. **Run `grep -rn '\belse\b' --include='*.go' --exclude='*_test.go' --exclude='mock_*.go'`** to confirm zero `else` blocks remain in non-test Go code.

## 6. Verification Criteria

- **`go vet ./...`** — exits 0 with no warnings.
- **`go test ./internal/config/... -v`** — all tests pass, specifically:
  - `TestWriteAllThenReadAll_OsacPluginRoundTrips` — PASS
  - `TestWriteAllThenReadAll_RhbkPluginRoundTrips` — PASS
  - `TestWriteAll_NoPluginFiles_WhenFieldsEmpty` — PASS
  - `TestWriteAll_RemovesPluginFilesWhenFieldsCleared` — PASS
- **`go test ./... -count=1`** — full test suite passes (no regressions).
- **`grep -rn '\belse\b' --include='*.go' --exclude='*_test.go' --exclude='mock_*.go'`** — returns no results.
- The resulting `WriteAll()` function is functionally identical: same files are written/removed under the same conditions.

## 7. Risks and Edge Cases

- **Risk: Logic change on accident.** The two inverted `if` blocks must be mutually exclusive (one checks `== nil`, the other `!= nil`). Since `osacCfg`/`rhbkCfg` are not modified between the checks, this is safe. However, the implementer should verify that no goroutine or side effect modifies these pointers between the two `if` statements — inspection of the function confirms they are local variables set once at the top of `WriteAll()`, so this is safe.
- **Risk: Test-only `else` blocks.** The ticket says "backend API" cleanup. Test files (`*_test.go`) currently have zero `else` blocks, so no action is needed there. If the scope is intended to include test code, none exists.
- **Minimal blast radius.** Only one file changes, and it has thorough test coverage for both branches.
