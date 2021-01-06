package consts

import (
	"github.com/annismckenzie/k3os-config-operator/pkg/internal/consts"
)

// constants for node secrets
const (
	NodeConfigSecretName = "k3os-nodes"
	NodeNameEnvName      = "NODE_NAME" // see config/manager/manager.yaml
)

// GetAddedLabelsNodeAnnotation returns the annotation where labels that the operator added are kept.
func GetAddedLabelsNodeAnnotation() string {
	return consts.AddedLabelsNodeAnnotation
}

// GetAddedTaintsNodeAnnotation returns the annotation where taints that the operator added are kept.
func GetAddedTaintsNodeAnnotation() string {
	return consts.AddedTaintsNodeAnnotation
}
