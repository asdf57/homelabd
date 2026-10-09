#!/usr/bin/env bash
set -euo pipefail
[[ -f /.dockerenv && $EUID == 0 ]] || { echo 'Run in a disposable root container' >&2; exit 1; }
live=$(mktemp -d)
installed=$(mktemp -d)
bash /assets/management/install.sh "$live"
bash "$live/usr/share/homelabd/setup/management/install.sh" "$installed"
while IFS= read -r -d '' source; do
    relative=${source#/assets/}
    cmp "$source" "$live/usr/share/homelabd/setup/$relative"
    cmp "$source" "$installed/usr/share/homelabd/setup/$relative"
done < <(find /assets/agent /assets/management /assets/systemd /assets/sshd -type f -print0)
bash "$installed/usr/share/homelabd/setup/agent/install.sh" "$installed" /bin/true https://api.example
echo 'PASS: complete setup bundle survives both provisioning handoffs'
