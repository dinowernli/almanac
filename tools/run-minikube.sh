#!/bin/bash

set -e
set -x

# Start minikube and configure docker to talk to the minikube docker daemon.
minikube start --vm-driver=xhyve
eval $(minikube docker-env)

# Build the almanac container and start a deployment.
docker build -t almanac:v1 .
kubectl create -f kube/deployment.yaml

# Create a service which exposes the binary.
kubectl expose deployment almanac --type=NodePort
minikube service almanac --url