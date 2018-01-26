#!/bin/bash

set -e
set -x

DOCKERTAG=almanac-dev
docker build -t $DOCKERTAG .
docker run -p 12345:12345 --tmpfs /tmp $DOCKERTAG
