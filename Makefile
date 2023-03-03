.PHONY: all clean build

all: build

build-all: build

build:
	go build -trimpath -o ii-tool ./ii-tool
	go build -trimpath -o ii-node ./ii-node
	go build -trimpath -o ii-gemini ./ii-gemini

install-all: install

install:
	go install ./ii-tool
	go install ./ii-node
	go install ./ii-gemini

clean:
	cd ii-node && go clean
	cd ii-tool && go clean
	cd ii-gemini && go clean
