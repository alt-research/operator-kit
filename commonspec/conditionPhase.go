package commonspec

import (
	"github.com/alt-research/operator-kit/array"
	mapset "github.com/deckarep/golang-set/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PhaseType string

var SuccessPhases = mapset.NewSet[PhaseType]()

func (p PhaseType) IsSucceed() bool {
	return SuccessPhases.Contains(p)
}

func (p PhaseType) String() string {
	return string(p)
}

func RegisterSuccessPhases(phases ...PhaseType) {
	for _, p := range phases {
		SuccessPhases.Add(p)
	}
}

const (
	PhaseInitializing PhaseType = "Initializing"
	PhasePending      PhaseType = "Pending"
	PhaseFinalize     PhaseType = "Finalizing"
	PhaseIdle         PhaseType = "Idle"

	PhaseError             PhaseType = "Error"
	PhaseInvalid           PhaseType = "Invalid"
	PhaseFinalizationError PhaseType = "FinalizationError"

	PhaseReady     PhaseType = "Ready"
	PhaseRunning   PhaseType = "Running"
	PhaseCompleted PhaseType = "Completed"
)

// ConditionPhase is the common struct to store metav1.Conditions
// and a status Phase Text
type ConditionPhase struct {
	// Conditions is representing the status of each step in the controller reconciliation
	// or can be used to represent some special status that needs extral message attached
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// Phase is a string representing the current status of the controller
	// Mostly used for displaying and checking resource status
	Phase PhaseType `json:"phase,omitempty"`
}

func (c *ConditionPhase) CheckReady(okPhase PhaseType, forConditions ...string) {
	if okPhase == "" {
		okPhase = PhaseReady
	}
	for _, cond := range c.Conditions {
		if len(forConditions) > 0 && !array.Contains(forConditions, cond.Type) {
			continue
		}
		if cond.Status != metav1.ConditionTrue {
			return
		}
	}
	c.Phase = okPhase
}

func (c *ConditionPhase) GetConditions() []metav1.Condition {
	return c.Conditions
}

func init() {
	RegisterSuccessPhases(PhaseCompleted, PhaseRunning, PhaseReady)
}
