# almanac

[![Build Status](https://travis-ci.org/dinowernli/almanac.svg?branch=master)](https://travis-ci.org/dinowernli/almanac)
[![Go Report Card](https://goreportcard.com/badge/github.com/dinowernli/almanac)](https://goreportcard.com/report/github.com/dinowernli/almanac)

A distributed log storage and serving system.

Design goals:
* Easy to deploy on cloud infrastructure such as GCP or AWS.
* Minimal operational burden, i.e., deployments should be easy to upgrade, restart, modify, etc.
* System cost scales with usage rather than uptime, making the system viable for small and large deployments.
* Does not require operating a resilient and fault-tolerant storage system.

## Design

The design doc for the system can be found [here](https://docs.google.com/document/d/1yVTRtSZQ2ulSV9CGwExn2l2E2kJqyssCMB7ZM7FNhnc/edit). As parts of the design go from being under discussion to being more consolidated, the design will gradually move into markdown in this repo.

## Building and running

If you have a working go environment, you will need to run the following as one-time setup:

* `./tools/fetch-deps.sh`
* `dep ensure`

### Running the demo

Run the demo binary by executing:

`go run ./cmd/almanac/almanac.go`

This will start a single-process cluster and will print the locations of a few relevant web pages which can be used to play around manually. By default, the demo runs against an in-memory storage implementation. In order to use an actual GCS bucket, execute:

`GOOGLE_APPLICATION_CREDENTIALS=<path> go run ./cmd/almanac/almanac.go --storage=gcs --gcs.bucket=<bucket>`

### Running tests

To run all the tests, execute:

`go test ./...`
