package discovery

const (
	RegionLabel     = "topology.kubernetes.io/region"
	MasterNodeLabel = "node-role.kubernetes.io/master"
)

var ClusterSourcePrefixes = map[string]string{
	"cloud.google.com/gke":   "gcp",
	"eks.amazonaws.com/":     "aws",
	"kubernetes.azure.com/":  "azure",
	"doks.digitalocean.com/": "digitalocean",
	"oke.oraclecloud.com/":   "oraclecloud",
}
