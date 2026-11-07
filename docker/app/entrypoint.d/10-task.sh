#!/bin/sh
# go-task: a job source task-ui detects, and the `task` a Taskfile job runs.
#
# @latest, and on every boot. GOBIN is nothing this image sets, so the
# binary lands in the default /go/bin, which is already first on PATH.
set -eu

CGO_ENABLED=0 go install github.com/go-task/task/v3/cmd/task@latest
