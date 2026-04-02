#!/usr/bin/env bash
set -euo pipefail

# GoMud deploy script
# Builds from source and deploys to /opt/gomud.
# Usage: ./deploy/deploy.sh [--no-restart] [--update-data]

INSTALL_DIR="/opt/gomud"
DATA_DIR="/opt/gomud/data"
USER="gomud"
SERVICE_NAME="gomud"

RESTART=true
UPDATE_DATA=false

for arg in "$@"; do
    case $arg in
        --no-restart)   RESTART=false ;;
        --update-data)  UPDATE_DATA=true ;;
        --help|-h)
            echo "Usage: $0 [--no-restart] [--update-data]"
            echo "  --no-restart    Build and copy binary but don't restart service"
            echo "  --update-data   Also update data files (templates, buffs, etc.)"
            echo "                  Does NOT overwrite config.yaml or world save data"
            exit 0 ;;
        *) echo "Unknown arg: $arg"; exit 1 ;;
    esac
done

SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

echo "=== GoMud Deploy ==="
echo "Source: $SCRIPT_DIR"
echo "Target: $INSTALL_DIR"

# Verify setup was run
if [[ ! -d "$INSTALL_DIR/bin" ]]; then
    echo "Error: $INSTALL_DIR/bin not found. Run setup.sh first."
    exit 1
fi

# Build
echo "Building..."
cd "$SCRIPT_DIR"
go build -o "$SCRIPT_DIR/gomud" .
echo "Build complete."

# Stop service before replacing binary or updating data
# Must stop before data sync to prevent server re-saving stale room instances
if ($RESTART || $UPDATE_DATA) && systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
    echo "Stopping $SERVICE_NAME..."
    sudo systemctl stop "$SERVICE_NAME"
    sleep 1
fi

# Copy binary
echo "Installing binary..."
sudo cp "$SCRIPT_DIR/gomud" "$INSTALL_DIR/bin/gomud"
sudo chown "$USER:$USER" "$INSTALL_DIR/bin/gomud"
sudo chmod 755 "$INSTALL_DIR/bin/gomud"
rm -f "$SCRIPT_DIR/gomud"

# Optionally update data files (but preserve config and save data)
if $UPDATE_DATA; then
    echo "Updating data files..."

    # Run validation before syncing
    echo "Validating world data..."
    if ! python3 "$SCRIPT_DIR/_datafiles/world/insideout/tools/validate.py" \
         --base-dir "$SCRIPT_DIR/_datafiles/world/insideout" 2>&1; then
        echo "ERROR: Validation failed. Fix errors before deploying."
        exit 1
    fi

    # Templates, buffs, spells, items definitions, etc. — safe to overwrite
    for subdir in html localize guides; do
        if [[ -d "$SCRIPT_DIR/_datafiles/$subdir" ]]; then
            sudo rsync -a --delete "$SCRIPT_DIR/_datafiles/$subdir/" "$DATA_DIR/$subdir/"
        fi
    done

    # World default data (templates, combat messages, color patterns, etc.)
    # Only update files that ship with the engine, NOT user-created rooms/mobs/items
    WORLD_DEFAULT="$SCRIPT_DIR/_datafiles/world/default"
    if [[ -d "$WORLD_DEFAULT" ]]; then
        for subdir in templates buffs spells races conversations biomes combat-messages; do
            if [[ -d "$WORLD_DEFAULT/$subdir" ]]; then
                sudo rsync -a "$WORLD_DEFAULT/$subdir/" "$DATA_DIR/world/default/$subdir/"
            fi
        done

        # Update engine-shipped files in world root (ansi-aliases, color-patterns, audio, etc.)
        for f in "$WORLD_DEFAULT"/*.yaml; do
            if [[ -f "$f" ]]; then
                sudo cp "$f" "$DATA_DIR/world/default/$(basename "$f")"
            fi
        done
    fi

    # Insideout world data — sync all content directories but NEVER touch
    # users/ or rooms.instances/ (runtime data created by the server)
    WORLD_INSIDEOUT="$SCRIPT_DIR/_datafiles/world/insideout"
    if [[ -d "$WORLD_INSIDEOUT" ]]; then
        echo "Syncing insideout world data..."

        # Ensure target directories exist
        sudo mkdir -p "$DATA_DIR/world/insideout/rooms.instances"

        # Clear room instances cache — these are runtime saves that override
        # room templates. Stale instances cause exits/descriptions to not update.
        if [[ -d "$DATA_DIR/world/insideout/rooms.instances" ]]; then
            echo "Clearing room instances cache..."
            sudo find "$DATA_DIR/world/insideout/rooms.instances" -name "*.yaml" -delete
        fi

        # Content directories — safe to sync (no --delete to preserve server-added files)
        for subdir in rooms mobs items buffs spells races quests templates \
                      conversations combat-messages mutators pets biomes \
                      tools plugin-data keywords.yaml; do
            src="$WORLD_INSIDEOUT/$subdir"
            dst="$DATA_DIR/world/insideout/$subdir"
            if [[ -e "$src" ]]; then
                if [[ -d "$src" ]]; then
                    sudo rsync -a "$src/" "$dst/"
                else
                    sudo cp "$src" "$dst"
                fi
            fi
        done

        # Root-level YAML files (ansi-aliases, audio, color-patterns, etc.)
        for f in "$WORLD_INSIDEOUT"/*.yaml; do
            if [[ -f "$f" ]]; then
                sudo cp "$f" "$DATA_DIR/world/insideout/$(basename "$f")"
            fi
        done

        echo "Insideout world data synced."
        echo "  Preserved: users/, config.yaml"
        echo "  Cleared: rooms.instances/ (stale room cache)"
    fi

    sudo chown -R "$USER:$USER" "$DATA_DIR"
    echo "Data files updated."
fi

# Also deploy gomud-agent if it exists
AGENT_DIR="/home/ktritz/projects/gomud-agent"
AGENT_SERVICE="gomud-agent"
if [[ -d "$AGENT_DIR" ]]; then
    echo "Building gomud-agent..."
    # Kill any rogue agent processes (from manual go run / background launches)
    ROGUE_PIDS=$(pgrep -f 'gomud-agent connect' | grep -v "^$(cat /tmp/gomud-agent/agent.pid 2>/dev/null)$" || true)
    if [[ -n "$ROGUE_PIDS" ]]; then
        echo "Killing rogue agent processes: $ROGUE_PIDS"
        kill $ROGUE_PIDS 2>/dev/null || true
        sleep 1
    fi

    # Stop agent service if running (binary will be busy otherwise)
    AGENT_WAS_RUNNING=false
    if systemctl is-active --quiet "$AGENT_SERVICE" 2>/dev/null; then
        AGENT_WAS_RUNNING=true
        sudo systemctl stop "$AGENT_SERVICE"
        sleep 1
    fi
    cd "$AGENT_DIR"
    go build -o "$AGENT_DIR/gomud-agent" .
    sudo cp "$AGENT_DIR/gomud-agent" "$INSTALL_DIR/bin/gomud-agent"
    sudo chown "$USER:$USER" "$INSTALL_DIR/bin/gomud-agent"
    sudo chmod 755 "$INSTALL_DIR/bin/gomud-agent"
    rm -f "$AGENT_DIR/gomud-agent"
    echo "gomud-agent installed."
fi

# Restart service (always restart after data updates to clear cached rooms/items/etc.)
if $RESTART; then
    if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
        echo "Restarting $SERVICE_NAME..."
        sudo systemctl restart "$SERVICE_NAME"
    else
        echo "Starting $SERVICE_NAME..."
        sudo systemctl start "$SERVICE_NAME"
    fi
    sleep 1
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        echo "Service is running."
    else
        echo "WARNING: Service failed to start. Check logs:"
        echo "  sudo journalctl -u $SERVICE_NAME -n 20"
        exit 1
    fi

    # Restart agent if it was running before deploy
    if [[ "${AGENT_WAS_RUNNING:-false}" == "true" ]]; then
        echo "Restarting $AGENT_SERVICE..."
        sleep 3  # Give server time to fully initialize
        sudo systemctl start "$AGENT_SERVICE"
    fi
fi

echo ""
echo "=== Deploy complete ==="
echo "  Binary:  $INSTALL_DIR/bin/gomud"
echo "  Status:  systemctl status $SERVICE_NAME"
echo "  Logs:    tail -f $INSTALL_DIR/logs/gomud.log"
