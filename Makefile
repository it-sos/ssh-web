.PHONY: build run clean test

build:
	go build -o ssh-web ./cmd/server

run: build
	./ssh-web

clean:
	rm -f ssh-web config.yaml

test:
	go test ./... -v
