package scopes

import "go/ast"

type names struct {
	parent *names
	data   map[string]*ast.Ident
}

func newNames(parent *names) *names {
	return &names{
		parent: parent,
		data:   make(map[string]*ast.Ident),
	}
}

func (ns *names) has(name string) bool {
	_, has := ns.data[name]
	if !has && ns.parent != nil {
		return ns.parent.has(name)
	}
	return has
}

func (ns *names) set(name string, ident *ast.Ident) {
	ns.data[name] = ident
}

func (ns *names) uniq(name string) *ast.Ident {
	if ns.has(name) {
		return ns.uniq(name + "_")
	} else {
		ident := ast.NewIdent(name)
		ns.set(name, ident)
		return ident
	}
}
