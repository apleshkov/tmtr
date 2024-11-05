package scopes

import (
	"go/ast"
	"text/template/parse"
)

type RootScope struct {
	ns  *names
	dot *ast.Ident
}

func NewRootScope(node *parse.ListNode) *RootScope {
	ns := newNames(nil)
	procList(node, ns)
	return &RootScope{
		ns:  ns,
		dot: ns.uniq("data"),
	}
}

func (s *RootScope) names() *names {
	return s.ns
}

func (s *RootScope) Dot() *ast.Ident {
	return s.dot
}

func (s *RootScope) Dollar() *ast.Ident {
	return s.dot
}
