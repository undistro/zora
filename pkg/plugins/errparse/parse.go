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

type pluginPattern struct {
	regexp regexp.Regexp
	msgf   func([][]byte) string
}

// The message patterns are ordered by priority.
var patterns = map[string][]pluginPattern{
	"popeye": {
		{regexp: *regexp.MustCompile(`(?m)^panic:\s+.{3}\[38;5;196m(.*).\[0m\n`), msgf: firstGroup},
		{regexp: *regexp.MustCompile(`(?m)^Boom!\s+.{3}\[38;5;196m(.*).\[0m\n`), msgf: firstGroup},
	},
	"marvin": {
		{regexp: *regexp.MustCompile(`(?m)^Error:\s(.*)\n`), msgf: firstGroup},
		{
			regexp: *regexp.MustCompile(`(?m)^E.*]\s*"msg"="(.*)"\s*"error"="(.*?)"`),
			msgf: func(matches [][]byte) string {
				var err string
				if len(matches) >= 3 {
					err = ": " + string(matches[2])
				}
				return string(matches[1]) + err
			},
		},
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
		matches := p.regexp.FindSubmatch(b)
		if len(matches) >= 2 {
			return p.msgf(matches), nil
		}
	}
	return "", fmt.Errorf("unable to match on <%s> error output", plugin)
}

func firstGroup(matches [][]byte) string {
	return string(matches[1])
}
