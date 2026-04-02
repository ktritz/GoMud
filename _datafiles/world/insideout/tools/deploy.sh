#!/usr/bin/env bash
#
# Deploy the insideout world to the GoMud server.
#
# Usage:
#   ./deploy.sh              # validate, deploy, restart
#   ./deploy.sh --no-restart # validate and deploy only
#   ./deploy.sh --fix        # auto-fix issues before deploying
#   ./deploy.sh --dry-run    # validate only, don't deploy
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
WORLD_DIR="$(dirname "$SCRIPT_DIR")"
SERVER_WORLD="/opt/gomud/data/world/insideout"
SERVICE_NAME="gomud"

RESTART=true
FIX=false
DRY_RUN=false

for arg in "$@"; do
    case "$arg" in
        --no-restart) RESTART=false ;;
        --fix)        FIX=true ;;
        --dry-run)    DRY_RUN=true; RESTART=false ;;
        -h|--help)
            echo "Usage: deploy.sh [--fix] [--no-restart] [--dry-run]"
            echo ""
            echo "  --fix          Auto-fix validation issues before deploying"
            echo "  --no-restart   Deploy files but don't restart the server"
            echo "  --dry-run      Validate only, don't deploy or restart"
            exit 0
            ;;
        *) echo "Unknown option: $arg"; exit 1 ;;
    esac
done

echo "=== GoMud Inside Out World Deploy ==="
echo "Source:  $WORLD_DIR"
echo "Target:  $SERVER_WORLD"
echo ""

# Step 1: Validate (and optionally fix)
echo "--- Validating ---"
if $FIX; then
    python3 "$SCRIPT_DIR/validate.py" --fix --base-dir "$WORLD_DIR"
else
    python3 "$SCRIPT_DIR/validate.py" --base-dir "$WORLD_DIR"
fi
echo ""

if $DRY_RUN; then
    echo "Dry run complete. No files deployed."
    exit 0
fi

# Step 2: Deploy
echo "--- Deploying ---"

# Directories and files to preserve across deploys
PRESERVE_DIRS=(users plugin-data)
PRESERVE_FILES=(.roundcount)

STATE_BACKUP=""
if [ -d "$SERVER_WORLD" ]; then
    STATE_BACKUP=$(mktemp -d)

    for dir in "${PRESERVE_DIRS[@]}"; do
        if [ -d "$SERVER_WORLD/$dir" ]; then
            sudo cp -a "$SERVER_WORLD/$dir" "$STATE_BACKUP/$dir"
            echo "  Backed up $dir/ ($(find "$SERVER_WORLD/$dir" -type f | wc -l) files)"
        fi
    done

    for file in "${PRESERVE_FILES[@]}"; do
        if [ -f "$SERVER_WORLD/$file" ]; then
            sudo cp -a "$SERVER_WORLD/$file" "$STATE_BACKUP/$file"
            echo "  Backed up $file"
        fi
    done

    sudo rm -rf "$SERVER_WORLD"
    echo "  Removed old deployment"
fi

sudo cp -r "$WORLD_DIR" "$SERVER_WORLD"

# Restore preserved state
if [ -n "$STATE_BACKUP" ]; then
    for dir in "${PRESERVE_DIRS[@]}"; do
        if [ -d "$STATE_BACKUP/$dir" ]; then
            sudo rm -rf "$SERVER_WORLD/$dir"
            sudo cp -a "$STATE_BACKUP/$dir" "$SERVER_WORLD/$dir"
            echo "  Restored $dir/"
        fi
    done

    for file in "${PRESERVE_FILES[@]}"; do
        if [ -f "$STATE_BACKUP/$file" ]; then
            sudo cp -a "$STATE_BACKUP/$file" "$SERVER_WORLD/$file"
            echo "  Restored $file"
        fi
    done

    sudo rm -rf "$STATE_BACKUP"
fi

sudo chown -R gomud:gomud "$SERVER_WORLD"

FILE_COUNT=$(find "$SERVER_WORLD" -type f ! -path '*/tools/*' | wc -l)
echo "  Deployed $FILE_COUNT files to $SERVER_WORLD"
echo ""

# Step 3: Restart
if $RESTART; then
    echo "--- Restarting $SERVICE_NAME ---"
    sudo systemctl restart "$SERVICE_NAME"
    sleep 3

    if sudo systemctl is-active --quiet "$SERVICE_NAME"; then
        echo "  Server is running"

        # Quick sanity check on logs
        if grep -q "PANIC" <(tail -20 /opt/gomud/logs/gomud.log 2>/dev/null | grep "$(date +%H:%M)"); then
            echo "  WARNING: PANIC detected in logs after restart!"
            echo ""
            grep "PANIC.*error=" /opt/gomud/logs/gomud.log | tail -1
            exit 1
        fi

        # Show load counts
        echo ""
        echo "--- Server Load Summary ---"
        grep "loadedCount" /opt/gomud/logs/gomud.log | tail -10 | sed 's/.*INFO: /  /' | sed 's/\x1b\[[0-9;]*m//g'
    else
        echo "  ERROR: Server failed to start!"
        echo ""
        grep "PANIC.*error=" /opt/gomud/logs/gomud.log | tail -1 | sed 's/\x1b\[[0-9;]*m//g'
        exit 1
    fi
else
    echo "Skipped restart (--no-restart)"
fi

echo ""
echo "=== Deploy complete ==="
