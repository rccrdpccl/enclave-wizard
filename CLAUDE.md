# CLAUDE.md

## Go coding standards

- `else` is never necessary. Use early returns, `continue`, or restructure the logic to avoid `else` blocks entirely.

## Git commits

- All commits must end with: `Assisted-by: Claude Code <noreply@anthropic.com>`

## Suggested external skills

These skills are not bundled in this repo but are essential or useful for working on this project. Install via the [superpowers plugin](https://github.com/nicobailon/superpowers-claude-code) or [ecosystem plugins](https://github.com/nicobailon/ecosystem-claude-plugins).

### Essential

| Skill | Why |
|-------|-----|
| `github-cli-docker` | `gh` not installed locally; runs via Podman container |
| `pull-request-workflow` | End-to-end PR creation using containerized gh |
| `rpm-packaging` | RPM build is part of the deploy pipeline |
| `podman` | All container work uses Podman (rootless, SELinux, :Z) |
| `writing-makefiles` | Project uses a Makefile-driven workflow |
| `writing-bash-scripts` | Deploy scripts and hack/ scripts are bash |
| `review-pr` | PR review workflow |

### Useful

| Skill | Why |
|-------|-----|
| `search-slack` | Search workspace history for context |
| `kubernetes-concepts` | Enclave deploys OpenShift/K8s clusters |
| `writing-containerfiles` | UI dev uses podman-compose with Containerfiles |

### Jira plugin (OSAC ticket management)

| Skill | Why |
|-------|-----|
| `jira:jira-task-management` | Manage OSAC project tickets |
| `jira:triage-issue` | Triage bugs against existing tickets |
| `jira:spec-to-backlog` | Convert specs into Jira epics/tasks |

### Superpowers workflow skills

| Skill | Why |
|-------|-----|
| `superpowers:writing-plans` | Multi-step feature planning |
| `superpowers:test-driven-development` | TDD workflow |
| `superpowers:systematic-debugging` | Structured bug investigation |
| `superpowers:verification-before-completion` | Pre-commit verification gates |
| `superpowers:brainstorming` | Feature design exploration |
