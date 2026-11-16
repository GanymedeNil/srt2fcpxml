.PHONY: build test web clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

build:
	go build -o build/srt2fcpxml ./cmd/srt2fcpxml
	go build -o build/fcpxml2srt ./cmd/fcpxml2srt

test:
	go test ./...

# The web version: a static site in web/dist.
web:
	mkdir -p web/dist
	GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o web/dist/srt2fcpxml.wasm ./cmd/wasm
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" web/dist/
	cp web/index.html web/style.css web/app.js web/dist/

clean:
	rm -rf build dist web/dist
