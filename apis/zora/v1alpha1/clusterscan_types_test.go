package v1alpha1

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSyncStatus(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus *ClusterScanStatus
		plugins       map[string]*PluginScanStatus
		want          *ClusterScanStatus
	}{
		{
			name: "complete + complete",
			plugins: map[string]*PluginScanStatus{
				"popeye": {
					LastScheduleTime:     mustParseTime("2022-08-08T21:00:00Z"),
					LastFinishedTime:     mustParseTime("2022-08-08T21:00:06Z"),
					LastSuccessfulTime:   mustParseTime("2022-08-08T21:00:06Z"),
					NextScheduleTime:     mustParseTime("2022-08-08T22:00:00Z"),
					LastScanID:           "9da315be-b5a1-4f1a-952b-915cc19fe446",
					LastSuccessfulScanID: "9da315be-b5a1-4f1a-952b-915cc19fe446",
					LastStatus:           string(batchv1.JobComplete),
					LastFinishedStatus:   string(batchv1.JobComplete),
				},
				"kubescape": {
					LastScheduleTime:     mustParseTime("2022-08-08T21:00:00Z"),
					LastFinishedTime:     mustParseTime("2022-08-08T21:00:03Z"),
					LastSuccessfulTime:   mustParseTime("2022-08-08T21:00:03Z"),
					NextScheduleTime:     mustParseTime("2022-08-08T22:00:00Z"),
					LastScanID:           "ce34e6fc-768d-49d0-91b5-65df89ed147d",
					LastSuccessfulScanID: "ce34e6fc-768d-49d0-91b5-65df89ed147d",
					LastStatus:           string(batchv1.JobComplete),
					LastFinishedStatus:   string(batchv1.JobComplete),
				},
			},
			want: &ClusterScanStatus{
				LastScheduleTime:   mustParseTime("2022-08-08T21:00:00Z"),
				LastFinishedTime:   mustParseTime("2022-08-08T21:00:06Z"),
				LastSuccessfulTime: mustParseTime("2022-08-08T21:00:06Z"),
				NextScheduleTime:   mustParseTime("2022-08-08T22:00:00Z"),
				LastStatus:         string(batchv1.JobComplete),
				LastFinishedStatus: string(batchv1.JobComplete),
				PluginNames:        "kubescape,popeye",
			},
		},
		{
			name: "complete + active",
			plugins: map[string]*PluginScanStatus{
				"popeye": {
					LastFinishedTime:     mustParseTime("2022-08-08T20:00:06Z"),
					LastSuccessfulTime:   mustParseTime("2022-08-08T20:00:06Z"),
					LastScheduleTime:     mustParseTime("2022-08-08T21:00:00Z"),
					LastStatus:           "Active",
					NextScheduleTime:     mustParseTime("2022-08-08T22:00:00Z"),
					LastScanID:           "9da315be-b5a1-4f1a-952b-915cc19fe446",
					LastSuccessfulScanID: "9da315be-b5a1-4f1a-952b-915cc19fe446",
					LastFinishedStatus:   string(batchv1.JobComplete),
				},
				"kubescape": {
					LastScheduleTime:     mustParseTime("2022-08-08T21:00:00Z"),
					LastFinishedTime:     mustParseTime("2022-08-08T21:00:03Z"),
					LastSuccessfulTime:   mustParseTime("2022-08-08T21:00:03Z"),
					NextScheduleTime:     mustParseTime("2022-08-08T22:00:00Z"),
					LastScanID:           "ce34e6fc-768d-49d0-91b5-65df89ed147d",
					LastSuccessfulScanID: "ce34e6fc-768d-49d0-91b5-65df89ed147d",
					LastStatus:           string(batchv1.JobComplete),
					LastFinishedStatus:   string(batchv1.JobComplete),
				},
			},
			want: &ClusterScanStatus{
				LastScheduleTime:   mustParseTime("2022-08-08T21:00:00Z"),
				LastFinishedTime:   mustParseTime("2022-08-08T21:00:03Z"),
				LastSuccessfulTime: mustParseTime("2022-08-08T21:00:03Z"),
				NextScheduleTime:   mustParseTime("2022-08-08T22:00:00Z"),
				LastStatus:         "Active",
				LastFinishedStatus: string(batchv1.JobComplete),
				PluginNames:        "kubescape,popeye",
			},
		},
		{
			name: "active + active",
			plugins: map[string]*PluginScanStatus{
				"popeye": {
					LastFinishedTime:     mustParseTime("2022-08-08T20:00:06Z"),
					LastSuccessfulTime:   mustParseTime("2022-08-08T20:00:06Z"),
					LastScheduleTime:     mustParseTime("2022-08-08T21:00:00Z"),
					LastStatus:           "Active",
					NextScheduleTime:     mustParseTime("2022-08-08T22:00:00Z"),
					LastScanID:           "9da315be-b5a1-4f1a-952b-915cc19fe446",
					LastSuccessfulScanID: "9da315be-b5a1-4f1a-952b-915cc19fe446",
					LastFinishedStatus:   string(batchv1.JobComplete),
				},
				"kubescape": {
					LastScheduleTime:     mustParseTime("2022-08-08T21:00:00Z"),
					LastFinishedTime:     mustParseTime("2022-08-08T20:00:03Z"),
					LastSuccessfulTime:   mustParseTime("2022-08-08T20:00:03Z"),
					NextScheduleTime:     mustParseTime("2022-08-08T22:00:00Z"),
					LastScanID:           "ce34e6fc-768d-49d0-91b5-65df89ed147d",
					LastSuccessfulScanID: "ce34e6fc-768d-49d0-91b5-65df89ed147d",
					LastStatus:           "Active",
					LastFinishedStatus:   string(batchv1.JobComplete),
				},
			},
			want: &ClusterScanStatus{
				LastScheduleTime:   mustParseTime("2022-08-08T21:00:00Z"),
				LastFinishedTime:   mustParseTime("2022-08-08T20:00:06Z"),
				LastSuccessfulTime: mustParseTime("2022-08-08T20:00:06Z"),
				NextScheduleTime:   mustParseTime("2022-08-08T22:00:00Z"),
				LastStatus:         "Active",
				LastFinishedStatus: string(batchv1.JobComplete),
				PluginNames:        "kubescape,popeye",
			},
		},
		{
			name: "active 1st + active 1st",
			plugins: map[string]*PluginScanStatus{
				"popeye": {
					LastScheduleTime: mustParseTime("2022-08-08T21:00:00Z"),
					LastStatus:       "Active",
					NextScheduleTime: mustParseTime("2022-08-08T22:00:00Z"),
					LastScanID:       "9da315be-b5a1-4f1a-952b-915cc19fe446",
				},
				"kubescape": {
					LastScheduleTime: mustParseTime("2022-08-08T21:00:00Z"),
					NextScheduleTime: mustParseTime("2022-08-08T22:00:00Z"),
					LastScanID:       "ce34e6fc-768d-49d0-91b5-65df89ed147d",
					LastStatus:       "Active",
				},
			},
			want: &ClusterScanStatus{
				LastScheduleTime: mustParseTime("2022-08-08T21:00:00Z"),
				NextScheduleTime: mustParseTime("2022-08-08T22:00:00Z"),
				LastStatus:       "Active",
				PluginNames:      "kubescape,popeye",
			},
		},
		{
			name: "active + active and never successful",
			plugins: map[string]*PluginScanStatus{
				"popeye": {
					LastFinishedTime:     mustParseTime("2022-08-08T20:00:06Z"),
					LastSuccessfulTime:   mustParseTime("2022-08-08T20:00:06Z"),
					LastScheduleTime:     mustParseTime("2022-08-08T21:00:00Z"),
					LastStatus:           "Active",
					NextScheduleTime:     mustParseTime("2022-08-08T22:00:00Z"),
					LastScanID:           "9da315be-b5a1-4f1a-952b-915cc19fe446",
					LastSuccessfulScanID: "9da315be-b5a1-4f1a-952b-915cc19fe446",
					LastFinishedStatus:   string(batchv1.JobComplete),
				},
				"kubescape": {
					LastScheduleTime:   mustParseTime("2022-08-08T21:00:00Z"),
					LastFinishedTime:   mustParseTime("2022-08-08T20:00:03Z"),
					NextScheduleTime:   mustParseTime("2022-08-08T22:00:00Z"),
					LastScanID:         "ce34e6fc-768d-49d0-91b5-65df89ed147d",
					LastStatus:         "Active",
					LastFinishedStatus: string(batchv1.JobFailed),
					LastErrorMsg:       "failed connecting to Kubernetes cluster",
				},
			},
			want: &ClusterScanStatus{
				LastScheduleTime:   mustParseTime("2022-08-08T21:00:00Z"),
				LastFinishedTime:   mustParseTime("2022-08-08T20:00:06Z"),
				LastSuccessfulTime: mustParseTime("2022-08-08T20:00:06Z"),
				NextScheduleTime:   mustParseTime("2022-08-08T22:00:00Z"),
				LastStatus:         "Active",
				LastFinishedStatus: string(batchv1.JobFailed),
				PluginNames:        "kubescape,popeye",
			},
		},
		{
			name: "active and never successful + active and never successful",
			plugins: map[string]*PluginScanStatus{
				"popeye": {
					LastScheduleTime:   mustParseTime("2022-08-08T21:00:00Z"),
					LastFinishedTime:   mustParseTime("2022-08-08T20:00:06Z"),
					LastStatus:         "Active",
					LastFinishedStatus: string(batchv1.JobFailed),
					NextScheduleTime:   mustParseTime("2022-08-08T22:00:00Z"),
					LastScanID:         "9da315be-b5a1-4f1a-952b-915cc19fe446",
					LastErrorMsg:       `Get "http://localhost:8081/version?timeout=30s": dial tcp 127.0.0.1:8081: connect: connection refused`,
				},
				"kubescape": {
					LastScheduleTime:   mustParseTime("2022-08-08T21:00:00Z"),
					LastFinishedTime:   mustParseTime("2022-08-08T20:00:03Z"),
					LastStatus:         "Active",
					LastFinishedStatus: string(batchv1.JobFailed),
					NextScheduleTime:   mustParseTime("2022-08-08T22:00:00Z"),
					LastScanID:         "ce34e6fc-768d-49d0-91b5-65df89ed147d",
					LastErrorMsg:       "failed connecting to Kubernetes cluster",
				},
			},
			want: &ClusterScanStatus{
				LastScheduleTime:   mustParseTime("2022-08-08T21:00:00Z"),
				LastFinishedTime:   mustParseTime("2022-08-08T20:00:06Z"),
				NextScheduleTime:   mustParseTime("2022-08-08T22:00:00Z"),
				LastStatus:         "Active",
				LastFinishedStatus: string(batchv1.JobFailed),
				PluginNames:        "kubescape,popeye",
			},
		},
		{
			name: "complete + active and never successful",
			plugins: map[string]*PluginScanStatus{
				"popeye": {
					LastScheduleTime:     mustParseTime("2022-08-08T21:00:00Z"),
					LastFinishedTime:     mustParseTime("2022-08-08T21:00:06Z"),
					LastSuccessfulTime:   mustParseTime("2022-08-08T21:00:06Z"),
					NextScheduleTime:     mustParseTime("2022-08-08T22:00:00Z"),
					LastScanID:           "9da315be-b5a1-4f1a-952b-915cc19fe446",
					LastSuccessfulScanID: "9da315be-b5a1-4f1a-952b-915cc19fe446",
					LastStatus:           string(batchv1.JobComplete),
					LastFinishedStatus:   string(batchv1.JobComplete),
				},
				"kubescape": {
					LastScheduleTime:   mustParseTime("2022-08-08T21:00:00Z"),
					LastFinishedTime:   mustParseTime("2022-08-08T20:00:03Z"),
					NextScheduleTime:   mustParseTime("2022-08-08T22:00:00Z"),
					LastScanID:         "ce34e6fc-768d-49d0-91b5-65df89ed147d",
					LastStatus:         "Active",
					LastFinishedStatus: string(batchv1.JobFailed),
					LastErrorMsg:       "failed connecting to Kubernetes cluster",
				},
			},
			want: &ClusterScanStatus{
				LastScheduleTime:   mustParseTime("2022-08-08T21:00:00Z"),
				LastFinishedTime:   mustParseTime("2022-08-08T21:00:06Z"),
				LastSuccessfulTime: mustParseTime("2022-08-08T21:00:06Z"),
				NextScheduleTime:   mustParseTime("2022-08-08T22:00:00Z"),
				LastStatus:         "Active",
				LastFinishedStatus: string(batchv1.JobFailed),
				PluginNames:        "kubescape,popeye",
			},
		},
		{
			name: "complete + active 1st",
			plugins: map[string]*PluginScanStatus{
				"popeye": {
					LastScheduleTime: mustParseTime("2022-08-08T21:00:00Z"),
					LastStatus:       "Active",
					NextScheduleTime: mustParseTime("2022-08-08T22:00:00Z"),
					LastScanID:       "9da315be-b5a1-4f1a-952b-915cc19fe446",
				},
				"kubescape": {
					LastScheduleTime:     mustParseTime("2022-08-08T21:00:00Z"),
					LastFinishedTime:     mustParseTime("2022-08-08T21:00:03Z"),
					LastSuccessfulTime:   mustParseTime("2022-08-08T21:00:03Z"),
					NextScheduleTime:     mustParseTime("2022-08-08T22:00:00Z"),
					LastScanID:           "ce34e6fc-768d-49d0-91b5-65df89ed147d",
					LastSuccessfulScanID: "ce34e6fc-768d-49d0-91b5-65df89ed147d",
					LastStatus:           string(batchv1.JobComplete),
					LastFinishedStatus:   string(batchv1.JobComplete),
				},
			},
			want: &ClusterScanStatus{
				LastScheduleTime:   mustParseTime("2022-08-08T21:00:00Z"),
				LastFinishedTime:   mustParseTime("2022-08-08T21:00:03Z"),
				LastSuccessfulTime: mustParseTime("2022-08-08T21:00:03Z"),
				NextScheduleTime:   mustParseTime("2022-08-08T22:00:00Z"),
				LastStatus:         "Active",
				LastFinishedStatus: string(batchv1.JobComplete),
				PluginNames:        "kubescape,popeye",
			},
		},
		{
			name: "complete + always failed",
			plugins: map[string]*PluginScanStatus{
				"popeye": {
					LastScheduleTime:     mustParseTime("2022-08-08T21:00:00Z"),
					LastFinishedTime:     mustParseTime("2022-08-08T21:00:03Z"),
					LastSuccessfulTime:   mustParseTime("2022-08-08T21:00:03Z"),
					NextScheduleTime:     mustParseTime("2022-08-08T22:00:00Z"),
					LastScanID:           "9da315be-b5a1-4f1a-952b-915cc19fe446",
					LastSuccessfulScanID: "9da315be-b5a1-4f1a-952b-915cc19fe446",
					LastStatus:           string(batchv1.JobComplete),
					LastFinishedStatus:   string(batchv1.JobComplete),
				},
				"kubescape": {
					LastScheduleTime:   mustParseTime("2022-08-08T21:00:00Z"),
					LastFinishedTime:   mustParseTime("2022-08-08T21:00:06Z"),
					NextScheduleTime:   mustParseTime("2022-08-08T22:00:00Z"),
					LastScanID:         "ce34e6fc-768d-49d0-91b5-65df89ed147d",
					LastStatus:         string(batchv1.JobFailed),
					LastFinishedStatus: string(batchv1.JobFailed),
					LastErrorMsg:       `failed to discover API server information. error: Get "https://15.446.40.219/version?timeout=39s": x509: certificate signed by unknown authority`,
				},
			},
			want: &ClusterScanStatus{
				LastScheduleTime:   mustParseTime("2022-08-08T21:00:00Z"),
				LastFinishedTime:   mustParseTime("2022-08-08T21:00:06Z"),
				LastSuccessfulTime: mustParseTime("2022-08-08T21:00:03Z"),
				NextScheduleTime:   mustParseTime("2022-08-08T22:00:00Z"),
				LastFinishedStatus: string(batchv1.JobFailed),
				LastStatus:         string(batchv1.JobFailed),
				PluginNames:        "kubescape,popeye",
			},
		},
		{
			name: "failed + always failed",
			plugins: map[string]*PluginScanStatus{
				"popeye": {
					LastScheduleTime:     mustParseTime("2022-08-08T21:00:00Z"),
					LastFinishedTime:     mustParseTime("2022-08-08T21:00:03Z"),
					NextScheduleTime:     mustParseTime("2022-08-08T22:00:00Z"),
					LastScanID:           "9da315be-b5a1-4f1a-952b-915cc19fe446",
					LastSuccessfulTime:   mustParseTime("2022-08-08T20:00:03Z"),
					LastSuccessfulScanID: "ab8b751d-a0ab-40ac-9980-0cd2133a43f8",
					LastStatus:           string(batchv1.JobFailed),
					LastFinishedStatus:   string(batchv1.JobFailed),
					LastErrorMsg:         "the server has asked for the client to provide credentials",
				},
				"kubescape": {
					LastScheduleTime:   mustParseTime("2022-08-08T21:00:00Z"),
					LastFinishedTime:   mustParseTime("2022-08-08T21:00:06Z"),
					NextScheduleTime:   mustParseTime("2022-08-08T22:00:00Z"),
					LastScanID:         "ce34e6fc-768d-49d0-91b5-65df89ed147d",
					LastStatus:         string(batchv1.JobFailed),
					LastFinishedStatus: string(batchv1.JobFailed),
					LastErrorMsg:       `failed to discover API server information. error: Get "https://35.236.51.220/version?timeout=32s": x509: certificate signed by unknown authority`,
				},
			},
			want: &ClusterScanStatus{
				LastScheduleTime:   mustParseTime("2022-08-08T21:00:00Z"),
				LastFinishedTime:   mustParseTime("2022-08-08T21:00:06Z"),
				LastSuccessfulTime: mustParseTime("2022-08-08T20:00:03Z"),
				NextScheduleTime:   mustParseTime("2022-08-08T22:00:00Z"),
				LastFinishedStatus: string(batchv1.JobFailed),
				LastStatus:         string(batchv1.JobFailed),
				PluginNames:        "kubescape,popeye",
			},
		},
		{
			name: "always failed",
			currentStatus: &ClusterScanStatus{
				NextScheduleTime: mustParseTime("2022-08-12T12:00:00Z"),
			},
			plugins: map[string]*PluginScanStatus{
				"brutus": {
					LastScheduleTime:   mustParseTime("2022-08-12T13:00:00Z"),
					LastFinishedTime:   mustParseTime("2022-08-12T13:00:03Z"),
					NextScheduleTime:   mustParseTime("2022-08-12T14:00:00Z"),
					LastScanID:         "886938da-f1e5-438c-8ceb-be9dbd15c8e",
					LastStatus:         string(batchv1.JobFailed),
					LastFinishedStatus: string(batchv1.JobFailed),
					LastErrorMsg:       `Exec failed unknown flag: --Xforce-exit-zero`,
				},
			},
			want: &ClusterScanStatus{
				LastScheduleTime:   mustParseTime("2022-08-12T13:00:00Z"),
				LastFinishedTime:   mustParseTime("2022-08-12T13:00:03Z"),
				NextScheduleTime:   mustParseTime("2022-08-12T14:00:00Z"),
				LastFinishedStatus: string(batchv1.JobFailed),
				LastStatus:         string(batchv1.JobFailed),
				PluginNames:        "brutus",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.want.Plugins = tt.plugins
			css := &ClusterScanStatus{}
			if tt.currentStatus != nil {
				css = tt.currentStatus
			}
			css.Plugins = tt.plugins
			css.SyncStatus()
			if !reflect.DeepEqual(css, tt.want) {
				t.Errorf("SyncStatus() = %s", cmp.Diff(css, tt.want))
			}
		})
	}
}

func mustParseTime(v string) *metav1.Time {
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		panic(fmt.Sprintf("mustParseTime(%s): %s", v, err.Error()))
	}
	return &metav1.Time{Time: t}
}
