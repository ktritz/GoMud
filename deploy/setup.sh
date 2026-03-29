#!/usr/bin/env bash
set -euo pipefail

# GoMud initial setup script
# Run once on a fresh server: sudo ./deploy/setup.sh
# Creates user, directories, systemd service, and initial config.

INSTALL_DIR="/opt/gomud"
DATA_DIR="/opt/gomud/data"
USER="gomud"
GROUP="gomud"
SERVICE_NAME="gomud"

echo "=== GoMud Setup ==="

# Must be root
if [[ $EUID -ne 0 ]]; then
    echo "Run as root: sudo $0"
    exit 1
fi

# Create service user
if ! id "$USER" &>/dev/null; then
    useradd --system --home-dir "$INSTALL_DIR" --shell /usr/sbin/nologin "$USER"
    echo "Created user: $USER"
else
    echo "User $USER already exists"
fi

# Create directory structure
mkdir -p "$INSTALL_DIR/bin"
mkdir -p "$DATA_DIR"
mkdir -p "$INSTALL_DIR/logs"

# If this is a fresh setup, copy default data files
SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
if [[ ! -f "$DATA_DIR/config.yaml" ]]; then
    echo "Copying default data files..."
    cp -r "$SCRIPT_DIR/_datafiles/"* "$DATA_DIR/"

    # Patch config paths for deployment layout
    sed -i "s|DataFiles: _datafiles/world/default|DataFiles: $DATA_DIR/world/default|" "$DATA_DIR/config.yaml"
    sed -i "s|PublicHtml: _datafiles/html/public|PublicHtml: $DATA_DIR/html/public|" "$DATA_DIR/config.yaml"
    sed -i "s|AdminHtml: _datafiles/html/admin|AdminHtml: $DATA_DIR/html/admin|" "$DATA_DIR/config.yaml"

    echo "Default data files installed to $DATA_DIR"
    echo "Edit $DATA_DIR/config.yaml to customize ports, settings, etc."
else
    echo "Config already exists at $DATA_DIR/config.yaml — skipping data copy"
fi

# Set ownership
chown -R "$USER:$GROUP" "$INSTALL_DIR"

# Install systemd service
cat > "/etc/systemd/system/${SERVICE_NAME}.service" << EOF
[Unit]
Description=GoMud Game Server
After=network.target

[Service]
Type=simple
User=$USER
Group=$GROUP
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/bin/gomud
Restart=on-failure
RestartSec=5
StandardOutput=append:$INSTALL_DIR/logs/gomud.log
StandardError=append:$INSTALL_DIR/logs/gomud.log

# Hardening
NoNewPrivileges=true
ProtectSystem=strict
ReadWritePaths=$DATA_DIR $INSTALL_DIR/logs
ProtectHome=true

# Allow binding to low ports (80, 443) without root
AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable "$SERVICE_NAME"

echo ""
echo "=== Setup complete ==="
echo "  Install dir:  $INSTALL_DIR"
echo "  Data dir:     $DATA_DIR"
echo "  Config:       $DATA_DIR/config.yaml"
echo "  Logs:         $INSTALL_DIR/logs/gomud.log"
echo "  Service:      systemctl {start|stop|restart|status} $SERVICE_NAME"
echo ""
echo "Next: build and deploy with ./deploy/deploy.sh"
