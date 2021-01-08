# k3OS Config Operator

This operator will keep all fields of k3OS's `config.yaml` files in sync. Specifically, it's written to:
- sync node labels
- sync node taints

It runs as a DaemonSet on each node in the cluster.

It was written to address [this part of the k3OS README](https://github.com/rancher/k3os#configuration):

> The /var/lib/rancher/k3os/config.yaml or config.d/* files are intended to be used at runtime. These files can be manipulated manually, through scripting, or _managed with the Kubernetes operator_.

That Kubernetes operator doesn't exist. At least, it didn't until now. 🤠


## Prerequisites

1. A k3OS cluster that's running [nominally](https://joshdance.medium.com/what-does-nominal-mean-when-spacex-mission-control-says-it-39c2d249da27#:~:text=performing%20or%20achieved%20within%20expected,within%20expected%20and%20acceptable%20limits.).
2. A local clone of https://github.com/sgielen/picl-k3os-image-generator.
3. The `config` directory with YAML files as describes in https://github.com/sgielen/picl-k3os-image-generator#getting-started:
```
├── config
│  └── dc:a6:32:xx:xx:xx.yaml
│  └── dc:a6:32:xx:xx:xx.yaml
│  └── dc:a6:32:xx:xx:xx.yaml
```
3. Your local `kubectl` configured to push YAMLs to your k3OS cluster.
4. Execute `make deploy-k3os-config` in your local checkout of `picl-k3os-image-generator`. This will generate the configuration and push it into the cluster.
5. Continue on with the installation steps outlined below.


## Installation

```sh
  kubectl apply -f https://raw.githubusercontent.com/annismckenzie/k3os-config-operator/v0.1.0/deploy/operator.yaml
```
