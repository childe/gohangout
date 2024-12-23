hash:=$(shell git describe --tags --always)
buildTime:=$(shell git log -1 --format="%cI")
tag:=$(shell git rev-parse --short HEAD)

.PHONY: gohangout all clean check test docker

gohangout:
	CGO_ENABLED=0 go build -ldflags "-X main.version=$(hash) -X main.buildTime=$(buildTime)" -o gohangout

docker:
	docker build -t gohangout .

all:
	@echo $(hash)
	mkdir -p build/

	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.version=$(hash) -X main.buildTime=$(buildTime)" -o build/gohangout-windows-x64-$(tag).exe
	GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -ldflags "-X main.version=$(hash) -X main.buildTime=$(buildTime)" -o build/gohangout-windows-386-$(tag).exe
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.version=$(hash) -X main.buildTime=$(buildTime)" -o build/gohangout-linux-x64-$(tag)
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -ldflags "-X main.version=$(hash) -X main.buildTime=$(buildTime)" -o build/gohangout-linux-386-$(tag)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-X main.version=$(hash) -X main.buildTime=$(buildTime)" -o build/gohangout-linux-arm64-$(tag)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.version=$(hash) -X main.buildTime=$(buildTime)" -o build/gohangout-darwin-x64-$(tag)

clean:
	rm -rf gohangout

test:
	for _ in {1..5} ; do go test -v -count=1 -gcflags="all=-N -l" ./... && break ; done
