# k3OS Config Operator

This operator will keep all fields of k3OS's `config.yaml` files in sync. Specifically, it's written to:
- sync node labels
- sync node taints

It runs as a DaemonSet on each node in the cluster.

## Prerequisites

1. k3OS cluster that's running nominally
2. A (local) directory of YAML files as describes in https://github.com/sgielen/picl-k3os-image-generator#getting-started:
```
├── config
│  └── dc:a6:32:xx:xx:xx.yaml
│  └── dc:a6:32:xx:xx:xx.yaml
│  └── dc:a6:32:xx:xx:xx.yaml
```
3. your local `kubectl` configured to push YAMLs to your k3OS master

## Installation
TODO
