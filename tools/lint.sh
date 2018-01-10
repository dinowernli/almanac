#!/bin/bash

set -e

if [[ "$1" == "--fix" ]]; then
  find . -name '*.go' | xargs gofmt -s -w
  exit 0
else
  BAD=`find . -name '*.go' | xargs gofmt -l`
  echo $BAD
  if [[ -z $BAD ]]; then
    echo 'lint successful'
    exit 0
  else
    echo 'lint failed, please run with --fix'
    exit 1
  fi
fi
