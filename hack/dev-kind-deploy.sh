#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT=$(realpath "$0")
PROJECT_ROOT=$(dirname "$(dirname "$SCRIPT")")
PROJECT_NAME=$(basename "$PROJECT_ROOT")

KIND_CLUSTER_NAME=$PROJECT_NAME
VERSION=$(git describe --tags --match "v*")
IMG=$PROJECT_NAME-controller:$VERSION

cd "$PROJECT_ROOT"

make undeploy || true
make uninstall || true

IMG=$IMG make docker-build
kind load docker-image "$IMG" -n "$KIND_CLUSTER_NAME"

IMG=$IMG make install
IMG=$IMG make deploy