hash:=$(shell git rev-parse --short HEAD)

.PHONY: gohangout all clean check test docker linux-binary

gohangout:
	mkdir -p build/
	go build -o build/gohangout

linux-binary:
	mkdir -p build
	@echo "Building gohangout binary in docker container"
	@if [ -n "$(GOPATH)" ]; then\
		docker run -e CGO_ENABLED=0 -v $(GOPATH):/go -v $(PWD):/gohangout -w /gohangout golang:1.17 go build -ldflags "-X main.version=$(hash)" -o build/gohangout;\
	else\
		docker run -e CGO_ENABLED=0 -v $(PWD):/gohangout -w /gohangout golang:1.17 go build -ldflags "-X main.version=$(hash)" -o build/gohangout;\
	fi

docker: linux-binary
	docker build -t gohangout .

all: check
	@echo $(hash)
	mkdir -p build/

	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(hash)" -o build/gohangout-windows-x64-$(hash).exe
	GOOS=windows GOARCH=386 go build -ldflags "-X main.version=$(hash)" -o build/gohangout-windows-386-$(hash).exe
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(hash)" -o build/gohangout-linux-x64-$(hash)
	GOOS=linux GOARCH=386 go build -ldflags "-X main.version=$(hash)" -o build/gohangout-linux-386-$(hash)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(hash)" -o build/gohangout-darwin-x64-$(hash)

clean:
	rm -rf build/*

check:
	git diff-index --quiet HEAD --

test:
	go test input/*.go
	go test output/*.go
	go test filter/*.go
	go test condition_filter/*.go
	go test value_render/*.go
