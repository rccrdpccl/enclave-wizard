---
name: osac-jira
description: Use when creating, updating, or managing Jira issues in the OSAC project for enclave-wizard work. Also use when the user asks to file a ticket, check issue status, or manage epics in this project.
---

# OSAC Jira Task Management

## Overview

All enclave-wizard work is tracked in the **OSAC** Jira project on Red Hat Jira (`redhat.atlassian.net`). Issues must always be linked to an epic and labeled `OSAC`.

## Defaults for This Project

| Field       | Value                           |
|-------------|---------------------------------|
| Project     | `OSAC` (pass `--project OSAC`)  |
| Label       | `OSAC` (always `-l OSAC`)       |
| Component   | `Enclave`                       |
| Epic        | `OSAC-809` (Enclave Configuration Wizard) unless user specifies otherwise |
| Assignee    | `rpiccoli@redhat.com` (`jira me`) unless user specifies otherwise |

**jira-cli defaults to MGMT project.** Always pass `--project OSAC` on create commands or the issue lands in the wrong project.

## Creating Issues

### Task

```bash
jira issue create -tTask -s "Summary" -b "Description" \
  -P OSAC-809 -a "rpiccoli@redhat.com" -l OSAC \
  --project OSAC --no-input
```

### Story

```bash
jira issue create -tStory -s "Summary" -b "Description" \
  -P OSAC-809 -a "rpiccoli@redhat.com" -l OSAC \
  --project OSAC --no-input
```

### Bug

```bash
jira issue create -tBug -s "Bug title" \
  -b $'**Description of the problem:**\n\n<describe>\n\n**How reproducible:**\n\n<rate>\n\n**Steps to reproduce:**\n\n1. <step>\n\n**Expected result:**\n\n<expected>\n\n**Actual result:**\n\n<actual>' \
  -P OSAC-809 -a "rpiccoli@redhat.com" -l OSAC \
  --project OSAC --no-input
```

### Sub-task

```bash
jira issue create -tSub-task -s "Summary" -P <PARENT-KEY> -l OSAC \
  --project OSAC --no-input
```

For JSON output (to capture the key), add `--raw`.

## Statuses

Available transitions: `New` → `In Progress` → `Review` → `Closed`

Also valid: `Backlog`, `To Do`, `Release Pending`, `Verified`

```bash
jira issue move OSAC-XXX "In Progress" --comment "Reason for transition"
```

**Always add a `--comment` explaining why you are changing status.**

## IMPORTANT: Confirm Before Closing

Never transition an issue to `Closed` without explicit user confirmation, even if the codebase suggests work is complete. Code being committed does not mean the work is done.

## Quick Reference

```bash
# View issue
jira issue view OSAC-XXX --plain

# List open issues assigned to me
jira issue list --jql 'project = OSAC AND assignee = currentUser() AND status not in (Closed, Done)' --plain

# List issues under wizard epic
jira issue list --jql 'project = OSAC AND "Epic Link" = OSAC-809' --plain

# Assign to self
jira issue assign OSAC-XXX $(jira me)

# Add comment
jira issue comment add OSAC-XXX "Comment text"

# Search
jira issue list --jql 'project = OSAC AND labels = OSAC AND text ~ "search term"' --plain
```
