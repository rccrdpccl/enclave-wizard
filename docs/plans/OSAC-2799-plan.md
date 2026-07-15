STATUS: READY

# OSAC-2799: Clean up golang code

## 1. Goal

Eliminate all `else` blocks in Go source files (violating the project's CLAUDE.md convention), replace raw `fmt.Print*` calls in library/server code with structured `slog` logging, use `errors.Is` for error identity checks instead of string matching, and remove duplicated helper functions across packages.

## 2. Context

### `else` block violations (CLAUDE.md rule: "else is never necessary")

- `internal/config/writer.go:57-69` — Two `if/else` blocks in `WriteAll` that write plugin config or remove existing files. The `else` branches call `removeIfExists`.

### Raw `fmt.Print*` in library/server code (should use `slog`)

All of these are in non-CLI, non-startup-banner contexts where structured logging is the project norm:

- `internal/config/writer.go:77` — `fmt.Printf("WARNING: ...")` in `removeIfExists`
- `internal/validation/validator.go:25,31,38,53,77` — `fmt.Println("WARNING: ...")` and `fmt.Println("Schema validation enabled")` and `fmt.Println("Patched ...")` and `fmt.Println("Created ...")`

Note: `main.go` uses `fmt.Print*` for CLI startup banners (password display, listening address, shutdown message) and `cmd/schemagen/main.go` for CLI tool output — these are appropriate and should be left as-is.

### String-based error check instead of `errors.Is`

- `internal/api/stream.go:96-98` — `isNotFound` function uses `strings.Contains(err.Error(), "not found")` instead of `errors.Is(err, runner.ErrNotFound)`.

### Duplicated `addonPluginsFromConfig` function

- `internal/api/tasks.go:203-231` — `TasksHandler.addonPluginsFromConfig()`
- `internal/deploy/deployer.go:233-261` — `Deployer.addonPluginsFromConfig()`

These two methods are identical in logic: read config, filter for addon plugins, sort by order, return names. This duplication should be extracted to a shared function.

### Duplicated progress calculation

- `internal/api/deploy.go:235-280` — `DeployHandler.calculateProgress()`
- `internal/deploy/deployer.go:149-200` — `Deployer.GetProgress()`

Near-identical logic for counting events and computing deployment progress percentage. Should be extracted to a shared function.

## 3. Files to Modify

| File | Rationale |
|------|-----------|
| `internal/config/writer.go` | Remove two `else` blocks; replace `fmt.Printf` in `removeIfExists` with `slog.Warn` |
| `internal/validation/validator.go` | Replace all `fmt.Println` calls with `slog` equivalents |
| `internal/api/stream.go` | Replace `isNotFound` string-matching with `errors.Is(err, runner.ErrNotFound)` |
| `internal/api/tasks.go` | Replace `addonPluginsFromConfig` method with call to shared helper |
| `internal/api/deploy.go` | Replace `calculateProgress` with call to shared helper |
| `internal/deploy/deployer.go` | Replace `addonPluginsFromConfig` and progress calculation with calls to shared helpers |

## 4. New Files to Create

| File | Purpose |
|------|---------|
| `internal/deploy/addons.go` | Shared `AddonPluginsFromConfig(reader, registry)` function extracted from the duplicated method |
| `internal/deploy/progress.go` | Shared `CalculateProgress(dep, events)` function extracted from the duplicated progress calculation |

## 5. Implementation Steps

### Step 1: Remove `else` blocks in `internal/config/writer.go`

In `WriteAll()`, lines 57-69, restructure the two `if/else` blocks to eliminate `else`. The current pattern is:

```go
if osacCfg != nil {
    if err := writeYAMLFile(...); err != nil {
        return ...
    }
} else {
    removeIfExists(...)
}
```

Refactor to:

```go
if osacCfg == nil {
    removeIfExists(filepath.Join(pluginsDir, "osac.yaml"))
} else if err := writeYAMLFile(filepath.Join(pluginsDir, "osac.yaml"), osacCfg); err != nil {
    return fmt.Errorf("writing osac.yaml: %w", err)
}
```

Wait — that introduces `else if`. Better approach using early action + continue/return style:

Since both branches are in a sequential block (not a loop), use two separate `if` blocks with negated conditions:

```go
if osacCfg == nil {
    removeIfExists(filepath.Join(pluginsDir, "osac.yaml"))
}
if osacCfg != nil {
    if err := writeYAMLFile(filepath.Join(pluginsDir, "osac.yaml"), osacCfg); err != nil {
        return fmt.Errorf("writing osac.yaml: %w", err)
    }
}
```

Or more idiomatic — use a helper approach with early handling:

Actually the cleanest approach given the project style: handle the nil case first with early action, then handle the non-nil case. Since these don't return early on the nil path, use a `switch` or simply restructure:

```go
if osacCfg != nil {
    if err := writeYAMLFile(filepath.Join(pluginsDir, "osac.yaml"), osacCfg); err != nil {
        return fmt.Errorf("writing osac.yaml: %w", err)
    }
}
if osacCfg == nil {
    removeIfExists(filepath.Join(pluginsDir, "osac.yaml"))
}
```

Apply the same pattern for `rhbkCfg`.

### Step 2: Replace `fmt.Printf` with `slog.Warn` in `internal/config/writer.go`

In `removeIfExists` (line 77), change:
```go
fmt.Printf("WARNING: failed to remove %s: %v\n", filepath.Base(path), err)
```
to:
```go
slog.Warn("failed to remove plugin config file", "file", filepath.Base(path), "error", err)
```

Add `"log/slog"` to the import block and remove `"fmt"` if it's no longer used (it is still used by `fmt.Errorf`, so keep it).

### Step 3: Replace `fmt.Println` with `slog` in `internal/validation/validator.go`

Replace these calls:

| Line | Current | Replacement |
|------|---------|-------------|
| 25 | `fmt.Println("WARNING: schema validation unavailable (task runner not available)")` | `slog.Warn("schema validation unavailable", "reason", "task runner not available")` |
| 31 | `fmt.Println("WARNING: schema validation unavailable (playbook not found)")` | `slog.Warn("schema validation unavailable", "reason", "playbook not found")` |
| 38 | `fmt.Println("Schema validation enabled")` | `slog.Info("schema validation enabled")` |
| 53 | `fmt.Println("Patched validation playbook: disabled no_log for error visibility")` | `slog.Info("patched validation playbook", "change", "disabled no_log for error visibility")` |
| 77 | `fmt.Println("Created plugin validation playbook wrapper")` | `slog.Info("created plugin validation playbook wrapper")` |

Update imports: add `"log/slog"`, remove `"fmt"` (still needed for `fmt.Sprintf` on lines 105, 115, 195 — keep it).

### Step 4: Fix `isNotFound` in `internal/api/stream.go` to use `errors.Is`

Replace:
```go
func isNotFound(err error) bool {
    return err != nil && strings.Contains(err.Error(), "not found")
}
```
with:
```go
func isNotFound(err error) bool {
    return errors.Is(err, runner.ErrNotFound)
}
```

Update imports: add `"errors"`, add `"github.com/rh-ecosystem-edge/enclave-wizard/internal/runner"`, remove `"strings"` (still used by `extractTaskID` — keep it).

### Step 5: Extract shared `AddonPluginsFromConfig` into `internal/deploy/addons.go`

Create `internal/deploy/addons.go` with:

```go
package deploy

import (
    "sort"

    "github.com/rh-ecosystem-edge/enclave-wizard/internal/config"
    "github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
    "github.com/rh-ecosystem-edge/enclave-wizard/internal/plugins"
)

func AddonPluginsFromConfig(reader *config.Reader, registry *plugins.Registry) []string {
    if reader == nil {
        return nil
    }
    cfg, err := reader.ReadAll()
    if err != nil {
        return nil
    }
    type addonInfo struct {
        name  string
        order int
    }
    var addons []addonInfo
    for _, name := range cfg.Global.EnabledPlugins {
        p, ok := registry.Get(name)
        if !ok {
            continue
        }
        if p.Type == models.PluginTypeAddon {
            addons = append(addons, addonInfo{name: p.Name, order: p.Order})
        }
    }
    sort.Slice(addons, func(i, j int) bool { return addons[i].order < addons[j].order })
    result := make([]string, len(addons))
    for i, a := range addons {
        result[i] = a.name
    }
    return result
}
```

Then update callers:
- `internal/deploy/deployer.go`: Replace `d.addonPluginsFromConfig()` calls with `AddonPluginsFromConfig(d.configReader, d.registry)` and remove the private `addonPluginsFromConfig` method.
- `internal/api/tasks.go`: Replace `h.addonPluginsFromConfig()` call with `deploy.AddonPluginsFromConfig(h.configReader, h.registry)` and remove the private `addonPluginsFromConfig` method. Add import for `"github.com/rh-ecosystem-edge/enclave-wizard/internal/deploy"`.

### Step 6: Extract shared `CalculateProgress` into `internal/deploy/progress.go`

Create `internal/deploy/progress.go` with:

```go
package deploy

import (
    "encoding/json"
    "strings"

    "github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
)

func CalculateProgress(dep *models.Deployment, events []json.RawMessage) models.DeploymentProgress {
    completedTasks := 0
    currentTask := ""

    for _, raw := range events {
        var ev struct {
            Event     string `json:"event"`
            EventData struct {
                Task string `json:"task"`
            } `json:"event_data"`
        }
        if json.Unmarshal(raw, &ev) != nil {
            continue
        }
        switch {
        case strings.HasPrefix(ev.Event, "runner_on_"):
            completedTasks++
            currentTask = ev.EventData.Task
        case ev.Event == "playbook_on_task_start" && ev.EventData.Task != "":
            currentTask = ev.EventData.Task
        }
    }

    total := dep.TotalTasks
    if total == 0 {
        total = 350
    }
    pct := completedTasks * 100 / total
    if pct > 99 && dep.Status == models.TaskStatusRunning {
        pct = 99
    }
    if dep.Status == models.TaskStatusSuccessful {
        pct = 100
    }

    return models.DeploymentProgress{
        Completed:    completedTasks,
        Total:        total,
        Percentage:   pct,
        CurrentPhase: "",
        CurrentTask:  currentTask,
    }
}
```

Then update callers:
- `internal/deploy/deployer.go:GetProgress` — call `CalculateProgress(dep, events)` instead of inline logic.
- `internal/api/deploy.go:calculateProgress` — call `deploy.CalculateProgress(dep, events)` instead of inline logic. Add import for `"github.com/rh-ecosystem-edge/enclave-wizard/internal/deploy"` (if not already present — it isn't). Remove unused imports (`"encoding/json"`, `"strings"`) from `deploy.go` if they become unused.

## 6. Verification Criteria

1. **Build**: `go build ./...` succeeds (note: requires either `ui/apps/wizard/dist` to exist for the embed directive, or use `go build -tags dev ./internal/...` to skip main).

2. **Tests**: `go test ./internal/...` passes. All existing tests should continue to pass with no modifications needed since the refactoring preserves all behavior.

3. **Vet**: `go vet ./internal/...` reports no issues.

4. **Grep verification**:
   - `grep -rn 'else {' --include='*.go' | grep -v '_test.go' | grep -v mock_` returns no results.
   - `grep -rn 'fmt\.Print' --include='*.go' | grep -v '_test.go' | grep -v mock_ | grep -v 'cmd/' | grep -v 'main.go'` returns no results (all non-CLI `fmt.Print` eliminated).
   - `grep -rn 'strings.Contains.*err.*Error.*not found' --include='*.go'` returns no results.

5. **No duplicated functions**: `grep -rn 'func.*addonPluginsFromConfig' --include='*.go'` returns exactly one result (the new shared function in `internal/deploy/addons.go`).

## 7. Risks and Edge Cases

1. **Circular import risk**: `internal/api/tasks.go` calling `deploy.AddonPluginsFromConfig` introduces a dependency from `api` → `deploy`. Currently `api` does NOT import `deploy`, but `deploy` imports `config`, `models`, `plugins`, `runner` — none of which import `api`, so there is no circular dependency risk. Verify by checking that `deploy` does not import `api`.

2. **`isNotFound` behavior change**: The current `isNotFound` in `stream.go` catches *any* error containing "not found" (including wrapped errors with that substring). Replacing it with `errors.Is(err, runner.ErrNotFound)` is more precise but could miss errors from other layers. In practice, `runner.Stream()` only returns `runner.ErrNotFound` as its sentinel, so this is safe. The `runner.Runner` interface's `Stream` method only returns `ErrNotFound` or nil.

3. **Progress calculation divergence**: The `Deployer.GetProgress` version has a slightly wider event data struct (includes `Play` field) while `DeployHandler.calculateProgress` does not. The `Play` field is parsed but never used, so the shared function can safely omit it. However, if the `Deployer` version later needs the `Play` field for the `CurrentPhase` output, the shared function's struct may need updating. For now, omitting it is correct since `CurrentPhase` is always `""`.

4. **Test coverage**: The `removeIfExists` and `isNotFound` functions are not directly tested in the existing test suite. The refactored versions maintain identical external behavior, so no new tests are strictly needed, but the implementer should verify `stream_test.go` passes.

5. **The `fmt` import** may not be fully removable from `validation/validator.go` because `fmt.Sprintf` is used in `parseFailedEvents` and `runPlaybook`. Keep `fmt` in imports there.

## 8. Chunks

- Chunk 1: Steps 1-3 — "Replace else blocks and raw fmt.Print with slog logging"
- Chunk 2: Steps 4-6 — "Fix error checking, extract shared helpers for addon plugins and progress calculation"
