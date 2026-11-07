# Final image dependencies: the docker CLI with the buildx and compose
# plugins, and the daemon behind TASK_UI_DOCKER=dind. Without buildx,
# `docker build` on a current daemon falls back to the deprecated legacy
# builder, and BuildKit features error out.
FROM docker:29-cli AS docker
FROM docker:29-dind AS dind

# Build project
FROM golang:1.27-bookworm AS builder

# bash runs the jobs in the runner and api tests, make is a convention
# they detect.
RUN apt-get update && apt-get install -y --no-install-recommends \
		bash \
		git \
		make \
	&& rm -rf /var/lib/apt/lists/*

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

# The final image environment. Debian, not Alpine: an entrypoint.d script
# installs software with `apt-get`, which is half of what this base is
# for. The other half is the Go toolchain the same scripts build with.
FROM golang:1.27-bookworm

# iptables is dockerd's; mailcap is /etc/mime.types, which an artifact
# download answers with; tzdata is the scheduler's zone. The rest is what
# a job expects to find.
RUN apt-get update && apt-get install -y --no-install-recommends \
		bash \
		ca-certificates \
		colordiff \
		curl \
		git \
		iptables \
		mailcap \
		openssh-client \
		tzdata \
	&& rm -rf /var/lib/apt/lists/*

# Task UI and the docker CLI. task and atkins are not here:
# /app/entrypoint.d installs them at boot, into /go/bin, which is already
# first on PATH.
COPY --from=builder /out/task-ui /usr/local/bin
COPY --from=docker /usr/local/bin/docker /usr/local/bin
COPY --from=docker /usr/local/libexec/docker/cli-plugins /usr/local/libexec/docker/cli-plugins

# The daemon, started only for TASK_UI_DOCKER=dind. The release binaries
# are static, so the ones built for Alpine run here. docker 29 carries
# containerd inside dockerd, so the shim, runc and the two helpers are
# the whole runtime; modprobe is the release's own shim, which answers 0
# for a module the kernel already holds.
COPY --from=dind \
	/usr/local/bin/dockerd \
	/usr/local/bin/containerd-shim-runc-v2 \
	/usr/local/bin/runc \
	/usr/local/bin/docker-init \
	/usr/local/bin/docker-proxy \
	/usr/local/bin/modprobe \
	/usr/local/bin/

# Configure the final image with whatever.
ADD /docker /

# /app is HOME and holds the image's own configuration: the entrypoint
# scripts, ~/.ssh and ~/.docker, where docker reads the registry logins
# from config.json. Mount that directory to keep the host's logins, or
# let the Docker Registry integration write it.
#
# /data holds the database, .htpasswd and the workspace folder,
# /data/workspaces; /keys every key of the instance. Mount both.
WORKDIR /app
RUN mkdir -p /data /keys /app/.docker && chmod 700 /keys /app/.docker
VOLUME /data
VOLUME /keys

RUN docker --version \
	&& docker buildx version \
	&& docker compose version \
	&& dockerd --version

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
