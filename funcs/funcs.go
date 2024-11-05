package funcs

import (
	"fmt"
	"html/template"
	"io"
	"strings"
)

func Write(w io.Writer, v any, ew io.Writer) {
	if s, ok := v.(string); ok {
		writeString(w, s, ew)
	} else {
		writeString(w, fmt.Sprint(v), ew)
	}
}

func writeString(w io.Writer, s string, ew io.Writer) {
	if _, err := io.WriteString(w, s); err != nil && ew != nil {
		fmt.Fprintln(ew, err)
	}
}

func IsTrue(x any) bool {
	truth, _ := template.IsTrue(x)
	return truth
}

func IsNotTrue(x any) bool {
	return !IsTrue(x)
}

// Computes the Boolean AND of its arguments, returning the
// first false argument it encounters, or the last argument.
func And(args ...any) any {
	for _, a := range args {
		if !IsTrue(a) {
			return a
		}
	}
	l := len(args)
	return args[l-1]
}

// Computes the Boolean OR of its arguments, returning the
// first true argument it encounters, or the last argument.
func Or(args ...any) any {
	for _, a := range args {
		if IsTrue(a) {
			return a
		}
	}
	l := len(args)
	return args[l-1]
}

func MayBe[T any](ew io.Writer, fn func() (T, error)) T {
	v, err := fn()
	if err != nil && ew != nil {
		fmt.Fprintln(ew, err)
	}
	return v
}

func AttrEscaper(ew io.Writer, data ...any) string      { return attrEscaper(ew, data...) }
func CommentEscaper(ew io.Writer, data ...any) string   { return commentEscaper(ew, data...) }
func CSSEscaper(ew io.Writer, data ...any) string       { return cssEscaper(ew, data...) }
func CSSValueFilter(ew io.Writer, data ...any) string   { return cssValueFilter(ew, data...) }
func HTMLNameFilter(ew io.Writer, data ...any) string   { return htmlNameFilter(ew, data...) }
func HTMLEscaper(ew io.Writer, data ...any) string      { return htmlEscaper(ew, data...) }
func JSRegexpEscaper(ew io.Writer, data ...any) string  { return jsRegexpEscaper(ew, data...) }
func JSStrEscaper(ew io.Writer, data ...any) string     { return jsStrEscaper(ew, data...) }
func JSTmplLitEscaper(ew io.Writer, data ...any) string { return jsTmplLitEscaper(ew, data...) }
func JSValEscaper(ew io.Writer, data ...any) string     { return jsValEscaper(ew, data...) }
func HTMLNospaceEscaper(ew io.Writer, data ...any) string {
	return htmlNospaceEscaper(ew, data...)
}
func RCDataEscaper(ew io.Writer, data ...any) string { return rcdataEscaper(ew, data...) }
func SrcsetFilterAndEscaper(ew io.Writer, data ...any) string {
	return srcsetFilterAndEscaper(ew, data...)
}
func URLEscaper(ew io.Writer, data ...any) string    { return urlEscaper(ew, data...) }
func URLFilter(ew io.Writer, data ...any) string     { return urlFilter(ew, data...) }
func URLNormalizer(ew io.Writer, data ...any) string { return urlNormalizer(ew, data...) }
func EvalArgs(ew io.Writer, data ...any) string      { return evalArgs(ew, data...) }

var attrEscaper = newTmplFn("_html_template_attrescaper")
var commentEscaper = newTmplFn("_html_template_commentescaper")
var cssEscaper = newTmplFn("_html_template_cssescaper")
var cssValueFilter = newTmplFn("_html_template_cssvaluefilter")
var htmlNameFilter = newTmplFn("_html_template_htmlnamefilter")
var htmlEscaper = newTmplFn("_html_template_htmlescaper")
var jsRegexpEscaper = newTmplFn("_html_template_jsregexpescaper")
var jsStrEscaper = newTmplFn("_html_template_jsstrescaper")
var jsTmplLitEscaper = newTmplFn("_html_template_jstmpllitescaper")
var jsValEscaper = newTmplFn("_html_template_jsvalescaper")
var htmlNospaceEscaper = newTmplFn("_html_template_nospaceescaper")
var rcdataEscaper = newTmplFn("_html_template_rcdataescaper")
var srcsetFilterAndEscaper = newTmplFn("_html_template_srcsetescaper")
var urlEscaper = newTmplFn("_html_template_urlescaper")
var urlFilter = newTmplFn("_html_template_urlfilter")
var urlNormalizer = newTmplFn("_html_template_urlnormalizer")
var evalArgs = newTmplFn("_eval_args_")

var dummyFn = func(...any) string { return "" }

func newTmplFn(f string) func(ew io.Writer, data ...any) string {
	txt := fmt.Sprintf("{{%s .}}", f)
	tmpl := template.New(f).Funcs(template.FuncMap{
		// This dummy func needs to not fail the parsing, it will be redefined
		f: dummyFn,
	})
	tmpl = template.Must(tmpl.Parse(txt))
	return func(ew io.Writer, data ...any) string {
		var buf strings.Builder
		for _, a := range data {
			if err := tmpl.Execute(&buf, a); err != nil && ew != nil {
				fmt.Fprintln(ew, err)
			}
		}
		return buf.String()
	}
}
