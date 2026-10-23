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
RUN GOBIN=/usr/local/bin CGO_ENABLED=0 go install github.com/go-task/task/v3/cmd/task@latest
RUN GOBIN=/usr/local/bin CGO_ENABLED=0 go install github.com/titpetric/atkins@latest

WORKDIR /task-ui
ADD . /task-ui

RUN go test ./...
RUN GOBIN=/usr/local/bin CGO_ENABLED=0 go install .

# The final image environment.
FROM golang:1.27-alpine

RUN apk --no-cache add bash git colordiff openssh mailcap curl

# mailcap - adds /etc/mime.types for golang server
# colordiff - as it sounds like
# git, ssh - allows git clone from github, bitbucket,...
# bash - run bash scripts, commands

# Copy over Task, Atkins, Task UI and docker-cli
COPY --from=builder /usr/local/bin/task /usr/local/bin
COPY --from=builder /usr/local/bin/atkins /usr/local/bin
COPY --from=builder /usr/local/bin/task-ui /usr/local/bin
COPY --from=docker /usr/local/bin/docker /usr/local/bin
COPY --from=docker /usr/local/libexec/docker/cli-plugins /usr/local/libexec/docker/cli-plugins

# Configure the final image with whatever.
ADD /docker /

# /app holds the image's own configuration and nothing a run writes.
# /data holds the database, .htpasswd and the workspace folder,
# /data/workspaces; /keys every key of the instance. Mount both.
WORKDIR /app
RUN mkdir -p /data /keys && chmod 700 /keys
VOLUME /data
VOLUME /keys

RUN docker --version && docker buildx version && docker compose version

ENV GOPATH=/go \
	TASK_UI_LISTEN=:3000

EXPOSE 3000

ENTRYPOINT ["/usr/local/bin/task-ui"]
