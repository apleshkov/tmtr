package gen

import (
	"fmt"
	"go/ast"
	ht "html/template"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	tt "text/template"
	"text/template/parse"

	"github.com/apleshkov/tmtr/scopes"
)

type Mode int

const (
	ModeText Mode = iota
	ModeHTML
)

type NamedTemplateInfo struct {
	Name, DataType string
}

type GeneratorOptions struct {
	InFile, OutFile string
	Mode            Mode
	Package         string
	FnName          string
	DataType        string
	Tmpls           []NamedTemplateInfo
	Imports         []string
	Funcs           []string
}

type Generator struct {
	mode      Mode
	outIdent  *ast.Ident
	dataIdent *ast.Ident
	eoutIdent *ast.Ident
	imports   *imports
	usedTmpls map[string]bool
}

func GenerateFromFile(opts GeneratorOptions) (*ast.File, error) {
	file := opts.InFile
	if bytes, err := os.ReadFile(file); err != nil {
		return nil, err
	} else {
		name := filepath.Base(file)
		text := string(bytes)
		return generateFromText(name, text, opts)
	}
}

func generateFromText(name, text string, opts GeneratorOptions) (*ast.File, error) {
	if root, all, err := parseText(name, text, opts); err != nil {
		return nil, err
	} else {
		return generateFile(root, all, opts), nil
	}
}

func parseText(name, text string, opts GeneratorOptions) (*tmplWrapper, []*tmplWrapper, error) {
	switch opts.Mode {
	case ModeText:
		return parseTextTemplate(name, text, opts)
	case ModeHTML:
		return parseHTMLTemplate(name, text, opts)
	}
	return nil, nil, fmt.Errorf("parsing failed, unknown generator mode: %d", opts.Mode)
}

var dummyFn = func(...any) string { return "" }

func addDummyFuncs(opts GeneratorOptions, cb func(tt.FuncMap)) {
	fm := make(tt.FuncMap)
	if opts.Funcs != nil {
		for _, n := range opts.Funcs {
			fm[n] = dummyFn
		}
	}
	if opts.Imports != nil {
		for _, n := range opts.Imports {
			if _, n, ok := strings.Cut(n, "/"); ok {
				fm[n] = dummyFn
				continue
			}
			fm[n] = dummyFn
		}
	}
	fm["maybe"] = dummyFn
	cb(fm)
}

func parseTextTemplate(name, text string, opts GeneratorOptions) (root *tmplWrapper, all []*tmplWrapper, err error) {
	tmpl := tt.New(name)
	addDummyFuncs(opts, func(fm tt.FuncMap) { tmpl.Funcs(fm) })
	if _, err := tmpl.Parse(text); err != nil {
		return nil, nil, err
	}
	root, all = wrapTmpls(tmpl, nil, opts)
	return root, all, nil
}

func parseHTMLTemplate(name, text string, opts GeneratorOptions) (root *tmplWrapper, all []*tmplWrapper, err error) {
	tmpl := ht.New(name)
	addDummyFuncs(opts, func(fm tt.FuncMap) { tmpl.Funcs(fm) })
	if _, err := tmpl.Parse(text); err != nil {
		return nil, nil, err
	}
	root, all = wrapTmpls(nil, tmpl, opts)
	// The HTML template's `Execute` method calls the `escape` private
	// method inside. It enriches a template tree with the escaping
	// commands, so we don't need to re-implement it from scratch.
	// Thus we don't care about the actual execution and ignore all its
	// possible errors and panics.
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "[recovery] %v\n", r)
		}
	}()
	if opts.Tmpls != nil {
		// Add external templates to fix the execution
		for _, info := range opts.Tmpls {
			name := info.Name
			if t, err := ht.New(name).Parse(""); err != nil {
				fmt.Fprintf(os.Stderr, "[xtmpl] %v\n", err)
			} else {
				tmpl.AddParseTree(name, t.Tree)
			}
		}
	}
	_ = tmpl.Execute(io.Discard, "")
	return root, all, nil
}

func generateFile(rw *tmplWrapper, wrappers []*tmplWrapper, opts GeneratorOptions) *ast.File {
	scope := scopes.NewRootScope(rw.root)
	imports := newImports()
	if opts.Imports != nil {
		for _, s := range opts.Imports {
			if _, n, ok := strings.Cut(s, "/"); ok { // e.g. "net/http"
				imports.use(n, s, scope)
			} else { // e.g. "math"
				imports.use(s, s, scope)
			}
		}
	}
	decls := make([]ast.Decl, 0)
	for _, w := range wrappers {
		fn := generateFunction(
			w,
			scope,
			imports,
			opts.Mode,
		)
		decls = append(decls, fn)
	}
	decls = append([]ast.Decl{imports.decls()}, decls...)
	return &ast.File{
		Name:  ast.NewIdent(opts.Package),
		Decls: decls,
	}
}

func generateFunction(wrapper *tmplWrapper, rootScope *scopes.RootScope, imports *imports, mode Mode) *ast.FuncDecl {
	scope := scopes.NewListScope(rootScope, wrapper.root)
	g := &Generator{
		mode:      mode,
		outIdent:  scopes.Uniq(scope, "output"),
		dataIdent: scope.Dot(),
		eoutIdent: scopes.Uniq(scope, "errOutput"),
		imports:   imports,
		usedTmpls: make(map[string]bool),
	}
	body := g.listNodeStmt(wrapper.root, scope)
	iowr := &ast.SelectorExpr{
		X:   g.useIO(scope),
		Sel: ast.NewIdent("Writer"),
	}
	declArgs := []*ast.Field{
		{
			Names: []*ast.Ident{g.outIdent},
			Type:  iowr,
		},
		{
			Names: []*ast.Ident{g.dataIdent},
			Type:  ast.NewIdent(wrapper.dataType),
		},
	}
	usedTmpls := make([]NamedTemplateInfo, 0, len(g.usedTmpls))
	for name := range g.usedTmpls {
		if t, ok := wrapper.infos[name]; ok {
			usedTmpls = append(usedTmpls, t)
		} else {
			usedTmpls = append(usedTmpls, NamedTemplateInfo{
				Name:     name,
				DataType: wrapper.dataType,
			})
		}
	}
	slices.SortFunc(usedTmpls, func(a, b NamedTemplateInfo) int {
		return strings.Compare(a.Name, b.Name)
	})
	for _, t := range usedTmpls {
		args := []*ast.Field{{Type: iowr}}
		if len(t.DataType) > 0 {
			args = append(args, &ast.Field{Type: ast.NewIdent(t.DataType)})
		}
		args = append(args, &ast.Field{Type: iowr})
		declArgs = append(declArgs, &ast.Field{
			Names: []*ast.Ident{scopes.Uniq(scope, t.Name)},
			Type: &ast.FuncType{
				Params: &ast.FieldList{
					List: args,
				},
			},
		})
	}
	declArgs = append(declArgs, &ast.Field{
		Names: []*ast.Ident{g.eoutIdent},
		Type:  iowr,
	})
	return &ast.FuncDecl{
		Name: ast.NewIdent(wrapper.fnName),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: declArgs,
			},
		},
		Body: body,
	}
}

type tmplWrapper struct {
	text             *tt.Template
	html             *ht.Template
	root             *parse.ListNode
	fnName, dataType string
	infos            map[string]NamedTemplateInfo
}

func newTmplWrapper(
	text *tt.Template,
	html *ht.Template,
	fnName, dataType string,
	infos map[string]NamedTemplateInfo,
) *tmplWrapper {
	w := &tmplWrapper{
		text:     text,
		html:     html,
		fnName:   fnName,
		dataType: dataType,
		infos:    infos,
	}
	if text != nil {
		w.root = text.Root
	}
	if html != nil {
		w.root = html.Tree.Root
	}
	return w
}

func wrapTmpls(text *tt.Template, html *ht.Template, opts GeneratorOptions) (root *tmplWrapper, all []*tmplWrapper) {
	all = make([]*tmplWrapper, 0)
	infos := make(map[string]NamedTemplateInfo)
	if opts.Tmpls != nil {
		for _, v := range opts.Tmpls {
			infos[v.Name] = v
		}
	}
	if tmpl := text; tmpl != nil {
		for _, t := range tmpl.Templates() {
			tn := t.Name()
			infos[tn] = NamedTemplateInfo{
				Name:     tn,
				DataType: opts.DataType,
			}
			fn := opts.FnName
			if t != tmpl {
				fn += upperFirstLetter(tn)
			}
			w := newTmplWrapper(t, nil, fn, opts.DataType, infos)
			if t == tmpl {
				root = w
			}
			all = append(all, w)
		}
	}
	if tmpl := html; tmpl != nil {
		for _, t := range tmpl.Templates() {
			tn := t.Name()
			infos[tn] = NamedTemplateInfo{
				Name:     tn,
				DataType: opts.DataType,
			}
			fn := opts.FnName
			if t != tmpl {
				fn += upperFirstLetter(tn)
			}
			w := newTmplWrapper(nil, t, fn, opts.DataType, infos)
			if t == tmpl {
				root = w
			}
			all = append(all, w)
		}
	}
	slices.SortFunc(all, func(a, b *tmplWrapper) int {
		return strings.Compare(a.fnName, b.fnName)
	})
	return root, all
}

func upperFirstLetter(s string) string {
	if len(s) > 1 {
		return strings.ToUpper(s[:1]) + s[1:]
	}
	return strings.ToUpper(s)
}
