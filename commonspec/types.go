package commonspec

// GoDuration is a valid time duration that can be parsed by Go's time.ParseDuration() function.
// Supported units: h, m, s, ms
// Examples: `45ms`, `30s`, `1m`, `1h20m15s`
// +kubebuilder:validation:Pattern:="^(0|(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$"
type GoDuration string

type TemplateMeta struct {
	//+optional
	GenerateName string `json:"generateName,omitempty"`
	//+optional
	Labels map[string]string `json:"labels,omitempty"`
	//+optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

type NodeMode string

const (
	ModeFull      = "full"
	ModeValidator = "validator"
	ModeArchive   = "archive"
)
