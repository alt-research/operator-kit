package commonspec

import (
	corev1 "k8s.io/api/core/v1"
)

type ImageRef struct {
	//+optional
	Repository string `json:"repository"`
	//+optional
	Tag string `json:"tag,omitempty"`
	//+kubebuilder:default:=IfNotPresent
	//+kubebuilder:validation:Enum=Always;Never;IfNotPresent
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`
	//+optional
	//+patchMergeKey=name
	//+patchStrategy=merge
	PullSecrets []corev1.LocalObjectReference `json:"pullSecrets,omitempty"`
}

func (r ImageRef) Ref() string {
	if r.Repository == "" {
		return ""
	}
	if r.Tag == "" {
		return r.Repository + ":latest"
	}
	return r.Repository + ":" + r.Tag
}
