package payloads

import (
	"github.com/getupio-undistro/zora/apis/zora/v1alpha1"
)

type Issue struct {
	ID       string             `json:"id"`
	Message  string             `json:"message"`
	Severity string             `json:"severity"`
	Category string             `json:"category"`
	Plugin   string             `json:"plugin"`
	Clusters []ClusterReference `json:"clusters"`
	Url      string             `json:"url"`
	Original *Original          `json:"original"`
}

type Original struct {
	Message  *string `json:"message"`
	Severity *string `json:"severity"`
	Category *string `json:"category"`
}

type ClusterReference struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
	TotalResources int    `json:"totalResources"`
}

func NewIssue(clusterIssue v1alpha1.ClusterIssue) Issue {
	i := Issue{
		ID:       clusterIssue.Spec.ID,
		Message:  clusterIssue.Spec.Message,
		Severity: string(clusterIssue.Spec.Severity),
		Category: clusterIssue.Spec.Category,
		Plugin:   clusterIssue.Labels[v1alpha1.LabelPlugin],
		Url:      clusterIssue.Spec.Url,
	}
	if clusterIssue.HasOverride() {
		i.Original = &Original{
			Message:  clusterIssue.Status.OrigMessage,
			Category: clusterIssue.Status.OrigCategory,
		}
		if clusterIssue.Status.OrigSeverity != nil {
			s := string(*clusterIssue.Status.OrigSeverity)
			i.Original.Severity = &s
		}
	}
	return i
}

func NewIssues(clusterIssues []v1alpha1.ClusterIssue) []Issue {
	issuesByID := make(map[string]*Issue)
	clustersByIssue := make(map[string]map[string]*ClusterReference)
	for _, clusterIssue := range clusterIssues {
		if clusterIssue.Status.Hidden {
			continue
		}

		clusterRef := &ClusterReference{
			Name:           clusterIssue.Spec.Cluster,
			Namespace:      clusterIssue.Namespace,
			TotalResources: clusterIssue.Spec.TotalResources,
		}
		newIssue := NewIssue(clusterIssue)
		issueID := clusterIssue.Spec.ID
		issuesByID[issueID] = &newIssue
		if _, ok := clustersByIssue[issueID]; !ok {
			clustersByIssue[issueID] = make(map[string]*ClusterReference)
		}
		clustersByIssue[issueID][clusterRef.Name] = clusterRef
	}
	res := make([]Issue, 0, len(issuesByID))
	for id, i := range issuesByID {
		for _, ref := range clustersByIssue[id] {
			i.Clusters = append(i.Clusters, *ref)
		}

		res = append(res, *i)
	}
	return res
}
