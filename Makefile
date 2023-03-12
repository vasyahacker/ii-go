.PHONY: all clean build install

all: build

build-all: build

build:
	go build -trimpath -o ii-tool   ./cmd/ii-tool
	go build -trimpath -o ii-node   ./cmd/ii-node
	go build -trimpath -o ii-gemini ./cmd/ii-gemini

install-all: install

install:
	go install ./cmd/ii-tool
	go install ./cmd/ii-node
	go install ./cmd/ii-gemini

clean:
	rm -f ii-node ii-tool ii-gemini
