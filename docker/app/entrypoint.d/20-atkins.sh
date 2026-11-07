#!/bin/sh
# atkins: the pipeline runner most jobs of this instance are declared in.
#
# @latest, not a version, and it installs on every boot rather than only
# when absent: a pin is a number somebody has to remember to raise, and
# skipping the install when the binary exists means a restart never picks
# up a release. The module cache on /go makes the no-change case a proxy
# round trip and a cached build.
set -eu

CGO_ENABLED=0 go install github.com/titpetric/atkins@latest
