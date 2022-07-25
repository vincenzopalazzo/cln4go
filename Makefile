CC=go
FMT=gofmt
NAME=test
BASE_DIR=/script
OS=linux
ARCH=386
ARM=

default: fmt

fmt:
	cd comm; $(CC) fmt ./...
	cd client; $(CC) fmt ./...
	cd plugin; $(CC) fmt ./...

lint:
	cd comm; golangci-lint run
	cd client; golangci-lint run
	cd plugin; golangci-lint run

check:
	cd comm; $(CC) test -v ./...
	cd client; $(CC) test -v ./...
	cd plugin; $(CC) test -v ./...

check_fmt:
	gofmt -e ../.

build:
	cd plugin; make build

dep:
	$(CC) mod vendor
