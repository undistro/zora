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

package formats

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

func Memory(q resource.Quantity) string {
	return fmt.Sprintf("%vMi", q.Value()/(1024*1024))
}

func MemoryUsage(q resource.Quantity, percentage int32) string {
	return fmt.Sprintf("%s (%d%%)", Memory(q), percentage)
}

func CPU(q resource.Quantity) string {
	return fmt.Sprintf("%vm", q.MilliValue())
}

func CPUUsage(q resource.Quantity, percentage int32) string {
	return fmt.Sprintf("%s (%d%%)", CPU(q), percentage)
}
