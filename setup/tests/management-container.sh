#!/usr/bin/env bash
set -euo pipefail
# This test deliberately creates/deletes local identities. Never run on a host.
[[ -f /.dockerenv && $EUID == 0 ]] || { echo 'Run only inside a disposable root Docker container' >&2; exit 1; }
if [[ -f /etc/arch-release ]]; then
    pacman -Syu --noconfirm
    pacman -S --needed --noconfirm bash coreutils gawk shadow util-linux sudo openssh systemd
    sshd=/usr/bin/sshd
else
    apt-get update
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends bash coreutils gawk passwd util-linux sudo openssh-server openssh-client systemd
    sshd=/usr/sbin/sshd
fi
install -d -m 0755 /run/sshd /home
workspace=$(mktemp -d)
ssh-keygen -q -t ed25519 -N '' -f "$workspace/ca"
ssh-keygen -q -t ed25519 -N '' -f "$workspace/client"
ssh-keygen -q -t ed25519 -N '' -f "$workspace/untrusted-ca"
ssh-keygen -q -t ed25519 -N '' -f "$workspace/host"
bash /assets/management/install.sh / "$workspace/ca.pub"
identity=$(id -u ansible):$(id -g ansible)
bash /assets/management/install.sh / "$workspace/ca.pub"
[[ "$(id -u ansible):$(id -g ansible)" == "$identity" ]]
printf 'residual-home\n' > /home/ansible/retained-file
userdel ansible
/usr/local/libexec/ensure-ansible-user
[[ "$(id -u ansible):$(id -g ansible)" == "$identity" ]]
[[ $(cat /home/ansible/retained-file) == residual-home ]]
! pgrep -x homelabd
# Home symlinks and recycled IDs must cause safe failure, without chmod/chown.
mv /home/ansible "$workspace/residual-home"
ln -s "$workspace/residual-home" /home/ansible
if /usr/local/libexec/ensure-ansible-user; then echo 'Symlink home was accepted' >&2; exit 1; fi
unlink /home/ansible
mv "$workspace/residual-home" /home/ansible
userdel ansible
useradd --uid "${identity%:*}" --no-create-home intruder
if /usr/local/libexec/ensure-ansible-user; then echo 'UID collision was accepted' >&2; exit 1; fi
userdel intruder
/usr/local/libexec/ensure-ansible-user

systemd-analyze verify /etc/systemd/system/ansible-account.service /etc/systemd/system/ansible-account.timer
visudo -cf /etc/sudoers.d/ansible-management
configuration="$workspace/sshd.conf"
install -m 0644 /assets/sshd/10-homelabd.conf /etc/ssh/sshd_config.d/10-homelabd.conf
printf 'Port 2222\nListenAddress 127.0.0.1\nHostKey %s\nPidFile %s\nUsePAM yes\nLogLevel DEBUG\nInclude /etc/ssh/sshd_config.d/*.conf\n' "$workspace/host" "$workspace/sshd.pid" > "$configuration"
"$sshd" -t -f "$configuration"
"$sshd" -D -e -f "$configuration" > "$workspace/sshd.log" 2>&1 &
sshd_pid=$!
trap 'kill "$sshd_pid" 2>/dev/null || true' EXIT
printf '[127.0.0.1]:2222 %s\n' "$(cat "$workspace/host.pub")" > "$workspace/known_hosts"
ssh_args=(-p 2222 -o "UserKnownHostsFile=$workspace/known_hosts" -o StrictHostKeyChecking=yes -o BatchMode=yes -o IdentitiesOnly=yes -i "$workspace/client")
ssh-keygen -q -s "$workspace/ca" -I management-test -n ansible -V -1m:+5m "$workspace/client.pub"
for attempt in {1..20}; do
    if ssh "${ssh_args[@]}" ansible@127.0.0.1 'sudo -n true'; then break; fi
    sleep 0.2
done
ssh "${ssh_args[@]}" ansible@127.0.0.1 'sudo -n true'
# Model a compromised daemon writing raw keys for root and a sudo-capable user.
useradd --create-home --shell /bin/bash --password '*' privileged
printf 'privileged ALL=(ALL) NOPASSWD: ALL\n' > /etc/sudoers.d/test-privileged
mkdir -p /var/lib/homelab/authorized-keys
cp "$workspace/client.pub" /var/lib/homelab/authorized-keys/root
cp "$workspace/client.pub" /var/lib/homelab/authorized-keys/privileged
for login in root privileged; do
    if ssh "${ssh_args[@]}" -o CertificateFile=none "$login@127.0.0.1" true; then echo 'Daemon-controlled raw key was accepted' >&2; exit 1; fi
    effective_command=$("$sshd" -T -f "$configuration" -C "user=$login,host=localhost,addr=127.0.0.1" | awk '$1 == "authorizedkeyscommand" { print $2 }')
    # Some OpenSSH releases omit unset optional commands from -T output.
    [[ -z "$effective_command" || "$effective_command" == none ]] || { echo "Unexpected effective key command for $login: $effective_command" >&2; exit 1; }
done
if ssh "${ssh_args[@]}" -o CertificateFile=none ansible@127.0.0.1 true; then echo 'Raw-key fallback was accepted' >&2; exit 1; fi
ssh-keygen -q -s "$workspace/ca" -I wrong-principal -n root -V -1m:+5m "$workspace/client.pub"
if ssh "${ssh_args[@]}" ansible@127.0.0.1 true; then echo 'Wrong principal was accepted' >&2; exit 1; fi
ssh-keygen -q -s "$workspace/untrusted-ca" -I wrong-authority -n ansible -V -1m:+5m "$workspace/client.pub"
if ssh "${ssh_args[@]}" ansible@127.0.0.1 true; then echo 'Untrusted CA was accepted' >&2; exit 1; fi
"$sshd" -T -f "$configuration" -C user=ansible,host=localhost,addr=127.0.0.1 | awk '$1 == "passwordauthentication" || $1 == "kbdinteractiveauthentication" { if ($2 != "no") exit 1 }'
echo 'PASS: installation, runtime recovery, conflicts, certificate login, and fallback rejection'
