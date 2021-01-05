package consts

import (
	configv1alpha1 "github.com/annismckenzie/k3os-config-operator/apis/config/v1alpha1"
)

const (
	// NodeAnnotationValueSeperator is the string that is used as the separator for the added labels and taints annotations.
	NodeAnnotationValueSeperator = ","
)

// Even though this package is called consts and these are not constant these must be treated as constants regardless.
// This package is inside the internal package so users of this project as a library cannot update these variables.
var (
	// AnnotationPrefix denotes the prefix of annotations that the operator owns.
	AnnotationPrefix = "k3osconfigs." + configv1alpha1.GroupVersion.Group

	// AddedLabelsNodeAnnotation is the annotation where labels that the operator added are kept.
	AddedLabelsNodeAnnotation = AnnotationPrefix + "/labelsAdded"

	// AddedTaintsNodeAnnotation is the annotation where taints that the operator added are kept.
	AddedTaintsNodeAnnotation = AnnotationPrefix + "/taintsAdded"
)
