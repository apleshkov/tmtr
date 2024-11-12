package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/apleshkov/tmtr/gen"
	"github.com/apleshkov/tmtr/util"
)

func TestMain(m *testing.M) {
	runCommand(exec.Command("go", "build", "-C", "../", "-o", "./e2e/tmtr"))
	defer os.Remove("./tmtr")
	m.Run()
}

func TestText(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	util.TestEq(
		t,
		generate(
			`{{. | html}}`,
			gen.GeneratorOptions{
				Mode:     gen.ModeText,
				DataType: "string",
				FnName:   "render",
			},
			[]file{
				newBasicMainFile("render", fmt.Sprintf("%q", `O'Reilly: How are <i>you</i>?`)),
			},
		),
		`O&#39;Reilly: How are &lt;i&gt;you&lt;/i&gt;?`,
	)
}

func TestHTML(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	util.TestEq(
		t,
		generate(
			`<html><head><title>{{.}}</title></head><body><p>{{.}}</p></body></html>`,
			gen.GeneratorOptions{
				Mode:     gen.ModeHTML,
				DataType: "string",
				FnName:   "render",
			},
			[]file{
				newBasicMainFile("render", fmt.Sprintf("%q", `O'Reilly: How are <i>you</i>?`)),
			},
		),
		`<html><head><title>O&#39;Reilly: How are &lt;i&gt;you&lt;/i&gt;?</title></head><body><p>O&#39;Reilly: How are &lt;i&gt;you&lt;/i&gt;?</p></body></html>`,
	)
}

func TestMaybeValue(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	util.TestEq(
		t,
		generate(
			`{{maybe foo .}}`,
			gen.GeneratorOptions{
				Mode:     gen.ModeHTML,
				DataType: "string",
				Funcs:    []string{"foo"},
				FnName:   "render",
			},
			[]file{
				newBasicMainFile("render", fmt.Sprintf("%q", `bar`)),
				{
					name:    "foo.go",
					content: "package main\nfunc foo(s string) (string, error) { return s, nil }",
				},
			},
		),
		"bar",
	)
}

func TestMaybeError(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	util.TestEq(
		t,
		generate(
			`{{maybe foo .}}`,
			gen.GeneratorOptions{
				Mode:     gen.ModeHTML,
				DataType: "string",
				Funcs:    []string{"foo"},
				FnName:   "render",
			},
			[]file{
				newBasicMainFile("render", fmt.Sprintf("%q", `bar`)),
				{
					name:    "foo.go",
					content: "package main\nimport \"errors\"\nfunc foo(s string) (string, error) { return \"\", errors.New(\"invalid arg: \"+s) }",
				},
			},
		),
		"invalid arg: bar\n",
	)
}

func TestCallMethod(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	util.TestEq(
		t,
		generate(
			`{{call .foo}}`,
			gen.GeneratorOptions{
				Mode:     gen.ModeHTML,
				DataType: "data",
				FnName:   "render",
			},
			[]file{
				newBasicMainFile("render", "data{}"),
				{
					name:    "data.go",
					content: "package main\ntype data struct {}\nfunc (data) foo() string { return \"bar\" }",
				},
			},
		),
		"bar",
	)
}

type file struct {
	name, content string
}

func newMainFile(content string) file {
	return file{
		name:    "main.go",
		content: content,
	}
}

func newBasicMainFile(fn, data string) file {
	return newMainFile(
		fmt.Sprintf("package main\nimport \"os\"\nfunc main() { %s(os.Stdout, %s, os.Stdout) }\n", fn, data),
	)
}

func runCommand(cmd *exec.Cmd) {
	var ew strings.Builder
	cmd.Stderr = &ew
	if err := cmd.Run(); err != nil {
		panic(ew.String())
	}
}

func runCommandFunc(c func() *exec.Cmd) {
	runCommand(c())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func mustx[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func generate(tmpl string, opts gen.GeneratorOptions, files []file) string {
	var testName string
	if pc, _, _, ok := runtime.Caller(1); ok {
		n := runtime.FuncForPC(pc).Name()
		if i := strings.LastIndex(n, "."); i != -1 {
			n := n[(i + 1):]
			testName = strings.ToLower(n)
		}
	}
	if len(testName) == 0 {
		panic("invalid test name")
	}
	tmpdir := path.Join(".", testName)
	must(os.MkdirAll(tmpdir, os.ModePerm))
	defer func() {
		must(os.RemoveAll(tmpdir))
	}()
	var modeStr string
	switch opts.Mode {
	case gen.ModeHTML:
		modeStr = "html"
	case gen.ModeText:
		modeStr = "text"
	default:
		panic(fmt.Sprintf("invalid mode: %v", opts.Mode))
	}
	inFile := mustx(createInputFile(tmpdir, tmpl, modeStr))
	outFile := path.Join(tmpdir, "render.go")
	runCommandFunc(func() *exec.Cmd {
		args := []string{
			"-pkg", "main",
			"-mode", modeStr,
			"-fn", opts.FnName,
			"-type", opts.DataType,
			"-in", inFile,
			"-out", outFile,
		}
		for _, v := range opts.Tmpls {
			if len(v.DataType) > 0 {
				args = append(args, "-tpl", v.Name+":"+v.DataType)
			} else {
				args = append(args, "-tpl", v.Name)
			}
		}
		for _, v := range opts.Imports {
			args = append(args, "-import", v)
		}
		for _, v := range opts.Funcs {
			args = append(args, "-tplfn", v)
		}
		cmd := exec.Command("./tmtr", args...)
		return cmd
	})
	for _, f := range files {
		must(os.WriteFile(path.Join(tmpdir, f.name), []byte(f.content), os.ModePerm))
	}
	ver := strings.Replace(runtime.Version(), "go", "go ", 1)
	funcsdep := fmt.Sprintf("%s v0.0.0-unpublished", gen.FuncsPkgPath)
	gomod := fmt.Sprintf("module %s\n%s\nrequire %s\nreplace %s => ../../funcs", testName, ver, funcsdep, funcsdep)
	must(os.WriteFile(path.Join(tmpdir, "go.mod"), []byte(gomod), os.ModePerm))
	var buf strings.Builder
	runCommandFunc(func() *exec.Cmd {
		cmd := exec.Command("go", "run", "-C", tmpdir, ".")
		cmd.Stdout = &buf
		return cmd
	})
	return buf.String()
}

func createInputFile(dir, src string, ext string) (string, error) {
	p := path.Join(dir, "input."+ext)
	err := os.WriteFile(p, []byte(src), os.ModePerm)
	return p, err
}
