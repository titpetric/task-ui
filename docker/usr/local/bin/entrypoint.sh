#!/bin/sh
# entrypoint.sh: provision the container, then exec task-ui.
#
# Two directories of shell scripts run in order, lexical within each:
#
#   /app/entrypoint.d    this repository's, baked into the image. The Go
#                        tooling a run expects on its PATH lives here, so
#                        the Dockerfile holds no `go install` line. Every
#                        tool installs at @latest, or @main where there is
#                        no tag: a pinned version is a number somebody has
#                        to remember to raise, and this instance would
#                        rather a restart pick up a release.
#   /data/entrypoint.d   the operator's, on the mounted volume. Another
#                        repository drops scripts here to add software
#                        without rebuilding the image.
#
# /app is HOME, so a script here reaches ~/.docker and ~/.ssh as the
# server does.
#
# A missing directory is not an error and neither is an empty one: an
# instance with no provisioning starts.
#
# A script that fails stops the boot. A half-provisioned instance that
# serves is worse than one that does not come up: the jobs it runs would
# fail one at a time, somewhere else, with nothing naming the cause.
# Change `exit 1` below to `continue` to make a failure a warning.
set -eu

run_scripts() {
	dir="$1"
	[ -d "$dir" ] || return 0
	for script in "$dir"/*.sh; do
		[ -f "$script" ] || continue
		echo "entrypoint: $script"
		if ! sh -x "$script"; then
			echo "entrypoint: $script failed" >&2
			exit 1
		fi
	done
}

run_scripts /app/entrypoint.d
run_scripts /data/entrypoint.d

exec /usr/local/bin/task-ui "$@"
