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
	"context"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/undistro/zora/config/crd/bases"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"
)

var (
	log  = logf.Log.WithName("crds")
	CRDs []apiextensionsv1.CustomResourceDefinition
)

// Update updates Zora CRDs if needed
func Update(ctx context.Context, client *apiextensionsv1client.ApiextensionsV1Client) error {
	for _, crd := range CRDs {
		existing, err := client.CustomResourceDefinitions().Get(ctx, crd.Name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				log.Info("CRD not found", "name", crd.Name)
				continue
			}
			return err
		}
		obj, updatedFields := merge(*existing, crd)
		if len(updatedFields) == 0 {
			log.Info("Unchanged CRD", "name", crd.Name)
			continue
		}
		if _, err := client.CustomResourceDefinitions().Update(ctx, obj, metav1.UpdateOptions{}); err != nil {
			return err
		}
		log.Info("CRD updated", "name", crd.Name, "changes", updatedFields)
	}
	return nil
}

func merge(existing, desired apiextensionsv1.CustomResourceDefinition) (*apiextensionsv1.CustomResourceDefinition, []string) {
	existingVersions := make(map[string]apiextensionsv1.CustomResourceDefinitionVersion, len(existing.Spec.Versions))
	for _, v := range existing.Spec.Versions {
		existingVersions[v.Name] = v
	}
	result := existing.DeepCopy()
	var updatedFields []string

	if !equality.Semantic.DeepEqual(result.ObjectMeta.Annotations, desired.ObjectMeta.Annotations) {
		for k, v := range desired.ObjectMeta.Annotations {
			if result.ObjectMeta.Annotations == nil {
				result.ObjectMeta.Annotations = make(map[string]string, len(desired.ObjectMeta.Annotations))
			}
			result.ObjectMeta.Annotations[k] = v
		}
		updatedFields = append(updatedFields, "metadata.annotations")
	}

	if result.Spec.PreserveUnknownFields != desired.Spec.PreserveUnknownFields {
		result.Spec.PreserveUnknownFields = desired.Spec.PreserveUnknownFields
		updatedFields = append(updatedFields, "spec.preserveUnknownFields")
	}

	if !equality.Semantic.DeepEqual(conversionOrNone(result.Spec.Conversion), conversionOrNone(desired.Spec.Conversion)) {
		result.Spec.Conversion = desired.Spec.Conversion
		updatedFields = append(updatedFields, "spec.conversion")
	}

	sort.Strings(result.Spec.Names.ShortNames)
	sort.Strings(desired.Spec.Names.ShortNames)
	if !equality.Semantic.DeepEqual(result.Spec.Names.ShortNames, desired.Spec.Names.ShortNames) {
		result.Spec.Names.ShortNames = desired.Spec.Names.ShortNames
		updatedFields = append(updatedFields, "spec.names.shortNames")
	}

	for i, desiredVersion := range desired.Spec.Versions {
		existingVersion, exists := existingVersions[desiredVersion.Name]
		if !exists {
			// desired version doesn't exist in the existing CRD
			result.Spec.Versions = append(result.Spec.Versions, desiredVersion)
			updatedFields = append(updatedFields, fmt.Sprintf(`spec.versions[?(@.name==%q)]`, desiredVersion.Name))
			continue
		}

		if !equality.Semantic.DeepEqual(existingVersion.AdditionalPrinterColumns, desiredVersion.AdditionalPrinterColumns) {
			result.Spec.Versions[i].AdditionalPrinterColumns = desiredVersion.AdditionalPrinterColumns
			updatedFields = append(updatedFields, fmt.Sprintf(`spec.versions[?(@.name==%q)].additionalPrinterColumns`, desiredVersion.Name))
		}
		desiredSchemaStatus := desiredVersion.Schema.OpenAPIV3Schema.Properties["status"]
		if !equality.Semantic.DeepEqual(existingVersion.Schema.OpenAPIV3Schema.Properties["status"], desiredSchemaStatus) {
			result.Spec.Versions[i].Schema.OpenAPIV3Schema.Properties["status"] = desiredSchemaStatus
			updatedFields = append(updatedFields, fmt.Sprintf(`spec.versions[?(@.name==%q)].schema.openAPIV3Schema.properties.status`, desiredVersion.Name))
		}
		if existingVersion.Served != desiredVersion.Served {
			result.Spec.Versions[i].Served = desiredVersion.Served
			updatedFields = append(updatedFields, fmt.Sprintf(`spec.versions[?(@.name==%q)].served`, desiredVersion.Name))
		}
		if existingVersion.Storage != desiredVersion.Storage {
			result.Spec.Versions[i].Storage = desiredVersion.Storage
			updatedFields = append(updatedFields, fmt.Sprintf(`spec.versions[?(@.name==%q)].storage`, desiredVersion.Name))
		}
		if existingVersion.Deprecated != desiredVersion.Deprecated {
			result.Spec.Versions[i].Deprecated = desiredVersion.Deprecated
			updatedFields = append(updatedFields, fmt.Sprintf(`spec.versions[?(@.name==%q)].deprecated`, desiredVersion.Name))
		}
		if existingVersion.DeprecationWarning != desiredVersion.DeprecationWarning {
			result.Spec.Versions[i].DeprecationWarning = desiredVersion.DeprecationWarning
			updatedFields = append(updatedFields, fmt.Sprintf(`spec.versions[?(@.name==%q)].deprecationWarning`, desiredVersion.Name))
		}
	}
	return result, updatedFields
}

func conversionOrNone(c *apiextensionsv1.CustomResourceConversion) *apiextensionsv1.CustomResourceConversion {
	if c != nil {
		return c
	}
	return &apiextensionsv1.CustomResourceConversion{Strategy: apiextensionsv1.NoneConverter}
}

func init() {
	entries, err := bases.CRDsFS.ReadDir(".")
	if err != nil {
		panic(err)
	}
	crds := make([]apiextensionsv1.CustomResourceDefinition, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if filepath.Ext(name) != ".yaml" {
			continue
		}
		bs, err := bases.CRDsFS.ReadFile(name)
		if err != nil {
			panic(err)
		}
		crd := &apiextensionsv1.CustomResourceDefinition{}
		if err := yaml.Unmarshal(bs, crd); err != nil {
			panic(err)
		}
		crds = append(crds, *crd)
	}
	CRDs = crds
}
