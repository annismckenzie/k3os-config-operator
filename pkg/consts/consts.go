package consts

import (
	"os"

	"github.com/annismckenzie/k3os-config-operator/pkg/internal/consts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeConfigSecretName contains the name of the node Secret
const NodeConfigSecretName = "k3os-nodes"

// NodeConfigFileLocation contains the location of the mounted k3OS config file location.
const NodeConfigFileLocation = "/etc/k3osconfig.yaml" // see config/manager/nodeconfigfile_management.yaml

// environment variable names
const (
	nodeNameEnvName  = "NODE_NAME" // see config/manager/manager.yaml
	namespaceEnvName = "NAMESPACE"

	DevModeEnvName                 = "DEV_MODE"
	EnableNodeConfigFileManagement = "ENABLE_NODECONFIG_FILE_MANAGEMENT"
)

// Namespace returns the configured namespace.
// That this is fetched from the environment is an implementation detail.
func Namespace() string {
	return os.Getenv(namespaceEnvName)
}

// NodeName returns the node's name the operator is running on.
// That this is fetched from the environment is an implementation detail.
func NodeName() string {
	return os.Getenv(nodeNameEnvName)
}

// AddedLabelsNodeAnnotation returns the annotation where labels that the operator added are kept.
func AddedLabelsNodeAnnotation() string {
	return consts.AddedLabelsNodeAnnotation
}

// AddedTaintsNodeAnnotation returns the annotation where taints that the operator added are kept.
func AddedTaintsNodeAnnotation() string {
	return consts.AddedTaintsNodeAnnotation
}

// LabelSelectorForNodeConfigFileSecret returns the label selector for the k3OS node config file secret.
func LabelSelectorForNodeConfigFileSecret() metav1.LabelSelector {
	labelSelector := metav1.AddLabelToSelector(&metav1.LabelSelector{}, "app.kubernetes.io/managed-by", "k3os-config-operator")
	return *labelSelector
}
