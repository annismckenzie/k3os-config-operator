module github.com/annismckenzie/k3os-config-operator

go 1.16

require (
	github.com/go-logr/logr v0.4.0
	github.com/jessevdk/go-flags v1.5.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.14.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	sigs.k8s.io/cluster-api v0.4.0 // indirect; uses master, added to use testing helpers
	sigs.k8s.io/cluster-api/test v0.4.0
	sigs.k8s.io/controller-runtime v0.9.5
)

replace sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v0.4.0

// replace sigs.k8s.io/controller-runtime => ../../../sigs.k8s.io/controller-runtime
