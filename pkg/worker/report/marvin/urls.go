// Copyright 2023 Undistro Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package marvin

import marvin "github.com/undistro/marvin/pkg/types"

const (
	pssBaselineURL   = "https://kubernetes.io/docs/concepts/security/pod-security-standards/#baseline"
	pssRestrictedURL = "https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted"
	k8sResourcesURL  = "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/"
)

//nolint:lll
var urls = map[string]string{
	"M-100": pssBaselineURL,
	"M-101": pssBaselineURL,
	"M-102": pssBaselineURL,
	"M-103": pssBaselineURL,
	"M-104": pssBaselineURL,
	"M-105": pssBaselineURL,
	"M-106": pssBaselineURL,
	"M-107": pssBaselineURL,
	"M-108": pssBaselineURL,
	"M-109": pssBaselineURL,
	"M-110": pssBaselineURL,

	"M-111": pssRestrictedURL,
	"M-112": pssRestrictedURL,
	"M-113": pssRestrictedURL,
	"M-114": pssRestrictedURL,
	"M-115": pssRestrictedURL,
	"M-116": pssRestrictedURL,

	"M-201": "https://microsoft.github.io/Threat-Matrix-for-Kubernetes/mitigations/MS-M9026%20Avoid%20using%20plain%20text%20credentials%20in%20configuration%20files/",
	"M-202": "https://microsoft.github.io/Threat-Matrix-for-Kubernetes/mitigations/MS-M9025%20Disable%20Service%20Account%20Auto%20Mount/",
	"M-203": "https://microsoft.github.io/Threat-Matrix-for-Kubernetes/mitigations/MS-M9015%20Avoid%20Running%20Management%20Interface%20on%20Containers/",

	"M-300": "https://media.defense.gov/2022/Aug/29/2003066362/-1/-1/0/CTR_KUBERNETES_HARDENING_GUIDANCE_1.2_20220829.PDF#page=50",

	"M-400": "https://kubernetes.io/docs/concepts/containers/images/#image-names",
	"M-401": "https://kubernetes.io/docs/concepts/configuration/overview/#naked-pods-vs-replicasets-deployments-and-jobs",
	"M-402": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/",
	"M-403": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/",
	"M-404": k8sResourcesURL,
	"M-405": k8sResourcesURL,
	"M-406": k8sResourcesURL,
	"M-407": k8sResourcesURL,
}

func getURL(check *marvin.CheckResult) string {
	if u, ok := check.Labels["url"]; ok && !check.Builtin {
		return u
	}
	return urls[check.ID]
}
