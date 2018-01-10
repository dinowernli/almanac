#!/bin/bash

set -e

ALMANAC_ROOT=$GOPATH/src/dinowernli.me/almanac
TEMPLATES_DIR=$ALMANAC_ROOT/pkg/http/templates

BEFORESUM=`find $TEMPLATES_DIR -type f -exec shasum {} \; | shasum`

FILENAME=bindata.go
pushd $TEMPLATES_DIR
rm $FILENAME
go-bindata -nometadata -o=$FILENAME -pkg=templates *.tmpl
popd

AFTERSUM=`find $TEMPLATES_DIR -type f -exec shasum {} \; | shasum`

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
