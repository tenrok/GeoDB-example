.ONESHELL:
.PHONY: build build_linux build_windows clean all

ROOT_DIR:=$(shell dirname "$(realpath $(firstword $(MAKEFILE_LIST)))")
OUTPUT_DIR:=$(ROOT_DIR)/bin
BASENAME=geodb-example

GOOS=linux
GOARCH=amd64
CGO_ENABLED=0
TAGS=-tags "osusergo,netgo,sqlite_omit_load_extension"
LDLAGS=-ldflags "-s -w -extldflags '-static'"
EXT=
OUTPUT=$(OUTPUT_DIR)/$(BASENAME)-$(GOOS)-$(GOARCH)$(EXT)

TARGET=build_linux

ifeq ($(OS),Windows_NT)
	SHELL="$(PROGRAMFILES)/Git/bin/bash.exe"
	TARGET=build_windows
else
	SHELL=/usr/bin/bash
endif

build: $(TARGET)

build_linux: --compile

build_windows: GOOS=windows
build_windows: EXT=.exe
build_windows: --compile

clean:
	@find $(OUTPUT_DIR) -name '$(BASENAME)[-?][a-zA-Z0-9]*[-?][a-zA-Z0-9]*' -delete

all:
	@make clean
	@make linux
	@make windows

--compile:
	@cd "$(ROOT_DIR)"
	@cp -n ./configs/dist.example.yaml ./configs/example.yaml
	@export GOOS=$(GOOS)
	@export GOARCH=$(GOARCH)
	@export CGO_ENABLED=$(CGO_ENABLED)
	@go build -trimpath $(TAGS) $(LDLAGS) -o "$(OUTPUT)" "./cmd/example"
