#!/usr/bin/env bash
set -euo pipefail
[[ -f /.dockerenv && $EUID == 0 ]] || { echo 'Run in a disposable root container' >&2; exit 1; }
root=$(mktemp -d)
HOMELABD_API_TOKEN=$(printf '%064d' 1)
export HOMELABD_API_TOKEN
bash /assets/agent/install.sh "$root" /bin/true https://api.example
[[ $(stat -c '%u:%g:%a' "$root/etc/homelabd") == 0:0:700 ]]
[[ $(stat -c '%u:%g:%a' "$root/etc/homelabd/environment") == 0:0:600 ]]
[[ $(< "$root/etc/homelabd/environment") == "API_TOKEN=$HOMELABD_API_TOKEN" ]]
unset HOMELABD_API_TOKEN
bash /assets/agent/install.sh "$root" /bin/true https://api.example
[[ -s "$root/etc/homelabd/environment" ]]
fresh=$(mktemp -d)
bash /assets/agent/install.sh "$fresh" /bin/true https://api.example
[[ $(stat -c '%u:%g:%a' "$fresh/etc/homelabd") == 0:0:700 ]]
[[ ! -e "$fresh/etc/homelabd/environment" ]]
HOMELABD_API_TOKEN='invalid token'
export HOMELABD_API_TOKEN
if bash /assets/agent/install.sh "$fresh" /bin/true https://api.example; then exit 1; fi
echo 'Agent directory, embedded token permissions and validation passed'
