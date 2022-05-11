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

build:
	@echo "nothing yet"

update_utils:
	$(CC) mod vendor
