#!/usr/bin/env python3
"""Generate an enclave-mock directory from the real enclave repo.

Shallow-clones the enclave repo and transforms Ansible task files into
no-ops so the wizard can run locally without touching infrastructure.
"""

import argparse
import os
import shutil
import subprocess
import sys
import tempfile

import yaml


DEFAULT_REPO = "git@github.com:rccrdpccl/enclave.git"
DEFAULT_BRANCH = "main"
DEFAULT_OUTPUT = "enclave-mock"

# Ansible meta keys -- everything that is NOT the action module.
META_KEYS = frozenset({
    "name", "when", "loop", "loop_control", "register", "vars",
    "block", "rescue", "always", "tags", "become", "become_user",
    "environment", "listen", "notify", "timeout", "check_mode",
    "any_errors_fatal", "run_once", "delegate_to", "connection",
})

# Keys to strip from transformed tasks (retry/error handling that
# doesn't make sense on a debug noop).
STRIP_KEYS = frozenset({
    "retries", "delay", "until", "failed_when", "changed_when",
    "no_log", "ignore_errors",
})

# Modules that don't touch infrastructure -- keep as-is.
PASSTHROUGH_MODULES = frozenset({
    "ansible.builtin.include_tasks",
    "ansible.builtin.import_tasks",
    "ansible.builtin.include_role",
    "ansible.builtin.import_role",
    "ansible.builtin.include_vars",
    "ansible.builtin.set_fact",
    "ansible.builtin.debug",
    "ansible.builtin.assert",
    "ansible.builtin.fail",
    "ansible.builtin.meta",
    "ansible.builtin.find",
    "ansible.builtin.stat",
    "ansible.utils.validate",
})


def parse_args():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--branch", default=DEFAULT_BRANCH,
                    help="Enclave branch/tag/commit (default: %(default)s)")
    p.add_argument("--repo", default=DEFAULT_REPO,
                    help="Enclave repo URL (default: %(default)s)")
    p.add_argument("--output", default=DEFAULT_OUTPUT,
                    help="Output directory (default: %(default)s)")
    return p.parse_args()


def find_action_module(task):
    """Return the action module key in a task dict, or None."""
    for key in task:
        if key not in META_KEYS and key not in STRIP_KEYS:
            return key
    return None


def transform_task(task):
    """Transform a single Ansible task dict into a noop.

    Returns a new dict. Block tasks are handled recursively.
    """
    if not isinstance(task, dict):
        return task

    result = {}

    # Handle block/rescue/always recursively.
    if "block" in task:
        for key in task:
            if key in ("block", "rescue", "always"):
                result[key] = transform_task_list(task[key])
            elif key in STRIP_KEYS:
                continue
            else:
                result[key] = task[key]
        return result

    action = find_action_module(task)
    if action is None:
        return task

    # Passthrough modules are kept as-is (minus stripped keys).
    if action in PASSTHROUGH_MODULES:
        return {k: v for k, v in task.items() if k not in STRIP_KEYS}

    # Replace the action module with debug noop.
    task_name = task.get("name", action)
    for key in task:
        if key in STRIP_KEYS:
            continue
        if key == action:
            continue
        result[key] = task[key]

    result["ansible.builtin.debug"] = {"msg": f"NOOP: {task_name}"}
    return result


def transform_task_list(tasks):
    """Transform a list of Ansible tasks."""
    if not isinstance(tasks, list):
        return tasks
    return [transform_task(t) for t in tasks]


def should_transform_task_file(relpath):
    """Return True if the file is a task file that should be transformed."""
    parts = relpath.split(os.sep)

    # playbooks/tasks/*.yaml
    if len(parts) >= 3 and parts[0] == "playbooks" and parts[1] == "tasks":
        return True

    # plugins/*/tasks/*.yaml
    if (len(parts) >= 4 and parts[0] == "plugins"
            and parts[2] == "tasks" and parts[1] != "example"):
        return True

    return False


def should_transform_playbook(relpath):
    """Return True if the file is a top-level playbook to transform."""
    parts = relpath.split(os.sep)

    # playbooks/*.yaml but NOT playbooks/validation/* or playbooks/common/*
    if (len(parts) == 2 and parts[0] == "playbooks"
            and parts[1].endswith(".yaml")):
        return True

    return False


def transform_playbook_file(data):
    """Transform a top-level playbook (list of plays).

    Plays may have inline 'tasks', 'pre_tasks', 'post_tasks', 'handlers'
    which need transformation.
    """
    if not isinstance(data, list):
        return data

    result = []
    for play in data:
        if not isinstance(play, dict):
            result.append(play)
            continue

        new_play = {}
        for key, value in play.items():
            if key in ("tasks", "pre_tasks", "post_tasks", "handlers"):
                new_play[key] = transform_task_list(value)
            else:
                new_play[key] = value
        result.append(new_play)
    return result


def transform_file(filepath, relpath):
    """Read a YAML file, transform it, and write it back."""
    with open(filepath, "r") as f:
        raw = f.read()

    try:
        data = yaml.safe_load(raw)
    except yaml.YAMLError:
        return  # Skip files that aren't valid YAML

    if data is None:
        return

    if should_transform_task_file(relpath):
        transformed = transform_task_list(data)
    elif should_transform_playbook(relpath):
        transformed = transform_playbook_file(data)
    else:
        return

    with open(filepath, "w") as f:
        yaml.dump(transformed, f, default_flow_style=False, sort_keys=False,
                  allow_unicode=True, width=120)


def clone_enclave(repo, branch, dest):
    """Shallow-clone the enclave repo into dest."""
    print(f"Cloning {repo} (branch: {branch}) ...")
    subprocess.run(
        ["git", "clone", "--depth", "1", "--branch", branch, repo, dest],
        check=True,
    )
    # Remove .git -- we don't need history.
    git_dir = os.path.join(dest, ".git")
    if os.path.isdir(git_dir):
        shutil.rmtree(git_dir)


def transform_directory(root):
    """Walk the cloned directory and transform task/playbook files."""
    count = 0
    for dirpath, _, filenames in os.walk(root):
        for fname in filenames:
            if not fname.endswith(".yaml") and not fname.endswith(".yml"):
                continue
            filepath = os.path.join(dirpath, fname)
            relpath = os.path.relpath(filepath, root)
            if should_transform_task_file(relpath) or should_transform_playbook(relpath):
                transform_file(filepath, relpath)
                count += 1
    return count


def main():
    args = parse_args()
    output = os.path.abspath(args.output)

    if os.path.exists(output):
        print(f"Removing existing {output} ...")
        shutil.rmtree(output)

    with tempfile.TemporaryDirectory() as tmpdir:
        clone_dir = os.path.join(tmpdir, "enclave")
        clone_enclave(args.repo, args.branch, clone_dir)
        shutil.move(clone_dir, output)

    count = transform_directory(output)
    print(f"Transformed {count} files in {output}")
    print("Done. Use --enclave-dir enclave-mock to run the wizard.")


if __name__ == "__main__":
    main()
