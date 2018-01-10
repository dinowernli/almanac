# almanac

[![Build Status](https://travis-ci.org/dinowernli/almanac.svg?branch=master)](https://travis-ci.org/dinowernli/almanac)

A distributed log storage and serving system.

Design goals:
* Easy to run using cloud providers for the machines (e.g., GCE) and storage (e.g., GCS).
* Does not require looking after a resilient and fault-tolerant storage system.
* Simple cluster management because the system uses disk and memory solely for caching.

## Design

TODO(dino) link to doc.

## Building and running

### Running the demo

In order to run a demo, execute:

`bazel run main`

This will start a single-process cluster and will print the locations of a few relevant web pages which can be used to play around manually.

By default, the demo runs against an in-memory storage implementation. In order to use an actual GCS bucket, execute:

`GOOGLE_APPLICATION_CREDENTIALS=<path> bazel run main -- --storage=gcs --gcs.bucket=<bucket>`

### Running tests

In order to run the tests used for CI, execute:

`bazel test ...`

