#!/bin/bash

set -e
set -x

FINDCMD='find . -name "*.go" -and -not -path "*vendor*" -and -not -name "bindata.go" -and -not -path "*.pb.go"'

if [[ "$1" == "--fix" ]]; then
  eval $FINDCMD | xargs gofmt -s -w
  exit 0
else
  BAD=`eval $FINDCMD | xargs gofmt -l`
  echo $BAD
  if [[ -z $BAD ]]; then
    echo 'lint successful'
    exit 0
  else
    echo 'lint failed, please run with --fix'
    exit 1
  fi
fi
