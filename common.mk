# This contains Makefile logic that is common to several makefiles

BUILD_TAGS ?= pellemulator

#COMMIT_HASH := $(shell git rev-parse --short HEAD)
COMMIT_HASH := "v0.0.1" # TODO: update this to use the actual commit hash
LD_FLAGS = -X github.com/0xPellNetwork/pell-emulator/versioin/version.GitCommitHash=$(COMMIT_HASH)
BUILD_FLAGS = -mod=readonly -ldflags "$(LD_FLAGS)"
# allow users to pass additional flags via the conventional LDFLAGS variable
LD_FLAGS += $(LDFLAGS)

# handle nostrip
ifeq (,$(findstring nostrip,$(EMULATOR_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
  LD_FLAGS += -s -w
endif

# handle race
ifeq (race,$(findstring race,$(EMULATOR_BUILD_OPTIONS)))
  CGO_ENABLED=1
  BUILD_FLAGS += -race
endif

# handle cleveldb
ifeq (cleveldb,$(findstring cleveldb,$(EMULATOR_BUILD_OPTIONS)))
  CGO_ENABLED=1
  BUILD_TAGS += cleveldb
endif

# handle badgerdb
ifeq (badgerdb,$(findstring badgerdb,$(EMULATOR_BUILD_OPTIONS)))
  BUILD_TAGS += badgerdb
endif

# handle rocksdb
ifeq (rocksdb,$(findstring rocksdb,$(EMULATOR_BUILD_OPTIONS)))
  CGO_ENABLED=1
  BUILD_TAGS += rocksdb
endif

# handle boltdb
ifeq (boltdb,$(findstring boltdb,$(EMULATOR_BUILD_OPTIONS)))
  BUILD_TAGS += boltdb
endif
