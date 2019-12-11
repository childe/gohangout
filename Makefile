hash:=$(shell git rev-parse --short HEAD)

.PHONY: gohangout all clean check

GIT_COMMIT := $(shell git rev-list -1 HEAD)

gohangout:
	mkdir -p build/
	go build -o build/gohangout -ldflags "-X main.gitCommit=$(GIT_COMMIT)"

all:
	make check
	@echo $(hash)
	mkdir -p build/

	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.gitCommit=$(GIT_COMMIT)" -o build/gohangout-windows-x64-$(hash).exe
	GOOS=windows GOARCH=386 go build -ldflags "-X main.gitCommit=$(GIT_COMMIT)" -o build/gohangout-windows-386-$(hash).exe
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.gitCommit=$(GIT_COMMIT)" -o build/gohangout-linux-x64-$(hash)
	GOOS=linux GOARCH=386 go build -ldflags "-X main.gitCommit=$(GIT_COMMIT)" -o build/gohangout-linux-386-$(hash)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.gitCommit=$(GIT_COMMIT)" -o build/gohangout-darwin-x64-$(hash)

clean:
	rm -rf build/*

check:
	git diff-index --quiet HEAD --
