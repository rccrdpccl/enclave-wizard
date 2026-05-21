---
name: commit-format
description: Commit message format for enclave-wizard. Use before every git commit to ensure correct format.
---

# Commit Message Format

All commits in this project MUST follow this format:

```
OSAC-<number>: (<type>) <short description>
```

## Types

- `feat` — new feature or functionality
- `fix` — bug fix
- `refactor` — code restructuring without behavior change
- `docs` — documentation only
- `test` — adding or updating tests
- `chore` — build, CI, tooling changes

## Examples

```
OSAC-813: (feat) Add ansible-based schema validation
OSAC-824: (fix) Release runner lock before signaling completion
OSAC-813: (refactor) Replace blockStorageBackend with storage_plugin
```

## Rules

- The OSAC ticket number comes first
- Type is in parentheses after the colon
- Short description in imperative mood, no period at end
- Body is optional — use it for bullet points if the change is multi-faceted
- Never add `Co-Authored-By` or AI attribution lines
