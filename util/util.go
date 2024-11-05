package util

import (
	"go/ast"
)

func TrimDollarPrefix(s string) string {
	if len(s) > 1 && s[0] == '$' {
		return s[1:]
	} else {
		return s
	}
}

var (
	UnderscoreIdent = ast.NewIdent("_")
)
