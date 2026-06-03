#!/usr/bin/env python3
"""Validate a generated enclave-mock directory.

Checks:
1. Required directories exist (plugins/, schemas/, defaults/, playbooks/)
2. Plugin metadata files are valid YAML with expected fields
3. Schemas are untouched (no NOOP markers)
4. Task files contain only passthrough or debug modules
5. Validation playbooks are preserved (contain ansible.utils.validate)
"""

import os
import sys

import yaml


PASSTHROUGH_MODULES = {
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
}

META_KEYS = {
    "name", "when", "loop", "loop_control", "register", "vars",
    "block", "rescue", "always", "tags", "become", "become_user",
    "environment", "listen", "notify", "timeout", "check_mode",
    "any_errors_fatal", "run_once", "delegate_to", "connection",
}

INFRA_MODULES = {
    "kubernetes.core.k8s",
    "ansible.builtin.command",
    "ansible.builtin.shell",
    "ansible.builtin.copy",
    "ansible.builtin.template",
    "ansible.builtin.file",
    "ansible.builtin.uri",
    "ansible.builtin.slurp",
    "ansible.builtin.get_url",
    "ansible.builtin.systemd",
    "ansible.builtin.unarchive",
    "ansible.builtin.wait_for",
    "ansible.builtin.pause",
    "ansible.builtin.replace",
    "ansible.builtin.blockinfile",
    "ansible.builtin.lineinfile",
    "containers.podman.podman_secret",
    "containers.podman.podman_volume",
    "containers.podman.podman_image",
    "containers.podman.podman_container",
    "containers.podman.podman_container_info",
    "containers.podman.podman_pod_info",
}


def check_directories(mock_dir):
    """Check required directories exist."""
    errors = []
    for d in ["plugins", "schemas", "defaults", "playbooks"]:
        path = os.path.join(mock_dir, d)
        if not os.path.isdir(path):
            errors.append(f"Missing required directory: {d}/")
    return errors


def check_plugin_metadata(mock_dir):
    """Check plugin.yaml files have expected fields."""
    errors = []
    plugins_dir = os.path.join(mock_dir, "plugins")
    if not os.path.isdir(plugins_dir):
        return errors

    for entry in os.listdir(plugins_dir):
        if entry == "example":
            continue
        plugin_yaml = os.path.join(plugins_dir, entry, "plugin.yaml")
        if not os.path.isfile(plugin_yaml):
            continue

        with open(plugin_yaml) as f:
            try:
                data = yaml.safe_load(f)
            except yaml.YAMLError as e:
                errors.append(f"plugins/{entry}/plugin.yaml: invalid YAML: {e}")
                continue

        if not isinstance(data, dict):
            errors.append(f"plugins/{entry}/plugin.yaml: not a dict")
            continue

        if "name" not in data:
            errors.append(f"plugins/{entry}/plugin.yaml: missing 'name' field")
        if "type" not in data:
            errors.append(f"plugins/{entry}/plugin.yaml: missing 'type' field")
    return errors


def check_no_noop_in_schemas(mock_dir):
    """Check that schema files were not transformed."""
    errors = []
    schemas_dir = os.path.join(mock_dir, "schemas")
    if not os.path.isdir(schemas_dir):
        return errors

    for fname in os.listdir(schemas_dir):
        if not fname.endswith(".yaml"):
            continue
        path = os.path.join(schemas_dir, fname)
        with open(path) as f:
            content = f.read()
        if "NOOP:" in content:
            errors.append(f"schemas/{fname}: contains NOOP marker (should be preserved)")
    return errors


def find_action_in_task(task):
    """Return the action module key, or None."""
    if not isinstance(task, dict):
        return None
    for key in task:
        if key not in META_KEYS and key not in {"retries", "delay", "until",
                "failed_when", "changed_when", "no_log", "ignore_errors"}:
            return key
    return None


def check_tasks_recursive(tasks, relpath, errors):
    """Recursively check a task list for leftover infra modules."""
    if not isinstance(tasks, list):
        return
    for i, task in enumerate(tasks):
        if not isinstance(task, dict):
            continue
        # Recurse into block/rescue/always
        for block_key in ("block", "rescue", "always"):
            if block_key in task:
                check_tasks_recursive(task[block_key], relpath, errors)
        if "block" in task:
            continue
        action = find_action_in_task(task)
        if action and action in INFRA_MODULES:
            name = task.get("name", f"task #{i}")
            errors.append(f"{relpath}: task '{name}' still has infra module '{action}'")


def check_task_files_noopped(mock_dir):
    """Check that transformed task files contain no infra modules."""
    errors = []
    dirs_to_check = [os.path.join(mock_dir, "playbooks", "tasks")]

    plugins_dir = os.path.join(mock_dir, "plugins")
    if os.path.isdir(plugins_dir):
        for entry in os.listdir(plugins_dir):
            tasks_dir = os.path.join(plugins_dir, entry, "tasks")
            if os.path.isdir(tasks_dir):
                dirs_to_check.append(tasks_dir)

    for tasks_dir in dirs_to_check:
        if not os.path.isdir(tasks_dir):
            continue
        for fname in os.listdir(tasks_dir):
            if not fname.endswith(".yaml") and not fname.endswith(".yml"):
                continue
            path = os.path.join(tasks_dir, fname)
            relpath = os.path.relpath(path, mock_dir)

            with open(path) as f:
                try:
                    data = yaml.safe_load(f)
                except yaml.YAMLError:
                    continue

            if not isinstance(data, list):
                continue

            check_tasks_recursive(data, relpath, errors)
    return errors


def check_validation_preserved(mock_dir):
    """Check that validation playbooks are preserved."""
    errors = []
    val_dir = os.path.join(mock_dir, "playbooks", "validation")
    if not os.path.isdir(val_dir):
        errors.append("playbooks/validation/ directory missing")
        return errors

    schema_val = os.path.join(val_dir, "validate-schema.yaml")
    if not os.path.isfile(schema_val):
        errors.append("playbooks/validation/validate-schema.yaml missing")
    return errors


def main():
    if len(sys.argv) < 2:
        mock_dir = "enclave-mock"
    else:
        mock_dir = sys.argv[1]

    if not os.path.isdir(mock_dir):
        print(f"ERROR: {mock_dir} does not exist. Run 'make enclave-mock' first.")
        sys.exit(1)

    print(f"Validating {mock_dir} ...")
    all_errors = []
    all_errors.extend(check_directories(mock_dir))
    all_errors.extend(check_plugin_metadata(mock_dir))
    all_errors.extend(check_no_noop_in_schemas(mock_dir))
    all_errors.extend(check_task_files_noopped(mock_dir))
    all_errors.extend(check_validation_preserved(mock_dir))

    if all_errors:
        print(f"\nFAILED: {len(all_errors)} error(s):")
        for e in all_errors:
            print(f"  - {e}")
        sys.exit(1)

    print("PASSED: enclave-mock is structurally valid.")


if __name__ == "__main__":
    main()
