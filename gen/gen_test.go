package gen

import (
	"go/printer"
	"go/token"
	"strings"
	"testing"

	"github.com/apleshkov/tmtr/util"
)

func TestTextStmt(t *testing.T) {
	testFuncOutput(
		t, ModeText,
		"Hello",
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
			tmtr.Write(output, "Hello", errOutput)
		}`,
	)
	testFuncOutput(
		t, ModeHTML,
		"Hello\nWorld",
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
			tmtr.Write(output, "Hello\nWorld", errOutput)
		}`,
	)
}

func TestComments(t *testing.T) {
	testFuncOutput(
		t, ModeHTML,
		`{{/* a comment */}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{- /* a comment with white space trimmed from preceding and following text */ -}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
        }`,
	)
}

func TestAttrs(t *testing.T) {
	testFuncOutput(
		t, ModeHTML,
		`<span class="{{.}}">{{.}}</span>`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "<span class=\"", errOutput)
            tmtr.Write(output, tmtr.AttrEscaper(errOutput, data), errOutput)
            tmtr.Write(output, "\">", errOutput)
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, data), errOutput)
            tmtr.Write(output, "</span>", errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`<a href="{{.}}" style="{{.}}">{{.}}</a>`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "<a href=\"", errOutput)
            tmtr.Write(output, tmtr.AttrEscaper(errOutput, tmtr.URLNormalizer(errOutput, tmtr.URLFilter(errOutput, data))), errOutput)
            tmtr.Write(output, "\" style=\"", errOutput)
            tmtr.Write(output, tmtr.AttrEscaper(errOutput, tmtr.CSSValueFilter(errOutput, data)), errOutput)
            tmtr.Write(output, "\">", errOutput)
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, data), errOutput)
            tmtr.Write(output, "</a>", errOutput)
        }`,
	)
}

func TestActions(t *testing.T) {
	testFuncOutput(
		t, ModeText,
		`{{printf "%q" "output"}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, fmt.Sprintf("%q", "output"), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeText,
		`{{"output" | printf "%q"}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, fmt.Sprintf("%q", "output"), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeText,
		`{{printf "%q" (print "out" "put")}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, fmt.Sprintf("%q", fmt.Sprint("out", "put")), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeText,
		`{{"put" | printf "%s%s" "out" | printf "%q"}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, fmt.Sprintf("%q", fmt.Sprintf("%s%s", "out", "put")), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeText,
		`{{"output" | printf "%s" | printf "%q"}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, fmt.Sprintf("%q", fmt.Sprintf("%s", "output")), errOutput)
        }`,
	)
}

func TestFunctions(t *testing.T) {
	testFuncOutput(
		t, ModeHTML,
		`{{and 0 1 2}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, tmtr.And(0, 1, 2)), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{or 0 1 2}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, tmtr.Or(0, 1, 2)), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{call .X.Y 1 2}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, data.X.Y(1, 2)), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{call .}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, data()), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{html .}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, ht.HTMLEscaper(tmtr.EvalArgs(errOutput, data)), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{html 1 2 3}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, ht.HTMLEscaper(tmtr.EvalArgs(errOutput, 1, 2, 3)), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{. | html}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, ht.HTMLEscaper(data), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{index . 0}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, data[0]), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{index . 1 2 3}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, data[1][2][3]), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{slice . 1 2}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, data[1:2]), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{slice .}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, data[:]), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{slice . 1}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, data[1:]), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{slice . 1 2 3}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, data[1:2:3]), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{js .}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, ht.JSEscaper(data)), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{js 1 2 3}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, ht.JSEscaper(1, 2, 3)), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{len .}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, len(data)), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{not .}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, tmtr.IsNotTrue(data)), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{print 1 2 3}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, fmt.Sprint(1, 2, 3)), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{printf "%d %d %d" 1 2 3}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, fmt.Sprintf("%d %d %d", 1, 2, 3)), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{println 1 2 3}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, fmt.Sprintln(1, 2, 3)), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{urlquery 1 2 3}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, ht.URLQueryEscaper(tmtr.EvalArgs(errOutput, 1, 2, 3))), errOutput)
        }`,
	)
}

func TestMayBe(t *testing.T) {
	testFuncOutput(
		t, ModeText,
		`{{maybe .}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
	        tmtr.Write(output, tmtr.MayBe(errOutput, func() (any, error) {
	            return data
	        }), errOutput)
	    }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{maybe . 1 2 3}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
	        tmtr.Write(output, tmtr.HTMLEscaper(errOutput, tmtr.MayBe(errOutput, func() (any, error) {
	            return data(1, 2, 3)
	        })), errOutput)
	    }`,
	)
}

func TestComparisonOps(t *testing.T) {
	testFuncOutput(
		t, ModeHTML,
		`{{eq . 1}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, data == 1), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{eq . 1 2 3}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, data == 1 || data == 2 || data == 3), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{ne . 1}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, data != 1), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{lt . 1}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, data < 1), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{le . 1}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, data <= 1), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{gt . 1}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, data > 1), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{ge . 1}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, data >= 1), errOutput)
        }`,
	)
}

func TestVars(t *testing.T) {
	testFuncOutput(
		t, ModeText,
		"{{$x := .}}{{$x}}",
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            x := data
            tmtr.Write(output, x, errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeText,
		"{{$a = .}}{{$a}}",
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            a = data
            tmtr.Write(output, a, errOutput)
        }`,
	)
}

func TestIfStmt(t *testing.T) {
	testFuncOutput(
		t, ModeText,
		"{{if .}}{{.}}{{end}}",
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            if tmtr.IsTrue(data) {
                tmtr.Write(output, data, errOutput)
            }
        }`,
	)
	testFuncOutput(
		t, ModeText,
		"{{if eq . 1}}{{.}}{{end}}",
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            if tmtr.IsTrue(data == 1) {
                tmtr.Write(output, data, errOutput)
            }
        }`,
	)
	testFuncOutput(
		t, ModeText,
		"{{if .}}{{.}}{{else}}Empty{{end}}",
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            if tmtr.IsTrue(data) {
                tmtr.Write(output, data, errOutput)
            } else {
                tmtr.Write(output, "Empty", errOutput)
            }
        }`,
	)
	testFuncOutput(
		t, ModeText,
		"{{if eq . 1}}1{{else if eq . 2}}2{{else}}{{.}}{{end}}",
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            if tmtr.IsTrue(data == 1) {
                tmtr.Write(output, "1", errOutput)
            } else {
                if tmtr.IsTrue(data == 2) {
                    tmtr.Write(output, "2", errOutput)
                } else {
                    tmtr.Write(output, data, errOutput)
                }
            }
        }`,
	)
	testFuncOutput(
		t, ModeText,
		"{{if $x := .}}{{$x}}{{end}}",
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            if x := data; tmtr.IsTrue(x) {
                tmtr.Write(output, x, errOutput)
            }
        }`,
	)
}

func TestRangeStmt(t *testing.T) {
	testFuncOutput(
		t, ModeText,
		"{{range .}}{{.}}{{end}}",
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            if list := data; tmtr.IsTrue(list) {
                for _, elem := range list {
                    tmtr.Write(output, elem, errOutput)
                }
            }
        }`,
	)
	testFuncOutput(
		t, ModeText,
		"{{range $v := .Items}}{{$v}}{{.}}{{end}}",
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            if list := data.Items; tmtr.IsTrue(list) {
                for _, v := range list {
                    tmtr.Write(output, v, errOutput)
                    tmtr.Write(output, v, errOutput)
                }
            }
        }`,
	)
	testFuncOutput(
		t, ModeText,
		"{{range $i, $v := .Items}}{{$i}}{{$v}}{{.}}{{end}}",
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            if list := data.Items; tmtr.IsTrue(list) {
                for i, v := range list {
                    tmtr.Write(output, i, errOutput)
                    tmtr.Write(output, v, errOutput)
                    tmtr.Write(output, v, errOutput)
                }
            }
        }`,
	)
	testFuncOutput(
		t, ModeText,
		"{{range .Items}}{{.}}{{else}}Empty{{end}}",
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            if list := data.Items; tmtr.IsTrue(list) {
                for _, elem := range list {
                    tmtr.Write(output, elem, errOutput)
                }
            } else {
                tmtr.Write(output, "Empty", errOutput)
            }
        }`,
	)
	testFuncOutput(
		t, ModeText,
		"{{range .Items}}{{.}}{{else}}{{$list := 0}}{{$elem := 1}}{{end}}",
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            if list := data.Items; tmtr.IsTrue(list) {
                for _, elem := range list {
                    tmtr.Write(output, elem, errOutput)
                }
            } else {
                list := 0
                elem := 1
            }
        }`,
	)
	testFuncOutput(
		t, ModeText,
		"{{range .}}{{if .}}{{break}}{{else}}{{continue}}{{end}}{{end}}",
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            if list := data; tmtr.IsTrue(list) {
                for _, elem := range list {
                    if tmtr.IsTrue(elem) {
                        break
                    } else {
                        continue
                    }
                }
            }
        }`,
	)
}

func TestWithStmt(t *testing.T) {
	testFuncOutput(
		t, ModeText,
		`{{with .Foo}}{{.}}{{end}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            if with := data.Foo; tmtr.IsTrue(with) {
                tmtr.Write(output, with, errOutput)
            }
        }`,
	)
	testFuncOutput(
		t, ModeText,
		`{{with .Foo}}{{.}}{{else}}{{.}}{{end}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            if with := data.Foo; tmtr.IsTrue(with) {
                tmtr.Write(output, with, errOutput)
            } else {
                tmtr.Write(output, data, errOutput)
            }
        }`,
	)
	testFuncOutput(
		t, ModeText,
		`{{with .Foo}}{{.}}{{else with .Bar}}{{.}}{{end}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            if with := data.Foo; tmtr.IsTrue(with) {
                tmtr.Write(output, with, errOutput)
            } else {
                if with := data.Bar; tmtr.IsTrue(with) {
                    tmtr.Write(output, with, errOutput)
                }
            }
        }`,
	)
	testFuncOutput(
		t, ModeText,
		`{{with $foo := .Foo}}{{.}}{{end}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            if foo := data.Foo; tmtr.IsTrue(foo) {
                tmtr.Write(output, foo, errOutput)
            }
        }`,
	)
}

func TestParens(t *testing.T) {
	testFuncOutput(
		t, ModeText,
		`{{(len (call .).Field)}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, len(data().Field), errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeText,
		`{{$arg1 := 0}}{{$arg2 := 1}}{{print (.F1 $arg1) (.F2 $arg2) (.StructValuedMethod "arg").Field}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            arg1 := 0
            arg2 := 1
            tmtr.Write(output, fmt.Sprint(data.F1(arg1, data.F2(arg2, data.StructValuedMethod("arg").Field))), errOutput)
        }`,
	)
}

func TestParentData(t *testing.T) {
	testFuncOutput(
		t, ModeText,
		`{{$}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, data, errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeText,
		`{{if .}}{{$}}{{else}}{{.}}{{$}}{{end}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            if tmtr.IsTrue(data) {
                tmtr.Write(output, data, errOutput)
            } else {
                tmtr.Write(output, data, errOutput)
                tmtr.Write(output, data, errOutput)
            }
        }`,
	)
	testFuncOutput(
		t, ModeText,
		`{{range .}}{{.}}{{$}}{{else}}{{.}}{{$}}{{end}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            if list := data; tmtr.IsTrue(list) {
                for _, elem := range list {
                    tmtr.Write(output, elem, errOutput)
                    tmtr.Write(output, list, errOutput)
                }
            } else {
                tmtr.Write(output, data, errOutput)
                tmtr.Write(output, data, errOutput)
            }
        }`,
	)
	testFuncOutput(
		t, ModeText,
		`{{with .}}{{.}}{{$}}{{else}}{{.}}{{$}}{{end}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            if with := data; tmtr.IsTrue(with) {
                tmtr.Write(output, with, errOutput)
                tmtr.Write(output, data, errOutput)
            } else {
                tmtr.Write(output, data, errOutput)
                tmtr.Write(output, data, errOutput)
            }
        }`,
	)
	testFuncOutput(
		t, ModeText,
		`{{range $v := .}}
			{{with .}}
				{{.}}{{$}}
				{{if .}}
					{{.}}{{$}}
				{{else}}
					{{.}}{{$}}
				{{end}}
			{{else}}
				{{.}}{{$}}
			{{end}}
		{{else}}
			{{.}}{{$}}
		{{end}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            if list := data; tmtr.IsTrue(list) {
                for _, v := range list {
                    tmtr.Write(output, "\n\t\t\t", errOutput)
                    if with := v; tmtr.IsTrue(with) {
                        tmtr.Write(output, "\n\t\t\t\t", errOutput)
                        tmtr.Write(output, with, errOutput)
                        tmtr.Write(output, v, errOutput)
                        tmtr.Write(output, "\n\t\t\t\t", errOutput)
                        if tmtr.IsTrue(with) {
                            tmtr.Write(output, "\n\t\t\t\t\t", errOutput)
                            tmtr.Write(output, with, errOutput)
                            tmtr.Write(output, v, errOutput)
                            tmtr.Write(output, "\n\t\t\t\t", errOutput)
                        } else {
                            tmtr.Write(output, "\n\t\t\t\t\t", errOutput)
                            tmtr.Write(output, with, errOutput)
                            tmtr.Write(output, v, errOutput)
                            tmtr.Write(output, "\n\t\t\t\t", errOutput)
                        }
                        tmtr.Write(output, "\n\t\t\t", errOutput)
                    } else {
                        tmtr.Write(output, "\n\t\t\t\t", errOutput)
                        tmtr.Write(output, v, errOutput)
                        tmtr.Write(output, list, errOutput)
                        tmtr.Write(output, "\n\t\t\t", errOutput)
                    }
                    tmtr.Write(output, "\n\t\t", errOutput)
                }
            } else {
                tmtr.Write(output, "\n\t\t\t", errOutput)
                tmtr.Write(output, data, errOutput)
                tmtr.Write(output, data, errOutput)
                tmtr.Write(output, "\n\t\t", errOutput)
            }
        }`,
	)
}

func TestEscapers(t *testing.T) {
	// AttrEscaper
	testFuncOutput(
		t, ModeHTML,
		`<x y="{{.}}">`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "<x y=\"", errOutput)
            tmtr.Write(output, tmtr.AttrEscaper(errOutput, data), errOutput)
            tmtr.Write(output, "\">", errOutput)
        }`,
	)
	// CommentEscaper
	testFuncOutput(
		t, ModeHTML,
		`<!--{{.}}-->`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "", errOutput)
            tmtr.Write(output, tmtr.CommentEscaper(errOutput, data), errOutput)
            tmtr.Write(output, "", errOutput)
        }`,
	)
	// CSSEscaper
	testFuncOutput(
		t, ModeHTML,
		`<style>x { y: "{{.}}" }</style>`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "<style>x { y: \"", errOutput)
            tmtr.Write(output, tmtr.CSSEscaper(errOutput, tmtr.URLFilter(errOutput, data)), errOutput)
            tmtr.Write(output, "\" }</style>", errOutput)
        }`,
	)
	// CSSValueFilter
	testFuncOutput(
		t, ModeHTML,
		`<img style="{{.}}">`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "<img style=\"", errOutput)
            tmtr.Write(output, tmtr.AttrEscaper(errOutput, tmtr.CSSValueFilter(errOutput, data)), errOutput)
            tmtr.Write(output, "\">", errOutput)
        }`,
	)
	// HTMLNameFilter
	testFuncOutput(
		t, ModeHTML,
		`<x{{.}}>`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "<x", errOutput)
            tmtr.Write(output, tmtr.HTMLNameFilter(errOutput, data), errOutput)
            tmtr.Write(output, ">", errOutput)
        }`,
	)
	// HTMLEscaper
	testFuncOutput(
		t, ModeHTML,
		`<x>{{.}}</x>`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "<x>", errOutput)
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, data), errOutput)
            tmtr.Write(output, "</x>", errOutput)
		}`,
	)
	// JSRegexpEscaper
	testFuncOutput(
		t, ModeHTML,
		`<script>(/{{.}}/)</script>`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "<script>(/", errOutput)
            tmtr.Write(output, tmtr.JSRegexpEscaper(errOutput, data), errOutput)
            tmtr.Write(output, "/)</script>", errOutput)
        }`,
	)
	// JSStrEscaper
	testFuncOutput(
		t, ModeHTML,
		`<a onclick="'{{.}}'">`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "<a onclick=\"'", errOutput)
            tmtr.Write(output, tmtr.JSStrEscaper(errOutput, data), errOutput)
            tmtr.Write(output, "'\">", errOutput)
        }`,
	)
	// JSTmplLitEscaper
	testFuncOutput(
		t, ModeHTML,
		"<a onclick=\"`{{.}}`\">",
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "<a onclick=\"`+"`"+`", errOutput)
            tmtr.Write(output, tmtr.JSTmplLitEscaper(errOutput, data), errOutput)
            tmtr.Write(output, "`+"`"+`\">", errOutput)
        }`,
	)
	// JSValEscaper
	testFuncOutput(
		t, ModeHTML,
		`<script>{{.}}</script>`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "<script>", errOutput)
            tmtr.Write(output, tmtr.JSValEscaper(errOutput, data), errOutput)
            tmtr.Write(output, "</script>", errOutput)
        }`,
	)
	// HTMLNospaceEscaper
	testFuncOutput(
		t, ModeHTML,
		`<x y={{.}}>`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "<x y=", errOutput)
            tmtr.Write(output, tmtr.HTMLNospaceEscaper(errOutput, data), errOutput)
            tmtr.Write(output, ">", errOutput)
        }`,
	)
	// RCDataEscaper
	testFuncOutput(
		t, ModeHTML,
		`<title>{{.}}</title>`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "<title>", errOutput)
            tmtr.Write(output, tmtr.RCDataEscaper(errOutput, data), errOutput)
            tmtr.Write(output, "</title>", errOutput)
        }`,
	)
	// SrcsetFilterAndEscaper
	testFuncOutput(
		t, ModeHTML,
		`<x srcset="{{.}}">`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "<x srcset=\"", errOutput)
            tmtr.Write(output, tmtr.AttrEscaper(errOutput, tmtr.SrcsetFilterAndEscaper(errOutput, data)), errOutput)
            tmtr.Write(output, "\">", errOutput)
        }`,
	)
	// URLEscaper
	testFuncOutput(
		t, ModeHTML,
		`<x href="/?{{.}}">`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "<x href=\"/?", errOutput)
            tmtr.Write(output, tmtr.AttrEscaper(errOutput, tmtr.URLEscaper(errOutput, data)), errOutput)
            tmtr.Write(output, "\">", errOutput)
        }`,
	)
	// URLFilter
	testFuncOutput(
		t, ModeHTML,
		`<x href="{{.}}">`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "<x href=\"", errOutput)
            tmtr.Write(output, tmtr.AttrEscaper(errOutput, tmtr.URLNormalizer(errOutput, tmtr.URLFilter(errOutput, data))), errOutput)
            tmtr.Write(output, "\">", errOutput)
        }`,
	)
	// URLNormalizer
	testFuncOutput(
		t, ModeHTML,
		`<x href="/{{.}}">`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "<x href=\"/", errOutput)
            tmtr.Write(output, tmtr.AttrEscaper(errOutput, tmtr.URLNormalizer(errOutput, data)), errOutput)
            tmtr.Write(output, "\">", errOutput)
        }`,
	)
}

func TestExternalTemplates(t *testing.T) {
	infos := []NamedTemplateInfo{
		{Name: "foo", DataType: "string"},
		{Name: "bar", DataType: ""},
	}
	testOutputWithOpts(
		t, newTestGeneratorOpts(ModeHTML, infos, nil, nil),
		`{{template "foo" .}}`,
		`func RenderTest(output io.Writer, data any, foo func(io.Writer, string, io.Writer), errOutput io.Writer) {
            foo(output, data, errOutput)
        }`,
		true, 0,
	)
	testOutputWithOpts(
		t, newTestGeneratorOpts(ModeHTML, infos, nil, nil),
		`{{template "bar"}}`,
		`func RenderTest(output io.Writer, data any, bar func(io.Writer, io.Writer), errOutput io.Writer) {
            bar(output, errOutput)
        }`,
		true, 0,
	)
	testOutputWithOpts(
		t, newTestGeneratorOpts(ModeText, infos, nil, nil),
		`{{template "foo" .max 1 2 3}}{{if .flag}}{{template "bar"}}{{end}}`,
		`func RenderTest(output io.Writer, data any, bar func(io.Writer, io.Writer), foo func(io.Writer, string, io.Writer), errOutput io.Writer) {
	        foo(output, data.max(1, 2, 3), errOutput)
	        if tmtr.IsTrue(data.flag) {
	            bar(output, errOutput)
	        }
	    }`,
		true, 0,
	)
	testOutputWithOpts(
		t, newTestGeneratorOpts(ModeHTML, infos, nil, nil),
		`{{template "foo" .max 1 2 3}}{{if .flag}}{{template "bar"}}{{end}}`,
		`func RenderTest(output io.Writer, data any, bar func(io.Writer, io.Writer), foo func(io.Writer, string, io.Writer), errOutput io.Writer) {
            foo(output, data.max(1, 2, 3), errOutput)
            if tmtr.IsTrue(data.flag) {
                bar(output, errOutput)
            }
        }`,
		true, 0,
	)
}

func TestMangledTemplates(t *testing.T) {
	testFuncOutput( // unknown
		t, ModeHTML,
		`<html><head><title>{{template "title" .Content.Data}}</title></head><body></body></html>`,
		`func RenderTest(output io.Writer, data any, title func(io.Writer, any, io.Writer), errOutput io.Writer) {
            tmtr.Write(output, "<html><head><title>", errOutput)
            title(output, data.Content.Data, errOutput)
            tmtr.Write(output, "</title></head><body></body></html>", errOutput)
        }`,
	)
	testOutputWithOpts( // external
		t, newTestGeneratorOpts(ModeHTML, []NamedTemplateInfo{{Name: "title", DataType: "string"}}, nil, nil),
		`<html><head><title>{{template "title" .Content.Data}}</title></head><body></body></html>`,
		`func RenderTest(output io.Writer, data any, title func(io.Writer, string, io.Writer), errOutput io.Writer) {
            tmtr.Write(output, "<html><head><title>", errOutput)
            title(output, data.Content.Data, errOutput)
            tmtr.Write(output, "</title></head><body></body></html>", errOutput)
        }`,
		true, 0,
	)
}

func TestUnknownTemplate(t *testing.T) {
	testOutput(
		t, ModeText,
		`{{template "withArg" .}}{{template "noArg"}}`,
		`package main

        import io "io"

        func RenderTest(output io.Writer, data any, noArg func(io.Writer, any, io.Writer), withArg func(io.Writer, any, io.Writer), errOutput io.Writer) {
            withArg(output, data, errOutput)
            noArg(output, errOutput)
        }`,
	)
	testOutput(
		t, ModeHTML,
		`{{template "withArg" .}}{{template "noArg"}}`,
		`package main

        import io "io"

        func RenderTest(output io.Writer, data any, noArg func(io.Writer, any, io.Writer), withArg func(io.Writer, any, io.Writer), errOutput io.Writer) {
            withArg(output, data, errOutput)
            noArg(output, errOutput)
        }`,
	)
}

func TestDefinedTemplates(t *testing.T) {
	testFuncOutput(
		t, ModeText,
		`{{define "foo"}}<p>{{.}}</p>{{end}}{{template "foo" .}}{{define "bar"}}bar{{end}}{{template "bar" .}}`,
		`func RenderTest(output io.Writer, data any, bar func(io.Writer, any, io.Writer), foo func(io.Writer, any, io.Writer), errOutput io.Writer) {
            foo(output, data, errOutput)
            bar(output, data, errOutput)
        }
        func RenderTestBar(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "bar", errOutput)
        }
        func RenderTestFoo(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "<p>", errOutput)
            tmtr.Write(output, data, errOutput)
            tmtr.Write(output, "</p>", errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeHTML,
		`{{define "foo"}}<p>{{.}}</p>{{end}}{{template "foo" .}}{{define "bar"}}bar{{end}}{{template "bar" .}}`,
		`func RenderTest(output io.Writer, data any, bar func(io.Writer, any, io.Writer), foo func(io.Writer, any, io.Writer), errOutput io.Writer) {
            foo(output, data, errOutput)
            bar(output, data, errOutput)
        }
        func RenderTestBar(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "bar", errOutput)
        }
        func RenderTestFoo(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "<p>", errOutput)
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, data), errOutput)
            tmtr.Write(output, "</p>", errOutput)
        }`,
	)
}

func TestNestedTemplates(t *testing.T) {
	testFuncOutput(
		t, ModeHTML,
		`
		{{define "T1"}}ONE{{end}}
		{{define "T2"}}TWO: {{template "T1" .}}{{end}}
		{{define "T3"}}{{template "T1" .}} {{template "T2" .}}{{end}}
		{{template "T3" .}}
		`,
		`func RenderTest(output io.Writer, data any, T3 func(io.Writer, any, io.Writer), errOutput io.Writer) {
            tmtr.Write(output, "\n\t\t", errOutput)
            tmtr.Write(output, "\n\t\t", errOutput)
            tmtr.Write(output, "\n\t\t", errOutput)
            tmtr.Write(output, "\n\t\t", errOutput)
            T3(output, data, errOutput)
            tmtr.Write(output, "\n\t\t", errOutput)
        }
        func RenderTestT1(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "ONE", errOutput)
        }
        func RenderTestT2(output io.Writer, data any, T1 func(io.Writer, any, io.Writer), errOutput io.Writer) {
            tmtr.Write(output, "TWO: ", errOutput)
            T1(output, data, errOutput)
        }
        func RenderTestT3(output io.Writer, data any, T1 func(io.Writer, any, io.Writer), T2 func(io.Writer, any, io.Writer), errOutput io.Writer) {
            T1(output, data, errOutput)
            tmtr.Write(output, " ", errOutput)
            T2(output, data, errOutput)
        }`,
	)
}

func TestInternalImports(t *testing.T) {
	testFuncOutput(
		t, ModeText,
		`text`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, "text", errOutput)
        }`,
	)
	testFuncOutput(
		t, ModeText,
		`{{print .}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, fmt.Sprint(data), errOutput)
        }`,
	)
	testOutput(
		t, ModeText,
		`{{. | html}}`,
		`package main

        import (
            ht "html/template"
            io "io"
			tmtr "github.com/apleshkov/tmtr/funcs"
        )

        func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, ht.HTMLEscaper(data), errOutput)
        }`,
	)
}

func TestExternalImports(t *testing.T) {
	testOutputWithOpts(
		t, newTestGeneratorOpts(ModeText, nil, []string{"math", "net", "net/http", "unicode/utf16"}, nil),
		`{{math.Max 1 2}}`,
		`package main
        
        import (
            http "net/http"
            io "io"
            math "math"
            net "net"
			tmtr "github.com/apleshkov/tmtr/funcs"
            utf16 "unicode/utf16"
        )
        
        func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, math.Max(1, 2), errOutput)
        }`,
		false, 0,
	)
	testOutputWithOpts(
		t, newTestGeneratorOpts(ModeHTML, nil, []string{"io", "net/http", "net/http"}, nil),
		`{{http.MethodConnect}}`,
		`package main
        
        import (
            http "net/http"
            io "io"
			tmtr "github.com/apleshkov/tmtr/funcs"
        )
        
        func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, http.MethodConnect), errOutput)
        }`,
		false, 0,
	)
	// The "data" import conflicts with the "data" argument. The generator *doesn't*
	// rename "data.Data" to "data_.Data", because it doesn't know a user intention.
	testOutputWithOpts(
		t, newTestGeneratorOpts(ModeHTML, nil, []string{"output", "data"}, nil),
		`{{data.Data}}`,
		`package main

        import (
            data_ "data"
            io "io"
            output "output"
			tmtr "github.com/apleshkov/tmtr/funcs"
        )

        func RenderTest(output_ io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output_, tmtr.HTMLEscaper(errOutput, data.Data), errOutput)
        }`,
		false, 0,
	)
}

func TestExternalFuncs(t *testing.T) {
	testOutputWithOpts(
		t, newTestGeneratorOpts(ModeText, nil, nil, []string{"foo", "bar"}),
		`{{foo .}}{{. | bar}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, foo(data), errOutput)
            tmtr.Write(output, bar(data), errOutput)
        }`,
		true, 0,
	)
	testOutputWithOpts(
		t, newTestGeneratorOpts(ModeHTML, nil, nil, []string{"foo", "bar"}),
		`{{foo .}}{{. | bar}}`,
		`func RenderTest(output io.Writer, data any, errOutput io.Writer) {
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, foo(data)), errOutput)
            tmtr.Write(output, tmtr.HTMLEscaper(errOutput, bar(data)), errOutput)
        }`,
		true, 0,
	)
}

func newTestGeneratorOpts(mode Mode, tmpls []NamedTemplateInfo, imports []string, funcs []string) GeneratorOptions {
	return GeneratorOptions{
		Mode:     mode,
		Package:  "main",
		FnName:   "RenderTest",
		DataType: "any",
		Tmpls:    tmpls,
		Imports:  imports,
		Funcs:    funcs,
	}
}

func testOutput(t *testing.T, mode Mode, tmpl, expected string) {
	testOutputWithOpts(
		t, newTestGeneratorOpts(mode, nil, nil, nil),
		tmpl, expected, false,
		1,
	)
}

func testFuncOutput(t *testing.T, mode Mode, tmpl, expected string) {
	testOutputWithOpts(
		t, newTestGeneratorOpts(mode, nil, nil, nil),
		tmpl, expected, true,
		1,
	)
}

func testOutputWithOpts(t *testing.T, opts GeneratorOptions, tmpl, expected string, funcOnly bool, depth int) {
	f, err := generateFromText("test", tmpl, opts)
	if err != nil {
		t.Error(err)
		return
	}
	cfg := printer.Config{
		Mode:     printer.UseSpaces,
		Tabwidth: 4,
	}
	var buf strings.Builder
	if err := cfg.Fprint(&buf, token.NewFileSet(), f); err != nil {
		t.Error(err)
	}
	actual := buf.String()
	if funcOnly {
		actual = actual[strings.Index(actual, "func Render"):]
	}
	tab := strings.Repeat(" ", cfg.Tabwidth)
	exp := strings.Split(strings.ReplaceAll(expected, "\t", tab), "\n")
	for i, line := range exp {
		exp[i] = strings.TrimPrefix(line, tab+tab)
	}
	expected = strings.Join(exp, "\n")
	if actual != expected+"\n" {
		t.Logf("\n>>> INPUT >>>\n\n%s\n\n>>> ACTUAL >>>\n\n%s\n<<< EXPECTED <<<\n\n%s", tmpl, actual, expected)
		util.Fail(t, depth+1)
	}
}
