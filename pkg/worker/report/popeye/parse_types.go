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

import (
	"strconv"
	"strings"

	zorav1a1 "github.com/undistro/zora/api/zora/v1alpha1"
)

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
	// file <internal/issues/assets/codes.yml>.
	IssueIDtoGenericMsg = map[string]string{
		// Container
		"POP-105": "Unnamed probe port in use",
		"POP-108": "Unnamed port",
		"POP-109": "CPU reached request threshold",
		"POP-110": "Memory reached request threshold",
		"POP-111": "CPU reached user threshold",
		"POP-112": "Memory reached user threshold",
		"POP-113": "Container image not hosted on an allowed docker registry",

		// Pod
		"POP-200": "Pod is terminating",
		"POP-201": "Pod is terminating a process",
		"POP-202": "Pod is waiting",
		"POP-203": "Pod is waiting a process",
		"POP-204": "Pod is not ready",
		"POP-205": "Pod was restarted",
		"POP-207": "Pod is in an unhappy phase",
		"POP-209": "Pod is managed by multiple PodDisruptionBudgets",

		// Security
		"POP-304": "ServiceAccount references a secret which does not exist",
		"POP-305": "ServiceAccount references a docker-image pull secret which does not exist",
		"POP-307": "Reference to a non existing ServiceAccount",

		// General
		"POP-401": "Unable to locate key reference",
		"POP-403": "Deprecated API group",
		"POP-404": "Deprecation check failed",
		"POP-407": "References a resource which does not exist",

		// Deployment and StatefulSet
		"POP-501": "Unhealthy, mismatch between desired and available state",
		"POP-503": "At current load, CPU under allocated",
		"POP-504": "At current load, CPU over allocated",
		"POP-505": "At current load, Memory under allocated",
		"POP-506": "At current load, Memory over allocated",
		"POP-507": "Deployment references ServiceAccount which does not exist",
		"POP-508": "No pods match controller selector",

		// HPA
		"POP-600": "HPA references a Deployment which does not exist",
		"POP-602": "Replicas at burst will match or exceed cluster CPU capacity",
		"POP-603": "Replicas at burst will match or exceed cluster memory capacity",
		"POP-604": "If ALL HPAs are triggered, cluster CPU capacity will match or exceed threshold",
		"POP-605": "If ALL HPAs are triggered, cluster memory capacity will match or exceed threshold",

		// Node
		"POP-700": "Found taint that no pod can tolerate",
		"POP-704": "Insufficient memory on Node (MemoryPressure condition)",
		"POP-705": "Insufficient disk space on Node (DiskPressure condition)",
		"POP-706": "Insufficient PIDs on Node (PIDPressure condition)",
		"POP-707": "No network configured on Node (NetworkUnavailable condition)",
		"POP-709": "Node CPU threshold reached",
		"POP-710": "Node Memory threshold reached",

		// PodDisruptionBudget
		"POP-901": "MinAvailable is greater than the number of pods currently running",

		// Service
		"POP-1101": "Skip ports check. No explicit ports detected on pod",
		"POP-1102": "Unnamed service port in use",
		"POP-1106": "No target ports match service port",

		// ReplicaSet
		"POP-1120": "Unhealthy ReplicaSet",

		// NetworkPolicies
		"POP-1200": "No pods match pod selector",
		"POP-1201": "No namespaces match namespace selector",
		"POP-1202": "No pods match pod selector",
		"POP-1203": "Allow All or Deny All policy in effect",
		"POP-1204": "Pod is not secured by a network policy",
		"POP-1206": "No pods matched IPBlock",
		"POP-1207": "No pods matched except IPBlock",
		"POP-1208": "No pods match pod selector in namespace",

		// RBAC
		"POP-1300": "References a role which does not exist",

		// Ingress
		"POP-1400": "Ingress LoadBalancer port reported an error",
		"POP-1401": "Ingress references a service backend which does not exist",
		"POP-1402": "Ingress references a service port which is not defined",
		"POP-1403": "Ingress backend uses a port#, prefer a named port",

		// Cronjob
		"POP-1500": "Cronjob is suspended",

		// CiliumIdentity
		"POP-1601": "Unable to assert namespace label",
		"POP-1602": "References namespace which does not exists",
		"POP-1603": "Missing security namespace label",
		"POP-1604": "Namespace mismatch with security labels namespace",

		//CiliumEndpoint
		"POP-1700": "No cilium endpoints matched selector",
		"POP-1702": "References an unknown node IP",
		"POP-1703": "Pod owner is not in a running state",
		"POP-1704": "References an unknown owner ref",
	}

	// IssueIDtoUrl maps Popeye's issue codes to urls for wiki pages, blog
	// posts and other sources documenting the issue.
	//nolint:lll
	IssueIDtoUrl = map[string]string{
		// Container
		"POP-100": "https://kubernetes.io/docs/concepts/containers/images/#image-names",
		"POP-101": "https://kubernetes.io/docs/concepts/containers/images/#image-names",
		"POP-102": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/",
		"POP-103": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/",
		"POP-104": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-readiness-probes",
		"POP-105": "",
		"POP-106": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
		"POP-107": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
		"POP-108": "",
		"POP-109": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
		"POP-110": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
		"POP-111": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
		"POP-112": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
		"POP-113": "",

		// Pod
		"POP-200": "https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle",
		"POP-201": "https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle",
		"POP-202": "https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle",
		"POP-203": "https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle",
		"POP-204": "https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle",
		"POP-205": "https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle",
		"POP-206": "https://kubernetes.io/docs/concepts/workloads/pods/disruptions",
		"POP-207": "https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle",
		"POP-208": "https://kubernetes.io/docs/concepts/configuration/overview/#naked-pods-vs-replicasets-deployments-and-jobs",
		"POP-209": "https://kubernetes.io/docs/concepts/workloads/pods/disruptions",

		// Security
		"POP-300": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/",
		"POP-301": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/",
		"POP-302": "https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted",
		"POP-303": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/",
		"POP-304": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#add-imagepullsecrets-to-a-service-account",
		"POP-305": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#add-imagepullsecrets-to-a-service-account",
		"POP-306": "https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted",
		"POP-307": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/",

		// General
		"POP-400": "",
		"POP-401": "",
		"POP-402": "https://kubernetes.io/docs/tasks/debug/debug-cluster/resource-metrics-pipeline/#metrics-server",
		"POP-403": "https://kubernetes.io/docs/reference/using-api/deprecation-guide",
		"POP-404": "https://kubernetes.io/docs/reference/using-api/deprecation-guide",
		"POP-405": "https://kubernetes.io/docs/tasks/administer-cluster/cluster-upgrade/",
		"POP-406": "https://kubernetes.io/releases/",
		"POP-407": "",

		// Deployment and StatefulSet
		"POP-500": "https://kubernetes.io/docs/concepts/workloads/",
		"POP-501": "https://kubernetes.io/docs/concepts/workloads/",
		"POP-503": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
		"POP-504": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
		"POP-505": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
		"POP-506": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
		"POP-507": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/",
		"POP-508": "https://kubernetes.io/docs/concepts/workloads/",

		// HPA
		"POP-600": "https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/",
		"POP-601": "https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/",
		"POP-602": "https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/",
		"POP-603": "https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/",
		"POP-604": "https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/",
		"POP-605": "https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/",

		// Node
		"POP-700": "https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/",
		"POP-701": "https://kubernetes.io/docs/concepts/architecture/nodes/#node-status",
		"POP-702": "https://kubernetes.io/docs/concepts/architecture/nodes/#node-status",
		"POP-703": "https://kubernetes.io/docs/concepts/architecture/nodes/#node-status",
		"POP-704": "https://kubernetes.io/docs/concepts/architecture/nodes/#node-status",
		"POP-705": "https://kubernetes.io/docs/concepts/architecture/nodes/#node-status",
		"POP-706": "https://kubernetes.io/docs/concepts/architecture/nodes/#node-status",
		"POP-707": "https://kubernetes.io/docs/concepts/architecture/nodes/#node-status",
		"POP-708": "https://kubernetes.io/docs/tasks/debug/debug-cluster/resource-metrics-pipeline/",
		"POP-709": "https://kubernetes.io/docs/concepts/architecture/nodes/",
		"POP-710": "https://kubernetes.io/docs/concepts/architecture/nodes/",
		"POP-711": "https://kubernetes.io/docs/concepts/architecture/nodes/#manual-node-administration",
		"POP-712": "https://kubernetes.io/docs/concepts/overview/components/",

		// Namespace
		"POP-800": "https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/",

		// PodDisruptionBudget
		"POP-900": "https://kubernetes.io/docs/concepts/workloads/pods/disruptions/",
		"POP-901": "https://kubernetes.io/docs/concepts/workloads/pods/disruptions/",

		// PV and PVC
		"POP-1000": "https://kubernetes.io/docs/concepts/storage/persistent-volumes/",
		"POP-1001": "https://kubernetes.io/docs/concepts/storage/persistent-volumes/",
		"POP-1002": "https://kubernetes.io/docs/concepts/storage/persistent-volumes/",
		"POP-1003": "https://kubernetes.io/docs/concepts/storage/persistent-volumes/",
		"POP-1004": "https://kubernetes.io/docs/concepts/storage/persistent-volumes/",

		// Service
		"POP-1100": "https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service",
		"POP-1101": "https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service",
		"POP-1102": "https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service",
		"POP-1103": "https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer",
		"POP-1104": "https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport",
		"POP-1105": "https://kubernetes.io/docs/concepts/services-networking/service/#services-without-selectors",
		"POP-1106": "https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service",
		"POP-1107": "https://kubernetes.io/docs/concepts/services-networking/service/#external-traffic-policy",
		"POP-1108": "https://kubernetes.io/docs/concepts/services-networking/service/#external-traffic-policy",
		"POP-1109": "https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service",
		"POP-1110": "https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service",

		// ReplicaSet
		"POP-1120": "https://kubernetes.io/docs/concepts/workloads/controllers/replicaset/",

		// NetworkPolicies
		"POP-1200": "https://kubernetes.io/docs/concepts/services-networking/network-policies/#networkpolicy-resource",
		"POP-1201": "https://kubernetes.io/docs/concepts/services-networking/network-policies/#networkpolicy-resource",
		"POP-1202": "https://kubernetes.io/docs/concepts/services-networking/network-policies/#networkpolicy-resource",
		"POP-1203": "https://kubernetes.io/docs/concepts/services-networking/network-policies/#networkpolicy-resource",
		"POP-1204": "https://kubernetes.io/docs/concepts/services-networking/network-policies/#networkpolicy-resource",
		"POP-1205": "https://kubernetes.io/docs/concepts/services-networking/network-policies/#networkpolicy-resource",
		"POP-1206": "https://kubernetes.io/docs/concepts/services-networking/network-policies/#networkpolicy-resource",
		"POP-1207": "https://kubernetes.io/docs/concepts/services-networking/network-policies/#networkpolicy-resource",
		"POP-1208": "https://kubernetes.io/docs/concepts/services-networking/network-policies/#networkpolicy-resource",

		// RBAC
		"POP-1300": "https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding",

		// Ingress
		"POP-1400": "https://kubernetes.io/docs/concepts/services-networking/ingress/#load-balancing",
		"POP-1401": "https://kubernetes.io/docs/concepts/services-networking/ingress/#the-ingress-resource",
		"POP-1402": "https://kubernetes.io/docs/concepts/services-networking/ingress/#the-ingress-resource",
		"POP-1403": "https://kubernetes.io/docs/concepts/services-networking/ingress/#the-ingress-resource",
		"POP-1404": "https://kubernetes.io/docs/concepts/services-networking/ingress/#the-ingress-resource",

		// Cronjob
		"POP-1500": "https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/",
		"POP-1501": "https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/",
		"POP-1502": "https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/",

		// CiliumIdentity
		"POP-1600": "https://docs.cilium.io/en/stable/",
		"POP-1601": "https://docs.cilium.io/en/stable/",
		"POP-1602": "https://docs.cilium.io/en/stable/",
		"POP-1603": "https://docs.cilium.io/en/stable/",
		"POP-1604": "https://docs.cilium.io/en/stable/",

		// CiliumEndpoint
		"POP-1700": "https://docs.cilium.io/en/stable/",
		"POP-1701": "https://docs.cilium.io/en/stable/",
		"POP-1702": "https://docs.cilium.io/en/stable/",
		"POP-1703": "https://docs.cilium.io/en/stable/",
		"POP-1704": "https://docs.cilium.io/en/stable/",
	}
	categoriesInterval = map[int]string{
		1:  "Container",
		2:  "Pod",
		3:  "Security",
		4:  "General",
		5:  "Workloads",
		6:  "HorizontalPodAutoscaler",
		7:  "Node",
		8:  "Namespace",
		9:  "PodDisruptionBudget",
		10: "Volumes",
		11: "Service",
		// exception for "ReplicaSet" of POP-1120
		12: "NetworkPolicies",
		13: "RBAC",
		14: "Ingress",
		15: "Cronjob",
		16: "CiliumIdentity",
		17: "CiliumEndpoint",
	}
)

// getCategory returns a category from the given issue ID.
// https://github.com/derailed/popeye/blob/master/docs/codes.md
func getCategory(s string) string {
	ss := strings.SplitN(s, "-", 2)
	if len(ss) != 2 {
		return ""
	}

	id, err := strconv.Atoi(ss[1])
	if err != nil {
		return ""
	}

	if id == 1120 {
		return "ReplicaSet"
	}

	for i, category := range categoriesInterval {
		start := i * 100
		if id >= start && id < start+100 {
			return category
		}
	}

	return ""
}
