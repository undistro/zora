package discovery

const RegionLabel = "topology.kubernetes.io/region"

var ClusterSourcePrefixes = map[string]string{
	"cloud.google.com/gke": "gcp",
	"eks.amazonaws.com/":   "aws",
}
