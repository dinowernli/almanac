# almanac

[![Build Status](https://travis-ci.org/dinowernli/almanac.svg?branch=master)](https://travis-ci.org/dinowernli/almanac)
[![Go Report Card](https://goreportcard.com/badge/github.com/dinowernli/almanac)](https://goreportcard.com/report/github.com/dinowernli/almanac)

A distributed log storage and serving system.

Design goals:
* Easy to run using cloud providers for the machines (e.g., GCE) and storage (e.g., GCS).
* Does not require looking after a resilient and fault-tolerant storage system.
* Simple cluster management because the system uses disk and memory solely for caching.

## Design

TODO(dino) link to doc.

## Building and running

### Repo setup

If you have a working go environment, you will need to run the following as one-time setup:

* `./tools/fetch-deps.sh`
* `dep ensure`

### Running the demo

First, build the binary by executing:

`(cd main && go build)`

The binary can then be run using:

`./main/main`

This will start a single-process cluster and will print the locations of a few relevant web pages which can be used to play around manually. By default, the demo runs against an in-memory storage implementation. In order to use an actual GCS bucket, execute:

`GOOGLE_APPLICATION_CREDENTIALS=<path> ./main/main --storage=gcs --gcs.bucket=<bucket>`

### Running tests

To run all the tests, execute:

`go test ./...`

