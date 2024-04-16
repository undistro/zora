// Copyright 2024 Undistro Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crds

import (
	"reflect"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

var (
	exampleCRD = v1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: "examples.zora.undistro.io"},
		Spec: v1.CustomResourceDefinitionSpec{
			Group: "zora.undistro.io",
			Names: v1.CustomResourceDefinitionNames{
				Plural:     "examples",
				Singular:   "example",
				ShortNames: []string{"ex", "exs"},
				Kind:       "Example",
				ListKind:   "ExampleList",
			},
			Scope: "Namespaced",
			Versions: []v1.CustomResourceDefinitionVersion{
				{
					Name:         "v1alpha1",
					Served:       true,
					Storage:      false,
					Subresources: &v1.CustomResourceSubresources{Status: &v1.CustomResourceSubresourceStatus{}},
					AdditionalPrinterColumns: []v1.CustomResourceColumnDefinition{
						{
							JSONPath: `.status.conditions[?(@.type=="Ready")].status`,
							Name:     "Ready",
							Type:     "string",
						},
					},
					Schema: &v1.CustomResourceValidation{OpenAPIV3Schema: &v1.JSONSchemaProps{
						Properties: map[string]v1.JSONSchemaProps{
							"foo":    {Type: "string"},
							"status": {Type: "object", Properties: map[string]v1.JSONSchemaProps{"bar": {Type: "string"}}},
						},
					}},
				},
			},
			Conversion:            &v1.CustomResourceConversion{Strategy: v1.NoneConverter},
			PreserveUnknownFields: false,
		},
	}
	v1alpha2Version = v1.CustomResourceDefinitionVersion{
		Name:         "v1alpha2",
		Served:       true,
		Storage:      true,
		Subresources: &v1.CustomResourceSubresources{Status: &v1.CustomResourceSubresourceStatus{}},
		AdditionalPrinterColumns: []v1.CustomResourceColumnDefinition{
			{
				JSONPath: `.status.conditions[?(@.type=="Ready")].status`,
				Name:     "Ready",
				Type:     "string",
			},
		},
		Schema: &v1.CustomResourceValidation{OpenAPIV3Schema: &v1.JSONSchemaProps{
			Properties: map[string]v1.JSONSchemaProps{
				"foo":    {Type: "string"},
				"bar":    {Type: "string"},
				"status": {Type: "object", Properties: map[string]v1.JSONSchemaProps{"bar": {Type: "string"}}},
			},
		}},
	}
)

func TestMergeCRDs(t *testing.T) {
	type args struct {
		existing   v1.CustomResourceDefinition
		updateFunc func(*v1.CustomResourceDefinition)
	}
	tests := []struct {
		name   string
		args   args
		want   *v1.CustomResourceDefinition
		fields []string
	}{
		{
			name: "equal",
			args: args{
				existing: exampleCRD,
				updateFunc: func(crd *v1.CustomResourceDefinition) {
					// just sorting update
					crd.Spec.Names.ShortNames = []string{"exs", "ex"}
					// the same value
					crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["status"] = v1.JSONSchemaProps{Type: "object", Properties: map[string]v1.JSONSchemaProps{"bar": {Type: "string"}}}
					// the default value should be with None strategy
					crd.Spec.Conversion = nil
				},
			},
			want:   &exampleCRD,
			fields: nil,
		},
		{
			name: "ignored fields",
			args: args{
				existing: exampleCRD,
				updateFunc: func(crd *v1.CustomResourceDefinition) {
					crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["new"] = v1.JSONSchemaProps{Type: "string"}
					crd.Spec.Scope = "Cluster"
					crd.Spec.Group = "foo.bar"
					crd.Spec.Names.Kind = "Foo"
					crd.Spec.Names.ListKind = "FooList"
					crd.Spec.Names.Plural = "foos"
					crd.Spec.Names.Singular = "foo"
				},
			},
			want:   &exampleCRD,
			fields: nil,
		},
		{
			name: "allowed updates",
			args: args{
				existing: exampleCRD,
				updateFunc: func(crd *v1.CustomResourceDefinition) {
					crd.Spec.PreserveUnknownFields = true
					crd.Spec.Conversion = &v1.CustomResourceConversion{Strategy: v1.WebhookConverter}
					crd.Spec.Names.ShortNames = append(crd.Spec.Names.ShortNames, "new")
					crd.Spec.Versions[0].AdditionalPrinterColumns[0].Name = "Readyz"
					crd.Spec.Versions[0].Served = false
					crd.Spec.Versions[0].Storage = true
					crd.Spec.Versions[0].Deprecated = true
					crd.Spec.Versions[0].DeprecationWarning = pointer.String("deprecated version")
					crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["status"] = v1.JSONSchemaProps{Type: "string"}
					crd.Spec.Versions = append(crd.Spec.Versions, v1alpha2Version)
				},
			},
			fields: []string{
				"spec.preserveUnknownFields",
				"spec.conversion",
				"spec.names.shortNames",
				`spec.versions[?(@.name=="v1alpha1")].additionalPrinterColumns`,
				`spec.versions[?(@.name=="v1alpha1")].served`,
				`spec.versions[?(@.name=="v1alpha1")].storage`,
				`spec.versions[?(@.name=="v1alpha1")].deprecated`,
				`spec.versions[?(@.name=="v1alpha1")].deprecationWarning`,
				`spec.versions[?(@.name=="v1alpha1")].schema.openAPIV3Schema.properties.status`,
				`spec.versions[?(@.name=="v1alpha2")]`,
			},
			want: &v1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{Name: "examples.zora.undistro.io"},
				Spec: v1.CustomResourceDefinitionSpec{
					Group: "zora.undistro.io",
					Names: v1.CustomResourceDefinitionNames{
						Plural:     "examples",
						Singular:   "example",
						ShortNames: []string{"ex", "exs", "new"},
						Kind:       "Example",
						ListKind:   "ExampleList",
					},
					Scope: "Namespaced",
					Versions: []v1.CustomResourceDefinitionVersion{
						{
							Name:               "v1alpha1",
							Served:             false,
							Storage:            true,
							Deprecated:         true,
							DeprecationWarning: pointer.String("deprecated version"),
							Subresources:       &v1.CustomResourceSubresources{Status: &v1.CustomResourceSubresourceStatus{}},
							AdditionalPrinterColumns: []v1.CustomResourceColumnDefinition{
								{
									JSONPath: `.status.conditions[?(@.type=="Ready")].status`,
									Name:     "Readyz",
									Type:     "string",
								},
							},
							Schema: &v1.CustomResourceValidation{OpenAPIV3Schema: &v1.JSONSchemaProps{
								Properties: map[string]v1.JSONSchemaProps{
									"foo":    {Type: "string"},
									"status": {Type: "string"}},
							}},
						},
						v1alpha2Version,
					},
					Conversion:            &v1.CustomResourceConversion{Strategy: "Webhook"},
					PreserveUnknownFields: true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desired := tt.args.existing.DeepCopy()
			tt.args.updateFunc(desired)
			got, fields := merge(tt.args.existing, *desired)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("merge() mismatch (-want +got):\n%s", cmp.Diff(tt.want, got))
			}
			sort.Strings(fields)
			sort.Strings(tt.fields)
			if !reflect.DeepEqual(fields, tt.fields) {
				t.Errorf("merge() updated fields mismatch (-want +got):\n%s", cmp.Diff(tt.fields, fields))
			}
		})
	}
}
