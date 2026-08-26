// Copyright 2026 Undistro Authors
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

package plugins

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"

	"github.com/undistro/zora/api/zora/v1alpha1"
)

func TestCronJobMutatorAppliesPluginSecurityContextToSupportContainers(t *testing.T) {
	securityContext := &corev1.SecurityContext{
		RunAsNonRoot:             pointer.Bool(true),
		AllowPrivilegeEscalation: pointer.Bool(false),
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{"ALL"},
		},
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
	}

	mutator := &CronJobMutator{
		Plugin: &v1alpha1.Plugin{Spec: v1alpha1.PluginSpec{SecurityContext: securityContext}},
		ClusterScan: &v1alpha1.ClusterScan{Spec: v1alpha1.ClusterScanSpec{
			ClusterRef: corev1.LocalObjectReference{Name: "test-cluster"},
		}},
	}

	worker := mutator.workerContainer()
	checks := mutator.initContainer()

	if !reflect.DeepEqual(worker.SecurityContext, securityContext) {
		t.Fatalf("worker security context = %#v, want %#v", worker.SecurityContext, securityContext)
	}
	if !reflect.DeepEqual(checks.SecurityContext, securityContext) {
		t.Fatalf("checks security context = %#v, want %#v", checks.SecurityContext, securityContext)
	}

	worker.SecurityContext.Capabilities.Drop[0] = "CHOWN"
	if checks.SecurityContext.Capabilities.Drop[0] != "ALL" {
		t.Fatal("worker and checks security contexts share mutable state")
	}
}

func TestCronJobMutatorUsesSafeDefaultSecurityContext(t *testing.T) {
	mutator := &CronJobMutator{
		Plugin:      &v1alpha1.Plugin{},
		ClusterScan: &v1alpha1.ClusterScan{},
	}

	want := &corev1.SecurityContext{
		RunAsNonRoot:             pointer.Bool(true),
		AllowPrivilegeEscalation: pointer.Bool(false),
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{"ALL"},
		},
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
	}

	if got := mutator.workerContainer().SecurityContext; !reflect.DeepEqual(got, want) {
		t.Fatalf("worker security context = %#v, want %#v", got, want)
	}
	if got := mutator.initContainer().SecurityContext; !reflect.DeepEqual(got, want) {
		t.Fatalf("checks security context = %#v, want %#v", got, want)
	}
}
