module github.com/annismckenzie/k3os-config-operator

go 1.15

require (
	github.com/go-logr/logr v0.3.0
	github.com/jessevdk/go-flags v1.4.0
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.4
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/cluster-api v0.3.11-0.20210115191551-61dc332270dc // uses master, added to use testing helpers
	sigs.k8s.io/controller-runtime v0.8.1
)

// replace sigs.k8s.io/controller-runtime => ../../../sigs.k8s.io/controller-runtime
