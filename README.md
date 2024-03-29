# k3OS Config Operator

[![Coverage Status](https://coveralls.io/repos/github/annismckenzie/k3os-config-operator/badge.svg?branch=main)](https://coveralls.io/github/annismckenzie/k3os-config-operator?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/annismckenzie/k3os-config-operator)](https://goreportcard.com/report/github.com/annismckenzie/k3os-config-operator)

This operator will keep all fields of k3OS's `config.yaml` files in sync. Specifically, it's written to:
- sync node labels
- sync node taints
- keep the `/var/lib/rancher/k3os/config.yaml` on each node in sync

It runs as a DaemonSet on each node in the cluster.

It was written to address [this part of the k3OS README](https://github.com/rancher/k3os#configuration):

> The /var/lib/rancher/k3os/config.yaml or config.d/* files are intended to be used at runtime. These files can be manipulated manually, through scripting, or _managed with the Kubernetes operator_.

That Kubernetes operator doesn't exist. At least, it didn't until now. 🤠


## Prerequisites

1. A k3OS cluster that's running [nominally](https://joshdance.medium.com/what-does-nominal-mean-when-spacex-mission-control-says-it-39c2d249da27#:~:text=performing%20or%20achieved%20within%20expected,within%20expected%20and%20acceptable%20limits.).
2. A local clone of https://github.com/annismckenzie/picl-k3os-image-generator (my fork).
3. The `config` directory with YAML files as describes in https://github.com/sgielen/picl-k3os-image-generator#getting-started:
```
├── config
│  └── dc:a6:32:xx:xx:xx.yaml
│  └── dc:a6:32:xx:xx:xx.yaml
│  └── dc:a6:32:xx:xx:xx.yaml
```
3. Your local `kubectl` configured to push YAMLs to your k3OS cluster.
4. Execute `make update-k3osconfig-secret` in your local checkout of `picl-k3os-image-generator` to validate the YAML files in the `config` directory and to generate the secret for the operator.
5. Execute `make deploy-k3os-config` in your local checkout of `picl-k3os-image-generator`. This will generate the configuration and push it into the cluster.
5. Continue on with the installation steps outlined below.


## Installation

```sh
  kubectl apply -f https://raw.githubusercontent.com/annismckenzie/k3os-config-operator/v0.3.2/deploy/operator.yaml
```


## Releasing

1. Create `release-vx.y.z` branch. Update `config/release/kustomization.yaml` with the new version, commit.
2. Tag locally with `vx.y.z`. Make sure the working directory is clean.
3. Run `make render-static-manifests`. This will build and push the new multi-arch image.
4. Update the installation path above with the new tag.
5. Commit the updated `deploy/operator.yaml` and the update to the README.
6. Recreate the local tag, push the tag, then update the changelog with `make update-changelog`. Commit and push.
7. Merge the release PR.
