#!/usr/bin/env bash
# Install reusable assets into a live host or an offline target root.
# Account creation and systemd activation are performed by the caller.
set -euo pipefail
[[ $# == 1 || $# == 2 ]] || { echo 'Usage: install.sh <target-root> [public-ca-bundle]' >&2; exit 1; }
target=$(realpath -e -- "$1")
setup=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
[[ -d "$target" ]] || exit 1
# Keep the same reviewed assets available to provisioning in the live image.
for relative in agent/install.sh management/ensure-ansible-user management/install.sh management/ansible.sudoers systemd/ansible-account.service systemd/ansible-account.timer systemd/homelabd.service sshd/00-ansible-management.conf sshd/10-homelabd.conf; do
    destination="$target/usr/share/homelabd/setup/$relative"
    if [[ "$(realpath -m "$destination")" != "$setup/$relative" ]]; then
        install -D -o root -g root -m 0644 "$setup/$relative" "$destination"
    fi
done
install -D -o root -g root -m 0755 "$setup/management/ensure-ansible-user" "$target/usr/local/libexec/ensure-ansible-user"
if [[ "$target" == / ]]; then
    /usr/local/libexec/ensure-ansible-user
fi
for unit in ansible-account.service ansible-account.timer; do
    install -D -o root -g root -m 0644 "$setup/systemd/$unit" "$target/etc/systemd/system/$unit"
done
# Validate on the build/installation host before placing a privileged policy.
visudo -cf "$setup/management/ansible.sudoers"
install -D -o root -g root -m 0440 "$setup/management/ansible.sudoers" "$target/etc/sudoers.d/ansible-management"
install -d -o root -g root -m 0755 "$target/etc/systemd/system/multi-user.target.wants" "$target/etc/systemd/system/timers.target.wants"
ln -sf /etc/systemd/system/ansible-account.service "$target/etc/systemd/system/multi-user.target.wants/ansible-account.service"
ln -sf /etc/systemd/system/ansible-account.timer "$target/etc/systemd/system/timers.target.wants/ansible-account.timer"
# Run before SSH starts, without coupling later timer repairs to sshd state.
for ssh_unit in ssh.service sshd.service; do
    install -D -o root -g root -m 0644 /dev/stdin "$target/etc/systemd/system/$ssh_unit.d/10-ansible-account.conf" <<'EOF'
[Service]
ExecStartPre=/usr/local/libexec/ensure-ansible-user
EOF
done
if [[ $# == 2 ]]; then
    bundle=$(realpath -e -- "$2")
    count=0
    while IFS= read -r key || [[ -n "$key" ]]; do
        [[ -n "$key" ]] || continue
        [[ "$key" == ssh-ed25519\ * && "$key" != *$'\r'* ]] || { echo 'Only plain Ed25519 CA public keys are accepted' >&2; exit 1; }
        ssh-keygen -lf /dev/stdin <<< "$key" >/dev/null
        ((count+=1))
    done < "$bundle"
    ((count > 0)) || { echo 'The public CA bundle is empty' >&2; exit 1; }
    install -d -o root -g root -m 0755 "$target/etc/ssh"
    temporary_bundle=$(mktemp "$target/etc/ssh/.homelab-user-ca.XXXXXX")
    install -o root -g root -m 0644 "$bundle" "$temporary_bundle"
    mv -T "$temporary_bundle" "$target/etc/ssh/homelab-user-ca.pub"
    install -D -o root -g root -m 0644 /dev/stdin "$target/etc/ssh/ansible-principals" <<< ansible
    install -D -o root -g root -m 0644 "$setup/sshd/00-ansible-management.conf" "$target/etc/ssh/sshd_config.d/00-ansible-management.conf"
fi
