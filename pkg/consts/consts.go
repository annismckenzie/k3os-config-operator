package consts

import (
	"os"

	"github.com/annismckenzie/k3os-config-operator/pkg/internal/consts"
)

// constants for node secrets
const (
	NodeConfigSecretName = "k3os-nodes"
	nodeNameEnvName      = "NODE_NAME" // see config/manager/manager.yaml
)

// GetNodeName returns the node's name the operator is running on.
// That this is fetched from the environment is an implementation detail.
func GetNodeName() string {
	return os.Getenv(nodeNameEnvName)
}

// GetAddedLabelsNodeAnnotation returns the annotation where labels that the operator added are kept.
func GetAddedLabelsNodeAnnotation() string {
	return consts.AddedLabelsNodeAnnotation
}

// GetAddedTaintsNodeAnnotation returns the annotation where taints that the operator added are kept.
func GetAddedTaintsNodeAnnotation() string {
	return consts.AddedTaintsNodeAnnotation
}
