package popeye

import zorav1a1 "github.com/getupio-undistro/zora/apis/zora/v1alpha1"

var (
	// LevelToIssueSeverity maps Popeye's <Level> type to Inspect's
	// <ClusterIssueSeverity>.
	LevelToIssueSeverity = [4]zorav1a1.ClusterIssueSeverity{
		zorav1a1.SeverityNone,
		zorav1a1.SeverityLow,
		zorav1a1.SeverityMedium,
		zorav1a1.SeverityHigh,
	}

	// IssueIDtoGenericMsg maps Popeye's issue codes to generic versions of the
	// issue description. The original issues can be found on Popeye's source
	// file <internal/issues/assets/codes.yml>.
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
		"304": "References a secret which does not exist",
		"305": "References a docker-image pull secret which does not exist",

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
		"709": "CPU threshold reached",
		"710": "Memory threshold reached",

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
)
