#!/bin/bash

set -e

ALMANAC_ROOT=$GOPATH/src/dinowernli.me/almanac
PROTO_DIR=$ALMANAC_ROOT/proto

protoc \
  --go_out=$ALMANAC_ROOT \
  --go_out=plugins=grpc:. \
  --proto_path=$ALMANAC_ROOT \
  $PROTO_DIR/*.proto
