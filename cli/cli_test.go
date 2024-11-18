package cli

import (
	"flag"
	"io"
	"strings"
	"testing"

	"github.com/apleshkov/tmtr/util"

	"github.com/apleshkov/tmtr/gen"
)

func TestUsage(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var buf strings.Builder
	fs.SetOutput(&buf)
	_ = newParser(fs)
	fs.Usage()
	util.TestEq(
		t, buf.String(),
		`Usage of tmtr 0.1.12:
  tmtr [-pkg name] -fn name -type type -in file [-mode mode] [-out file] [-tpl name[:type] ...] [-import ...] [-tplfn ...]

Examples:
  # Basic usage:
  tmtr -pkg "main" -fn "RenderIndex" -type "any" -in "./index.html"

  # External templates:
  tmtr -pkg "main" -fn "RenderIndex" -type "any" -in "./index.html" -tpl "foo:Foo" -tpl "foo:map[string]any"

  # User template function and its neccessary import:
  tmtr -pkg "main" -fn "RenderIndex" -type "any" -in "./index.html" -import "strconv" -tplfn "strconv.Atoi"

For more information, see:
  https://github.com/apleshkov/tmtr

Flags:
  -fn string
    	[required] function name
  -import value
    	[multiple] additional imports, e.g. "net/http"; comma-separated is also supported, e.g. "fmt,strings"
  -in string
    	path to the template file
  -mode in
    	'text' or 'html'; optional: 'html' is used if in's extension ends with 'html' (e.g. 'foo.html', 'bar.gohtml'), 'text' otherwise
  -out in
    	path to the output *.go file; optional: adds '.go' to the in filename (e.g. 'foo.html' -> 'foo.html.go')
  -pkg string
    	package name; optional: $GOPACKAGE by default (is set by go:generate)
  -tpl value
    	[multiple] external template with type, e.g. "foo:Foo"; comma-separated is also supported, e.g. "baz:string,quux:[]int"
  -tplfn value
    	[multiple] user template functions; comma-separated is also supported, e.g. "foo,bar"
  -type string
    	[required] data type
  -h, -help
    	Prints this message
`,
	)
}

func TestBad(t *testing.T) {
	_, err := newTestParser()([]string{})
	util.TestAssert(t, err != nil)
	_, err = newTestParser()([]string{""})
	util.TestAssert(t, err != nil)
	_, err = newTestParser()([]string{"-pkg", "main"})
	util.TestAssert(t, err != nil)
	_, err = newTestParser()([]string{"-pkg", "main", "-fn", "Test"})
	util.TestAssert(t, err != nil)
	_, err = newTestParser()([]string{
		"-pkg", "main", "-fn", "Test", "-type", "any",
	})
	util.TestAssert(t, err != nil)
	_, err = newTestParser()([]string{
		"-pkg", "main", "-fn", "Test", "-type", "any",
		"-in", "./foo.txt", "-mode", "foo",
	})
	util.TestAssert(t, err != nil)
}

func TestGood(t *testing.T) {
	opts, _ := newTestParser()(testMinArgs)
	util.TestEq(t, opts.Package, "main")
	util.TestEq(t, opts.FnName, "Test")
	util.TestEq(t, opts.DataType, "any")
	util.TestEq(t, opts.InFile, "./foo.txt")
	util.TestEq(t, opts.OutFile, "./foo.txt.go")
	util.TestEq(t, opts.Mode, gen.ModeText)
}

func TestMode(t *testing.T) {
	opts, _ := newTestParser()(append(testMinArgs, "-mode", "text"))
	util.TestEq(t, opts.Mode, gen.ModeText)
	opts, _ = newTestParser()(append(testMinArgs, "-mode", "html"))
	util.TestEq(t, opts.Mode, gen.ModeHTML)
	opts, _ = newTestParser()(append(testMinArgs, "-in", "test.html"))
	util.TestEq(t, opts.Mode, gen.ModeHTML)
	opts, _ = newTestParser()(append(testMinArgs, "-in", "/path/to/test.gohtml"))
	util.TestEq(t, opts.Mode, gen.ModeHTML)
}

func TestTmpls(t *testing.T) {
	opts, _ := newTestParser()(append(testMinArgs, "-tpl", "foo:any"))
	util.TestEqSlice(t, opts.Tmpls, []gen.NamedTemplateInfo{
		{Name: "foo", DataType: "any"},
	})
	opts, _ = newTestParser()(append(testMinArgs, "-tpl", "  foo : []int", "-tpl", "bar:[][]bool   "))
	util.TestEqSlice(t, opts.Tmpls, []gen.NamedTemplateInfo{
		{Name: "foo", DataType: "[]int"},
		{Name: "bar", DataType: "[][]bool"},
	})
	opts, _ = newTestParser()(append(testMinArgs, "-tpl", "foo:map[string]int,bar : func(), baz:Baz", "-tpl", "quux:func(float32)float32"))
	util.TestEqSlice(t, opts.Tmpls, []gen.NamedTemplateInfo{
		{Name: "foo", DataType: "map[string]int"},
		{Name: "bar", DataType: "func()"},
		{Name: "baz", DataType: "Baz"},
		{Name: "quux", DataType: "func(float32)float32"},
	})
	opts, _ = newTestParser()(append(testMinArgs, "-tpl", "foo"))
	util.TestEqSlice(t, opts.Tmpls, []gen.NamedTemplateInfo{
		{Name: "foo", DataType: ""},
	})
	opts, _ = newTestParser()(append(testMinArgs, "-tpl", "foo:"))
	util.TestEqSlice(t, opts.Tmpls, []gen.NamedTemplateInfo{
		{Name: "foo", DataType: ""},
	})
}

func TestImports(t *testing.T) {
	opts, _ := newTestParser()(append(testMinArgs, "-import", "net"))
	util.TestEqSlice(t, opts.Imports, []string{"net"})
	opts, _ = newTestParser()(append(testMinArgs, "-import", "  net/http   ", "-import", " fmt "))
	util.TestEqSlice(t, opts.Imports, []string{"net/http", "fmt"})
	opts, _ = newTestParser()(append(testMinArgs, "-import", "  net/http , fmt  ", "-import", " unicode "))
	util.TestEqSlice(t, opts.Imports, []string{"net/http", "fmt", "unicode"})
}

func TestFuncs(t *testing.T) {
	opts, _ := newTestParser()(append(testMinArgs, "-tplfn", "foo"))
	util.TestEqSlice(t, opts.Funcs, []string{"foo"})
	opts, _ = newTestParser()(append(testMinArgs, "-tplfn", "  foo   ", "-tplfn", " Bar "))
	util.TestEqSlice(t, opts.Funcs, []string{"foo", "Bar"})
	opts, _ = newTestParser()(append(testMinArgs, "-tplfn", "  foo , Bar  ", "-tplfn", " BazBaz "))
	util.TestEqSlice(t, opts.Funcs, []string{"foo", "Bar", "BazBaz"})
}

func newTestParser() parseFn {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return newParser(fs)
}

var testMinArgs = []string{
	"-pkg", "main",
	"-fn", "Test",
	"-type", "any",
	"-in", "./foo.txt",
}
