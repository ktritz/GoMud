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

# Stop service before replacing binary
if $RESTART && systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
    echo "Stopping $SERVICE_NAME..."
    sudo systemctl stop "$SERVICE_NAME"
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

    sudo chown -R "$USER:$USER" "$DATA_DIR"
    echo "Data files updated."
fi

# Also deploy gomud-agent if it exists
AGENT_DIR="/home/ktritz/projects/gomud-agent"
if [[ -d "$AGENT_DIR" ]]; then
    echo "Building gomud-agent..."
    cd "$AGENT_DIR"
    go build -o "$AGENT_DIR/gomud-agent" .
    sudo cp "$AGENT_DIR/gomud-agent" "$INSTALL_DIR/bin/gomud-agent"
    sudo chown "$USER:$USER" "$INSTALL_DIR/bin/gomud-agent"
    sudo chmod 755 "$INSTALL_DIR/bin/gomud-agent"
    rm -f "$AGENT_DIR/gomud-agent"
    echo "gomud-agent installed."
fi

# Restart service
if $RESTART; then
    echo "Starting $SERVICE_NAME..."
    sudo systemctl start "$SERVICE_NAME"
    sleep 1
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        echo "Service is running."
    else
        echo "WARNING: Service failed to start. Check logs:"
        echo "  sudo journalctl -u $SERVICE_NAME -n 20"
        exit 1
    fi
fi

echo ""
echo "=== Deploy complete ==="
echo "  Binary:  $INSTALL_DIR/bin/gomud"
echo "  Status:  systemctl status $SERVICE_NAME"
echo "  Logs:    tail -f $INSTALL_DIR/logs/gomud.log"
