package commonspec

import (
	"github.com/alt-research/operator-kit/must"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PersistenceSpec struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	//+optional
	Size resource.Quantity `json:"size"`
	//+optional
	// TODO: support autoscale
	AutoScale *PersistenceAutoScale `json:"autoScale,omitempty"`
	//+optional
	ExternalClaimName *string `json:"externalClaim"`
	//+optional
	StorageClassName *string `json:"storageClassName,omitempty"`
	//+optional
	AccessModes []corev1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`
	//+optional
	EmptyDir           bool `json:"emptyDir,omitempty"`
	DeleteOnFinalizing bool `json:"deleteOnFinalizing,omitempty"`
}

// PersistenceAutoScale defines how to scale the volume
// Ref: https://github.com/DevOps-Nirvana/Kubernetes-Volume-Autoscaler#per-volume-configuration-via-annotations
type PersistenceAutoScale struct {
	//+kubebuilder:validation:ExclusiveMaximum:=true
	//+kubebuilder:validation:Maximum:=100
	//+kubebuilder:validation:Minimum:=20
	//+kubebuilder:default:=80
	PercentThreshold int `json:"percentThreshold,omitempty"`
	//+kubebuilder:default:=20
	ScaleUpPercent      int                `json:"scaleUpPercent,omitempty"`
	MaxSize             *resource.Quantity `json:"maxSize,omitempty"`
	ScaleAfterIntervals metav1.Duration    `json:"scaleAfterIntervals,omitempty"`
}

func (p *PersistenceSpec) SetDefaults() {
	p.AccessModes = must.Default(p.AccessModes, []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce})
	if p.Size.IsZero() {
		p.Size = must.Two(resource.ParseQuantity("10Gi"))
	}
}
