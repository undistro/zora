// Copyright 2022 Undistro Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Status is the minimally expected status subresource.
type Status struct {
	// ObservedGeneration is the 'Generation' of the resource that
	// was last processed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions the latest available observations of a resource's current state.
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// GetCondition fetches the condition of the specified type.
func (s *Status) GetCondition(t string) *metav1.Condition {
	return meta.FindStatusCondition(s.Conditions, t)
}

// ConditionIsTrue return true if the condition of specified type has status 'True'
func (s *Status) ConditionIsTrue(t string) bool {
	return meta.IsStatusConditionTrue(s.Conditions, t)
}

// SetCondition sets the newCondition in conditions.
//  1. if the condition of the specified type already exists, all fields of the existing condition are updated to
//     newCondition, LastTransitionTime is set to now if the new status differs from the old status
//  2. if a condition of the specified type does not exist, LastTransitionTime is set to now() if unset and newCondition is appended
func (s *Status) SetCondition(newCondition metav1.Condition) {
	if s.Conditions == nil {
		s.Conditions = []metav1.Condition{}
	}
	meta.SetStatusCondition(&s.Conditions, newCondition)
}
