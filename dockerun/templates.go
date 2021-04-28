package dockerun

import (
	"regexp"
	"strings"
	"text/template"
)

var rex = regexp.MustCompile(`[^\=\<\>\s\/]+`)
var tFuncs = template.FuncMap{
	"name": tName,
}

func tName(pkg string) string {
	pkg = strings.Trim(pkg, " /")
	parts := strings.Split(pkg, "/")
	if len(parts) == 0 {
		return pkg
	}
	pkg = parts[len(parts)-1]

	pkg = rex.FindString(pkg)
	return pkg
}
