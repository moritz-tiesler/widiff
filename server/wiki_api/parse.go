package wikiapi

import (
	"fmt"
	"strings"
)

var htmlPrefix string = "<tr><td colspan=\"4\"><pre>"
var htmlSuffix string = "</pre></td></tr>"

func ParseDiffText(diff string) (string, error) {
	diff, ok := strings.CutPrefix(diff, htmlPrefix)
	if !ok {
		return "", fmt.Errorf("could not strip prefix")
	}

	diff, ok = strings.CutSuffix(diff, htmlSuffix)
	if !ok {
		return "", fmt.Errorf("could not strip suffix")
	}

	return diff, nil
}
