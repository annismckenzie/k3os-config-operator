package nodes

import (
	"errors"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	corev1informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

var nodesFactory corev1informer.NodeInformer

// NewNodeInformer starts a new node informer.
func NewNodeInformer(clientset *kubernetes.Clientset, stopCh <-chan struct{}) error {
	factory := informers.NewSharedInformerFactory(clientset, 0)
	nodesFactory = factory.Core().V1().Nodes()
	nodeInformer := nodesFactory.Informer()
	go func() {
		defer runtime.HandleCrash()
		nodeInformer.Run(stopCh)
		<-stopCh
	}()
	if !cache.WaitForCacheSync(stopCh, nodeInformer.HasSynced) {
		return errors.New("timed out waiting for caches to sync")
	}
	return nil
}

// NewNodeLister returns a new node lister.
func NewNodeLister() listersv1.NodeLister {
	return nodesFactory.Lister()
}
