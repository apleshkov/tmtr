package gen

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
	"text/template/parse"

	"github.com/apleshkov/tmtr/scopes"
	"github.com/apleshkov/tmtr/util"
)

func (g *Generator) nodeStmt(n parse.Node, scope scopes.Scope) ast.Stmt {
	if n, ok := n.(*parse.TextNode); ok {
		text := string(n.Text)
		return g.writeUnescapedExprStmt(&ast.BasicLit{
			Kind:  token.STRING,
			Value: fmt.Sprintf("%q", text),
		}, scope)
	}
	if _, ok := n.(*parse.CommentNode); ok {
		return &ast.EmptyStmt{}
	}
	if n, ok := n.(*parse.ActionNode); ok {
		return g.actionNodeStmt(n, scope)
	}
	if _, ok := n.(*parse.BreakNode); ok {
		return &ast.BranchStmt{Tok: token.BREAK}
	}
	if _, ok := n.(*parse.ContinueNode); ok {
		return &ast.BranchStmt{Tok: token.CONTINUE}
	}
	if n, ok := n.(*parse.IfNode); ok {
		ifScope := scopes.NewIfScope(scope, n)
		thenScope := ifScope.ThenScope
		body := g.listNodeStmt(n.List, thenScope)
		pipe := n.Pipe
		var stmt *ast.IfStmt
		if pipe.Decl != nil {
			assign := g.pipeAssignStmt(pipe, scope)
			cond := g.nonEmptyCond(assign.Lhs[0], scope)
			for _, x := range assign.Lhs[1:] {
				cond = &ast.BinaryExpr{
					Op: token.LAND,
					X:  cond,
					Y:  g.nonEmptyCond(x, scope),
				}
			}
			stmt = &ast.IfStmt{
				Init: assign,
				Cond: cond,
				Body: body,
			}
		} else {
			stmt = &ast.IfStmt{
				Cond: g.nonEmptyCond(
					g.cmdsExpr(pipe.Cmds, thenScope),
					thenScope,
				),
				Body: body,
			}
		}
		if n.ElseList != nil {
			stmt.Else = g.listNodeStmt(n.ElseList, ifScope.ElseScope)
		}
		return stmt
	}
	if n, ok := n.(*parse.RangeNode); ok {
		iter := g.cmdsExpr(n.Pipe.Cmds, scope)
		scope := scopes.NewRangeScope(scope, n)
		x := scope.List()
		stmt := &ast.IfStmt{
			Init: &ast.AssignStmt{
				Tok: token.DEFINE,
				Lhs: []ast.Expr{x},
				Rhs: []ast.Expr{
					iter,
				},
			},
			Cond: g.nonEmptyCond(x, scope),
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.RangeStmt{
						Key:   scope.Key(),
						Value: scope.Value(),
						Tok:   token.DEFINE,
						X:     x,
						Body:  g.listNodeStmt(n.List, scope),
					},
				},
			},
		}
		if n.ElseList != nil {
			stmt.Else = g.listNodeStmt(n.ElseList, scope.ElseScope)
		}
		return stmt
	}
	if n, ok := n.(*parse.WithNode); ok {
		expr := g.cmdsExpr(n.Pipe.Cmds, scope)
		scope := scopes.NewWithScope(scope, n)
		x := scope.Dot()
		stmt := &ast.IfStmt{
			Init: &ast.AssignStmt{
				Tok: token.DEFINE,
				Lhs: []ast.Expr{x},
				Rhs: []ast.Expr{expr},
			},
			Cond: g.nonEmptyCond(x, scope),
			Body: g.listNodeStmt(n.List, scope),
		}
		if n.ElseList != nil {
			stmt.Else = g.listNodeStmt(n.ElseList, scope.ElseScope)
		}
		return stmt
	}
	if n, ok := n.(*parse.TemplateNode); ok {
		name := n.Name
		if s, _, ok := strings.Cut(name, "$htmltemplate_"); ok {
			// unmangle template name
			name = s
		}
		g.usedTmpls[name] = true
		args := []ast.Expr{g.outIdent}
		if n.Pipe != nil {
			args = append(args, g.cmdsExpr(n.Pipe.Cmds, scope))
		}
		args = append(args, g.eoutIdent)
		expr := &ast.CallExpr{
			Fun:  ast.NewIdent(name),
			Args: args,
		}
		return exprStmt(expr)
	}
	return g.writeExprStmt(g.nodeExpr(n, scope), scope)
}

func (g *Generator) writeUnescapedExprStmt(expr ast.Expr, scope scopes.Scope) ast.Stmt {
	return exprStmt(&ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   g.useFuncs(scope),
			Sel: writeIdent,
		},
		Args: []ast.Expr{g.outIdent, expr, g.eoutIdent},
	})
}

func (g *Generator) writeExprStmt(expr ast.Expr, scope scopes.Scope) ast.Stmt {
	return g.writeUnescapedExprStmt(expr, scope)
}

func (g *Generator) listNodeStmt(list *parse.ListNode, scope scopes.Scope) *ast.BlockStmt {
	body := make([]ast.Stmt, len(list.Nodes))
	for i, n := range list.Nodes {
		body[i] = g.nodeStmt(n, scope)
	}
	return &ast.BlockStmt{
		List: body,
	}
}

func (g *Generator) pipeAssignStmt(pipe *parse.PipeNode, scope scopes.Scope) *ast.AssignStmt {
	if pipe.Decl == nil {
		return nil
	}
	lhs := make([]ast.Expr, 0, len(pipe.Decl))
	for _, decl := range pipe.Decl {
		for _, s := range decl.Ident {
			s = util.TrimDollarPrefix(s)
			lhs = append(lhs, ast.NewIdent(s))
		}
	}
	rhs := g.cmdsExpr(pipe.Cmds, scope)
	tok := token.DEFINE
	if pipe.IsAssign {
		tok = token.ASSIGN
	}
	return &ast.AssignStmt{
		Tok: tok,
		Lhs: lhs,
		Rhs: []ast.Expr{rhs},
	}
}

func (g *Generator) actionNodeStmt(node *parse.ActionNode, scope scopes.Scope) ast.Stmt {
	pipe := node.Pipe
	if pipe.Decl != nil {
		return g.pipeAssignStmt(pipe, scope)
	} else {
		expr := g.cmdsExpr(pipe.Cmds, scope)
		return g.writeExprStmt(expr, scope)
	}
}

func (g *Generator) nonEmptyCond(x ast.Expr, scope scopes.Scope) ast.Expr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   g.useFuncs(scope),
			Sel: isTrueIdent,
		},
		Args: []ast.Expr{x},
	}
}

func exprStmt(expr ast.Expr) ast.Stmt {
	return &ast.ExprStmt{X: expr}
}
