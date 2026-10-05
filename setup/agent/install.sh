#!/usr/bin/env bash
# Shared offline daemon installation. Never start services on the build host.
set -euo pipefail
[[ $# == 3 ]] || { echo 'Usage: install.sh <target-root> <binary> <api-endpoint>' >&2; exit 1; }
target=$(realpath -e -- "$1")
binary=$(realpath -e -- "$2")
setup=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
endpoint=$3
[[ "$endpoint" =~ ^https?://[a-zA-Z0-9.:/_-]+$ ]] || { echo 'Unsafe API endpoint' >&2; exit 1; }
install -D -o root -g root -m 0755 "$binary" "$target/usr/local/bin/homelabd"
install -D -o root -g root -m 0644 "$setup/systemd/homelabd.service" "$target/etc/systemd/system/homelabd.service"
install -D -o root -g root -m 0644 "$setup/sshd/10-homelabd.conf" "$target/etc/ssh/sshd_config.d/10-homelabd.conf"
sed -i "s|^Environment=\"API_ENDPOINT=.*\"$|Environment=\"API_ENDPOINT=$endpoint\"|" "$target/etc/systemd/system/homelabd.service"
install -D -o root -g root -m 0644 /dev/stdin "$target/usr/lib/sysusers.d/homelabd.conf" <<'EOF'
g lldpd -
u homelabd - "Homelab agent" /var/lib/homelab -
m homelabd lldpd
EOF
install -D -o root -g root -m 0644 /dev/stdin "$target/usr/lib/tmpfiles.d/homelabd.conf" <<'EOF'
d /var/lib/homelab 0755 root root -
EOF
install -d -o root -g root -m 0755 "$target/etc/systemd/system/multi-user.target.wants"
ln -sf /etc/systemd/system/homelabd.service "$target/etc/systemd/system/multi-user.target.wants/homelabd.service"
ln -sf /usr/lib/systemd/system/lldpd.service "$target/etc/systemd/system/multi-user.target.wants/lldpd.service"
