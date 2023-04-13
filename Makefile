CC=go
FMT=gofmt
NAME=test

default: fmt lint

fmt:
	cd comm; $(CC) fmt ./...
	cd client; $(CC) fmt ./...
	cd plugin; $(CC) fmt ./...
	cd bench; $(CC) fmt ./...

lint:
	cd comm; golangci-lint run
	cd client; golangci-lint run
	cd plugin; golangci-lint run
	cd bench; golangci-lint run

check:
	cd comm; $(CC) test -v ./...
	cd client; $(CC) test -v ./...
	cd plugin; $(CC) test -v ./...

check_fmt:
	gofmt -e ../.

build:
	cd plugin; make build

dep:
	cd bench; $(CC) get -u all

bench_check:
	cd bench; $(CC) run main.go
