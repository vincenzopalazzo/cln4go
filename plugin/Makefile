CC=go
FMT=gofmt
NAME=test
BASE_DIR=/script
OS=linux
ARCH=386
ARM=

default: fmt

fmt:
	$(CC) fmt ./...

lint:
	golangci-lint run

check:
	$(CC) test -v ./...

check_fmt:
	gofmt -e ../.

build:
	$(CC) build -o example/simple_plugin example/simple_plugin.go

dep:
	$(CC) mod vendor
