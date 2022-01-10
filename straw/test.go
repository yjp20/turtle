package straw

import (
	"os"
	"strings"
	"regexp"
)

var source, _ = os.ReadFile("example.straw")

var testcases []string
var matcher, _ = regexp.Compile("^# [A-Z]")

func init() {
	text := string(source)
	tc := ""
	for _, line := range strings.Split(text, "\n") {
		if matcher.MatchString(line) {
			if tc != "" {
				testcases = append(testcases, tc)
			}
			tc = ""
		}
		tc += line + "\n"
	}
	debug(testcases)
}
