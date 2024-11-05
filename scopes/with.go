package scopes

import (
	"go/ast"
	"text/template/parse"
)

type WithScope struct {
	ns        *names
	dot       *ast.Ident
	dollar    *ast.Ident
	ElseScope *ListScope
}

func NewWithScope(parent Scope, node *parse.WithNode) *WithScope {
	ns := newNames(parent.names())
	scope := &WithScope{
		ns:     ns,
		dot:    ns.uniq("with"),
		dollar: parent.Dot(),
	}
	if vars := procPipe(node.Pipe, ns); vars != nil {
		scope.dot = vars[0]
	}
	if node.ElseList != nil {
		scope.ElseScope = NewListScope(parent, node.ElseList)
	}
	return scope
}

func (s *WithScope) names() *names {
	return s.ns
}

func (s *WithScope) Dot() *ast.Ident {
	return s.dot
}

func (s *WithScope) Dollar() *ast.Ident {
	return s.dollar
}
