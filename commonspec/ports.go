// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package commonspec

type ServicePortsPort struct {
	Port int32 `json:"port"`
	// +optional
	NodePort int32 `json:"nodePort,omitempty"`
}

type PodSpecPort struct {
	// +optional
	HostPort      int32 `json:"hostPort,omitempty"`
	ContainerPort int32 `json:"containerPort"`
	// +optional
	HostIP string `json:"hostIP,omitempty"`
}
