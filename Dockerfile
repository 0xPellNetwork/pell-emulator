FROM golang:1.23-bullseye AS build

ARG HTTP_PROXY
ARG HTTPS_PROXY

## Setup env
WORKDIR /app/pell-emulator

# TODO(jimmy): remove this once the deps are public
RUN --mount=type=secret,id=github_token \
    git config --global url."https://$(cat /run/secrets/github_token):x-oauth-basic@github.com/0xPellNetwork".insteadOf "https://github.com/0xPellNetwork"

COPY go.mod go.sum ./
RUN --mount=type=cache,target="/go/pkg/mod" \
    go mod download

COPY . .
RUN --mount=type=cache,target="/go/pkg/mod" \
    --mount=type=cache,target="/root/.cache/go-build" \
    make build

########## Setup runtime env ##########
RUN cp /app/pell-emulator/build/pell-emulator /usr/bin/pell-emulator

WORKDIR /root

CMD ["pell-emulator", "start", "--home", "/root/.pell-emulator", "--auto-update-connector", "true"]
