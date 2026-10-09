#!/usr/bin/env bash
# Real package/socket smoke test, only in a disposable Debian root container.
set -euo pipefail
[[ -f /.dockerenv && $EUID == 0 && -f /etc/debian_version ]] || exit 1
apt-get update -qq
apt-get install -y --no-install-recommends lldpd systemd util-linux sudo
# Exercise the assets provisioning actually takes from the live image.
bash /assets/management/install.sh /
bash /usr/share/homelabd/setup/agent/install.sh / /bin/true https://api.example
bash /usr/share/homelabd/setup/agent/configure-debian-lldp.sh /
[[ $(stat -c '%U:%G:%a' /usr/sbin/lldpcli) == root:_lldpd:750 ]]
! id -nG homelabd | tr ' ' '\n' | grep -qx adm
/usr/sbin/lldpd -d >/tmp/lldpd-test.log 2>&1 &
daemon_pid=$!
trap 'kill "$daemon_pid" 2>/dev/null || true' EXIT
for attempt in {1..30}; do
    [[ -S /run/lldpd.socket ]] && break
    sleep 0.1
done
setpriv --reuid=homelabd --regid=homelabd --init-groups --no-new-privs \
    env PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
    lldpcli -f json0 show neighbors details
echo 'PASS: Debian LLDP discovery without root, setuid or adm membership'
