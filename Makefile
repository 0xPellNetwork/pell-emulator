include common.mk

PACKAGES=$(shell go list ./...)
BUILDDIR?=$(CURDIR)/build
OUTPUT?=$(BUILDDIR)/pell-emulator
CMDDIR?=$(CURDIR)/cmd/pell-emulator

HTTPS_GIT := https://github.com/0xPellNetwork/pell_emulator.git
CGO_ENABLED ?= 0

# Process Docker environment varible TARGETPLATFORM
# in order to build binary with correspondent ARCH
# by default will always build for linux/amd64
TARGETPLATFORM ?=
GOOS ?= linux
GOARCH ?= amd64
GOARM ?=

ifeq (linux/arm,$(findstring linux/arm,$(TARGETPLATFORM)))
	GOOS=linux
	GOARCH=arm
	GOARM=7
endif

ifeq (linux/arm/v6,$(findstring linux/arm/v6,$(TARGETPLATFORM)))
	GOOS=linux
	GOARCH=arm
	GOARM=6
endif

ifeq (linux/arm64,$(findstring linux/arm64,$(TARGETPLATFORM)))
	GOOS=linux
	GOARCH=arm64
	GOARM=7
endif

ifeq (linux/386,$(findstring linux/386,$(TARGETPLATFORM)))
	GOOS=linux
	GOARCH=386
endif

ifeq (linux/amd64,$(findstring linux/amd64,$(TARGETPLATFORM)))
	GOOS=linux
	GOARCH=amd64
endif

ifeq (linux/mips,$(findstring linux/mips,$(TARGETPLATFORM)))
	GOOS=linux
	GOARCH=mips
endif

ifeq (linux/mipsle,$(findstring linux/mipsle,$(TARGETPLATFORM)))
	GOOS=linux
	GOARCH=mipsle
endif

ifeq (linux/mips64,$(findstring linux/mips64,$(TARGETPLATFORM)))
	GOOS=linux
	GOARCH=mips64
endif

ifeq (linux/mips64le,$(findstring linux/mips64le,$(TARGETPLATFORM)))
	GOOS=linux
	GOARCH=mips64le
endif

ifeq (linux/riscv64,$(findstring linux/riscv64,$(TARGETPLATFORM)))
	GOOS=linux
	GOARCH=riscv64
endif

#? all: Run target build, test and install
all: build install
.PHONY: all

###############################################################################
###                                Build                            ###
###############################################################################

#? build: Build
build:
	CGO_ENABLED=$(CGO_ENABLED) go build $(BUILD_FLAGS) -tags '$(BUILD_TAGS)' -o $(OUTPUT) $(CMDDIR)
.PHONY: build

#? install: Install  to GOBIN
install:
	CGO_ENABLED=$(CGO_ENABLED) go install $(BUILD_FLAGS) -tags $(BUILD_TAGS) $(CMDDIR)
.PHONY: install


###############################################################################
###                              Distribution                               ###
###############################################################################


#? go-mod-cache: Download go modules to local cache
go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download
.PHONY: go-mod-cache

#? go.sum: Ensure dependencies have not been modified
go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify
	@go mod tidy

###############################################################################
###                  Formatting, linting, and vetting                       ###
###############################################################################

format:
	find . -name '*.go' -type f -not -path "*.git*" -not -name '*.pb.go' -not -name '*pb_test.go' | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "*.git*"  -not -name '*.pb.go' -not -name '*pb_test.go' | xargs goimports -w -local github.com/0xPellNetwork/pelldvs
.PHONY: format

#? lint: Run latest golangci-lint linter
lint:
	@echo "--> Running linter"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run
.PHONY: lint

#? vulncheck: Run latest govulncheck
vulncheck:
	@go run golang.org/x/vuln/cmd/govulncheck@latest ./...
.PHONY: vulncheck

#? lint-typo: Run codespell to check typos
lint-typo:
	which codespell || pip3 install codespell
	@codespell
.PHONY: lint-typo

#? lint-typo: Run codespell to auto fix typos
lint-fix-typo:
	@codespell -w
.PHONY: lint-fix-typo

lint-yaml:
	@yamllint -c .github/linters/yaml-lint.yml .

DESTINATION = ./index.html.md


###############################################################################
###                           Documentation                                 ###
###############################################################################



###############################################################################
###                       Local testnet using docker                        ###
###############################################################################

#? build-linux: Build linux binary on other platforms
build-linux:
	GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) $(MAKE) build
.PHONY: build-linux


# Implements test splitting and running. This is pulled directly from
# the github action workflows for better local reproducibility.

GO_TEST_FILES != find $(CURDIR) -name "*_test.go"

# default to four splits by default
NUM_SPLIT ?= 4

$(BUILDDIR):
	mkdir -p $@

# The format statement filters out all packages that don't have tests.
# Note we need to check for both in-package tests (.TestGoFiles) and
# out-of-package tests (.XTestGoFiles).
$(BUILDDIR)/packages.txt:$(GO_TEST_FILES) $(BUILDDIR)
	go list -f "{{ if (or .TestGoFiles .XTestGoFiles) }}{{ .ImportPath }}{{ end }}" ./... | sort > $@

split-test-packages:$(BUILDDIR)/packages.txt
ifeq ($(UNAME_S),Linux)
	split -d -n l/$(NUM_SPLIT) $< $<.
else
	total_lines=$$(wc -l < $<); \
	lines_per_file=$$((total_lines / $(NUM_SPLIT) + 1)); \
	split -d -l $$lines_per_file $< $<.
endif
test-group-%:split-test-packages
	cat $(BUILDDIR)/packages.txt.$* | xargs go test -mod=readonly -timeout=15m -race -coverprofile=$(BUILDDIR)/$*.profile.out

#? help: Get more info on make commands.
help: Makefile
	@echo " Choose a command run in pell emulator:"
	@sed -n 's/^#?//p' $< | column -t -s ':' |  sort | sed -e 's/^/ /'
.PHONY: help

test:
	go test -v ./...
.PHONY: test

check-env-gh-token:
	@if [ -z "$${GITHUB_TOKEN}" ]; then \
		echo "Error: GITHUB_TOKEN is not set in environment"; \
		exit 1; \
	else \
	  	echo "$${GITHUB_TOKEN}" > ./.env.github_token.txt ; \
		echo "GITHUB_TOKEN is set."; \
	fi

docker-build: check-env-gh-token
	docker compose build

docker-up-all:
	docker compose down -v && docker compose up -d

docker-down:
	docker compose down -v

# Run goimports-reviser to lint and format imports
lint-imports:
	@find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" | while read -r file; do \
		goimports-reviser -company-prefixes github.com/0xPellNetwork/pell-emulator -rm-unused -format "$$file"; \
	done