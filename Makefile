hash:=$(shell git rev-parse --short HEAD)
tag?=$(hash)

.PHONY: gohangout all clean check test docker

gohangout:
	CGO_ENABLED=0 go build -ldflags "-X main.version=$(hash)" -o gohangout

docker:
	docker build -t gohangout .

all: check
	@echo $(hash)
	mkdir -p build/

	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.version=$(hash)" -o build/gohangout-windows-x64-$(tag).exe
	GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -ldflags "-X main.version=$(hash)" -o build/gohangout-windows-386-$(tag).exe
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.version=$(hash)" -o build/gohangout-linux-x64-$(tag)
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -ldflags "-X main.version=$(hash)" -o build/gohangout-linux-386-$(tag)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-X main.version=$(hash)" -o build/gohangout-linux-arm64-$(tag)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.version=$(hash)" -o build/gohangout-darwin-x64-$(tag)

clean:
	rm -rf gohangout

check:
	git diff-index --quiet HEAD --

test:
	go test -v -count=1 -gcflags="all=-N -l" ./... 
