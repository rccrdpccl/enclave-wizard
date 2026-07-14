#!/usr/bin/env bash
set -euo pipefail

readonly PID_FILE="/tmp/enclave-wizard-demo.pid"

main() {
    if [[ -f "$PID_FILE" ]]; then
        local pid
        pid=$(cat "$PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            printf 'Stopping demo (PID %s)...\n' "$pid"
            kill "$pid"
            sleep 1
            kill -0 "$pid" 2>/dev/null && kill -9 "$pid" 2>/dev/null || true
            echo "Stopped."
        else
            echo "Demo not running (stale PID file)."
        fi
        rm -f "$PID_FILE"
        return
    fi

    if pkill -f "enclave-wizard.*--demo-deploy" 2>/dev/null; then
        echo "Stopped demo (found by process name)."
        return
    fi

    echo "No demo running."
}

main "$@"
