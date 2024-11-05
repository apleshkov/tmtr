package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/apleshkov/tmtr/gen"
)

func newUsage(fs *flag.FlagSet) func() {
	wr := fs.Output()
	return func() {
		fmt.Fprintf(wr, "Usage of tmtr:\n")
		fmt.Fprintf(wr, "  tmtr [-pkg name] -fn name -type type -in file [-mode mode] [-out file] [-tpl name[:type] ...] [-import ...] [-tplfn ...]\n")
		fmt.Fprintf(wr, "\nExamples:\n")
		fmt.Fprintf(wr, "  # Basic usage:\n")
		fmt.Fprintf(wr, "  tmtr -pkg \"main\" -fn \"RenderIndex\" -type \"any\" -in \"./index.html\"\n")
		fmt.Fprintf(wr, "\n  # External templates:\n")
		fmt.Fprintf(wr, "  tmtr -pkg \"main\" -fn \"RenderIndex\" -type \"any\" -in \"./index.html\" -tpl \"foo:Foo\" -tpl \"foo:map[string]any\"\n")
		fmt.Fprintf(wr, "\n  # User template function and its neccessary import:\n")
		fmt.Fprintf(wr, "  tmtr -pkg \"main\" -fn \"RenderIndex\" -type \"any\" -in \"./index.html\" -import \"strconv\" -tplfn \"strconv.Atoi\"\n")
		fmt.Fprintf(wr, "\nFor more information, see:\n")
		fmt.Fprintf(wr, "  https://github.com/apleshkov/tmtr\n")
		fmt.Fprintf(wr, "\nFlags:\n")
		fs.PrintDefaults()
		fmt.Fprintf(wr, "  -h, -help\n    \tPrints this message\n")
	}
}

var Usage = newUsage(flag.CommandLine)

type strsVar []string

func (sv *strsVar) String() string {
	return strings.Join(*sv, ",")
}

func (sv *strsVar) Set(raw string) error {
	list := strings.Split(raw, ",")
	for _, s := range list {
		*sv = append(*sv, strings.TrimSpace(s))
	}
	return nil
}

type parseFn func(args []string) (*gen.GeneratorOptions, error)

func newParser(fs *flag.FlagSet) parseFn {
	fs.Usage = newUsage(fs)
	pkg := fs.String("pkg", os.Getenv("GOPACKAGE"), "package name; optional: $GOPACKAGE by default (is set by go:generate)")
	fnName := fs.String("fn", "", "[required] function name")
	dataType := fs.String("type", "", "[required] data type")
	inPath := fs.String("in", "", "path to the template file")
	modeStr := fs.String("mode", "", "'text' or 'html'; optional: 'html' is used if `in`'s extension ends with 'html' (e.g. 'foo.html', 'bar.gohtml'), 'text' otherwise")
	outPath := fs.String("out", "", "path to the output *.go file; optional: adds '.go' to the `in` filename (e.g. 'foo.html' -> 'foo.html.go')")
	var tpl strsVar
	fs.Var(&tpl, "tpl", `[multiple] external template with type, e.g. "foo:Foo"; comma-separated is also supported, e.g. "baz:string,quux:[]int"`)
	var imports strsVar
	fs.Var(&imports, "import", `[multiple] additional imports, e.g. "net/http"; comma-separated is also supported, e.g. "fmt,strings"`)
	var funcs strsVar
	fs.Var(&funcs, "tplfn", `[multiple] user template functions; comma-separated is also supported, e.g. "foo,bar"`)
	return func(args []string) (*gen.GeneratorOptions, error) {
		err := fs.Parse(args)
		if err != nil {
			return nil, err
		}
		if len(*pkg) == 0 {
			return nil, newBadFlag("no `pkg` provided")
		}
		if len(*fnName) == 0 {
			return nil, newBadFlag("no `fn` provided")
		}
		if len(*dataType) == 0 {
			return nil, newBadFlag("no `type` provided")
		}
		*dataType = strings.TrimSpace(*dataType)
		if len(*inPath) == 0 {
			return nil, newBadFlag("no `in` provided")
		}
		if len(*modeStr) == 0 {
			*modeStr = "text"
			if strings.HasSuffix(*inPath, "html") {
				*modeStr = "html"
			}
		}
		if len(*outPath) == 0 {
			*outPath = strings.TrimSpace(*inPath) + ".go"
		}
		var mode gen.Mode
		switch *modeStr {
		case "text":
			mode = gen.ModeText
		case "html":
			mode = gen.ModeHTML
		default:
			return nil, newBadFlag("unknown `mode`: " + *modeStr)
		}
		tmpls := make([]gen.NamedTemplateInfo, 0, len(tpl))
		for _, s := range tpl {
			if n, dt, ok := strings.Cut(s, ":"); ok {
				tmpls = append(tmpls, gen.NamedTemplateInfo{
					Name:     strings.TrimSpace(n),
					DataType: strings.TrimSpace(dt),
				})
				continue
			}
			tmpls = append(tmpls, gen.NamedTemplateInfo{
				Name:     strings.TrimSpace(s),
				DataType: "",
			})
		}
		return &gen.GeneratorOptions{
			InFile:   *inPath,
			OutFile:  *outPath,
			Mode:     mode,
			Package:  *pkg,
			FnName:   *fnName,
			DataType: *dataType,
			Tmpls:    tmpls,
			Imports:  imports,
			Funcs:    funcs,
		}, nil
	}
}

var parseCommandLine = newParser(flag.CommandLine)

func Parse() (*gen.GeneratorOptions, error) {
	return parseCommandLine(os.Args[1:])
}
