module github.com/annismckenzie/k3os-config-operator

go 1.15

require (
	github.com/go-logr/logr v0.4.0
	github.com/jessevdk/go-flags v1.4.0
	github.com/onsi/ginkgo v1.15.0
	github.com/onsi/gomega v1.10.5
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.3
	sigs.k8s.io/cluster-api v0.3.11-0.20210115191551-61dc332270dc // uses master, added to use testing helpers
	sigs.k8s.io/controller-runtime v0.8.2
)

// replace sigs.k8s.io/controller-runtime => ../../../sigs.k8s.io/controller-runtime
