package wiki_api

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"widiff/wiki"
)

var htmlPrefix string = "<tr><td colspan=\"4\"><pre>"
var htmlSuffix string = "</pre></td></tr>"

func ParseDiffText(c wiki.Comparison) (string, error) {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("diff --git a/%s a/%s\n\n", c.FromTitle, c.FromTitle))

	diff, ok := strings.CutPrefix(c.Body, htmlPrefix)
	if !ok {
		return "", fmt.Errorf("could not strip prefix")
	}

	diff, ok = strings.CutSuffix(diff, htmlSuffix)
	if !ok {
		return "", fmt.Errorf("could not strip suffix")
	}
	builder.WriteString(diff)

	return builder.String(), nil
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func LongestChange(changes []wiki.RecentChange) (wiki.RecentChange, int) {
	longest := slices.MaxFunc(changes, func(a, b wiki.RecentChange) int {
		return cmp.Compare(Abs(a.OldLen-a.NewLen), Abs(b.OldLen-b.NewLen))
	})

	return longest, Abs(longest.OldLen - longest.NewLen)
}
