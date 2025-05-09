# Final image dependency
FROM docker:20.10 AS docker

# Build project
FROM golang:1.24-alpine AS builder

RUN GOBIN=/usr/local/bin CGO_ENABLED=0 go install github.com/go-task/task/v3/cmd/task@latest

WORKDIR /task-ui
ADD . /task-ui

RUN apk --no-cache add git
RUN go test ./...
RUN GOBIN=/usr/local/bin go install .

# Ttyrec build image
FROM alpine:latest AS ttyrecord

WORKDIR /tmp
RUN apk --no-cache add build-base
RUN apk --no-cache add git
RUN git clone https://github.com/ovh/ovh-ttyrec.git
WORKDIR /tmp/ovh-ttyrec
RUN STATIC=1 ./configure --bindir=/usr/local/bin && make && make install

# The final image environment.
FROM golang:1.24-alpine

RUN apk --no-cache add bash git colordiff openssh mailcap curl

# mailcap - adds /etc/mime.types for golang server
# colordiff - as it sounds like
# git, ssh - allows git clone from github, bitbucket,...
# bash - run bash scripts, commands
# curl - install compose v2

# Copy over Task, Task UI and docker-cli
COPY --from=builder /usr/local/bin/task /usr/local/bin
COPY --from=builder /usr/local/bin/task-ui /usr/local/bin
COPY --from=docker /usr/local/bin/docker /usr/local/bin
COPY --from=ttyrecord /usr/local/bin/* /usr/local/bin

# Configure the final image with whatever.
ADD /docker /

WORKDIR /app

RUN	DOCKER_CONFIG=${DOCKER_CONFIG:-$HOME/.docker} && \
	mkdir -p $DOCKER_CONFIG/cli-plugins && \
	curl -sSL https://github.com/docker/compose/releases/download/v2.2.3/docker-compose-linux-x86_64 -o $DOCKER_CONFIG/cli-plugins/docker-compose && \
	chmod +x $DOCKER_CONFIG/cli-plugins/docker-compose

ENV GOPATH=/go

ENTRYPOINT ["/usr/local/bin/task-ui"]