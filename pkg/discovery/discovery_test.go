package discovery

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestNodeResources(t *testing.T) {
	type args struct {
		nodes       []corev1.Node
		nodeMetrics []v1beta1.NodeMetrics
	}
	tests := []struct {
		name string
		args args
		want []NodeInfo
	}{
		{
			name: "OK",
			args: args{
				nodes: []corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "node1",
							Labels: map[string]string{"foo": "bar"},
						},
						Status: corev1.NodeStatus{
							Allocatable: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("2000m"),
								corev1.ResourceMemory: resource.MustParse("4Gi"),
							},
							Conditions: []corev1.NodeCondition{
								{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "node2",
							Labels: map[string]string{"foo": "foo"},
						},
						Status: corev1.NodeStatus{
							Allocatable: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("2000m"),
								corev1.ResourceMemory: resource.MustParse("4Gi"),
							},
							Conditions: []corev1.NodeCondition{
								{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
							},
						},
					},
				},
				nodeMetrics: []v1beta1.NodeMetrics{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "node1",
							Labels: map[string]string{"foo": "bar"},
						},
						Usage: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    resource.MustParse("1000m"),
							corev1.ResourceMemory: resource.MustParse("2Gi"),
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "node2",
							Labels: map[string]string{"foo": "foo"},
						},
						Usage: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    resource.MustParse("1000m"),
							corev1.ResourceMemory: resource.MustParse("2Gi"),
						},
					},
				},
			},
			want: []NodeInfo{
				{
					Name:   "node1",
					Labels: map[string]string{"foo": "bar"},
					Ready:  true,
					Resources: map[corev1.ResourceName]Resources{
						corev1.ResourceCPU: {
							Available:       resource.MustParse("2000m"),
							Usage:           resource.MustParse("1000m"),
							UsagePercentage: 50,
						},
						corev1.ResourceMemory: {
							Available:       resource.MustParse("4Gi"),
							Usage:           resource.MustParse("2Gi"),
							UsagePercentage: 50,
						},
					},
				},
				{
					Name:   "node2",
					Labels: map[string]string{"foo": "foo"},
					Ready:  true,
					Resources: map[corev1.ResourceName]Resources{
						corev1.ResourceCPU: {
							Available:       resource.MustParse("2000m"),
							Usage:           resource.MustParse("1000m"),
							UsagePercentage: 50,
						},
						corev1.ResourceMemory: {
							Available:       resource.MustParse("4Gi"),
							Usage:           resource.MustParse("2Gi"),
							UsagePercentage: 50,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nodeResources(tt.args.nodes, tt.args.nodeMetrics); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nodeResources() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSumNodeResources(t *testing.T) {
	type args struct {
		nodes []NodeInfo
	}
	tests := []struct {
		name string
		args args
		want map[corev1.ResourceName]Resources
	}{
		{
			name: "OK",
			args: args{
				nodes: []NodeInfo{
					{
						Name:   "node1",
						Labels: map[string]string{"foo": "bar"},
						Ready:  true,
						Resources: map[corev1.ResourceName]Resources{
							corev1.ResourceCPU: {
								Available:       resource.MustParse("2000m"),
								Usage:           resource.MustParse("1000m"),
								UsagePercentage: 50,
							},
							corev1.ResourceMemory: {
								Available:       resource.MustParse("4Gi"),
								Usage:           resource.MustParse("2Gi"),
								UsagePercentage: 50,
							},
						},
					},
					{
						Name:   "node2",
						Labels: map[string]string{"foo": "foo"},
						Ready:  true,
						Resources: map[corev1.ResourceName]Resources{
							corev1.ResourceCPU: {
								Available:       resource.MustParse("2000m"),
								Usage:           resource.MustParse("1000m"),
								UsagePercentage: 50,
							},
							corev1.ResourceMemory: {
								Available:       resource.MustParse("4Gi"),
								Usage:           resource.MustParse("2Gi"),
								UsagePercentage: 50,
							},
						},
					},
				},
			},
			want: map[corev1.ResourceName]Resources{
				corev1.ResourceCPU: {
					Available:       resource.MustParse("4000m"),
					Usage:           resource.MustParse("2000m"),
					UsagePercentage: 50,
				},
				corev1.ResourceMemory: {
					Available:       resource.MustParse("8Gi"),
					Usage:           resource.MustParse("4Gi"),
					UsagePercentage: 50,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sumNodeResources(tt.args.nodes); !resourcesAreEqual(got, tt.want) {
				t.Errorf("sumNodeResources() = %v, want %v", got, tt.want)
			}
		})
	}
}

func resourcesAreEqual(x, y map[corev1.ResourceName]Resources) bool {
	if x == nil || y == nil {
		return x == nil && y == nil
	}
	for name, res := range x {
		if !y[name].Available.Equal(res.Available) {
			return false
		}
		if !y[name].Usage.Equal(res.Usage) {
			return false
		}
		if y[name].UsagePercentage != res.UsagePercentage {
			return false
		}
	}

	return true
}

func TestProvider(t *testing.T) {
	type args struct {
		nodes []corev1.Node
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "eks",
			args: args{[]corev1.Node{{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"beta.kubernetes.io/arch":                  "amd64",
						"beta.kubernetes.io/instance-type":         "t3a.medium",
						"beta.kubernetes.io/os":                    "linux",
						"eks.amazonaws.com/capacityType":           "ON_DEMAND",
						"eks.amazonaws.com/nodegroup-image":        "ami-0e1b6f116a3733fef",
						"eks.amazonaws.com/nodegroup":              "cluster-mp-0",
						"failure-domain.beta.kubernetes.io/region": "us-east-1",
						"failure-domain.beta.kubernetes.io/zone":   "us-east-1a",
						"kubernetes.io/arch":                       "amd64",
						"kubernetes.io/hostname":                   "ip-10-0-107-23.ec2.internal",
						"kubernetes.io/os":                         "linux",
						"node.kubernetes.io/instance-type":         "t3a.medium",
						"topology.kubernetes.io/region":            "us-east-1",
						"topology.kubernetes.io/zone":              "us-east-1a",
					},
				},
			}}},
			want: "aws",
		},
		{
			name: "gke",
			args: args{[]corev1.Node{{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"beta.kubernetes.io/arch":                  "amd64",
						"beta.kubernetes.io/instance-type":         "g1-small",
						"beta.kubernetes.io/os":                    "linux",
						"cloud.google.com/gke-boot-disk":           "pd-standard",
						"cloud.google.com/gke-container-runtime":   "containerd",
						"cloud.google.com/gke-nodepool":            "default-pool",
						"cloud.google.com/gke-os-distribution":     "cos",
						"cloud.google.com/machine-family":          "g1",
						"failure-domain.beta.kubernetes.io/region": "us-central1",
						"failure-domain.beta.kubernetes.io/zone":   "us-central1-c",
						"kubernetes.io/arch":                       "amd64",
						"kubernetes.io/hostname":                   "gke-cluster-default-pool-7d29f8fc-wkg5",
						"kubernetes.io/os":                         "linux",
						"node.kubernetes.io/instance-type":         "g1-small",
						"topology.kubernetes.io/region":            "us-central1",
						"topology.kubernetes.io/zone":              "us-central1-c",
					},
				},
			}}},
			want: "gcp",
		},
		{
			name: "aks",
			args: args{[]corev1.Node{{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"agentpool":                                "default",
						"beta.kubernetes.io/arch":                  "amd64",
						"beta.kubernetes.io/instance-type":         "Standard_B2s",
						"beta.kubernetes.io/os":                    "linux",
						"failure-domain.beta.kubernetes.io/region": "australiaeast",
						"failure-domain.beta.kubernetes.io/zone":   "0",
						"kubernetes.azure.com/agentpool":           "default",
						"kubernetes.azure.com/cluster":             "MC_azure-k8stest_k8stest_australiaeast",
						"kubernetes.azure.com/mode":                "system",
						"kubernetes.azure.com/node-image-version":  "AKSUbuntu-1804gen2containerd-2022.04.13",
						"kubernetes.azure.com/os-sku":              "Ubuntu",
						"kubernetes.azure.com/role":                "agent",
						"kubernetes.azure.com/storageprofile":      "managed",
						"kubernetes.azure.com/storagetier":         "Premium_LRS",
						"kubernetes.io/arch":                       "amd64",
						"kubernetes.io/hostname":                   "aks-default-32454763-vmss000000",
						"kubernetes.io/os":                         "linux",
						"kubernetes.io/role":                       "agent",
						"node-role.kubernetes.io/agent":            " ",
						"node.kubernetes.io/instance-type":         "Standard_B2s",
						"storageprofile":                           "managed",
						"storagetier":                              "Premium_LRS",
						"topology.disk.csi.azure.com/zone":         " ",
						"topology.kubernetes.io/region":            "australiaeast",
						"topology.kubernetes.io/zone":              "0",
					},
				},
			}}},
			want: "azure",
		},
		{
			name: "doks",
			args: args{[]corev1.Node{{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"beta.kubernetes.io/arch":                  "amd64",
						"beta.kubernetes.io/instance-type":         "s-2vcpu-2gb",
						"beta.kubernetes.io/os":                    "linux",
						"doks.digitalocean.com/node-id":            "1cf0f149-1502-4fbf-83e5-7d454f474ee5",
						"doks.digitalocean.com/node-pool-id":       "843e43f9-8ef9-48fb-9806-a5a9e880a7d2",
						"doks.digitalocean.com/node-pool":          "default",
						"doks.digitalocean.com/version":            "1.22.8-do.1",
						"failure-domain.beta.kubernetes.io/region": "nyc1",
						"kubernetes.io/arch":                       "amd64",
						"kubernetes.io/hostname":                   "default-cj7j3",
						"kubernetes.io/os":                         "linux",
						"node.kubernetes.io/instance-type":         "s-2vcpu-2gb",
						"region":                                   "nyc1",
						"topology.kubernetes.io/region":            "nyc1",
					},
				},
			}}},
			want: "digitalocean",
		},
		{
			name: "oke",
			args: args{[]corev1.Node{{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"beta.kubernetes.io/arch":                      "amd64",
						"beta.kubernetes.io/instance-type":             "VM.Standard.E3.Flex",
						"beta.kubernetes.io/os":                        "linux",
						"displayName":                                  "oke-c7ayj7ifseq-nker2wfokza-sxq6xdxqcra-0",
						"failure-domain.beta.kubernetes.io/region":     "sa-vinhedo-1",
						"failure-domain.beta.kubernetes.io/zone":       "SA-VINHEDO-1-AD-1",
						"hostname":                                     "oke-c7ayj7ifseq-nker2wfokza-sxq6xdxqcra-0",
						"internal_addr":                                "10.0.10.130",
						"kubernetes.io/arch":                           "amd64",
						"kubernetes.io/hostname":                       "10.0.10.130",
						"kubernetes.io/os":                             "linux",
						"last-migration-failure":                       "get_kubesvc_failure",
						"name":                                         "cluster1",
						"node-role.kubernetes.io/node":                 "",
						"node.info.ds_proxymux_client":                 "true",
						"node.info/compartment.id_prefix":              "ocid1.tenancy.oc1",
						"node.info/compartment.id_suffix":              "aaaaaaaa2pajrajctgxzjwujd24bz6zvdytxnbzmer5ner4r4uwellipwx6q",
						"node.info/compartment.name":                   "mfariam",
						"node.info/kubeletVersion":                     "v1.22",
						"node.kubernetes.io/instance-type":             "VM.Standard.E3.Flex",
						"oci.oraclecloud.com/fault-domain":             "FAULT-DOMAIN-1",
						"oke.oraclecloud.com/node.info.private_subnet": "true",
						"oke.oraclecloud.com/node.info.private_worker": "true",
						"oke.oraclecloud.com/tenant_agent.version":     "1.42.6-bae5d92f49-820",
						"topology.kubernetes.io/region":                "sa-vinhedo-1",
						"topology.kubernetes.io/zone":                  "SA-VINHEDO-1-AD-1",
					},
				},
			}}},
			want: "oraclecloud",
		},
		{
			name: "kind",
			args: args{[]corev1.Node{{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"beta.kubernetes.io/arch":                                 "amd64",
						"beta.kubernetes.io/os":                                   "linux",
						"kubernetes.io/arch":                                      "amd64",
						"kubernetes.io/hostname":                                  "kind-control-plane",
						"kubernetes.io/os":                                        "linux",
						"node-role.kubernetes.io/control-plane":                   "",
						"node-role.kubernetes.io/master":                          "",
						"node.kubernetes.io/exclude-from-external-load-balancers": "",
					},
				},
			}}},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &clusterDiscovery{}
			if got := r.provider(tt.args.nodes); got != tt.want {
				t.Errorf("Provider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegion(t *testing.T) {
	type args struct {
		nodes []corev1.Node
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "OK",
			args: args{nodes: []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "node1",
						Labels: map[string]string{},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node2",
						Labels: map[string]string{
							RegionLabel: "us-east-1",
						},
					},
				},
			}},
			want: "us-east-1",
		},
		{
			name: "multi-region",
			args: args{nodes: []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "node1",
						Labels: map[string]string{},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "node2",
						Labels: map[string]string{},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node3",
					},
				},
			}},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &clusterDiscovery{}
			if got := r.region(tt.args.nodes); got != tt.want {
				t.Errorf("Region() got = %v, want %v", got, tt.want)
			}
		})
	}
}
