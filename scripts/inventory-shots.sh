#!/usr/bin/env bash
# Take the screenshots of the Inventory page.
#
#   scripts/inventory-shots.sh
#
# Builds task-ui and internal/cmd/fakehub, a Beszel hub of three hosts
# (pve with an extra disk, dev1, docker1 down), puts an ssh in the PATH
# of the instance that answers for pve and dev1 with df and docker ps,
# sets the Beszel integration up through the API and captures the page,
# then with pve open, in both themes. Writes docs/assets/inventory-*.png.
#
# Needs chromium and curl in PATH.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

for tool in chromium curl; do
  command -v "$tool" >/dev/null || { echo "$tool not found in PATH" >&2; exit 1; }
done

work="$(mktemp -d)"
pids=()
cleanup() {
  for pid in "${pids[@]}"; do kill "$pid" 2>/dev/null || true; done
  rm -rf "$work"
}
trap cleanup EXIT

go build -o "$work/task-ui" .
go build -o "$work/fakehub" ./internal/cmd/fakehub

hub=3191
port=3192
"$work/fakehub" -listen "127.0.0.1:${hub}" &
pids+=("$!")

# The ssh of the instance: debian signs in to pve and dev1, which print
# their filesystems and containers; docker1 is down.
mkdir -p "$work/bin" "$work/demo"
cat >"$work/bin/ssh" <<'SSH'
#!/bin/sh
for last; do :; done
login=$(eval echo \${$(($# - 1))})
case "$login" in
  debian@10.0.0.2) host=pve ;;
  debian@10.0.0.11) host=dev1 ;;
  *) echo "ssh: connect to host ${login#*@} port 22: No route to host" >&2; exit 255 ;;
esac
case "$last" in *"df -PT"*) ;; *) exit 0 ;; esac
echo 'Filesystem     Type     1-blocks         Used    Available Capacity Mounted on'
echo 'udev           devtmpfs 33641611264            0  33641611264       0% /dev'
echo 'tmpfs          tmpfs     6731628544      2240512   6729388032       1% /run'
echo 'tmpfs          tmpfs    33658142720            0  33658142720       0% /dev/shm'
if [ "$host" = pve ]; then
  echo '/dev/nvme0n1p2 ext4    100822970368  44238602240  51428372480      47% /'
  echo '/dev/nvme0n1p1 vfat       535805952      6197248    529608704       2% /boot/efi'
  echo '/dev/sda1      ext4   2000035000320 1185736835072 712648785920      63% /mnt/data'
  echo '/dev/sda1      ext4   2000035000320 1185736835072 712648785920      63% /srv/backup'
  echo 'overlay        overlay 100822970368  44238602240  51428372480      47% /var/lib/docker/overlay2/3f1c0a/merged'
  echo 'overlay        overlay 100822970368  44238602240  51428372480      47% /var/lib/docker/overlay2/9be2d4/merged'
else
  echo '/dev/vda1      ext4    210237997056  94916268032 106712686592      48% /'
  echo '/dev/vda15     vfat       129718272     12731392    116986880      10% /boot/efi'
  echo 'overlay        overlay 210237997056  94916268032 106712686592      48% /var/lib/docker/overlay2/c2a9e1/merged'
fi
echo '@@docker'
if [ "$host" = pve ]; then
  echo '{"Names":"postgres","Image":"postgres:17","State":"running","Status":"Up 12 days (healthy)","Size":"63.2MB (virtual 501MB)"}'
  echo '{"Names":"prometheus","Image":"prom/prometheus:v3.1.0","State":"running","Status":"Up 12 days","Size":"0B (virtual 297MB)"}'
  echo '{"Names":"grafana","Image":"grafana/grafana:11.4.0","State":"running","Status":"Up 12 days","Size":"1.21MB (virtual 487MB)"}'
  echo '{"Names":"minio","Image":"minio/minio:latest","State":"running","Status":"Up 3 days","Size":"0B (virtual 175MB)"}'
  echo '{"Names":"loki","Image":"grafana/loki:3.3.2","State":"running","Status":"Up 12 days","Size":"12.4kB (virtual 82.1MB)"}'
  echo '{"Names":"caddy","Image":"caddy:2.9","State":"running","Status":"Up 12 days","Size":"4.1kB (virtual 48.2MB)"}'
  echo '{"Names":"backup-1729","Image":"restic/restic:0.17.3","State":"exited","Status":"Exited (0) 6 hours ago","Size":"1.5kB (virtual 42.6MB)"}'
else
  echo '{"Names":"task-ui","Image":"registry.cdn.si/task-ui:latest","State":"running","Status":"Up 2 hours","Size":"18.4MB (virtual 312MB)"}'
  echo '{"Names":"forgejo","Image":"codeberg.org/forgejo/forgejo:9","State":"running","Status":"Up 5 days","Size":"2.03MB (virtual 196MB)"}'
  echo '{"Names":"redis","Image":"redis:7-alpine","State":"running","Status":"Up 5 days","Size":"0B (virtual 41.4MB)"}'
  echo '{"Names":"caddy","Image":"caddy:2.9","State":"running","Status":"Up 5 days","Size":"4.1kB (virtual 48.2MB)"}'
fi
echo '@@exit 0'
SSH
chmod +x "$work/bin/ssh"

env PATH="$work/bin:$PATH" \
  TASK_UI_WORKSPACE="$work/demo" \
  TASK_UI_LISTEN="127.0.0.1:${port}" \
  TASK_UI_DATA="$work/data" \
  TASK_UI_SSH_DIR="$work/ssh" \
  TASK_UI_ADMINS=admin \
  TASK_UI_AUTOLOGIN=admin \
  "$work/task-ui" >"$work/task-ui.log" 2>&1 &
pids+=("$!")
for _ in $(seq 1 50); do
  curl -fs -o /dev/null "http://127.0.0.1:${port}/healthz" && break
  sleep 0.1
done
export TASK_UI_BASE="http://127.0.0.1:${port}"

# api calls the instance the way the pages do.
api() {
  local method="$1" path="$2"
  shift 2
  curl -fsS -o "$work/api.out" -X "$method" -H "Origin: ${TASK_UI_BASE}" -H "Content-Type: application/json" "$@" "${TASK_UI_BASE}/api${path}" \
    || { echo "api ${method} ${path} failed:" >&2; cat "$work/api.out" >&2; cat "$work/task-ui.log" >&2; exit 1; }
}

# The integrations keep their tokens encrypted with the instance's key,
# which the start made: the hub.
api PUT /integrations/beszel -d "{\"values\":{\"url\":\"http://127.0.0.1:${hub}\",\"token\":\"demo\"}}"

# A first view finds the logins and asks the hosts for their facts; the
# shots come after they answered.
curl -fsS -o /dev/null "${TASK_UI_BASE}/inventory"
sleep 1
curl -fsS -o /dev/null "${TASK_UI_BASE}/inventory"

export TASK_UI_SHOTS="$work"
shots=()
for theme in dark light; do
  shots+=("$(scripts/chromium.sh /inventory "inventory-${theme}" "$theme" 1440x1400)")
  shots+=("$(scripts/chromium.sh "/inventory#host=pve" "inventory-host-${theme}" "$theme" 1440x2850 5000)")
done

mkdir -p docs/assets
for file in "${shots[@]}"; do
  name="$(basename "$file" .png)"
  out="docs/assets/${name#task-ui-}.png"
  if command -v magick >/dev/null 2>&1; then
    magick "$file" -strip -colors 256 -define png:compression-level=9 "$out"
  else
    cp "$file" "$out"
  fi
  echo "$out"
done
