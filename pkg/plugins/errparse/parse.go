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
)

// The message patterns are ordered by priority.
var patterns = map[string][]*regexp.Regexp{
	"popeye": {
		regexp.MustCompile(`(?m)^panic:\s+.{3}\[38;5;196m(.*).\[0m\n`),
		regexp.MustCompile(`(?m)^Boom!\s+.{3}\[38;5;196m(.*).\[0m\n`),
	},
	"marvin": {
		regexp.MustCompile(`(?m)^Error:\s(.*)\n`),
		regexp.MustCompile(`(?m)^E.*]\s*"msg"="(.*)"\s*"error"=`),
	},
}

// Parse extracts an error message from a given <io.Reader> pointing to a Zora
// plugin error output. It uses regular expressions as heuristics to find the
// message, whereby the first match is returned.
func Parse(r io.Reader, plugin string) (string, error) {
	if _, ok := patterns[plugin]; !ok {
		return "", fmt.Errorf("invalid plugin: <%s>", plugin)
	}
	if r == nil {
		return "", fmt.Errorf("invalid reader")
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("unable to read <%s> error data: %w", plugin, err)
	}
	for _, p := range patterns[plugin] {
		mats := p.FindSubmatch(b)
		if len(mats) >= 2 {
			return string(mats[1]), nil
		}
	}
	return "", fmt.Errorf("unable to match on <%s> error output", plugin)
}
