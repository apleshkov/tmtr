package gen

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"text/template/parse"

	"github.com/apleshkov/tmtr/scopes"
	"github.com/apleshkov/tmtr/util"
)

func (g *Generator) nodeExpr(n parse.Node, scope scopes.Scope) ast.Expr {
	if n, ok := n.(*parse.CommandNode); ok {
		return g.cmdExpr(n, scope)
	}
	if n, ok := n.(*parse.IdentifierNode); ok {
		s := n.Ident
		switch s {
		case "and":
			return &ast.SelectorExpr{
				X:   g.useFuncs(scope),
				Sel: andIdent,
			}
		case "or":
			return &ast.SelectorExpr{
				X:   g.useFuncs(scope),
				Sel: orIdent,
			}
		case "html":
			return &ast.SelectorExpr{
				X:   g.useHTMLTemplate(scope),
				Sel: htmlEscaperIdent,
			}
		case "js":
			return &ast.SelectorExpr{
				X:   g.useHTMLTemplate(scope),
				Sel: jsEscaperIdent,
			}
		case "not":
			return &ast.SelectorExpr{
				X:   g.useFuncs(scope),
				Sel: isNotTrueIdent,
			}
		case "print", "_eval_args_":
			return &ast.SelectorExpr{
				X:   g.useFmt(scope),
				Sel: sprintIdent,
			}
		case "printf":
			return &ast.SelectorExpr{
				X:   g.useFmt(scope),
				Sel: sprintfIdent,
			}
		case "println":
			return &ast.SelectorExpr{
				X:   g.useFmt(scope),
				Sel: sprintlnIdent,
			}
		case "urlquery":
			return &ast.SelectorExpr{
				X:   g.useHTMLTemplate(scope),
				Sel: urlQueryEscaperIdent,
			}
		case "_html_template_attrescaper":
			return &ast.SelectorExpr{
				X:   g.useFuncs(scope),
				Sel: escapeHTMLAttrIdent,
			}
		case "_html_template_commentescaper":
			return &ast.SelectorExpr{
				X:   g.useFuncs(scope),
				Sel: escapeCommentIdent,
			}
		case "_html_template_cssescaper":
			return &ast.SelectorExpr{
				X:   g.useFuncs(scope),
				Sel: escapeCSSIdent,
			}
		case "_html_template_cssvaluefilter":
			return &ast.SelectorExpr{
				X:   g.useFuncs(scope),
				Sel: filterCSSIdent,
			}
		case "_html_template_htmlnamefilter":
			return &ast.SelectorExpr{
				X:   g.useFuncs(scope),
				Sel: filterHTMLTagContentIdent,
			}
		case "_html_template_htmlescaper":
			return &ast.SelectorExpr{
				X:   g.useFuncs(scope),
				Sel: escapeHTMLIdent,
			}
		case "_html_template_jsregexpescaper":
			return &ast.SelectorExpr{
				X:   g.useFuncs(scope),
				Sel: escapeJSRegexpIdent,
			}
		case "_html_template_jsstrescaper":
			return &ast.SelectorExpr{
				X:   g.useFuncs(scope),
				Sel: escapeJSStrIdent,
			}
		case "_html_template_jstmpllitescaper":
			return &ast.SelectorExpr{
				X:   g.useFuncs(scope),
				Sel: escapeJSTmplLitIdent,
			}
		case "_html_template_jsvalescaper":
			return &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   g.useFuncs(scope),
					Sel: escapeJSIdent,
				},
				Args: []ast.Expr{g.eoutIdent},
			}
		case "_html_template_nospaceescaper":
			return &ast.SelectorExpr{
				X:   g.useFuncs(scope),
				Sel: escapeUnquotedHTMLAttrIdent,
			}
		case "_html_template_rcdataescaper":
			return &ast.SelectorExpr{
				X:   g.useFuncs(scope),
				Sel: escapeRCDataIdent,
			}
		case "_html_template_srcsetescaper":
			return &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   g.useFuncs(scope),
					Sel: filterAndEscapeSrcsetIdent,
				},
				Args: []ast.Expr{g.eoutIdent},
			}
		case "_html_template_urlescaper":
			return &ast.SelectorExpr{
				X:   g.useFuncs(scope),
				Sel: escapeURLIdent,
			}
		case "_html_template_urlfilter":
			return &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   g.useFuncs(scope),
					Sel: filterURLIdent,
				},
				Args: []ast.Expr{g.eoutIdent},
			}
		case "_html_template_urlnormalizer":
			return &ast.SelectorExpr{
				X:   g.useFuncs(scope),
				Sel: normalizeURLIdent,
			}
		default:
			return ast.NewIdent(s)
		}
	}
	if n, ok := n.(*parse.ChainNode); ok {
		expr := g.nodeExpr(n.Node, scope)
		for _, f := range n.Field {
			expr = &ast.SelectorExpr{
				X:   expr,
				Sel: ast.NewIdent(f),
			}
		}
		return expr
	}
	if n, ok := n.(*parse.PipeNode); ok {
		return g.cmdsExpr(n.Cmds, scope)
	}
	if n, ok := n.(*parse.VariableNode); ok {
		idents := n.Ident
		if len(idents) == 0 {
			return &ast.BadExpr{}
		}
		var expr ast.Expr
		name := idents[0]
		if name == "$" {
			expr = scope.Dollar()
		} else {
			name = util.TrimDollarPrefix(name)
			expr = ast.NewIdent(name)
		}
		for _, s := range idents[1:] {
			expr = &ast.SelectorExpr{
				X:   expr,
				Sel: ast.NewIdent(s),
			}
		}
		return expr
	}
	if _, ok := n.(*parse.DotNode); ok {
		return scope.Dot()
	}
	if _, ok := n.(*parse.NilNode); ok {
		return nilIdent
	}
	if n, ok := n.(*parse.FieldNode); ok {
		var c ast.Expr = scope.Dot()
		for _, s := range n.Ident {
			c = &ast.SelectorExpr{
				X:   c,
				Sel: ast.NewIdent(s),
			}
		}
		return c
	}
	if n, ok := n.(*parse.BoolNode); ok {
		return ast.NewIdent(n.String())
	}
	if n, ok := n.(*parse.NumberNode); ok {
		if n.IsInt {
			return &ast.BasicLit{
				Kind:  token.INT,
				Value: n.Text,
			}
		}
		if n.IsFloat {
			return &ast.BasicLit{
				Kind:  token.FLOAT,
				Value: n.Text,
			}
		}
		if n.IsComplex {
			return &ast.BasicLit{
				Kind:  token.IMAG,
				Value: n.Text,
			}
		}
	}
	if n, ok := n.(*parse.StringNode); ok {
		return &ast.BasicLit{
			Kind:  token.STRING,
			Value: n.Quoted,
		}
	}
	fmt.Fprintf(os.Stderr, "[expr] unknown node: %s (type: %d)\n", n.String(), n.Type())
	return &ast.BadExpr{}
}

func (g *Generator) indexExpr(args []parse.Node, scope scopes.Scope) ast.Expr {
	if len(args) < 2 {
		return &ast.BadExpr{}
	}
	expr := &ast.IndexExpr{
		X:     g.nodeExpr(args[0], scope),
		Index: g.nodeExpr(args[1], scope),
	}
	for _, a := range args[2:] {
		expr = &ast.IndexExpr{
			X:     expr,
			Index: g.nodeExpr(a, scope),
		}
	}
	return expr
}

func (g *Generator) sliceExpr(args []parse.Node, scope scopes.Scope) ast.Expr {
	if len(args) < 1 {
		return &ast.BadExpr{}
	}
	x := g.nodeExpr(args[0], scope)
	args = args[1:]
	switch len(args) {
	case 0:
		return &ast.SliceExpr{X: x}
	case 1:
		return &ast.SliceExpr{
			X:   x,
			Low: g.nodeExpr(args[0], scope),
		}
	case 2:
		return &ast.SliceExpr{
			X:    x,
			Low:  g.nodeExpr(args[0], scope),
			High: g.nodeExpr(args[1], scope),
		}
	case 3:
		return &ast.SliceExpr{
			X:      x,
			Low:    g.nodeExpr(args[0], scope),
			High:   g.nodeExpr(args[1], scope),
			Max:    g.nodeExpr(args[2], scope),
			Slice3: true,
		}
	default:
		return &ast.BadExpr{}
	}
}

func (g *Generator) binExpr(op token.Token, args []parse.Node, scope scopes.Scope) ast.Expr {
	if op == token.EQL && len(args) > 2 {
		x := g.nodeExpr(args[0], scope)
		var prev ast.Expr
		for _, a := range args[1:] {
			eq := &ast.BinaryExpr{
				Op: op,
				X:  x,
				Y:  g.nodeExpr(a, scope),
			}
			if prev != nil {
				prev = &ast.BinaryExpr{
					Op: token.LOR,
					X:  prev,
					Y:  eq,
				}
			} else {
				prev = eq
			}
		}
		return prev
	}
	if len(args) != 2 {
		return &ast.BadExpr{}
	}
	return &ast.BinaryExpr{
		Op: op,
		X:  g.nodeExpr(args[0], scope),
		Y:  g.nodeExpr(args[1], scope),
	}
}

func (g *Generator) maybeExpr(args []parse.Node, scope scopes.Scope) ast.Expr {
	res := make([]ast.Expr, 0)
	if len(args) > 0 {
		expr := g.nodesExpr(args, scope)
		res = append(res, expr)
	}
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   g.useFuncs(scope),
			Sel: maybeIdent,
		},
		Args: []ast.Expr{
			g.eoutIdent,
			&ast.FuncLit{
				Type: &ast.FuncType{
					Results: &ast.FieldList{
						List: []*ast.Field{
							{Type: ast.NewIdent("any")},
							{Type: ast.NewIdent("error")},
						},
					},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: res,
						},
					},
				},
			},
		},
	}
}

func (g *Generator) nodesExpr(nodes []parse.Node, scope scopes.Scope) ast.Expr {
	if len(nodes) > 1 {
		var root, curr *ast.CallExpr
		for i, a := range nodes {
			expr := g.nodeExpr(a, scope)
			if id, ok := expr.(*ast.Ident); ok {
				if id.Name == "call" {
					continue
				}
				if id.Name == "index" {
					return g.indexExpr(nodes[(i+1):], scope)
				}
				if id.Name == "slice" {
					return g.sliceExpr(nodes[(i+1):], scope)
				}
				if id.Name == "eq" {
					return g.binExpr(token.EQL, nodes[(i+1):], scope)
				}
				if id.Name == "ne" {
					return g.binExpr(token.NEQ, nodes[(i+1):], scope)
				}
				if id.Name == "lt" {
					return g.binExpr(token.LSS, nodes[(i+1):], scope)
				}
				if id.Name == "le" {
					return g.binExpr(token.LEQ, nodes[(i+1):], scope)
				}
				if id.Name == "gt" {
					return g.binExpr(token.GTR, nodes[(i+1):], scope)
				}
				if id.Name == "ge" {
					return g.binExpr(token.GEQ, nodes[(i+1):], scope)
				}
				if id.Name == "maybe" {
					return g.maybeExpr(nodes[(i+1):], scope)
				}
			}
			if root == nil {
				if call, ok := expr.(*ast.CallExpr); ok {
					root = call
				} else {
					root = &ast.CallExpr{
						Fun:  expr,
						Args: make([]ast.Expr, 0),
					}
				}
				curr = root
			} else {
				curr.Args = append(curr.Args, expr)
				if call, ok := expr.(*ast.CallExpr); ok {
					curr = call
				}
			}
		}
		return root
	}
	if len(nodes) == 1 {
		return g.nodeExpr(nodes[0], scope)
	}
	return &ast.BadExpr{}
}

func (g *Generator) cmdExpr(cmd *parse.CommandNode, scope scopes.Scope) ast.Expr {
	args := cmd.Args
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "[expr] empty command: %s\n", cmd.String())
		return &ast.BadExpr{}
	}
	return g.nodesExpr(args, scope)
}

func (g *Generator) cmdsExpr(cmds []*parse.CommandNode, scope scopes.Scope) ast.Expr {
	if len(cmds) > 1 {
		var prev ast.Expr
		for _, cmd := range cmds {
			expr := g.cmdExpr(cmd, scope)
			if prev != nil {
				if call, ok := expr.(*ast.CallExpr); ok {
					call.Args = append(call.Args, prev)
				} else {
					expr = &ast.CallExpr{
						Fun:  expr,
						Args: []ast.Expr{prev},
					}
				}
			}
			prev = expr
		}
		return prev
	}
	if len(cmds) == 1 {
		return g.cmdExpr(cmds[0], scope)
	}
	fmt.Fprintln(os.Stderr, "[expr] empty commands")
	return &ast.BadExpr{}
}
