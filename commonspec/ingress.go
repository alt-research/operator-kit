package commonspec

import netv1 "k8s.io/api/networking/v1"

type IngressSpec struct {
	ClassName   *string            `json:"className,omitempty"`
	Annotations map[string]string  `json:"annotations,omitempty"`
	Labels      map[string]string  `json:"labels,omitempty"`
	TLS         []netv1.IngressTLS `json:"tls,omitempty"`
	Hosts       []IngresHost       `json:"hosts,omitempty"`
}

type IngresHost struct {
	//+optional
	Host string `json:"host,omitempty"`
	//+listType=atomic
	Paths []IngresHostPath `json:"paths,omitempty"`
}

type IngresHostPath struct {
	//+optional
	Path         string `json:"path,omitempty"`
	IgnorePrefix bool   `json:"ignorePrefix,omitempty"`
	//+optional
	//+kubebuilder:default:="ImplementationSpecific"
	PathType *netv1.PathType `json:"pathType,omitempty"`
	//+optional
	Backend *netv1.IngressBackend `json:"backend,omitempty"`
}
