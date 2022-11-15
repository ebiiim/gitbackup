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

# KIND v0.17.0 images https://github.com/kubernetes-sigs/kind/releases/tag/v0.17.0
KIND_IMAGE_125="kindest/node:v1.25.3@sha256:f52781bc0d7a19fb6c405c2af83abfeb311f130707a0e219175677e366cc45d1"
KIND_IMAGE_124="kindest/node:v1.24.7@sha256:577c630ce8e509131eab1aea12c022190978dd2f745aac5eb1fe65c0807eb315"
KIND_IMAGE_123="kindest/node:v1.23.13@sha256:ef453bb7c79f0e3caba88d2067d4196f427794086a7d0df8df4f019d5e336b61"
KIND_IMAGE_122="kindest/node:v1.22.15@sha256:7d9708c4b0873f0fe2e171e2b1b7f45ae89482617778c1c875f1053d4cef2e41"
KIND_IMAGE_121="kindest/node:v1.21.14@sha256:9d9eb5fb26b4fbc0c6d95fa8c790414f9750dd583f5d7cee45d92e8c26670aa1"
KIND_IMAGE_120="kindest/node:v1.20.15@sha256:a32bf55309294120616886b5338f95dd98a2f7231519c7dedcec32ba29699394"
KIND_IMAGE_119="kindest/node:v1.19.16@sha256:476cb3269232888437b61deca013832fee41f9f074f9bed79f57e4280f7c48b7"

CERT_MANAGER_YAML="https://github.com/cert-manager/cert-manager/releases/download/v1.10.0/cert-manager.yaml"

function run {
    _kind_image=$1
    _cert_manager_yaml=$2

    KIND_IMAGE=$_kind_image CERT_MANAGER_YAML=$_cert_manager_yaml ./hack/dev-kind-reset-cluster.sh
    sleep 50

    kind load docker-image "$IMG" -n "$KIND_CLUSTER_NAME"
    make deploy IMG="$IMG"
    sleep 30

    ./hack/dev-kind-samples.sh
    sleep 5

    if test "$(kubectl get cronjob | wc -l)" -eq 0; then
        kubectl get pod -A
        kubectl get repo
        kubectl get cronjob
        kubectl get configmap
        kubectl get secret
        exit 1;
    else
        echo "OK"
    fi
}

cd "$PROJECT_ROOT"

make docker-build IMG="$IMG"

run $KIND_IMAGE_125 $CERT_MANAGER_YAML
run $KIND_IMAGE_124 $CERT_MANAGER_YAML
run $KIND_IMAGE_123 $CERT_MANAGER_YAML
run $KIND_IMAGE_122 $CERT_MANAGER_YAML
run $KIND_IMAGE_121 $CERT_MANAGER_YAML
# run $KIND_IMAGE_120 $CERT_MANAGER_YAML # no CronJob.batch v1 support
# run $KIND_IMAGE_119 $CERT_MANAGER_YAML # no CronJob.batch v1 support
