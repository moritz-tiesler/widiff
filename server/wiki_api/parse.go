package wikiapi

import (
	"fmt"
	"strings"
)

var htmlPrefix string = "<tr><td colspan=\"4\"><pre>"
var htmlSuffix string = "</pre></td></tr>"

func ParseDiffText(c Comparison) (string, error) {
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
