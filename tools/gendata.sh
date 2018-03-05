#!/bin/bash

set -x

RELATIVE_TEMPLATES_DIR=pkg/http/templates

ALMANAC_ROOT=$GOPATH/src/github.com/dinowernli/almanac
TEMPLATES_DIR=$ALMANAC_ROOT/$RELATIVE_TEMPLATES_DIR

TMPDIR=`mktemp -d`
BACKUP_ROOT=$TMPDIR/almanac
cp -R $ALMANAC_ROOT $BACKUP_ROOT
BACKUP_TEMPLATES_DIR=$BACKUP_ROOT/$RELATIVE_TEMPLATES_DIR

# Do the actual generation.
FILENAME=bindata.go
pushd $TEMPLATES_DIR
rm $FILENAME
go-bindata -nometadata -o=$FILENAME -pkg=templates *.tmpl
BINDATA_OUT=$?
gofmt -s -w $FILENAME
popd

DIFF=`diff -rq $BACKUP_TEMPLATES_DIR $TEMPLATES_DIR`
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
