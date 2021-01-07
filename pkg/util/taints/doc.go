/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package taints implements utilities for working with taints.
package taints

// The contents of taints.go and taints_test.go were taken from
// https://github.com/kubernetes/kubernetes/tree/653c710b0db8f092c2d46201e178f866c82d0b58/pkg/util/taints
// because they weren't available in any other k8s.io library (even kubectl has a copy of those exact
// files but unexported all of the methods). I'm not going to reimplement all of this functionality
// because it's quite a lot.
