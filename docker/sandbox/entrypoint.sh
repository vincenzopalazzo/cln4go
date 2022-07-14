#!/bin/bash
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
go version
./run-bitcoin.sh
cd code || exit 1
# FIXME: in the future we will build the plugin here and they need to be compiled
# before run it
make
cd .. || exit 1
./run-clightning.sh
cd code || exit 1
ls -la
CLN_UNIX=/workdir/lightning_dir_one/regtest/lightning-rpc make check