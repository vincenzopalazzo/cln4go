#!/bin/bash
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
go version
./run-bitcoin.sh
cd code || exit 1
make build
cd .. || exit 1
./run-clightning.sh
cd code || exit 1
ls -la
CLN_UNIX_SOCKET=/workdir/lightning_dir_one/regtest/lightning-rpc make check