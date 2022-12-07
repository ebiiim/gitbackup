#!/usr/bin/env bash

# scripts must be run from project root
. hack/0-env.sh || exit 1

# consts

VERSION=$(git describe --tags --match "v*")
IMG=$PROJECT_NAME-controller:$VERSION
REGISTRY=${REGISTRY:-""}
if [ -n "$REGISTRY" ]; then
    IMG=$REGISTRY/$PROJECT_NAME-controller:$VERSION
fi
DIST_FILE=gitbackup.yaml

make kustomize

cd config/manager && "$PROJECT_ROOT/bin/kustomize" edit set image controller="$IMG"
cd "$PROJECT_ROOT" && "$PROJECT_ROOT/bin/kustomize" build config/default > $DIST_FILE
