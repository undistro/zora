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

package popeye

import zorav1a1 "github.com/undistro/zora/apis/zora/v1alpha1"

var (
	// LevelToIssueSeverity maps Popeye's <Level> type to Zora's
	// <ClusterIssueSeverity>.
	LevelToIssueSeverity = [4]zorav1a1.ClusterIssueSeverity{
		zorav1a1.SeverityUnknown,
		zorav1a1.SeverityLow,
		zorav1a1.SeverityMedium,
		zorav1a1.SeverityHigh,
	}

	// IssueIDtoGenericMsg maps Popeye's issue codes to generic versions of the
	// issue description. The original issues can be found on Popeye's source
	// file <internal/issues/assets/codes.yml>. The codes do not contain the
	// "POP-" preffix.
	IssueIDtoGenericMsg = map[string]string{
		// Container
		"105": "Unnamed probe port in use",
		"108": "Unnamed port",
		"109": "CPU reached request threshold",
		"110": "Memory reached request threshold",
		"111": "CPU reached user threshold",
		"112": "Memory reached user threshold",
		"113": "Container image not hosted on an allowed docker registry",

		// Pod
		"200": "Pod is terminating",
		"201": "Pod is terminating a process",
		"202": "Pod is waiting",
		"203": "Pod is waiting a process",
		"204": "Pod is not ready",
		"205": "Pod was restarted",
		"207": "Pod is in an unhappy phase",

		// Security
		"304": "ServiceAccount references a secret which does not exist",
		"305": "ServiceAccount references a docker-image pull secret which does not exist",

		// General
		"401": "Unable to locate key reference",
		"402": "No metric-server detected",
		"403": "Deprecated API group",
		"404": "Deprecation check failed",

		// Deployment and StatefulSet
		"501": "Unhealthy, mismatch between desired and available state",
		"503": "At current load, CPU under allocated",
		"504": "At current load, CPU over allocated",
		"505": "At current load, Memory under allocated",
		"506": "At current load, Memory over allocated",
		"507": "Deployment references ServiceAccount which does not exist",

		// HPA
		"600": "HPA references a Deployment which does not exist",
		"601": "HPA references a StatefulSet which does not exist",
		"602": "Replicas at burst will match or exceed cluster CPU capacity",
		"603": "Replicas at burst will match or exceed cluster memory capacity",
		"604": "If ALL HPAs are triggered, cluster CPU capacity will match or exceed threshold",
		"605": "If ALL HPAs are triggered, cluster memory capacity will match or exceed threshold",

		// Node
		"700": "Found taint that no pod can tolerate",
		"704": "Insufficient memory on Node (MemoryPressure condition)",
		"705": "Insufficient disk space on Node (DiskPressure condition)",
		"706": "Insufficient PIDs on Node (PIDPressure condition)",
		"707": "No network configured on Node (NetworkUnavailable condition)",
		"709": "Node CPU threshold reached",
		"710": "Node Memory threshold reached",

		// PodDisruptionBudget
		"901": "MinAvailable is greater than the number of pods currently running",

		// Service
		"1101": "Skip ports check. No explicit ports detected on pod",
		"1102": "Unnamed service port in use",
		"1106": "No target ports match service port",

		// ReplicaSet
		"1120": "Unhealthy ReplicaSet",

		// NetworkPolicies
		"1200": "No pods match pod selector",
		"1201": "No namespaces match namespace selector",

		// RBAC
		"1300": "References a role which does not exist",
	}

	// IssueIDtoUrl maps Popeye's issue codes to urls for wiki pages, blog
	// posts and other sources documenting the issue. The codes do not contain
	// the "POP-" preffix.
	IssueIDtoUrl = map[string]string{
		// Container
		"100": "https://kubernetes.io/docs/concepts/containers/images/#image-names",
		"101": "https://kubernetes.io/docs/concepts/containers/images/#image-names",
		"102": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/",
		"103": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/",
		"104": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-readiness-probes",
		"105": "",
		"106": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
		"107": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
		"108": "",
		"109": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
		"110": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
		"111": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
		"112": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
		"113": "",

		// Pod
		"200": "https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle",
		"201": "https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle",
		"202": "https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle",
		"203": "https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle",
		"204": "https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle",
		"205": "https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle",
		"206": "https://kubernetes.io/docs/concepts/workloads/pods/disruptions",
		"207": "https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle",
		"208": "https://kubernetes.io/docs/concepts/configuration/overview/#naked-pods-vs-replicasets-deployments-and-jobs",

		// Security
		"300": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/",
		"301": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/",
		"302": "https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted",
		"303": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/",
		"304": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#add-imagepullsecrets-to-a-service-account",
		"305": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#add-imagepullsecrets-to-a-service-account",
		"306": "https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted",

		// General
		"400": "",
		"401": "",
		"402": "https://kubernetes.io/docs/tasks/debug/debug-cluster/resource-metrics-pipeline/#metrics-server",
		"403": "https://kubernetes.io/docs/reference/using-api/deprecation-guide",
		"404": "https://kubernetes.io/docs/reference/using-api/deprecation-guide",
		"405": "https://kubernetes.io/docs/tasks/administer-cluster/cluster-upgrade/",
		"406": "https://kubernetes.io/releases/",

		// Deployment and StatefulSet
		"500": "https://kubernetes.io/docs/concepts/workloads/",
		"501": "https://kubernetes.io/docs/concepts/workloads/",
		"503": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
		"504": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
		"505": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
		"506": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
		"507": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/",

		// HPA
		"600": "https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/",
		"601": "https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/",
		"602": "https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/",
		"603": "https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/",
		"604": "https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/",
		"605": "https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/",

		// Node
		"700": "https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/",
		"701": "https://kubernetes.io/docs/concepts/architecture/nodes/#node-status",
		"702": "https://kubernetes.io/docs/concepts/architecture/nodes/#node-status",
		"703": "https://kubernetes.io/docs/concepts/architecture/nodes/#node-status",
		"704": "https://kubernetes.io/docs/concepts/architecture/nodes/#node-status",
		"705": "https://kubernetes.io/docs/concepts/architecture/nodes/#node-status",
		"706": "https://kubernetes.io/docs/concepts/architecture/nodes/#node-status",
		"707": "https://kubernetes.io/docs/concepts/architecture/nodes/#node-status",
		"708": "https://kubernetes.io/docs/tasks/debug/debug-cluster/resource-metrics-pipeline/",
		"709": "https://kubernetes.io/docs/concepts/architecture/nodes/",
		"710": "https://kubernetes.io/docs/concepts/architecture/nodes/",
		"711": "https://kubernetes.io/docs/concepts/architecture/nodes/#manual-node-administration",
		"712": "https://kubernetes.io/docs/concepts/overview/components/",

		// Namespace
		"800": "https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/",

		// PodDisruptionBudget
		"900": "https://kubernetes.io/docs/concepts/workloads/pods/disruptions/",
		"901": "https://kubernetes.io/docs/concepts/workloads/pods/disruptions/",

		// PV and PVC
		"1000": "https://kubernetes.io/docs/concepts/storage/persistent-volumes/",
		"1001": "https://kubernetes.io/docs/concepts/storage/persistent-volumes/",
		"1002": "https://kubernetes.io/docs/concepts/storage/persistent-volumes/",
		"1003": "https://kubernetes.io/docs/concepts/storage/persistent-volumes/",
		"1004": "https://kubernetes.io/docs/concepts/storage/persistent-volumes/",

		// Service
		"1100": "https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service",
		"1101": "https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service",
		"1102": "https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service",
		"1103": "https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer",
		"1104": "https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport",
		"1105": "https://kubernetes.io/docs/concepts/services-networking/service/#services-without-selectors",
		"1106": "https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service",
		"1107": "https://kubernetes.io/docs/concepts/services-networking/service/#external-traffic-policy",
		"1108": "https://kubernetes.io/docs/concepts/services-networking/service/#external-traffic-policy",
		"1109": "https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service",

		// ReplicaSet
		"1120": "https://kubernetes.io/docs/concepts/workloads/controllers/replicaset/",

		// NetworkPolicies
		"1200": "https://kubernetes.io/docs/concepts/services-networking/network-policies/#networkpolicy-resource",
		"1201": "https://kubernetes.io/docs/concepts/services-networking/network-policies/#networkpolicy-resource",

		// RBAC
		"1300": "https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding",
	}
)
