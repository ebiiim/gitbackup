#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT=$(realpath "$0")
PROJECT_ROOT=$(dirname "$(dirname "$SCRIPT")")
PROJECT_NAME=$(basename "$PROJECT_ROOT")

KIND_CLUSTER_NAME=$PROJECT_NAME
KIND_IMAGE="kindest/node:v1.25.3@sha256:f52781bc0d7a19fb6c405c2af83abfeb311f130707a0e219175677e366cc45d1"
CERT_MANAGER_YAML="https://github.com/cert-manager/cert-manager/releases/download/v1.10.0/cert-manager.yaml"

cd "$PROJECT_ROOT"

kind create cluster --name "$KIND_CLUSTER_NAME" --image="$KIND_IMAGE" --config=- <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
EOF

kubectl apply -f "$CERT_MANAGER_YAML"
