package scopes

import (
	"go/ast"
	"text/template/parse"
)

type IfScope struct {
	ThenScope, ElseScope *ListScope
}

func NewIfScope(parent Scope, node *parse.IfNode) *IfScope {
	ns := newNames(parent.names())
	if pipe := node.Pipe; pipe.Decl != nil {
		procPipe(pipe, ns)
	}
	parent = &ifInitScope{
		parent: parent,
		ns:     ns,
	}
	res := &IfScope{
		ThenScope: NewListScope(parent, node.List),
	}
	if node.ElseList != nil {
		res.ElseScope = NewListScope(parent, node.ElseList)
	}
	return res
}

type ifInitScope struct {
	parent Scope
	ns     *names
}

func (s *ifInitScope) names() *names {
	return s.ns
}

func (s *ifInitScope) Dot() *ast.Ident {
	return s.parent.Dot()
}

func (s *ifInitScope) Dollar() *ast.Ident {
	return s.parent.Dollar()
}
