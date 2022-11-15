# Git Backup Operator

A [Kubernetes Operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) for scheduled backup of Git repositories.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Git Backup Operator](#git-backup-operator)
  - [Overview](#overview)
  - [Getting Started](#getting-started)
    - [Installation](#installation)
    - [Deploy a `Repository` resource](#deploy-a-repository-resource)
    - [Uninstallation](#uninstallation)
  - [Developing](#developing)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Overview

1. You create a `Repository` resource.
2. The controller creates a `CronJob` resource from it.
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

### Installation

1. Make sure you have [cert-manager](https://cert-manager.io/) installed, as it is used to generate webhook certificates.

```sh
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.10.0/cert-manager.yaml
```

2. Install the controller with the following command. It creates `gitbackup-system` namespace and deploys CRDs, controllers and other resources.

```sh
# TODO(user)
kubectl apply -f https://...
```

### Deploy a `Repository` resource

1. Create a `Secret` resource that contains `.git-credentials`.
	
```sh
kubectl create secret generic repo1-secret --from-file=$HOME/.git-credentials
```

2. Create a `Repository` resource.

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

3. Confirm that resources has been created.

```
$ kubectl get repos
NAME    AGE
repo1   5s

$ kubectl get cronjobs
NAME              SCHEDULE    SUSPEND   ACTIVE   LAST SCHEDULE   AGE
gitbackup-repo1   0 6 * * *   False     0        <none>          5s
```

Note: You can test the `CronJob` by manually triggering it.

```sh
kubectl create job --from=cronjob/<name> <job-name>
```

### Uninstallation

1. Delete all `Repository` resources.

```sh
kubectl delete --all repos -A
```

2. Delete the Operator.

```sh
# TODO(user)
kubectl delete -f https://...
```

## Developing

This Operator uses [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder), so we basically follow the Kubebuilder way. See the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html) for details.


Note: You can run it with [KIND](https://sigs.k8s.io/kind) with the following command.

```sh
./hack/dev-kind-reset-cluster.sh
./hack/dev-kind-deploy.sh
```
