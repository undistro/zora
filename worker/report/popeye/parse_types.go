package popeye

import zorav1a1 "github.com/getupio-undistro/zora/apis/zora/v1alpha1"

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

	// IssueIDtoUrl maps Popeye's issue codes to urls for wiki pages, blog
	// posts and other sources documenting the issue. The codes do not contain
	// the "POP-" preffix.
	IssueIDtoUrl = map[string]string{
		// Container
		"100": "",
		"101": "",
		"102": "",
		"103": "",
		"104": "",
		"105": "",
		"106": "",
		"107": "",
		"108": "",
		"109": "",
		"110": "",
		"111": "",
		"112": "",
		"113": "",

		// Pod
		"200": "",
		"201": "",
		"202": "",
		"203": "",
		"204": "",
		"205": "",
		"206": "",
		"207": "",
		"208": "",

		// Security
		"300": "",
		"301": "",
		"302": "",
		"303": "",
		"304": "",
		"305": "",
		"306": "",

		// General
		"400": "",
		"401": "",
		"402": "",
		"403": "",
		"404": "",
		"405": "",
		"406": "",

		// Deployment and StatefulSet
		"500": "",
		"501": "",
		"503": "",
		"504": "",
		"505": "",
		"506": "",
		"507": "",

		// HPA
		"600": "",
		"601": "",
		"602": "",
		"603": "",
		"604": "",
		"605": "",

		// Node
		"700": "",
		"701": "",
		"702": "",
		"703": "",
		"704": "",
		"705": "",
		"706": "",
		"707": "",
		"708": "",
		"709": "",
		"710": "",
		"711": "",
		"712": "",

		// Namespace
		"800": "",

		// PodDisruptionBudget
		"900": "",
		"901": "",

		// PV and PVC
		"1000": "",
		"1001": "",
		"1002": "",
		"1003": "",
		"1004": "",

		// Service
		"1100": "",
		"1101": "",
		"1102": "",
		"1103": "",
		"1104": "",
		"1105": "",
		"1106": "",
		"1107": "",
		"1108": "",
		"1109": "",

		// ReplicaSet
		"1120": "",

		// NetworkPolicies
		"1200": "",
		"1201": "",

		// RBAC
		"1300": "",
	}
)
