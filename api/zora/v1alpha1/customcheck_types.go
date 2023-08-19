// Copyright 2023 Undistro Authors
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
	"encoding/json"
	"fmt"

	marvin "github.com/undistro/marvin/pkg/types"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CustomCheckSpec defines the desired state of CustomCheck
type CustomCheckSpec struct {
	Match       Match        `json:"match"`
	Validations []Validation `json:"validations"`
	Message     string       `json:"message"`
	Category    string       `json:"category"`
	URL         string       `json:"url,omitempty"`

	// Parameters to be used in validations
	Params *apiextensionsv1.JSON `json:"params,omitempty"`

	//+kubebuilder:validation:Type=string
	//+kubebuilder:validation:Enum=Low;Medium;High
	Severity string `json:"severity"`
}

func (r *CustomCheck) GetParams() map[string]interface{} {
	var values map[string]interface{}
	if r.Spec.Params != nil {
		_ = json.Unmarshal(r.Spec.Params.Raw, &values)
	}
	return values
}

type Match marvin.Match

type Validation marvin.Validation

// CustomCheckStatus defines the observed state of CustomCheck
type CustomCheckStatus struct {
	Status `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName={check,checks}
//+kubebuilder:printcolumn:name="Message",type="string",JSONPath=".spec.message",priority=0
//+kubebuilder:printcolumn:name="Severity",type="string",JSONPath=".spec.severity",priority=0
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",priority=0

// CustomCheck is the Schema for the customchecks API
type CustomCheck struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomCheckSpec   `json:"spec,omitempty"`
	Status CustomCheckStatus `json:"status,omitempty"`
}

func (r *CustomCheck) FileName() string {
	return fmt.Sprintf("%s.yaml", r.Name)
}

func (r *CustomCheck) ToMarvin() *marvin.Check {
	mv := make([]marvin.Validation, len(r.Spec.Validations))
	for i, v := range r.Spec.Validations {
		mv[i] = marvin.Validation(v)
	}
	return &marvin.Check{
		ID:          r.Name,
		Match:       marvin.Match(r.Spec.Match),
		Validations: mv,
		Params:      r.GetParams(),
		Severity:    marvin.ParseSeverity(r.Spec.Severity),
		Message:     r.Spec.Message,
		Labels:      map[string]string{"category": r.Spec.Category, "url": r.Spec.URL},
	}
}

func (r *CustomCheck) SetReadyStatus(ready bool, reason, msg string) {
	status := metav1.ConditionFalse
	if ready {
		status = metav1.ConditionTrue
	}
	r.Status.SetCondition(metav1.Condition{
		Type:               "Ready",
		Status:             status,
		ObservedGeneration: r.Generation,
		Reason:             reason,
		Message:            msg,
	})
}

//+kubebuilder:object:root=true

// CustomCheckList contains a list of CustomCheck
type CustomCheckList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomCheck `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomCheck{}, &CustomCheckList{})
}
