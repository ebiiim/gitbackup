#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT=$(realpath "$0")
PROJECT_ROOT=$(dirname "$(dirname "$SCRIPT")")
PROJECT_NAME=$(basename "$PROJECT_ROOT")

VERSION=$(git describe --tags --match "v*")
IMG=$PROJECT_NAME-controller:$VERSION
REGISTRY=${REGISTRY:-""}
if [ -n "$REGISTRY" ]; then
    IMG=$REGISTRY/$PROJECT_NAME-controller:$VERSION
fi
DIST_FILE=gitbackup.yaml

cd "$PROJECT_ROOT"

make kustomize

cd config/manager && "$PROJECT_ROOT/bin/kustomize" edit set image controller="$IMG"
cd "$PROJECT_ROOT" && "$PROJECT_ROOT/bin/kustomize" build config/default > $DIST_FILE
