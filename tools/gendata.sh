#!/bin/bash

set -e

ALMANAC_ROOT=$GOPATH/src/dinowernli.me/almanac
TEMPLATES_DIR=$ALMANAC_ROOT/pkg/http/templates

pushd $TEMPLATES_DIR
go-bindata -pkg=templates *.tmpl
popd

