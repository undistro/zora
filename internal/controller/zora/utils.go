// Copyright 2024 Undistro Authors
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

package zora

import "fmt"

func truncateName(name string, length int) string {
	nameLen := len(name)
	if nameLen <= length {
		return name
	} else {
		maxLength := length - 3
		suffixLen := maxLength / 2
		prefixLen := maxLength - suffixLen
		return fmt.Sprintf("%s---%s", name[0:prefixLen], name[nameLen-suffixLen:])
	}
}
