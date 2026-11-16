.PHONY: build test clean

build:
	go build -o build/srt2fcpxml ./cmd/srt2fcpxml

test:
	go test ./...

clean:
	rm -rf build dist
