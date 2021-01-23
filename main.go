/*
MIT License

Copyright (c) 2021 Daniel Lohse

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
	"os"

	"github.com/annismckenzie/k3os-config-operator/config"
	"github.com/annismckenzie/k3os-config-operator/pkg/nodes"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	configv1alpha1 "github.com/annismckenzie/k3os-config-operator/apis/config/v1alpha1"
	configcontroller "github.com/annismckenzie/k3os-config-operator/controllers/config"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(configv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	configuration, err := config.InitializeConfiguration()
	if err != nil {
		zap.New().WithName("initializeConfig").Error(err, "unable to initialize configuration")
		os.Exit(1)
	}

	ctrl.SetLogger(zap.New(zap.UseDevMode(configuration.EnableDevMode())))
	ctx := ctrl.SetupSignalHandler()

	if nodeName := configuration.NodeName; nodeName == "" {
		setupLog.Info("unable to determine node name (is the NODE_NAME environment variable set?)")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                        scheme,
		Namespace:                     configuration.Namespace,
		MetricsBindAddress:            configuration.MetricsAddr,
		Port:                          configuration.BindPort,
		LeaderElection:                true, // this operator does not work without leader election
		LeaderElectionID:              "8a68cfa7.operators.annismckenzie.github.com",
		LeaderElectionNamespace:       configuration.Namespace,
		LeaderElectionResourceLock:    resourcelock.LeasesResourceLock,
		LeaderElectionReleaseOnCancel: true, // make the leader step down voluntarily when the manager ends
	})
	if err != nil {
		setupLog.Error(err, "unable to create manager")
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to build clientset")
		os.Exit(1)
	}
	if err = nodes.NewNodeInformer(ctx, clientset); err != nil {
		setupLog.Error(err, "unable to start node informer")
		os.Exit(1)
	}

	opts := []configcontroller.Option{
		configcontroller.WithNodeLister(nodes.NewNodeLister()),
		configcontroller.WithConfiguration(configuration),
	}
	if err = (&configcontroller.K3OSConfigReconciler{}).SetupWithManager(ctx, mgr, append(opts, configcontroller.RequireLeaderElection())...); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", configv1alpha1.K3OSConfigKind, "leader", true)
		os.Exit(1)
	}
	if err = (&configcontroller.K3OSConfigReconciler{}).SetupWithManager(ctx, mgr, opts...); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", configv1alpha1.K3OSConfigKind, "leader", false)
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
