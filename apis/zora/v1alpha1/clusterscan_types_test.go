// Copyright 2022 Undistro Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
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

func TestSplitStartTime(t *testing.T) {
	now := time.Now().UTC()
	cases := []struct {
		description string
		sch         *Schedule
		res         []int
	}{
		{
			description: "Start at 1 hour and 20 minutes",
			sch: &Schedule{
				StartTime: pointer.String("1:20"),
			},
			res: []int{1, 20},
		},
		{
			description: "Start at 1 hour and 13 minutes",
			sch: &Schedule{
				StartTime: pointer.String("5:13"),
			},
			res: []int{5, 13},
		},
		{
			description: "Start at 13 hour and 0 minutes",
			sch: &Schedule{
				StartTime: pointer.String("13:00"),
			},
			res: []int{13, 0},
		},
		{
			description: "Start at 18 hour and 44 minutes",
			sch: &Schedule{
				StartTime: pointer.String("18:44"),
			},
			res: []int{18, 44},
		},
		{
			description: "Start at 23 hour and 0 minutes",
			sch: &Schedule{
				StartTime: pointer.String("23:00"),
			},
			res: []int{23, 0},
		},
		{
			description: "Start at 23 hour and 7 minutes",
			sch: &Schedule{
				StartTime: pointer.String("23:07"),
			},
			res: []int{23, 7},
		},
		{
			description: "Empty start time",
			sch: &Schedule{
				StartTime: pointer.String(""),
			},
			res: []int{now.Hour(), now.Minute()},
		},
		{
			description: "Unset start time",
			sch:         &Schedule{},
			res:         []int{now.Hour(), now.Minute()},
		},
		{
			description: "Unset schedule",
			res:         []int{now.Hour(), now.Minute()},
		},
	}

	for _, c := range cases {
		if h, m := c.sch.SplitStartTime(); h != c.res[0] || m != c.res[1] {
			t.Errorf("Case: %s\n", c.description)
			t.Errorf(
				"Expected %d hours and %d minutes, but got %d hours and %d minutes",
				c.res[0], h, c.res[1], m,
			)
		}
	}
}

func TestHourlyRepetitions(t *testing.T) {
	now := time.Now().UTC()
	cases := []struct {
		description string
		sch         *Schedule
		res         string
	}{
		{
			description: "Repeat on start time's hour",
			sch: &Schedule{
				HourlyRep: 0,
				StartTime: pointer.String("1:20"),
			},
			res: "1",
		},
		{
			description: "Repeat every hour",
			sch: &Schedule{
				HourlyRep: 1,
			},
			res: "*",
		},
		{
			description: "Repeat every 2 hours",
			sch: &Schedule{
				HourlyRep: 2,
				StartTime: pointer.String("0:0"),
			},
			res: "0,2,4,6,8,10,12,14,16,18,20,22",
		},
		{
			description: "Repeat every 5 hours",
			sch: &Schedule{
				HourlyRep: 5,
				StartTime: pointer.String("0:0"),
			},
			res: "0,5,10,15,20",
		},
		{
			description: "Repeat every 7 hours",
			sch: &Schedule{
				HourlyRep: 7,
				StartTime: pointer.String("0:0"),
			},
			res: "0,7,14,21",
		},
		{
			description: "Repeat every 18 hours",
			sch: &Schedule{
				HourlyRep: 18,
				StartTime: pointer.String("0:0"),
			},
			res: "0,18",
		},
		{
			description: "Repeat every hour, passing at 1:00h",
			sch: &Schedule{
				HourlyRep: 1,
				StartTime: pointer.String("1:20"),
			},
			res: "*",
		},
		{
			description: "Repeat every 2 hours, passing at 13:00h",
			sch: &Schedule{
				HourlyRep: 2,
				StartTime: pointer.String("13:00"),
			},
			res: "1,3,5,7,9,11,13,15,17,19,21,23",
		},
		{
			description: "Repeat every 5 hours, passing at 18:00h",
			sch: &Schedule{
				HourlyRep: 5,
				StartTime: pointer.String("18:44"),
			},
			res: "3,8,13,18,23",
		},
		{
			description: "Repeat every 7 hours, passing at 23:00h",
			sch: &Schedule{
				HourlyRep: 7,
				StartTime: pointer.String("23:00"),
			},
			res: "2,9,16,23",
		},
		{
			description: "Repeat every 18 hours, passing at 23:00h",
			sch: &Schedule{
				HourlyRep: 18,
				StartTime: pointer.String("23:07"),
			},
			res: "5,23",
		},
		{
			description: "Repeat on current hour",
			sch:         &Schedule{},
			res:         strconv.Itoa(now.Hour()),
		},
		{
			description: "No repetition",
			res:         "",
		},
	}

	for _, c := range cases {
		if res := c.sch.HourlyRepSeries(); res != c.res {
			t.Errorf("Case: %s\n", c.description)
			t.Errorf("\nExpected: %s\nBut got:  %s", c.res, res)
		}
	}
}

func TestDayOfWeekSeries(t *testing.T) {
	cases := []struct {
		description string
		sch         *Schedule
		res         string
	}{
		{
			description: "Repeat every Sunday",
			sch: &Schedule{
				DaysOfWeek: []DayOfWeek{0},
			},
			res: "0",
		},
		{
			description: "Repeat from Sunday to Tuesday",
			sch: &Schedule{
				DaysOfWeek: []DayOfWeek{0, 1, 2},
			},
			res: "0,1,2",
		},
		{
			description: "Repeat from Sunday to Tuesday, repeated input",
			sch: &Schedule{
				DaysOfWeek: []DayOfWeek{0, 1, 1, 1, 2},
			},
			res: "0,1,2",
		},
		{
			description: "Repeat every day",
			sch: &Schedule{
				DaysOfWeek: []DayOfWeek{0, 1, 2, 3, 4, 5, 6},
			},
			res: "0,1,2,3,4,5,6",
		},
		{
			description: "Repeat every day when none is specified",
			sch:         &Schedule{},
			res:         "*",
		},
		{
			description: "No repetition",
			res:         "",
		},
	}

	for _, c := range cases {
		if res := c.sch.DayOfWeekSeries(); c.res != res {
			t.Errorf("Case: %s\n", c.description)
			t.Errorf("\nExpected: %s\nBut got:  %s", c.res, res)
		}
	}
}

func TestCronExpr(t *testing.T) {
	now := time.Now().UTC()
	cases := []struct {
		description string
		sch         *Schedule
		res         string
	}{
		{
			description: "Repeat on start time's hour and minute, every day",
			sch: &Schedule{
				HourlyRep: 0,
				StartTime: pointer.String("1:20"),
			},
			res: "20 1 * * *",
		},
		{
			description: "Repeat every hour, starting at current minute, every day",
			sch: &Schedule{
				HourlyRep: 1,
			},
			res: fmt.Sprintf("%d * * * *", now.Minute()),
		},
		{
			description: "Repeat every 2 hours, every day",
			sch: &Schedule{
				HourlyRep: 2,
				StartTime: pointer.String("0:0"),
			},
			res: "0 0,2,4,6,8,10,12,14,16,18,20,22 * * *",
		},
		{
			description: "Repeat every 5 hours, every day",
			sch: &Schedule{
				HourlyRep: 5,
				StartTime: pointer.String("0:0"),
			},
			res: "0 0,5,10,15,20 * * *",
		},
		{
			description: "Repeat every 7 hours, every day",
			sch: &Schedule{
				HourlyRep: 7,
				StartTime: pointer.String("0:0"),
			},
			res: "0 0,7,14,21 * * *",
		},
		{
			description: "Repeat every 18 hours, every day",
			sch: &Schedule{
				HourlyRep: 18,
				StartTime: pointer.String("0:0"),
			},
			res: "0 0,18 * * *",
		},
		{
			description: "Repeat every hour at the 20th minute, passing at 1:20h, every day",
			sch: &Schedule{
				HourlyRep: 1,
				StartTime: pointer.String("1:20"),
			},
			res: "20 * * * *",
		},
		{
			description: "Repeat every 2 hours at the 0th minute, passing at 13:00h, every day",
			sch: &Schedule{
				HourlyRep: 2,
				StartTime: pointer.String("13:00"),
			},
			res: "0 1,3,5,7,9,11,13,15,17,19,21,23 * * *",
		},
		{
			description: "Repeat every 5 hours at the 44th minute, passing at 18:44h, every day",
			sch: &Schedule{
				HourlyRep: 5,
				StartTime: pointer.String("18:44"),
			},
			res: "44 3,8,13,18,23 * * *",
		},
		{
			description: "Repeat every 7 hours at the 0th minute, passing at 23:00h, every day",
			sch: &Schedule{
				HourlyRep: 7,
				StartTime: pointer.String("23:00"),
			},
			res: "0 2,9,16,23 * * *",
		},
		{
			description: "Repeat every 18 hours at the 7th minute, passing at 23:07h, every day",
			sch: &Schedule{
				HourlyRep: 18,
				StartTime: pointer.String("23:07"),
			},
			res: "7 5,23 * * *",
		},

		{
			description: "Repeat on start time's hour and minute, every Sunday",
			sch: &Schedule{
				HourlyRep:  0,
				StartTime:  pointer.String("1:20"),
				DaysOfWeek: []DayOfWeek{0},
			},
			res: "20 1 * * 0",
		},
		{
			description: "Repeat every hour, starting at current minute, every Sunday",
			sch: &Schedule{
				HourlyRep:  1,
				DaysOfWeek: []DayOfWeek{0},
			},
			res: fmt.Sprintf("%d * * * 0", now.Minute()),
		},
		{
			description: "Repeat every 2 hours, every Sunday",
			sch: &Schedule{
				HourlyRep:  2,
				StartTime:  pointer.String("0:0"),
				DaysOfWeek: []DayOfWeek{0},
			},
			res: "0 0,2,4,6,8,10,12,14,16,18,20,22 * * 0",
		},
		{
			description: "Repeat every 5 hours, every Sunday",
			sch: &Schedule{
				HourlyRep:  5,
				StartTime:  pointer.String("0:0"),
				DaysOfWeek: []DayOfWeek{0},
			},
			res: "0 0,5,10,15,20 * * 0",
		},
		{
			description: "Repeat every 7 hours, every Sunday",
			sch: &Schedule{
				HourlyRep:  7,
				StartTime:  pointer.String("0:0"),
				DaysOfWeek: []DayOfWeek{0},
			},
			res: "0 0,7,14,21 * * 0",
		},
		{
			description: "Repeat every 18 hours, every Sunday",
			sch: &Schedule{
				HourlyRep:  18,
				StartTime:  pointer.String("0:0"),
				DaysOfWeek: []DayOfWeek{0},
			},
			res: "0 0,18 * * 0",
		},
		{
			description: "Repeat every hour at the 20th minute, passing at 1:20h, every Sunday",
			sch: &Schedule{
				HourlyRep:  1,
				StartTime:  pointer.String("1:20"),
				DaysOfWeek: []DayOfWeek{0},
			},
			res: "20 * * * 0",
		},
		{
			description: "Repeat every 2 hours at the 0th minute, passing at 13:00h, every Sunday",
			sch: &Schedule{
				HourlyRep:  2,
				StartTime:  pointer.String("13:00"),
				DaysOfWeek: []DayOfWeek{0},
			},
			res: "0 1,3,5,7,9,11,13,15,17,19,21,23 * * 0",
		},
		{
			description: "Repeat every 5 hours at the 44th minute, passing at 18:44h, every Sunday",
			sch: &Schedule{
				HourlyRep:  5,
				StartTime:  pointer.String("18:44"),
				DaysOfWeek: []DayOfWeek{0},
			},
			res: "44 3,8,13,18,23 * * 0",
		},
		{
			description: "Repeat every 7 hours at the 0th minute, passing at 23:00h, every Sunday",
			sch: &Schedule{
				HourlyRep:  7,
				StartTime:  pointer.String("23:00"),
				DaysOfWeek: []DayOfWeek{0},
			},
			res: "0 2,9,16,23 * * 0",
		},
		{
			description: "Repeat every 18 hours at the 7th minute, passing at 23:07h, every Sunday",
			sch: &Schedule{
				HourlyRep:  18,
				StartTime:  pointer.String("23:07"),
				DaysOfWeek: []DayOfWeek{0},
			},
			res: "7 5,23 * * 0",
		},

		{
			description: "Repeat on start time's hour and minute, from Sunday to Wednesday",
			sch: &Schedule{
				HourlyRep:  0,
				StartTime:  pointer.String("1:20"),
				DaysOfWeek: []DayOfWeek{0, 1, 2, 3},
			},
			res: "20 1 * * 0,1,2,3",
		},
		{
			description: "Repeat every hour, starting at current minute, from Sunday to Wednesday",
			sch: &Schedule{
				HourlyRep:  1,
				DaysOfWeek: []DayOfWeek{0, 1, 2, 3},
			},
			res: fmt.Sprintf("%d * * * 0,1,2,3", now.Minute()),
		},
		{
			description: "Repeat every 2 hours, from Sunday to Wednesday",
			sch: &Schedule{
				HourlyRep:  2,
				StartTime:  pointer.String("0:0"),
				DaysOfWeek: []DayOfWeek{0, 1, 2, 3},
			},
			res: "0 0,2,4,6,8,10,12,14,16,18,20,22 * * 0,1,2,3",
		},
		{
			description: "Repeat every 5 hours, from Sunday to Wednesday",
			sch: &Schedule{
				HourlyRep:  5,
				StartTime:  pointer.String("0:0"),
				DaysOfWeek: []DayOfWeek{0, 1, 2, 3},
			},
			res: "0 0,5,10,15,20 * * 0,1,2,3",
		},
		{
			description: "Repeat every 7 hours, from Sunday to Wednesday",
			sch: &Schedule{
				HourlyRep:  7,
				StartTime:  pointer.String("0:0"),
				DaysOfWeek: []DayOfWeek{0, 1, 2, 3},
			},
			res: "0 0,7,14,21 * * 0,1,2,3",
		},
		{
			description: "Repeat every 18 hours, from Sunday to Wednesday",
			sch: &Schedule{
				HourlyRep:  18,
				StartTime:  pointer.String("0:0"),
				DaysOfWeek: []DayOfWeek{0, 1, 2, 3},
			},
			res: "0 0,18 * * 0,1,2,3",
		},
		{
			description: "Repeat every hour at the 20th minute, passing at 1:20h, from Sunday to Wednesday",
			sch: &Schedule{
				HourlyRep:  1,
				StartTime:  pointer.String("1:20"),
				DaysOfWeek: []DayOfWeek{0, 1, 2, 3},
			},
			res: "20 * * * 0,1,2,3",
		},
		{
			description: "Repeat every 2 hours at the 0th minute, passing at 13:00h, from Sunday to Wednesday",
			sch: &Schedule{
				HourlyRep:  2,
				StartTime:  pointer.String("13:00"),
				DaysOfWeek: []DayOfWeek{0, 1, 2, 3},
			},
			res: "0 1,3,5,7,9,11,13,15,17,19,21,23 * * 0,1,2,3",
		},
		{
			description: "Repeat every 5 hours at the 44th minute, passing at 18:44h, from Sunday to Wednesday",
			sch: &Schedule{
				HourlyRep:  5,
				StartTime:  pointer.String("18:44"),
				DaysOfWeek: []DayOfWeek{0, 1, 2, 3},
			},
			res: "44 3,8,13,18,23 * * 0,1,2,3",
		},
		{
			description: "Repeat every 7 hours at the 0th minute, passing at 23:00h, from Sunday to Wednesday",
			sch: &Schedule{
				HourlyRep:  7,
				StartTime:  pointer.String("23:00"),
				DaysOfWeek: []DayOfWeek{0, 1, 2, 3},
			},
			res: "0 2,9,16,23 * * 0,1,2,3",
		},
		{
			description: "Repeat every 18 hours at the 7th minute, passing at 23:07h, from Sunday to Wednesday",
			sch: &Schedule{
				HourlyRep:  18,
				StartTime:  pointer.String("23:07"),
				DaysOfWeek: []DayOfWeek{0, 1, 2, 3},
			},
			res: "7 5,23 * * 0,1,2,3",
		},

		{
			description: "Repeat on current hour and minute, every day",
			sch:         &Schedule{},
			res:         fmt.Sprintf("%d %d * * *", now.Minute(), now.Hour()),
		},
		{
			description: "No repetition",
			res:         "",
		},
	}

	for _, c := range cases {
		if res := c.sch.CronExpr(); c.res != res {
			t.Errorf("Case: %s\n", c.description)
			t.Errorf("\nExpected: %s\nBut got:  %s", c.res, res)
		}
	}
}
