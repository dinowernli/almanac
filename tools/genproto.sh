#!/bin/bash

set -e

ALMANAC_ROOT=$GOPATH/src/dinowernli.me/almanac
PROTO_DIR=$ALMANAC_ROOT/proto

BEFORESUM=`find -s $PROTO_DIR -type f -exec shasum {} \; | shasum`

rm $PROTO_DIR/*.pb.go
protoc \
  --go_out=$ALMANAC_ROOT \
  --go_out=plugins=grpc:. \
  --proto_path=$ALMANAC_ROOT \
  $PROTO_DIR/*.proto

AFTERSUM=`find -s $PROTO_DIR -type f -exec shasum {} \; | shasum`

if [[ "$1" == "--check" ]]; then
  if [ "$BEFORESUM" == "$AFTERSUM" ]; then
    echo 'check succeeded, directories were the same'
    exit 0
  else
    echo 'check failed, directories were different'
    echo 'please rerun with --check' 
    exit 1
  fi
fi
