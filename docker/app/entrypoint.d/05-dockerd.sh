#!/bin/sh
# dockerd: the container's own daemon, so a job builds images with no
# socket from the host mounted in.
#
# Off unless TASK_UI_DOCKER=dind, and a no-op when a socket is already
# mounted: the mount is the operator saying which daemon to use, and a
# second daemon on the same path would hide it.
#
# The container needs `privileged: true`. Without it dockerd starts and
# fails on the first container it tries to run, which is a job failing
# minutes later with nothing naming the cause.
#
# /var/lib/docker needs a volume, and not for the cache: the layers it
# writes are overlayfs, and an overlayfs upperdir may not itself sit on
# overlayfs. Without one dockerd starts, pulls, and fails every container with
# `failed to mount ... fstype: overlay ... invalid argument`.
set -eu

[ "${TASK_UI_DOCKER:-}" = "dind" ] || exit 0

if [ -S /var/run/docker.sock ]; then
	echo "dockerd: /var/run/docker.sock is mounted, using it"
	exit 0
fi

dockerd >/var/log/dockerd.log 2>&1 &

# The daemon answers in about a second; 60 is the ceiling, not the wait.
i=0
while [ "$i" -lt 60 ]; do
	if docker version >/dev/null 2>&1; then
		exit 0
	fi
	i=$((i + 1))
	sleep 1
done

echo "dockerd: no answer in 60s, privileged: true is what it needs" >&2
tail -20 /var/log/dockerd.log >&2
exit 1
