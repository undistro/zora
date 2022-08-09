package errparse

import (
	"fmt"
	"io"
	"regexp"
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
		regexp.MustCompile(`\[fatal\]\s(.*)\n`),
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
			return string(mats[1]), nil
		}
	}
	return "", fmt.Errorf("Unable to match on <%s> error output", plug)
}
