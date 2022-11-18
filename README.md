# Git Backup Operator

![GitHub](https://img.shields.io/github/license/ebiiim/gitbackup)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/ebiiim/gitbackup)](https://github.com/ebiiim/gitbackup/releases/latest)
[![CI](https://github.com/ebiiim/gitbackup/actions/workflows/ci.yaml/badge.svg)](https://github.com/ebiiim/gitbackup/actions/workflows/ci.yaml)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/ebiiim/gitbackup)
[![Go Report Card](https://goreportcard.com/badge/github.com/ebiiim/gitbackup)](https://goreportcard.com/report/github.com/ebiiim/gitbackup)

A [Kubernetes Operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) for scheduled backup of Git repositories.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Overview](#overview)
- [Getting Started](#getting-started)
  - [Installation](#installation)
  - [Deploy a `Repository` resource](#deploy-a-repository-resource)
  - [Uninstallation](#uninstallation)
- [Developing](#developing)

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

Make sure you have [cert-manager](https://cert-manager.io/) installed, as it is used to generate webhook certificates.

```sh
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.10.0/cert-manager.yaml
```

Install the Operator with the following command. It creates `gitbackup-system` namespace and deploys CRDs, controllers and other resources.

```sh
kubectl apply -f https://github.com/ebiiim/gitbackup/releases/download/v0.1.0/gitbackup.yaml
```

### Deploy a `Repository` resource

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

### Uninstallation

Delete the Operator and resources with the following command.

```sh
kubectl delete -f https://github.com/ebiiim/gitbackup/releases/download/v0.1.0/gitbackup.yaml
```

## Developing

This Operator uses [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder), so we basically follow the Kubebuilder way. See the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html) for details.


NOTE: You can run it with [kind](https://kind.sigs.k8s.io/) with the following command.

```sh
./hack/dev-kind-reset-cluster.sh
./hack/dev-kind-deploy.sh
```
