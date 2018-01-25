#!/bin/bash

set -e
set -x

BINNAME=almanac-linux-static
DOCKERTAG=almanac-dev

CGO_ENABLED=0 GOOS=linux go build -o $BINNAME -a -ldflags '-extldflags "-static"' cmd/almanac/almanac.go
docker build -t $DOCKERTAG .
docker run -p 12345:12345 --tmpfs /tmp almanac-dev