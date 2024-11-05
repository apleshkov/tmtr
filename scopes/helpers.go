package scopes

import (
	"go/ast"
	"text/template/parse"

	"github.com/apleshkov/tmtr/util"
)

func procPipe(pipe *parse.PipeNode, ns *names) []*ast.Ident {
	if pipe.IsAssign || pipe.Decl == nil {
		return nil
	}
	idents := make([]*ast.Ident, 0)
	for _, decl := range pipe.Decl {
		for _, s := range decl.Ident {
			s := util.TrimDollarPrefix(s)
			id := ast.NewIdent(s)
			ns.set(s, id)
			idents = append(idents, id)
		}
	}
	return idents
}

func procList(list *parse.ListNode, ns *names) {
	for _, n := range list.Nodes {
		if n, ok := n.(*parse.ActionNode); ok {
			procPipe(n.Pipe, ns)
		}
	}
}
