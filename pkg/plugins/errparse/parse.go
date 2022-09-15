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

package errparse

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

// The message patterns are ordered by priority.
var patterns = map[string][]*regexp.Regexp{
	"popeye": {
		regexp.MustCompile(`panic:\s+.{3}\[38;5;196m(.*).\[0m\n`),
		regexp.MustCompile(`Boom!\s+.{3}\[38;5;196m(.*).\[0m\n`),
	},
	"kubescape": {
		regexp.MustCompile(`panic:\s+(.*)\n`),
		regexp.MustCompile(`\[error\]\s(.*)\n`),
		regexp.MustCompile(`\{"level":"error","ts":"\S+","msg":"(.*)"\}`),
		regexp.MustCompile(`\[fatal\]\s(.*)\n`),
		regexp.MustCompile(`\{"level":"fatal","ts":"\S+","msg":"(.*)"\}`),
	},
}

// Parse extracts an error message from a given <io.Reader> pointing to a Zora
// plugin error output. It uses regular expressions as heuristics to find the
// message, whereby the first match is returned.
func Parse(r io.Reader, plug string) (string, error) {
	if _, ok := patterns[plug]; !ok {
		return "", fmt.Errorf("Invalid plugin: <%s>", plug)
	}
	fc, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("Unable to read <%s> error data: %w", plug, err)
	}
	for _, p := range patterns[plug] {
		mats := p.FindSubmatch(fc)
		if len(mats) >= 2 {
			if plug == "kubescape" {
				return strings.ReplaceAll(string(mats[1]), `\"`, `"`), nil
			} else {
				return string(mats[1]), nil
			}
		}
	}
	return "", fmt.Errorf("Unable to match on <%s> error output", plug)
}
