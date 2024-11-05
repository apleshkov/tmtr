package scopes

import (
	"go/ast"
)

type Scope interface {
	names() *names
	Dot() *ast.Ident    // .
	Dollar() *ast.Ident // $
}

func Uniq(s Scope, name string) *ast.Ident {
	return s.names().uniq(name)
}
