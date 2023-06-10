git_rev    = $(shell git rev-parse --short HEAD)
git_tag    = $(shell git describe --tags --abbrev=0)
git_branch = $(shell git rev-parse --abbrev-ref HEAD)
app_name   = "headscale-panel"

BuildArch = $(shell go env GOARCH)
BuildOS = $(shell go env GOOS)

#BuildVersion := $(git_branch)_$(git_rev)
BuildVersion := $(shell echo $Branch)
BuildTime := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BuildCommit := $(shell git rev-parse --short HEAD)
BuildGoVersion := $(shell go version)

# in detached HEAD state
ifeq ($(git_branch), HEAD)
	git_branch = $(shell git show-ref | grep $(shell git show HEAD | sed -n 1p | cut -d " " -f 2) | sed 's|.*/\(.*\)|\1|' | grep -v HEAD | sort | uniq | head -n 1)
	# when git checkout <<tag>>, branch may still be empty
	ifeq ($(git_branch), )
		git_branch := $(git_tag)
	endif
	BuildVersion := $(git_branch)_$(git_rev)
endif

ifeq ($(git_branch), dev)
#	BuildVersion := develop_$(git_rev)
	BuildBranch := dev
endif

ifeq ($(git_branch), master|main)
#	BuildVersion := release_$(git_tag)_$(git_rev)
	BuildBranch := release
endif

# -ldflag parameters
GOLDFLAGS = -s -w -X 'headscale-panel/version.Version=$(git_tag)'
GOLDFLAGS += -X 'headscale-panel/version.BuildTime=$(BuildTime)'
#GOLDFLAGS += -X 'main.BuildCommit=$(BuildCommit)'
GOLDFLAGS += -X 'headscale-panel/version.BuildGoVersion=$(BuildGoVersion)'
GOLDFLAGS += -X 'headscale-panel/version.Branch=$(BuildBranch)'
GOLDFLAGS += -X 'headscale-panel/version.OS=$(BuildOS)'
GOLDFLAGS += -X 'headscale-panel/version.Arch=$(BuildArch)'

.PHONY: mod build

# go mod
mod:
	go mod tidy


build:
	go build -o "bin/$(app_name)" -ldflags "$(GOLDFLAGS)" -gcflags "-trimpath -m"

build-all: build-linux-x86 build-linux-x64 build-linux-arm64 build-mac-intel build-mac-silicon


build-linux-x86:
	GOOS=linux GOARCH=386 go build -o "bin/$(app_name)_linux_x86" -ldflags "$(GOLDFLAGS)" -gcflags "-trimpath -m"

build-linux-x64:
	GOOS=linux GOARCH=amd64 go build -o "bin/$(app_name)_linux_x64" -ldflags "$(GOLDFLAGS)" -gcflags "-trimpath -m"

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o "bin/$(app_name)_linux_arm64" -ldflags "$(GOLDFLAGS)" -gcflags "-trimpath -m"

build-mac-intel:
	GOOS=darwin GOARCH=amd64 go build -o "bin/$(app_name)_darwin_x64" -ldflags "$(GOLDFLAGS)" -gcflags "-trimpath -m"

build-mac-silicon:
	GOOS=darwin GOARCH=amd64 go build -o "bin/$(app_name)_darwin_arm64" -ldflags "$(GOLDFLAGS)" -gcflags "-trimpath -m"
