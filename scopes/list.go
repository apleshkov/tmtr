package scopes

import (
	"go/ast"
	"text/template/parse"
)

type ListScope struct {
	parent Scope
	ns     *names
}

func NewListScope(parent Scope, node *parse.ListNode) *ListScope {
	ns := newNames(parent.names())
	procList(node, ns)
	return &ListScope{
		parent: parent,
		ns:     ns,
	}
}

func (s *ListScope) names() *names {
	return s.ns
}

func (s *ListScope) Dot() *ast.Ident {
	return s.parent.Dot()
}

func (s *ListScope) Dollar() *ast.Ident {
	return s.parent.Dollar()
}
