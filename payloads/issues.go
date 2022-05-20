package payloads

import "github.com/getupio-undistro/inspect/apis/inspect/v1alpha1"

type Issue struct {
	ID       string             `json:"id,omitempty"`
	Message  string             `json:"message,omitempty"`
	Severity string             `json:"severity,omitempty"`
	Category string             `json:"category,omitempty"`
	Clusters []ClusterReference `json:"clusters,omitempty"`
}

type ClusterReference struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
	TotalResources int    `json:"totalResources"`
}

func NewIssue(clusterIssue v1alpha1.ClusterIssue) Issue {
	return Issue{
		ID:       clusterIssue.Spec.ID,
		Message:  clusterIssue.Spec.Message,
		Severity: string(clusterIssue.Spec.Severity),
		Category: clusterIssue.Spec.Category,
	}
}

func NewIssues(issues []v1alpha1.ClusterIssue) []Issue {
	clustersByIssueID := make(map[string]*Issue)
	for _, issue := range issues {
		clusterRef := ClusterReference{
			Name:           issue.Spec.Cluster,
			Namespace:      issue.Namespace,
			TotalResources: issue.Spec.TotalResources,
		}
		i, ok := clustersByIssueID[issue.Spec.ID]
		if !ok {
			newIssue := NewIssue(issue)
			newIssue.Clusters = append(newIssue.Clusters, clusterRef)
			clustersByIssueID[issue.Spec.ID] = &newIssue
		} else {
			i.Clusters = append(i.Clusters, clusterRef)
		}
	}
	res := make([]Issue, 0, len(clustersByIssueID))
	for _, i := range clustersByIssueID {
		res = append(res, *i)
	}
	return res
}
