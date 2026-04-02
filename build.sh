#!/usr/bin/env bash
#
# Build GoMud and optionally deploy to /opt/gomud/bin/
#
# Usage:
#   ./build.sh           # build and deploy
#   ./build.sh --no-deploy   # build only
#

set -euo pipefail

cd "$(dirname "$0")"

DEPLOY=true
for arg in "$@"; do
    case "$arg" in
        --no-deploy) DEPLOY=false ;;
    esac
done

echo "--- Building GoMud ---"
go clean -cache
go build -o gomud .
echo "  Built: gomud"

go vet ./...
echo "  Vet: passed"

if $DEPLOY; then
    echo "--- Deploying binary ---"
    sudo systemctl stop gomud 2>/dev/null || true
    sleep 1
    sudo cp gomud /opt/gomud/bin/gomud
    sudo chown gomud:gomud /opt/gomud/bin/gomud
    sudo systemctl start gomud
    sleep 3
    if sudo systemctl is-active --quiet gomud; then
        echo "  Server: running"
    else
        echo "  Server: FAILED TO START"
        exit 1
    fi
fi

rm -f gomud
echo "--- Done ---"
