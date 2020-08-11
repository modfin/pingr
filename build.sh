#!/bin/bash

rm -rf /go/src/pingr/ui && mkdir -p /go/src/pingr/ui/build
go get -u github.com/go-bindata/go-bindata/...
cd /go/src/pingr/ui && go-bindata -o fs.go -prefix "build/" -pkg ui build/

cd /go/src/pingr; go mod download

go build -o /pingrd ./cmd/pingrd/pingrd.go

./pingrd