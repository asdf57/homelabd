#!/usr/bin/env bash
set -euo pipefail

# create logger
log() {
    local level="$1"
    shift
    printf '%s [%s] %s\n' \
        "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
        "$level" \
        "$*" >&2
}

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/.." && pwd)"

HOMELAB_USER="homelabd"
HOMELAB_GROUP="homelabd"

sudo -v

# Create service account.
if ! getent group "$HOMELAB_GROUP" >/dev/null; then
    log "INFO" "Creating group $HOMELAB_GROUP"
    sudo groupadd --system "$HOMELAB_GROUP"
fi

if ! id "$HOMELAB_USER" >/dev/null 2>&1; then
    log "INFO" "Creating user $HOMELAB_USER"
    sudo useradd \
        --system \
        --gid "$HOMELAB_GROUP" \
        --home-dir /var/lib/homelab \
        --shell /usr/bin/nologin \
        "$HOMELAB_USER"
fi

# Build the go binary.
log "INFO" "Building homelabd binary"
mkdir -p "$REPO_ROOT/build"
go build -o "$REPO_ROOT/build/homelabd" "$REPO_ROOT"

log "INFO" "Installing homelabd binary to /usr/local/bin"
sudo install \
    -o root \
    -g root \
    -m 0755 \
    "$REPO_ROOT/build/homelabd" \
    /usr/local/bin/homelabd

log "INFO" "Removing build artifacts"
rm -f "$REPO_ROOT/build/homelabd"
rmdir "$REPO_ROOT/build" 2>/dev/null || true

log "INFO" "Creating /var/lib/homelab directory"
sudo install -d -o root -g root -m 0755 /var/lib/homelab

log "INFO" "Creating /var/lib/homelab/authorized-keys directory"
sudo install -d \
    -o "$HOMELAB_USER" \
    -g "$HOMELAB_GROUP" \
    -m 0750 \
    /var/lib/homelab/authorized-keys

log "INFO" "Configuring sshd to read homelabd-managed keys"
sudo install -d -o root -g root -m 0755 /etc/ssh/sshd_config.d
sudo install \
    -o root \
    -g root \
    -m 0644 \
    "$REPO_ROOT/setup/sshd/10-homelabd.conf" \
    /etc/ssh/sshd_config.d/10-homelabd.conf
if ! sudo sshd -t; then
    log "ERROR" "The sshd configuration is invalid; removing the homelabd drop-in"
    sudo rm -f /etc/ssh/sshd_config.d/10-homelabd.conf
    exit 1
fi

log "INFO" "Installing systemd service"
sudo install \
    -o root \
    -g root \
    -m 0644 \
    "$REPO_ROOT/setup/systemd/homelabd.service" \
    /etc/systemd/system/homelabd.service

log "INFO" "Reloading systemd daemon"
sudo systemctl daemon-reload

log "INFO" "Enabling homelabd service"
sudo systemctl enable homelabd.service

log "INFO" "Starting homelabd service"
if ! sudo systemctl restart homelabd.service; then
    log "ERROR" "homelabd service failed to start"
    sudo journalctl -u homelabd.service --no-pager
    exit 1
fi

log "INFO" "Reloading sshd"
sudo systemctl reload sshd.service

log "INFO" "homelabd service is running"
