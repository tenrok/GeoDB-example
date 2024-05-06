.ONESHELL: build
.PHONY: build windows linux
.DEFAULT_GOAL=build

ifeq ($(OS),Windows_NT)
SHELL=C:/Program Files/Git/bin/bash.exe
build: windows
else
SHELL=/usr/bin/bash
build: linux
endif

prepare:
	cp -n ./configs/dist.example.yaml ./configs/example.yaml

windows: prepare
	export GOOS=windows
	export GOARCH=amd64
	export CGO_ENABLED=1
	#go build -a -tags "osusergo,netgo,sqlite_omit_load_extension" -trimpath -ldflags '-s -w -extldflags "-static"' -o example.exe ./cmd/example
	go build -tags "osusergo,netgo,sqlite_omit_load_extension" -trimpath -o example.exe ./cmd/example

linux: prepare
	export GOOS=linux
	export GOARCH=amd64
	export CGO_ENABLED=1
	#go build -a -tags "osusergo,netgo,sqlite_omit_load_extension" -trimpath -ldflags '-s -w -extldflags "-static"' -o example ./cmd/example
	go build -tags="osusergo,netgo,sqlite_omit_load_extension" -trimpath -o example ./cmd/example
