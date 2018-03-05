#!/bin/bash

set -x

ALMANAC_ROOT=$GOPATH/src/github.com/dinowernli/almanac
PROTO_DIR=$ALMANAC_ROOT/proto

TMPDIR=`mktemp -d`
BACKUP_ROOT=$TMPDIR/almanac
cp -R $ALMANAC_ROOT $BACKUP_ROOT
BACKUP_PROTO_DIR=$BACKUP_ROOT/proto

# Do the actual generation.
protoc \
  --go_out=$ALMANAC_ROOT \
  --go_out=plugins=grpc:. \
  --proto_path=$ALMANAC_ROOT \
  $PROTO_DIR/*.proto
PROTOC_OUT=$?
find $PROTO_DIR -name '*.go' | xargs gofmt -s -w

DIFF=`diff -ry $BACKUP_PROTO_DIR $PROTO_DIR`
rm -rf $TMPDIR

# Definitely error out if protoc failed.
if [[ $PROTOC_OUT -ne 0 ]]; then
  echo 'protoc failed'
  exit 1
fi

# Error out in check mode if the diff didn't match.
if [[ "$1" == "--check" ]]; then
  if [[ -z $DIFF ]]; then
    echo "check succeeded, directories were the same"
    exit 0
  else
    echo "check failed, directories were different"
    echo "diff: $DIFF"
    exit 1
  fi
fi
