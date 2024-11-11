package gen

import (
	"go/ast"
	"go/token"
	"sort"

	"github.com/apleshkov/tmtr/scopes"
)

type imports struct {
	m map[string]*ast.ImportSpec
}

func newImports() *imports {
	return &imports{
		m: make(map[string]*ast.ImportSpec),
	}
}

func (imp *imports) use(key, path string, in scopes.Scope) *ast.Ident {
	if s, ok := imp.m[key]; ok {
		return s.Name
	}
	n := scopes.Uniq(in, key)
	s := &ast.ImportSpec{
		Name: n,
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: "\"" + path + "\"",
		},
	}
	imp.m[key] = s
	return n
}

func (imp *imports) decls() *ast.GenDecl {
	ks := make([]string, 0, len(imp.m))
	for k := range imp.m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	specs := make([]ast.Spec, len(ks))
	for i, k := range ks {
		specs[i] = imp.m[k]
	}
	return &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: specs,
	}
}

func (g *Generator) useIO(in scopes.Scope) *ast.Ident {
	return g.imports.use(ioPkg, ioPkg, in)
}

func (g *Generator) useFmt(in scopes.Scope) *ast.Ident {
	return g.imports.use(fmtPkg, fmtPkg, in)
}

func (g *Generator) useHTMLTemplate(in scopes.Scope) *ast.Ident {
	return g.imports.use(htPkg, htPkgPath, in)
}

func (g *Generator) useFuncs(in scopes.Scope) *ast.Ident {
	return g.imports.use(funcsPkg, FuncsPkgPath, in)
}

const (
	ioPkg        = "io"
	fmtPkg       = "fmt"
	htPkg        = "ht"
	htPkgPath    = "html/template"
	funcsPkg     = "tmtr"
	FuncsPkgPath = "github.com/apleshkov/tmtr/funcs"
)
