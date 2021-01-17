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

package config

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/annismckenzie/k3os-config-operator/config"
	"github.com/annismckenzie/k3os-config-operator/pkg/consts"
	"github.com/annismckenzie/k3os-config-operator/pkg/nodes"
	"github.com/annismckenzie/k3os-config-operator/pkg/util/taints"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/cluster-api/test/framework"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	configv1alpha1 "github.com/annismckenzie/k3os-config-operator/apis/config/v1alpha1"
	// +kubebuilder:scaffold:imports
)

// define utility constants for namespace name and testing timeouts/durations and intervals
const (
	k3OSConfigNamespaceName = "test-k3osconfig-namespace"
	dummyNodeName           = "dummy1"

	timeout  = time.Second * 10
	interval = time.Millisecond * 250
)

var cfg *rest.Config
var k8sClient client.Client
var k8sManager manager.Manager
var testEnv *envtest.Environment
var dummyNode1 = &corev1.Node{}

var (
	ctx    context.Context
	cancel context.CancelFunc
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"K3OSConfig controller suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))

	// set environment variables so that the configuration can be initialized successfully
	err := os.Setenv("NODE_NAME", dummyNodeName)
	Expect(err).ToNot(HaveOccurred())
	err = os.Setenv("NAMESPACE", k3OSConfigNamespaceName)
	Expect(err).ToNot(HaveOccurred())

	customAPIServerFlags := []string{
		"--enable-admission-plugins=PodSecurityPolicy",
	}
	apiServerFlags := append(envtest.DefaultKubeAPIServerFlags, customAPIServerFlags...)

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:  []string{filepath.Join("..", "..", "config", "crd", "bases")},
		KubeAPIServerFlags: apiServerFlags,
	}

	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	Expect(configv1alpha1.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	ctx, cancel = context.WithCancel(context.Background())

	k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	go func() {
		Expect(k8sManager.Start(ctx)).ToNot(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	// create the test namespace
	input := framework.CreateNamespaceInput{
		Creator: k8sClient,
		Name:    k3OSConfigNamespaceName,
	}
	_ = framework.CreateNamespace(ctx, input, timeout, interval)

	// create a dummy node
	dummyNode1 = createNode(&corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        dummyNodeName,
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
	})

	clientset, err := kubernetes.NewForConfig(k8sManager.GetConfig())
	Expect(err).ToNot(HaveOccurred())

	Expect(nodes.NewNodeInformer(ctx, clientset)).ToNot(HaveOccurred())

	close(done)
}, 60)

var _ = Describe("K3OSConfig controller", func() {
	const k3OSConfigName = "test-k3osconfig"
	var k3OSConfig *configv1alpha1.K3OSConfig
	var testConfiguration *config.Configuration
	var nodeConfigSecret *corev1.Secret
	var tempConfigFile *os.File
	var dummyNodeConfig *configv1alpha1.K3OSConfigFile

	BeforeEach(func() {
		var err error
		testConfiguration, err = config.InitializeConfiguration(flags.Default, flags.IgnoreUnknown)
		Expect(err).ToNot(HaveOccurred())
		testConfiguration.DevMode = true

		dummyNodeConfig = &configv1alpha1.K3OSConfigFile{
			Spec: configv1alpha1.K3OSConfigFileSpec{
				Hostname: dummyNodeName,
				K3OS: configv1alpha1.K3OSConfigFileSectionK3OS{
					Labels: map[string]string{
						"newLabel": "newLabelValue",
					},
					Taints: []string{"test:NoSchedule"},
				},
			},
		}
		dummyNodeConfig.Spec.Data, err = dummyNodeConfig.MarshalYAML()
		Expect(err).ToNot(HaveOccurred())

		nodeConfigSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testConfiguration.NodeConfigSecretName,
				Namespace: testConfiguration.Namespace,
				Labels: map[string]string{
					"app.kubernetes.io/managed-by": "k3os-config-operator", // FIXME: test this (operator shouldn't touch any other secret)
				},
			},
			Data: map[string][]byte{
				dummyNodeName: dummyNodeConfig.Spec.Data,
			},
		}

		k3OSConfig = &configv1alpha1.K3OSConfig{
			TypeMeta: metav1.TypeMeta{
				APIVersion: configv1alpha1.GroupVersion.String(),
				Kind:       configv1alpha1.K3OSConfigFileKind,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      k3OSConfigName,
				Namespace: k3OSConfigNamespaceName,
			},
			Spec: configv1alpha1.K3OSConfigSpec{
				SyncNodeLabels: true,
				SyncNodeTaints: true,
			},
		}

		tempConfigFile, err = ioutil.TempFile("", "k3osconfig.*.yaml")
		Expect(err).ToNot(HaveOccurred())
		err = tempConfigFile.Close()
		Expect(err).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		err := os.Remove(tempConfigFile.Name())
		Expect(err).ToNot(HaveOccurred())
	})

	JustBeforeEach(func() {
		createTestNodeConfigSecret(nodeConfigSecret)
		nodeListerOpt := WithNodeLister(nodes.NewNodeLister())
		configurationOpt := WithConfiguration(testConfiguration)
		err := (&K3OSConfigReconciler{}).SetupWithManager(ctx, k8sManager, nodeListerOpt, configurationOpt)
		Expect(err).ToNot(HaveOccurred())

		Expect(k8sClient.Create(ctx, k3OSConfig)).Should(Succeed())
		Eventually(func() bool { // retry getting the newly created K3OSConfig CR because creation may not happen immediately
			err := k8sClient.Get(ctx, types.NamespacedName{Name: k3OSConfigName, Namespace: k3OSConfigNamespaceName}, k3OSConfig)
			return err == nil
		}, timeout, interval).Should(BeTrue())
		Expect(k3OSConfig.Spec.SyncNodeLabels).Should(BeTrue())
	})

	Context("When creating a K3OSConfig CR", func() {
		// TODO: test each case separately (node config management, label sync, taint sync)
		BeforeEach(func() {
			testConfiguration.NodeConfigFileLocation = tempConfigFile.Name()
			testConfiguration.ManageNodeConfigFile = true
		})
		It("Should call Reconcile", func() {
			// TODO: extract this so it can be reused in every It
			updatedNode := &corev1.Node{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: dummyNode1.GetName()}, updatedNode); err != nil {
					return false
				}

				// FIXME: improve this conditional, maybe by adding a label or by watching the status of the CR? Nothing's updating that yet, though...
				if len(updatedNode.GetLabels()) > 0 && len(updatedNode.Spec.Taints) > 0 && len(updatedNode.GetAnnotations()) > 0 {
					return true
				}

				return false
			}, timeout, interval).Should(BeTrue())

			labels := updatedNode.GetLabels()
			annotations := updatedNode.GetAnnotations()

			Expect(labels["newLabel"]).To(Equal("newLabelValue"))
			checkTaint := &corev1.Taint{Key: "test", Effect: corev1.TaintEffectNoSchedule}
			Expect(taints.TaintExists(updatedNode.Spec.Taints, checkTaint)).To(BeTrue(), "expected taint test:NoSchedule to exist")
			Expect(annotations[consts.AddedLabelsNodeAnnotation()]).To(Equal("newLabel"))
			Expect(annotations[consts.AddedTaintsNodeAnnotation()]).To(Equal("test:NoSchedule"))

			updatedConfigFile, err := ioutil.ReadFile(tempConfigFile.Name())
			Expect(err).ToNot(HaveOccurred())
			Expect(updatedConfigFile).To(Equal(dummyNodeConfig.Spec.Data), fmt.Sprintf("expected config file to contain data:\n\n%v", string(dummyNodeConfig.Spec.Data)))
		})
	})
})

func createTestNodeConfigSecret(secret *corev1.Secret) {
	Expect(k8sClient.Create(ctx, secret)).Should(Succeed())

	Eventually(func() bool {
		return k8sClient.Get(ctx, types.NamespacedName{Name: secret.GetName(), Namespace: secret.GetNamespace()}, secret) == nil
	}, timeout, interval).Should(BeTrue())
}

func createNode(node *corev1.Node) *corev1.Node {
	createdNode := &corev1.Node{}
	Eventually(func() bool {
		return k8sClient.Create(ctx, node) == nil
	}, timeout, interval).Should(BeTrue())
	Eventually(func() bool {
		return k8sClient.Get(ctx, types.NamespacedName{Name: node.GetName()}, createdNode) == nil
	}, timeout, interval).Should(BeTrue())

	return createdNode
}

var _ = AfterSuite(func() {
	By("tearing down the test environment")

	// Cleanup the test namespace
	framework.DeleteNamespace(ctx, framework.DeleteNamespaceInput{Deleter: k8sClient, Name: k3OSConfigNamespaceName}, timeout, interval)

	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())

	cancel()
})
