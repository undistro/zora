package apis

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Status is the minimally expected status subresource.
// +k8s:deepcopy-gen=true
type Status struct {
	// ObservedGeneration is the 'Generation' of the resource that
	// was last processed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions the latest available observations of a resource's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// GetCondition fetches the condition of the specified type.
func (s *Status) GetCondition(t string) *metav1.Condition {
	for _, cond := range s.Conditions {
		if cond.Type == t {
			return &cond
		}
	}
	return nil
}

// ConditionIsTrue return true if the condition of specified type has status 'True'
func (s *Status) ConditionIsTrue(t string) bool {
	for _, cond := range s.Conditions {
		if cond.Type == t {
			return cond.Status == metav1.ConditionTrue
		}
	}
	return false
}
