package popeye

import (
	"io/ioutil"
	"reflect"
	"sort"
	"testing"

	inspectv1a1 "github.com/getupio-undistro/inspect/apis/inspect/v1alpha1"
	"github.com/google/go-cmp/cmp"
)

func TestPrepareIdAndMsg(t *testing.T) {
	cases := []struct {
		description string
		popmsg      string
		id          string
		msg         string
		toerr       bool
	}{
		{
			description: "Popeye generic issue 113",
			popmsg:      "[POP-113] Container image fake_img:latest is not hosted on an allowed docker registry",
			id:          "POP-113",
			msg:         "Container image not hosted on an allowed docker registry",
			toerr:       false,
		},
		{
			description: "Popeye issue 400",
			popmsg:      "[POP-400] Used? Unable to locate resource reference",
			id:          "POP-400",
			msg:         "Used? Unable to locate resource reference",
			toerr:       false,
		},
		{
			description: "Popeye issue 800",
			popmsg:      "[POP-800] Namespace is inactive",
			id:          "POP-800",
			msg:         "Namespace is inactive",
			toerr:       false,
		},
		{
			description: "Popeye generic issue 1109",
			popmsg:      "[POP-1109] Only one Pod associated with this endpoint",
			id:          "POP-1109",
			msg:         "Only one Pod associated with this endpoint",
			toerr:       false,
		},
		{
			description: "Popeye generic issue 1200",
			popmsg:      "[POP-1200] Unhealthy ReplicaSet 5 desired but have 2 ready",
			id:          "POP-1200",
			msg:         "Unhealthy ReplicaSet",
			toerr:       false,
		},
		{
			description: "Popeye message code without dash",
			popmsg:      "[POP666] Fake Popeye message",
			toerr:       true,
		},
		{
			description: "Popeye message code without brackets",
			popmsg:      "POP-666 Super fake Popeye message",
			toerr:       true,
		},
		{
			description: "Popeye message without description",
			popmsg:      "[POP-666]",
			toerr:       true,
		},
	}

	for _, c := range cases {
		if id, msg, err := prepareIdAndMsg(c.popmsg); (id != c.id || msg != c.msg) && err != nil && !c.toerr {
			t.Errorf("Setup failed on case: %s\n", c.description)
			t.Error(err)
		}
	}

}

func TestParse(t *testing.T) {
	cases := []struct {
		description string
		testrepname string
		cispecs     []*inspectv1a1.ClusterIssueSpec
		toerr       bool
	}{
		{
			description: "Single <ClusterIssueSpec> instance with many resources",
			testrepname: "test_data/test_report_1.json",
			cispecs: []*inspectv1a1.ClusterIssueSpec{
				{
					ID:       "POP-400",
					Message:  "Used? Unable to locate resource reference",
					Severity: LevelToIssueSeverity[1],
					Category: "clusterroles",
					Resources: map[string][]string{
						"rbac.authorization.k8s.io/v1/clusterroles": {
							"capi-kubeadm-control-plane-manager-role",
							"cert-manager-edit",
							"system:certificates.k8s.io:kube-apiserver-client-kubelet-approver",
							"system:persistent-volume-provisioner",
							"undistro-metrics-reader",
							"cert-manager-view",
							"system:heapster",
							"system:kube-aggregator",
							"admin",
							"system:metrics-server-aggregated-reader",
							"system:node-bootstrapper",
							"system:node-problem-detector",
							"view",
							"capi-manager-role",
							"system:certificates.k8s.io:kubelet-serving-approver",
							"system:certificates.k8s.io:legacy-unknown-approver",
						},
					},
					TotalResources: 16,
				},
			},
			toerr: false,
		},

		{
			description: "Four <ClusterIssueSpec> instance with many resources",
			testrepname: "test_data/test_report_2.json",
			cispecs: []*inspectv1a1.ClusterIssueSpec{
				{
					ID:       "POP-400",
					Message:  "Used? Unable to locate resource reference",
					Severity: LevelToIssueSeverity[1],
					Category: "clusterroles",
					Resources: map[string][]string{
						"rbac.authorization.k8s.io/v1/clusterroles": {"system:node-bootstrapper", "undistro-metrics-reader"},
					},
					TotalResources: 2,
				},
				{
					ID:       "POP-106",
					Message:  "No resources requests/limits defined",
					Severity: LevelToIssueSeverity[2],
					Category: "daemonsets",
					Resources: map[string][]string{
						"containers": {"kube-system/aws-node", "cert-manager/cert-manager"},
					},
					TotalResources: 2,
				},
				{
					ID:       "POP-107",
					Message:  "No resource limits defined",
					Severity: LevelToIssueSeverity[2],
					Category: "daemonsets",
					Resources: map[string][]string{
						"containers": {"kube-system/aws-node", "kube-system/kube-proxy"},
					},
					TotalResources: 2,
				},
				{
					ID:       "POP-108",
					Message:  "Unnamed port",
					Severity: LevelToIssueSeverity[1],
					Category: "deployments",
					Resources: map[string][]string{
						"containers": {"cert-manager/cert-manager"},
					},
					TotalResources: 1,
				},
			},
			toerr: false,
		},

		{
			description: "Invalid Popeye report",
			testrepname: "test_data/test_report_3.json",
			cispecs:     nil,
			toerr:       true,
		},
		{
			description: "Empty Popeye report",
			testrepname: "test_data/test_report_4.json",
			cispecs:     nil,
			toerr:       true,
		},
	}

	sfun := func(cis []*inspectv1a1.ClusterIssueSpec) {
		sort.Slice(cis, func(i, j int) bool {
			return cis[i].ID > cis[j].ID
		})
		for c := 0; c < len(cis); c++ {
			for r, _ := range cis[c].Resources {
				sort.Strings(cis[c].Resources[r])
			}
		}
	}
	for _, c := range cases {
		rep, err := ioutil.ReadFile(c.testrepname)
		if err != nil {
			t.Errorf("Setup failed on case: %s\n", c.description)
			t.Fatal(err)
		}
		cispecs, err := Parse(rep)
		sfun(c.cispecs)
		sfun(cispecs)
		if (err != nil) != c.toerr || !reflect.DeepEqual(c.cispecs, cispecs) {
			t.Errorf("Case: %s\n", c.description)
			t.Errorf("Mismatch between expected and obtained values: \n%s\n", cmp.Diff(c.cispecs, cispecs))
		}
	}

}
