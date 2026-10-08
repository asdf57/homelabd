#!/usr/bin/env bash
# Offline package access configuration; never start a service on the build host.
set -euo pipefail
[[ $# -le 1 ]] || { echo 'Usage: configure-debian-lldp.sh [target-root]' >&2; exit 1; }
target=$(realpath -e -- "${1:-/}")
[[ -f "$target/etc/debian_version" ]] || exit 0
[[ -f "$target/usr/sbin/lldpcli" ]] || { echo 'Missing required lldpd package' >&2; exit 1; }
# Debian uses _lldpd for its socket. Do not grant the unrelated adm group.
printf 'm homelabd _lldpd\n' | install -o root -g root -m 0644 /dev/stdin "$target/usr/lib/sysusers.d/homelabd-lldp.conf"
chroot "$target" systemd-sysusers
override=$(chroot "$target" dpkg-statoverride --list /usr/sbin/lldpcli || true)
if [[ "$override" == '_lldpd adm 4750 /usr/sbin/lldpcli' ]]; then
    chroot "$target" dpkg-statoverride --remove /usr/sbin/lldpcli
    override=''
fi
[[ -z "$override" || "$override" == 'root _lldpd 750 /usr/sbin/lldpcli' ]] || {
    echo 'Conflicting lldpcli permission override' >&2; exit 1;
}
if [[ -z "$override" ]]; then
    chroot "$target" dpkg-statoverride --update --add root _lldpd 0750 /usr/sbin/lldpcli
else
    chroot "$target" chown root:_lldpd /usr/sbin/lldpcli
    chroot "$target" chmod 0750 /usr/sbin/lldpcli
fi
# No setuid execution: access comes only from the LLDP socket group.
