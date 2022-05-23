package payloads

import (
	"reflect"
	"testing"

	"github.com/getupio-undistro/inspect/apis/inspect/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewIssues(t *testing.T) {
	type args struct {
		issues []v1alpha1.ClusterIssue
	}
	tests := []struct {
		name string
		args args
		want []Issue
	}{
		{
			name: "OK",
			args: args{
				issues: []v1alpha1.ClusterIssue{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "prd1-pop-106",
							Namespace: "prd",
						},
						Spec: v1alpha1.ClusterIssueSpec{
							Cluster:        "prd1",
							ID:             "POP-106",
							Message:        "No resources requests/limits defined",
							Severity:       "Medium",
							Category:       "Category",
							TotalResources: 10,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "prd2-pop-777",
							Namespace: "prd",
						},
						Spec: v1alpha1.ClusterIssueSpec{
							Cluster:        "prd2",
							ID:             "POP-777",
							Message:        "Message",
							Severity:       "Medium",
							Category:       "Category",
							TotalResources: 7,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "hml1-pop-106",
							Namespace: "hml",
						},
						Spec: v1alpha1.ClusterIssueSpec{
							Cluster:        "hml1",
							ID:             "POP-106",
							Message:        "No resources requests/limits defined",
							Severity:       "Medium",
							Category:       "Category",
							TotalResources: 17,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "dev1-pop-106",
							Namespace: "dev",
						},
						Spec: v1alpha1.ClusterIssueSpec{
							Cluster:        "dev1",
							ID:             "POP-106",
							Message:        "No resources requests/limits defined",
							Severity:       "Medium",
							Category:       "Category",
							TotalResources: 71,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "dev2-pop-777",
							Namespace: "dev",
						},
						Spec: v1alpha1.ClusterIssueSpec{
							Cluster:        "dev2",
							ID:             "POP-777",
							Message:        "Message",
							Severity:       "Medium",
							Category:       "Category",
							TotalResources: 27,
						},
					},
				},
			},
			want: []Issue{
				{
					ID:       "POP-106",
					Message:  "No resources requests/limits defined",
					Severity: "Medium",
					Category: "Category",
					Clusters: []ClusterReference{
						{
							Name:           "prd1",
							Namespace:      "prd",
							TotalResources: 10,
						},
						{
							Name:           "hml1",
							Namespace:      "hml",
							TotalResources: 17,
						},
						{
							Name:           "dev1",
							Namespace:      "dev",
							TotalResources: 71,
						},
					},
				},
				{
					ID:       "POP-777",
					Message:  "Message",
					Severity: "Medium",
					Category: "Category",
					Clusters: []ClusterReference{
						{
							Name:           "prd2",
							Namespace:      "prd",
							TotalResources: 7,
						},
						{
							Name:           "dev2",
							Namespace:      "dev",
							TotalResources: 27,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewIssues(tt.args.issues); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewIssues() = %v, want %v", got, tt.want)
			}
		})
	}
}
