// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package commonspec

// WellKnowns
// Ref: https://kubernetes.io/docs/reference/labels-annotations-taints
const (
	Component = "app.kubernetes.io/component"
	ManagedBy = "app.kubernetes.io/managed-by"
	Name      = "app.kubernetes.io/name"
	PartOf    = "app.kubernetes.io/part-of"
	Version   = "app.kubernetes.io/version"
)

// Alt Operator
const (
	Finalizer = "operator.altlayer.io/finalizer"

	ParentPrefix = "parents.operator.altlayer.io/"

	OperatorDomain = "operator.altlayer.io"
	Index          = "index"
	Mode           = "mode"
)
