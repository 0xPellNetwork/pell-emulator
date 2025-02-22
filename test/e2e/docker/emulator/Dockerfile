FROM golang:1.22-bullseye AS build-pelldvs

ARG HTTP_PROXY
ARG HTTPS_PROXY

WORKDIR /app

ARG DEBUG_ENABLED
ARG GITHUB_TOKEN
RUN if [ -z "$GITHUB_TOKEN" ]; then echo "GITHUB_TOKEN is not set" && exit 1; fi
RUN git config --global url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/0xPellNetwork".insteadOf "https://github.com/0xPellNetwork"

ARG PELLDVS_VERSION
RUN git clone https://github.com/0xPellNetwork/pelldvs.git -b $PELLDVS_VERSION --depth 1
# Build pelldvs
WORKDIR /app/pelldvs
RUN --mount=type=cache,target="/go/pkg/mod" \
    --mount=type=cache,target="/root/.cache/go-build" \
    make build

FROM golang:1.23-bullseye AS build-emulator

## Setup env
WORKDIR /app/emulator

ARG GITHUB_TOKEN
RUN if [ -z "$GITHUB_TOKEN" ]; then echo "GITHUB_TOKEN is not set" && exit 1; fi
RUN git config --global url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/0xPellNetwork".insteadOf "https://github.com/0xPellNetwork"

COPY go.mod go.sum ./
RUN --mount=type=cache,target="/go/pkg/mod" \
    go mod download

COPY . .
RUN --mount=type=cache,target="/go/pkg/mod" \
    --mount=type=cache,target="/root/.cache/go-build" \
    if [ "${DEBUG_ENABLED}" = "true" ]; then \
        make build-debug; \
    else \
        make build; \
    fi

########## Setup runtime env ##########
FROM golang:1.23-bullseye AS runtime
RUN apt-get update -yqq && apt-get install -yqq openssh-server curl jq less
RUN mkdir -p /run/sshd && chmod 0755 /run/sshd

RUN curl -L https://foundry.paradigm.xyz | bash
ENV PATH="$PATH:/root/.foundry/bin"
RUN foundryup

RUN go install github.com/mikefarah/yq/v4@latest
RUN go install github.com/go-delve/delve/cmd/dlv@latest
COPY --from=build-pelldvs /app/pelldvs/build/pelldvs /usr/bin/pelldvs
COPY --from=build-emulator /app/emulator/build/pell-emulator /usr/bin/pell-emulator
COPY ./test/e2e/docker/emulator/scripts /root/scripts

COPY ./test/e2e/docker/ssh /root/.ssh
RUN chmod 600 /root/.ssh/*


WORKDIR /root

# expose port
EXPOSE 22

CMD ["bash"]
