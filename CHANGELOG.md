# Changelog

## What comes next?

- Upgrade `kubebuilder`.
  - NOTE: We use `kube-rbac-proxy:v0.13.0` so [`gcr.io` retirement](https://github.com/kubernetes-sigs/kubebuilder/discussions/3907) affects us. The workaround is to use other image registry.

## 0.2.1 - 2023-01-05

### Changed

- Finished Jobs (regardless of completeness) will be deleted after 100 hours. Since this is a backup task, basically it should be fine as long as the latest run was successful.

## 0.2.0 - 2022-12-10

### Added

- Collection CRD.

### Changed

- Repository default ConfigMap name `gitbackup-gitconfig-{repoName}` -> `gitbackup-repository-{repoName}-gitconfig`

## 0.1.1 - 2022-12-07

### Fixed

- Minor improvements.

## 0.1.0 - 2022-11-15

### Added

- Repository CRD.
