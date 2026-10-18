# Final image dependency
FROM docker:20.10 AS docker

# Build project
FROM golang:1.27-alpine AS builder

RUN apk --no-cache add git
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
# curl - install compose v2

# Copy over Task, Atkins, Task UI and docker-cli
COPY --from=builder /usr/local/bin/task /usr/local/bin
COPY --from=builder /usr/local/bin/atkins /usr/local/bin
COPY --from=builder /usr/local/bin/task-ui /usr/local/bin
COPY --from=docker /usr/local/bin/docker /usr/local/bin

# Configure the final image with whatever.
ADD /docker /

WORKDIR /app
RUN mkdir -p /data
VOLUME /data

RUN	DOCKER_CONFIG=${DOCKER_CONFIG:-$HOME/.docker} && \
	mkdir -p $DOCKER_CONFIG/cli-plugins && \
	curl -sSL https://github.com/docker/compose/releases/download/v2.2.3/docker-compose-linux-x86_64 -o $DOCKER_CONFIG/cli-plugins/docker-compose && \
	chmod +x $DOCKER_CONFIG/cli-plugins/docker-compose

ENV GOPATH=/go \
	TASK_UI_LISTEN=:3000 \
	TASK_UI_ROOT=/app \
	TASK_UI_DATA=/data \
	TASK_UI_HTPASSWD=/app/.htpasswd \
	TASK_UI_SSH_DIR=/root/.ssh

EXPOSE 3000

ENTRYPOINT ["/usr/local/bin/task-ui"]
