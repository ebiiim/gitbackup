# Git Backup Operator

[![GitHub](https://img.shields.io/github/license/ebiiim/gitbackup)](https://github.com/ebiiim/gitbackup/blob/main/LICENSE)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/ebiiim/gitbackup)](https://github.com/ebiiim/gitbackup/releases/latest)
[![CI](https://github.com/ebiiim/gitbackup/actions/workflows/ci.yaml/badge.svg)](https://github.com/ebiiim/gitbackup/actions/workflows/ci.yaml)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/ebiiim/gitbackup)
[![Go Report Card](https://goreportcard.com/badge/github.com/ebiiim/gitbackup)](https://goreportcard.com/report/github.com/ebiiim/gitbackup)
[![codecov](https://codecov.io/gh/ebiiim/gitbackup/branch/main/graph/badge.svg)](https://codecov.io/gh/ebiiim/gitbackup)

A [Kubernetes Operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) for scheduled backup of Git repositories.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Overview](#overview)
- [Getting Started](#getting-started)
  - [Installation](#installation)
  - [Backup a Git repository with a `Repository` resource](#backup-a-git-repository-with-a-repository-resource)
  - [Backup many Git repositories with a `Collection` resource](#backup-many-git-repositories-with-a-collection-resource)
  - [Uninstallation](#uninstallation)
- [Developing](#developing)
  - [Prerequisites](#prerequisites)
  - [Run development clusters with kind](#run-development-clusters-with-kind)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Overview

1. You create a `Repository` resource.
2. The Operator creates a `CronJob` resource from it.
3. The `CronJob` does the actual work.

```yaml
apiVersion: gitbackup.ebiiim.com/v1beta1
kind: Repository
metadata:
  name: repo1
spec:
  src: https://github.com/ebiiim/gitbackup
  dst: https://gitlab.com/ebiiim/gitbackup
  schedule: "0 6 * * *"
  gitCredentials:
    name: repo1-secret # specify a Secret resource in the same namespace
```

## Getting Started

Supported Kubernetes versions: __1.21 or higher__

### Installation

Make sure you have [cert-manager](https://cert-manager.io/) deployed, as it is used to generate webhook certificates.

```sh
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.10.0/cert-manager.yaml
```

> ⚠️ You may have to wait a second for cert-manager to be ready.

Deploy the Operator with the following command. It creates `gitbackup-system` namespace and deploys CRDs, controllers and other resources.

```sh
kubectl apply -f https://github.com/ebiiim/gitbackup/releases/download/v0.2.0/gitbackup.yaml
```

### Backup a Git repository with a `Repository` resource

First, create a `Secret` resource that contains `.git-credentials`.
	
```sh
kubectl create secret generic repo1-secret --from-file=$HOME/.git-credentials
```

Next, create a `Repository` resource.

```yaml
apiVersion: gitbackup.ebiiim.com/v1beta1
kind: Repository
metadata:
  name: repo1
spec:
  src: https://github.com/ebiiim/gitbackup
  dst: https://gitlab.com/ebiiim/gitbackup
  schedule: "0 6 * * *"
  gitCredentials:
    name: repo1-secret
```

Finally, confirm that resources has been created.

```
$ kubectl get repos
NAME    AGE
repo1   5s

$ kubectl get cronjobs
NAME              SCHEDULE    SUSPEND   ACTIVE   LAST SCHEDULE   AGE
gitbackup-repo1   0 6 * * *   False     0        <none>          5s
```

NOTE: You can test the `CronJob` by manually triggering it.

```sh
kubectl create job --from=cronjob/<name> <job-name>
```

### Backup many Git repositories with a `Collection` resource

First, create a `Secret` resource that contains `.git-credentials`.
	
```sh
kubectl create secret generic repo1-secret --from-file=$HOME/.git-credentials
```

Next, create a `Collection` resource.

```yaml
apiVersion: gitbackup.ebiiim.com/v1beta1
kind: Collection
metadata:
  name: coll1
spec:
  schedule: "0 6 * * *"
  gitCredentials:
    name: coll1-secret
  repos:
    - name: gitbackup
      src: https://github.com/ebiiim/gitbackup
      dst: https://gitlab.com/ebiiim/gitbackup
    - name: foo
      src: https://example.com/src/foo
      dst: https://example.com/dst/foo
    - name: bar
      src: https://example.com/src/bar
      dst: https://example.com/dst/bar
```

Finally, confirm that resources has been created.

```
$ kubectl get colls
NAME    AGE
coll1   5s

$ kubectl get repos
NAME                AGE
coll1-bar           5s
coll1-foo           5s
coll1-gitbackup     5s

$ kubectl get cronjobs
NAME                        SCHEDULE    SUSPEND   ACTIVE   LAST SCHEDULE   AGE
gitbackup-coll1-bar         2 6 * * *   False     0        <none>          5s
gitbackup-coll1-foo         1 6 * * *   False     0        <none>          5s
gitbackup-coll1-gitbackup   0 6 * * *   False     0        <none>          5s
```

> 💡 NOTE: Each job runs one minute apart.

### Uninstallation

Delete the Operator and resources with the following command.

```sh
kubectl delete -f https://github.com/ebiiim/gitbackup/releases/download/v0.2.0/gitbackup.yaml
```

## Developing

This Operator uses [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder), so we basically follow the Kubebuilder way. See the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html) for details.

### Prerequisites

Make sure you have the following tools installed:

- Git
- Make
- Go
- Docker

### Run development clusters with [kind](https://kind.sigs.k8s.io/)

```sh
./hack/dev-kind-reset-clusters.sh # create a K8s cluster `kind-gitbackup`
./hack/dev-kind-deploy.sh # build and deploy the Operator
```
