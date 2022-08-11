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
		"100": "https://kubernetes.io/pt-br/docs/concepts/containers/images/#nomes-das-imagens",
		"101": "https://kubernetes.io/docs/concepts/containers/images/#image-names",
		"102": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-tcp-liveness-probe",
		"103": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-liveness-command",
		"104": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-liveness-command",
		"105": "",
		"106": "https://cloud.google.com/blog/products/containers-kubernetes/kubernetes-best-practices-resource-requests-and-limits",
		"107": "https://cloud.google.com/blog/products/containers-kubernetes/kubernetes-best-practices-resource-requests-and-limits",
		"108": "",
		"109": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#meaning-of-cpu",
		"110": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#meaning-of-memory",
		"111": "",
		"112": "",
		"113": "",

		// Pod
		"200": "https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle",
		"201": "https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle",
		"202": "https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle",
		"203": "https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle",
		"204": "https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle",
		"205": "https://www.ibm.com/docs/en/cloud-paks/cp-management/1.2.0?topic=issues-pods-restart-frequently",
		"206": "https://kubernetes.io/docs/concepts/workloads/pods/disruptions",
		"207": "https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle",
		"208": "https://kubernetes.io/pt-br/docs/concepts/scheduling-eviction/pod-overhead",

		// Security
		"300": "https://thenewstack.io/kubernetes-access-control-exploring-service-accounts/#:~:text=Service%20accounts%20are%20associated%20with,a%20ClusterIP%20service%20called%20Kubernete",
		"301": "https://kubernetes.io/docs/reference/access-authn-authz/bootstrap-tokens",
		"302": "https://medium.com/devzera/processos-em-containers-n%C3%A3o-devem-ser-executados-como-root-30755daff56f",
		"303": "https://kubernetes.io/docs/concepts/security/controlling-access",
		"304": "https://kubernetes.io/docs/concepts/configuration/secret/#using-a-secret",
		"305": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#add-imagepullsecrets-to-a-service-account",
		"306": "",

		// General
		"400": "",
		"401": "",
		"402": "https://www.ibm.com/docs/en/spp/10.1.8?topic=prerequisites-kubernetes-verifying-whether-metrics-server-is-running",
		"403": "https://kubernetes.io/docs/reference/using-api/deprecation-guide",
		"404": "https://opster.com/analysis/elasticsearch-nodes-failed-to-run-deprecation-checks",
		"405": "https://faun.pub/upgrade-your-kubernetes-cluster-without-upsetting-your-developers-e7d8559dee49",
		"406": "",

		// Deployment and StatefulSet
		"500": "https://dzone.com/articles/scale-to-zero-with-kubernetes",
		"501": "",
		"503": "",
		"504": "",
		"505": "",
		"506": "",
		"507": "https://medium.com/the-programmer/working-with-service-account-in-kubernetes-df129cb4d1cc",

		// HPA
		"600": "https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#implicit-maintenance-mode-deactivation",
		"601": "https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#implicit-maintenance-mode-deactivation",
		"602": "https://medium.com/omio-engineering/cpu-limits-and-aggressive-throttling-in-kubernetes-c6b20bd8a718",
		"603": "https://kubernetes.io/docs/tasks/configure-pod-container/assign-memory-resource",
		"604": "https://kubernetes.io/docs/tasks/configure-pod-container/assign-memory-resource",
		"605": "https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale",

		// Node
		"700": "https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/",
		"701": "https://kubernetes.io/docs/concepts/architecture/nodes/#node-controller",
		"702": "https://komodor.com/learn/how-to-fix-kubernetes-node-not-ready-error",
		"703": "https://www.ibm.com/docs/en/api-connect/10.0.x?topic=kubernetes-analytics-running-out-disk-space",
		"704": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers",
		"705": "https://kubernetes.io/docs/concepts/scheduling-eviction/node-pressure-eviction",
		"706": "https://containersolutions.github.io/runbooks/posts/kubernetes/0-nodes-available-insufficient",
		"707": "https://www.ibm.com/docs/en/csfdcd/7.1?topic=network-configuring-node",
		"708": "https://www.datadoghq.com/blog/how-to-collect-and-graph-kubernetes-metrics",
		"709": "https://www.ibm.com/docs/en/cloud-app-management/2019.4.0?topic=collector-kubernetes-metrics-thresholds",
		"710": "https://medium.com/@betz.mark/understanding-resource-limits-in-kubernetes-memory-6b41e9a955f9",
		"711": "https://kubernetes.io/docs/reference/scheduling/config",
		"712": "",

		// Namespace
		"800": "https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/#viewing-namespaces",

		// PodDisruptionBudget
		"900": "https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node",
		"901": "https://kubernetes.io/docs/tasks/run-application/configure-pdb",

		// PV and PVC
		"1000": "https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors",
		"1001": "https://hungsblog.de/en/technology/troubleshooting/kubernetes-pod-stuck-in-pending-status-nodes-had-no-available-volume",
		"1002": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-volume-storage",
		"1003": "https://www.ibm.com/docs/en/cloud-paks/1.0?topic=issues-persistentvolumeclaims-pvcs-are-in-pending-state",
		"1004": "",

		// Service
		"1100": "https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors",
		"1101": "https://www.gitpod.io/docs/config-ports",
		"1102": "https://www.bmc.com/blogs/kubernetes-port-targetport-nodeport/#:~:text=TargetPort%20is%20the%20port%20on,IP%20address%20and%20the%20NodePort",
		"1103": "https://kubernetes.io/docs/concepts/services-networking/_print/#load-balancing",
		"1104": "",
		"1105": "",
		"1106": "https://www.bmc.com/blogs/kubernetes-port-targetport-nodeport",
		"1107": "",
		"1108": "",
		"1109": "https://www.codegrepper.com/code-examples/shell/only+one+associated+endpoint+kubernetes",

		// ReplicaSet
		"1120": "",

		// NetworkPolicies
		"1200": "https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node",
		"1201": "https://kubernetes.io/docs/concepts/overview/working-with-objects/labels",

		// RBAC
		"1300": "",
	}
)
