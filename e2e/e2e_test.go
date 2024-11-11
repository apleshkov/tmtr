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
			gen.ModeText,
			newStringData(`O'Reilly: How are <i>you</i>?`),
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
			gen.ModeHTML,
			newStringData(`O'Reilly: How are <i>you</i>?`),
		),
		`<html><head><title>O&#39;Reilly: How are &lt;i&gt;you&lt;/i&gt;?</title></head><body><p>O&#39;Reilly: How are &lt;i&gt;you&lt;/i&gt;?</p></body></html>`,
	)
}

type data struct {
	ty, val string
}

func newStringData(s string) data {
	return data{
		ty:  "string",
		val: fmt.Sprintf("%q", s),
	}
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

func generate(tmpl string, mode gen.Mode, data data) string {
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
	inFile := mustx(createInputFile(tmpdir, tmpl, mode))
	outFile := path.Join(tmpdir, "render.go")
	runCommand(
		exec.Command("./tmtr", "-pkg", "main", "-fn", "render", "-type", data.ty, "-in", inFile, "-out", outFile),
	)
	main := fmt.Sprintf("package main\nimport \"os\"\nfunc main() { render(os.Stdout, %s, os.Stdout) }\n", data.val)
	must(os.WriteFile(path.Join(tmpdir, "main.go"), []byte(main), os.ModePerm))
	ver := strings.Replace(runtime.Version(), "go", "go ", 1)
	funcsdep := fmt.Sprintf("%s v0.0.0-unpublished", gen.FuncsPkgPath)
	gomod := fmt.Sprintf("module %s\n%s\nrequire %s\nreplace %s => ../../funcs", testName, ver, funcsdep, funcsdep)
	println(gomod)
	must(os.WriteFile(path.Join(tmpdir, "go.mod"), []byte(gomod), os.ModePerm))
	var buf strings.Builder
	runCommandFunc(func() *exec.Cmd {
		cmd := exec.Command("go", "run", "-C", tmpdir, ".")
		cmd.Stdout = &buf
		return cmd
	})
	return buf.String()
}

func createInputFile(dir, src string, mode gen.Mode) (string, error) {
	name := "input."
	switch mode {
	case gen.ModeText:
		name += "txt"
	case gen.ModeHTML:
		name += "html"
	default:
		return "", fmt.Errorf("invalid mode: %d", mode)
	}
	p := path.Join(dir, name)
	err := os.WriteFile(p, []byte(src), os.ModePerm)
	return p, err
}
