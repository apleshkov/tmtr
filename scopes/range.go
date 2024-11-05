package scopes

import (
	"go/ast"
	"text/template/parse"

	"github.com/apleshkov/tmtr/util"
)

type RangeScope struct {
	ns               *names
	list, key, value *ast.Ident
	ElseScope        *ListScope // nil if no ElseList
}

func NewRangeScope(parent Scope, node *parse.RangeNode) *RangeScope {
	ns := newNames(parent.names())
	procList(node.List, ns)
	var key, value *ast.Ident
	if vars := procPipe(node.Pipe, ns); vars != nil {
		if len(vars) == 2 {
			key = vars[0]
			value = vars[1]
		} else if len(vars) == 1 {
			key = util.UnderscoreIdent
			value = vars[0]
		} else {
			panic("invalid range declaration: " + node.String())
		}
	} else {
		key = util.UnderscoreIdent
		value = ns.uniq("elem")
	}
	scope := &RangeScope{
		ns:    ns,
		list:  ns.uniq("list"),
		key:   key,
		value: value,
	}
	if node.ElseList != nil {
		scope.ElseScope = NewListScope(parent, node.ElseList)
	}
	return scope
}

func (s *RangeScope) List() *ast.Ident {
	return s.list
}

func (s *RangeScope) Key() *ast.Ident {
	return s.key
}

func (s *RangeScope) Value() *ast.Ident {
	return s.value
}

func (s *RangeScope) names() *names {
	return s.ns
}

func (s *RangeScope) Dot() *ast.Ident {
	return s.value
}

func (s *RangeScope) Dollar() *ast.Ident {
	return s.list
}
