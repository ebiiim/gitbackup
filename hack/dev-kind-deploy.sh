#!/usr/bin/env bash

# scripts must be run from project root
. hack/1-bin.sh || exit 1

# consts

VERSION=$(git describe --tags --match "v*")
IMG=$PROJECT_NAME-controller:$VERSION

KIND_CLUSTER_NAME=$PROJECT_NAME
CLUSTER_NAME=kind-$KIND_CLUSTER_NAME

# main

"$KUBECTL" config use-context "$CLUSTER_NAME"

make
make manifests

make docker-build IMG="$IMG"
"$KIND" load docker-image "$IMG" -n "$KIND_CLUSTER_NAME"

make undeploy || true
make deploy IMG="$IMG"
