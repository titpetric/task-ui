# Final image dependency: the docker CLI with the buildx and compose
# plugins. Without buildx, `docker build` on a current daemon falls back
# to the deprecated legacy builder, and BuildKit features error out.
FROM docker:29-cli AS docker

# Build project
FROM golang:1.27-alpine AS builder

# bash runs the jobs in the runner and api tests, make is a convention
# they detect. coreutils for stty: busybox `stty size` answers from
# $COLUMNS/$LINES before the pty, so it misses a resize the tests make.
RUN apk --no-cache add bash coreutils git make

WORKDIR /task-ui
ADD . /task-ui

# No test here. The folder tests run task and atkins to detect a job, and
# neither is in this stage any more: the entrypoint installs them into the
# final image. ci:push runs the whole gate before it builds the image, so
# a test here was the tree tested twice and the second one needed tooling
# the image installs at boot.
#
# -o rather than `go install`, because GOBIN is nothing this image sets:
# what the entrypoint installs lands in the default /go/bin.
RUN CGO_ENABLED=0 go build -o /out/task-ui .

# The final image environment.
FROM golang:1.27-alpine

RUN apk --no-cache add bash git colordiff openssh mailcap curl

# mailcap - adds /etc/mime.types for golang server
# colordiff - as it sounds like
# git, ssh - allows git clone from github, bitbucket,...
# bash - run bash scripts, commands

# Task UI and docker-cli. task and atkins are not here: /root/entrypoint.d
# installs them at boot, into /go/bin, which is already first on PATH.
COPY --from=builder /out/task-ui /usr/local/bin
COPY --from=docker /usr/local/bin/docker /usr/local/bin
COPY --from=docker /usr/local/libexec/docker/cli-plugins /usr/local/libexec/docker/cli-plugins

# Configure the final image with whatever.
ADD /docker /

# /app is HOME and holds the image's own configuration: the entrypoint
# scripts, ~/.ssh and ~/.docker, where docker reads the registry logins
# from config.json. Mount that file to keep the host's logins, or let the
# Docker Registry integration write it.
#
# /data holds the database, .htpasswd and the workspace folder,
# /data/workspaces; /keys every key of the instance. Mount both.
WORKDIR /app
RUN mkdir -p /data /keys /app/.docker && chmod 700 /keys /app/.docker
VOLUME /data
VOLUME /keys

RUN docker --version && docker buildx version && docker compose version

# GOPATH and PATH are the base image's own, repeated here because the
# entrypoint scripts install into them: `go install` with no GOBIN writes
# to $GOPATH/bin, and a run finds what they wrote because it is on PATH.
#
# GOCACHE is moved under GOPATH so one mounted /go holds all three things
# a boot would otherwise redo: the binaries, the module downloads and the
# compiled objects. It defaults under HOME, which is not a volume, and an
# unmounted build cache is a hundred seconds of rebuilding tools that did
# not change.
ENV HOME=/app \
	GOPATH=/go \
	GOCACHE=/go/cache \
	PATH=/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
	TASK_UI_LISTEN=:3000

EXPOSE 3000

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
