package discovery

const RegionLabel = "topology.kubernetes.io/region"

var ClusterSourcePrefixes = map[string]map[string]string{
	"cloud.google.com/gke": map[string]string{
		"provider": "gcp", "flavor": "gke",
	},
	"eks.amazon.com/": map[string]string{
		"provider": "aws", "flavor": "eks",
	},
}
