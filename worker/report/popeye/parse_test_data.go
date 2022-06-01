package popeye

const (
	TestReport1 = `
{
	"popeye": {
		"sanitizers": [
		{
			"sanitizer": "cluster"
		},
		{
			"sanitizer": "clusterroles",
			"gvr": "rbac.authorization.k8s.io/v1/clusterroles",
			"issues": {
				"admin": [
				{
					"group": "__root__",
					"gvr": "rbac.authorization.k8s.io/v1/clusterroles",
					"level": 1,
					"message": "[POP-400] Used? Unable to locate resource reference"
				}
				],
				"aws-node": [],
				"capa-manager-role": [],
				"capa-proxy-role": [],
				"capi-kubeadm-control-plane-manager-role": [
				{
					"group": "__root__",
					"gvr": "rbac.authorization.k8s.io/v1/clusterroles",
					"level": 1,
					"message": "[POP-400] Used? Unable to locate resource reference"
				}
				],
				"capi-manager-role": [
				{
					"group": "__root__",
					"gvr": "rbac.authorization.k8s.io/v1/clusterroles",
					"level": 1,
					"message": "[POP-400] Used? Unable to locate resource reference"
				}
				],
				"cert-manager-edit": [
				{
					"group": "__root__",
					"gvr": "rbac.authorization.k8s.io/v1/clusterroles",
					"level": 1,
					"message": "[POP-400] Used? Unable to locate resource reference"
				}
				],
				"cert-manager-view": [
				{
					"group": "__root__",
					"gvr": "rbac.authorization.k8s.io/v1/clusterroles",
					"level": 1,
					"message": "[POP-400] Used? Unable to locate resource reference"
				}
				],
				"system:certificates.k8s.io:kube-apiserver-client-kubelet-approver": [
				{
					"group": "__root__",
					"gvr": "rbac.authorization.k8s.io/v1/clusterroles",
					"level": 1,
					"message": "[POP-400] Used? Unable to locate resource reference"
				}
				],
				"system:certificates.k8s.io:kubelet-serving-approver": [
				{
					"group": "__root__",
					"gvr": "rbac.authorization.k8s.io/v1/clusterroles",
					"level": 1,
					"message": "[POP-400] Used? Unable to locate resource reference"
				}
				],
				"system:certificates.k8s.io:legacy-unknown-approver": [
				{
					"group": "__root__",
					"gvr": "rbac.authorization.k8s.io/v1/clusterroles",
					"level": 1,
					"message": "[POP-400] Used? Unable to locate resource reference"
				}
				],
				"system:heapster": [
				{
					"group": "__root__",
					"gvr": "rbac.authorization.k8s.io/v1/clusterroles",
					"level": 1,
					"message": "[POP-400] Used? Unable to locate resource reference"
				}
				],
				"system:kube-aggregator": [
				{
					"group": "__root__",
					"gvr": "rbac.authorization.k8s.io/v1/clusterroles",
					"level": 1,
					"message": "[POP-400] Used? Unable to locate resource reference"
				}
				],
				"system:metrics-server-aggregated-reader": [
				{
					"group": "__root__",
					"gvr": "rbac.authorization.k8s.io/v1/clusterroles",
					"level": 1,
					"message": "[POP-400] Used? Unable to locate resource reference"
				}
				],
				"system:node-bootstrapper": [
				{
					"group": "__root__",
					"gvr": "rbac.authorization.k8s.io/v1/clusterroles",
					"level": 1,
					"message": "[POP-400] Used? Unable to locate resource reference"
				}
				],
				"system:node-problem-detector": [
				{
					"group": "__root__",
					"gvr": "rbac.authorization.k8s.io/v1/clusterroles",
					"level": 1,
					"message": "[POP-400] Used? Unable to locate resource reference"
				}
				],
				"system:persistent-volume-provisioner": [
				{
					"group": "__root__",
					"gvr": "rbac.authorization.k8s.io/v1/clusterroles",
					"level": 1,
					"message": "[POP-400] Used? Unable to locate resource reference"
				}
				],
				"undistro-metrics-reader": [
				{
					"group": "__root__",
					"gvr": "rbac.authorization.k8s.io/v1/clusterroles",
					"level": 1,
					"message": "[POP-400] Used? Unable to locate resource reference"
				}
				],
				"view": [
				{
					"group": "__root__",
					"gvr": "rbac.authorization.k8s.io/v1/clusterroles",
					"level": 1,
					"message": "[POP-400] Used? Unable to locate resource reference"
				}
				]
			}
		}
		]
	}
}

	`

	TestReport2 = `
	{
		"popeye": {
			"sanitizers": [
			{
				"sanitizer": "clusterroles",
				"issues": {
					"system:node": [],
					"system:node-bootstrapper": [
					{
						"gvr": "rbac.authorization.k8s.io/v1/clusterroles",
						"level": 1,
						"message": "[POP-400] Used? Unable to locate resource reference"
					}
					],
					"undistro-metrics-reader": [
					{
						"gvr": "rbac.authorization.k8s.io/v1/clusterroles",
						"level": 1,
						"message": "[POP-400] Used? Unable to locate resource reference"
					}
					]
				}
			},
			{
				"sanitizer": "daemonsets",
				"issues": {
					"kube-system/aws-node": [
					{
						"gvr": "containers",
						"level": 2,
						"message": "[POP-106] No resources requests/limits defined"
					},
					{
						"gvr": "containers",
						"level": 2,
						"message": "[POP-107] No resource limits defined"
					}
					],
					"kube-system/kube-proxy": [
					{
						"gvr": "containers",
						"level": 2,
						"message": "[POP-107] No resource limits defined"
					}
					]
				}
			},
			{
				"sanitizer": "deployments",
				"issues": {
					"cert-manager/cert-manager": [
					{
						"gvr": "containers",
						"level": 2,
						"message": "[POP-106] No resources requests/limits defined"
					},
					{
						"gvr": "containers",
						"level": 1,
						"message": "[POP-108] Unnamed port 9402"
					}
					]
				}
			}
			]
		}
	}
	`
	TestReport3 = "{[]}"
	TestReport4 = ""
)
