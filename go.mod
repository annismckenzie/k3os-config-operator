module github.com/annismckenzie/k3os-config-operator

go 1.16

require (
	github.com/containerd/containerd v1.5.7 // indirect
	github.com/go-logr/logr v0.4.0
	github.com/jessevdk/go-flags v1.5.0
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.17.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.22.4
	k8s.io/apimachinery v0.22.4
	k8s.io/client-go v0.22.3
	sigs.k8s.io/cluster-api v0.4.4 // indirect; uses master, added to use testing helpers
	sigs.k8s.io/cluster-api/test v0.4.4
	sigs.k8s.io/controller-runtime v0.10.3
)

replace sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v0.4.4

// replace sigs.k8s.io/controller-runtime => ../../../sigs.k8s.io/controller-runtime
