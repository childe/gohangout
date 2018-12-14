hash:=$(shell git rev-parse --short HEAD)

.PHONY: gohangout test ctrip all clean

gohangout:
	go build -o build/gohangout

test:
	mkdir -p build/
	GOOS=linux GOARCH=amd64 go build -o build/gohangout-linux-x64-$(hash)

ctrip:
	mkdir -p build/
	git checkout ctrip
	GOOS=linux GOARCH=amd64 go build -o build/gohangout-linux-x64-$(hash)

all:
	@echo $(hash)
	mkdir -p build/

	GOOS=windows GOARCH=amd64 go build -o build/gohangout-windows-x64-$(hash).exe
	GOOS=windows GOARCH=386 go build -o build/gohangout-windows-386-$(hash).exe
	GOOS=linux GOARCH=amd64 go build -o build/gohangout-linux-x64-$(hash)
	GOOS=linux GOARCH=386 go build -o build/gohangout-linux-386-$(hash)
	GOOS=darwin GOARCH=amd64 go build -o build/gohangout-darwin-x64-$(hash)

clean:
	rm -rf build/*

