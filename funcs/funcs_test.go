package funcs

import (
	"fmt"
	"html/template"
	"strings"
	"testing"
)

func TestAnd(t *testing.T) {
	data := []struct{ x, a any }{
		{true, And(true)},       // last
		{false, And(false)},     // first encounter
		{1, And(4, 3, 2, 1)},    // last
		{"", And("2", "", "1")}, // first encounter
	}
	for _, cs := range data {
		if cs.x != cs.a {
			t.Errorf("%v != %v", cs.x, cs.a)
		}
	}
}

func TestOr(t *testing.T) {
	data := []struct{ x, a any }{
		{true, Or(true)},       // first encounter
		{false, Or(false)},     // last
		{0, Or(0, 0, 0, 0)},    // last
		{"1", Or("", "1", "")}, // first encounter
	}
	for _, cs := range data {
		if cs.x != cs.a {
			t.Errorf("%v != %v", cs.x, cs.a)
		}
	}
}

func TestEscapeHTMLAttr(t *testing.T) {
	data := []struct {
		x    any
		a, b string
	}{
		{
			x: "Ab09 zZ9",
			a: "Ab09 zZ9",
			b: "Ab09 zZ9",
		},
		{
			x: " \t\n\f\r\000",
			a: " \t\n\f\r\uFFFD",
			b: " \t\n\f\r\uFFFD",
		},
		{
			x: "⌘",
			a: "⌘",
			b: "⌘",
		},
		{
			x: "\"`'<>",
			a: "&#34;`&#39;&lt;&gt;",
			b: "&#34;`&#39;&lt;&gt;",
		},
		{
			x: "alert(1)",
			a: "alert(1)",
			b: "alert(1)",
		},
		{
			x: "\"><script>alert('pwned')</script>",
			a: "&#34;&gt;&lt;script&gt;alert(&#39;pwned&#39;)&lt;/script&gt;",
			b: "&#34;&gt;&lt;script&gt;alert(&#39;pwned&#39;)&lt;/script&gt;",
		},
		{
			// `<img style="{{.}}">` -> `EscapeHTMLAttr(FilterCSS(...))`
			x: FilterCSS(template.CSS("color: #000; font-size: 110%")),
			a: "color: #000; font-size: 110%",
			b: "color: #000; font-size: 110%",
		},
		{
			// `<a onblur="{{.}}">` -> `EscapeHTMLAttr(EscapeJS(...))`
			x: EscapeJS(nil, template.JS("alert(1);")),
			a: "alert(1);",
			b: "alert(1);",
		},
		{
			// `<img srcset="{{.}}">` -> `EscapeHTMLAttr(FilterAndEscapeSrcset(...))`,
			x: FilterAndEscapeSrcset(nil, template.Srcset("https://path/to/img.png 640w")),
			a: "https://path/to/img.png 640w",
			b: "https://path/to/img.png 640w",
		},
		{
			// `<a href="/?{{.}}">` -> `EscapeHTMLAttr(EscapeURL(...))`
			x: EscapeURL(template.URL("https://example.com?x=1&y='2'")),
			a: "https://example.com?x=1&amp;y=&#39;2&#39;",
			b: "https://example.com?x=1&amp;y=&#39;2&#39;",
		},
		{
			// `<a href="{{.}}">` -> `EscapeHTMLAttr(NormalizeURL(FilterURL(...)))`
			x: NormalizeURL(FilterURL(nil, template.URL("https://example.com?x=1&y='2'"))),
			a: "https://example.com?x=1&amp;y=%272%27",
			b: "https://example.com?x=1&amp;y=%272%27",
		},
		{
			// `<a href="/{{.}}">` -> `EscapeHTMLAttr(NormalizeURL(...))`
			x: EscapeHTMLAttr(NormalizeURL(template.URL("https://example.com?x=1&y='2'"))),
			a: "https://example.com?x=1&amp;amp;y=&amp;#39;2&amp;#39;",
			b: "https://example.com?x=1&amp;amp;y=&amp;#39;2&amp;#39;",
		},
	}
	builtin := builtinEscaper("_html_template_attrescaper")
	for _, cs := range data {
		if a := EscapeHTMLAttr(cs.x); a != cs.a {
			t.Errorf("%q: %q != %q", cs.x, a, cs.a)
		}
		if b := builtin(cs.x); b != cs.b {
			t.Errorf("%q: %q != %q", cs.x, b, cs.b)
		}
	}
	for _, cs := range data {
		if xstr, ok := cs.x.(string); ok {
			if x := EscapeHTMLAttr(template.HTML(xstr)); x != xstr {
				t.Errorf("HTMLAttr: %q != %q", x, xstr)
			}
			if x := EscapeHTMLAttr(template.HTMLAttr(xstr)); x != xstr {
				t.Errorf("HTMLAttr: %q != %q", x, xstr)
			}
		}
	}
}

func TestEscapeUnquotedHTMLAttr(t *testing.T) {
	data := []struct {
		x, a, b string
	}{
		{
			x: "Ab09 zZ9",
			a: `"Ab09 zZ9"`,
			b: `Ab09&#32;zZ9`,
		},
		{
			x: " \t\n\f\r\000",
			a: `" \t\n\f\r�"`,
			b: `&#32;&#9;&#10;&#12;&#13;&#xfffd;`,
		},
		{
			x: "⌘",
			a: `"⌘"`,
			b: "⌘",
		},
		{
			x: "\"`'<>",
			a: "\"&#34;`&#39;&lt;&gt;\"",
			b: `&#34;&#96;&#39;&lt;&gt;`,
		},
		{
			x: "foo\u0020bar",
			a: `"foo bar"`,
			b: `foo&#32;bar`,
		},
		{
			x: "''><script>alert('pwned')</script>",
			a: `"&#39;&#39;&gt;&lt;script&gt;alert(&#39;pwned&#39;)&lt;/script&gt;"`,
			b: `&#39;&#39;&gt;&lt;script&gt;alert(&#39;pwned&#39;)&lt;/script&gt;`,
		},
	}
	builtin := builtinEscaper("_html_template_nospaceescaper")
	for _, cs := range data {
		if a := EscapeUnquotedHTMLAttr(cs.x); a != cs.a {
			t.Errorf("%q: `%v` != `%v`", cs.x, a, cs.a)
		}
		if b := builtin(cs.x); b != cs.b {
			t.Errorf("%q: `%v` != `%v`", cs.x, b, cs.b)
		}
	}
	for _, cs := range data {
		if x := EscapeUnquotedHTMLAttr(template.HTML(cs.x)); x != cs.x {
			t.Errorf("HTML: %q != %q", x, cs.x)
		}
		if x := EscapeUnquotedHTMLAttr(template.HTMLAttr(cs.x)); x != cs.x {
			t.Errorf("HTMLAttr: %q != %q", x, cs.x)
		}
	}
}

func TestEscapeComment(t *testing.T) {
	data := []string{"<!-- foobar -->"}
	builtin := builtinEscaper("_html_template_commentescaper")
	for _, x := range data {
		if a := EscapeComment(x); a != "" {
			t.Errorf("%q: %q != %q", x, a, x)
		}
		if b := builtin(x); b != "" {
			t.Errorf("%q: %q != %q", x, b, x)
		}
	}
}

func TestEscapeCSS(t *testing.T) {
	data := []struct {
		x, a, b string
	}{
		{
			x: "Ab09 zZ9",
			a: "Ab09 zZ9",
			b: "Ab09 zZ9",
		},
		{
			x: `"a'b'c"`,
			a: `\000022a\000027b\000027c\000022`,
			b: `\22 a\27 b\27 c\22 `,
		},
		{
			x: " \t\n\f\r\000",
			a: ` \000009\00000a\00000c\00000d`,
			b: ` \9 \a \c \d\0 `,
		},
		{
			x: "p { color: purple }",
			a: `p \00007b color\00003a purple \00007d`,
			b: `p \7b  color\3a  purple \7d `,
		},
		{
			x: `a[href=~"https:"].foo#bar`,
			a: `a\00005bhref\00003d\00007e\000022https\00003a\000022\00005d\00002efoo\000023bar`,
			b: `a[href=~\22https\3a\22].foo#bar`,
		},
		{
			x: "color: red; margin: 2px",
			a: `color\00003a red\00003b margin\00003a 2px`,
			b: `color\3a  red\3b  margin\3a  2px`,
		},
		{
			x: "rgba(0, 0, 255, 127)",
			a: `rgba\0000280\00002c 0\00002c 255\00002c 127\000029`,
			b: `rgba\28 0, 0, 255, 127\29 `,
		},
	}
	builtin := builtinEscaper("_html_template_cssescaper")
	for _, cs := range data {
		if a := EscapeCSS(cs.x); a != cs.a {
			t.Errorf("%q: `%v` != `%v`", cs.x, a, cs.a)
		}
		if b := builtin(cs.x); b != cs.b {
			t.Errorf("%q: `%v` != `%v`", cs.x, b, cs.b)
		}
	}
	for _, cs := range data {
		if x := EscapeCSS(template.CSS(cs.x)); x != cs.x {
			t.Errorf("CSS: %q != %q", x, cs.x)
		}
	}
}

func TestFilterCSS(t *testing.T) {
	data := []struct {
		x, a, b string
	}{
		{x: ""},
		{"10", "10", "10"},
		{"10px", "10px", "10px"},
		{"foo", "foo", "foo"},
		{
			x: "Ab09 zZ9",
			a: "Ab09 zZ9",
			b: "Ab09 zZ9",
		},
		{
			x: `"a'b'c"`,
			a: `ZgotmplZ`,
			b: `ZgotmplZ`,
		},
		{
			x: " \t\n\f\r\000",
			a: `ZgotmplZ`,
			b: `ZgotmplZ`,
		},
		{
			x: "p { color: purple }",
			a: `ZgotmplZ`,
			b: `ZgotmplZ`,
		},
		{
			x: `a[href=~"https:"].foo#bar`,
			a: `ZgotmplZ`,
			b: `ZgotmplZ`,
		},
		{
			x: "color: red; margin: 2px",
			a: `ZgotmplZ`,
			b: `ZgotmplZ`,
		},
		{
			x: "rgba(0, 0, 255, 127)",
			a: `ZgotmplZ`,
			b: `ZgotmplZ`,
		},
	}
	builtin := builtinEscaper("_html_template_cssvaluefilter")
	for _, cs := range data {
		if a := FilterCSS(cs.x); a != cs.a {
			t.Errorf("%q: %q != %q", cs.x, a, cs.a)
		}
		if b := builtin(cs.x); b != cs.b {
			t.Errorf("%q: %q != %q", cs.x, b, cs.b)
		}
	}
	for _, cs := range data {
		if x := FilterCSS(template.CSS(cs.x)); x != cs.x {
			t.Errorf("CSS: %q != %q", x, cs.x)
		}
	}
}

func TestFilterHTMLTagContent(t *testing.T) {
	data := []struct {
		x, a, b string
	}{
		{
			x: "",
			a: "ZgotmplZ",
			b: "ZgotmplZ",
		},
		{
			x: "foo",
			a: "foo",
			b: "foo",
		},
		{
			x: "foo bar",
			a: "ZgotmplZ",
			b: "ZgotmplZ",
		},
		{
			x: "fOoBaR",
			a: "fOoBaR",
			b: "foobar",
		},
		{
			x: " \t\n\f\r\000",
			a: "ZgotmplZ",
			b: "ZgotmplZ",
		},
		{
			x: "data-class",
			a: "ZgotmplZ",
			b: "ZgotmplZ",
		},
		{
			x: "foo:bar",
			a: "ZgotmplZ",
			b: "ZgotmplZ",
		},
		{
			x: "><a onclick=\"alert('pwned')\">Click me!</a>",
			a: "ZgotmplZ",
			b: "ZgotmplZ",
		},
		{
			x: "src=javascript:evil()",
			a: "ZgotmplZ",
			b: "ZgotmplZ",
		},
		{
			x: "=\"",
			a: "ZgotmplZ",
			b: "ZgotmplZ",
		},
	}
	builtin := builtinEscaper("_html_template_htmlnamefilter")
	for _, cs := range data {
		if a := FilterHTMLTagContent(cs.x); a != cs.a {
			t.Errorf("%q: %q != %q", cs.x, a, cs.a)
		}
		if b := builtin(cs.x); b != cs.b {
			t.Errorf("%q: %q != %q", cs.x, b, cs.b)
		}
	}
	for _, cs := range data {
		if x := FilterHTMLTagContent(template.HTML(cs.x)); x != cs.x {
			t.Errorf("HTML: %q != %q", x, cs.x)
		}
	}
}

func TestEscapeHTML(t *testing.T) {
	data := []struct {
		x, a string
	}{
		{"Ab09 zZ9", "Ab09 zZ9"},
		{" \t\n\f\r\000", " \t\n\f\r\uFFFD"},
		{"⌘", "⌘"},
		{"\"`'<>", "&#34;`&#39;&lt;&gt;"},
		{"alert(1)", "alert(1)"},
		{"\"><script>alert('pwned')</script>", "&#34;&gt;&lt;script&gt;alert(&#39;pwned&#39;)&lt;/script&gt;"},
	}
	builtin := builtinEscaper("_html_template_htmlescaper")
	for _, cs := range data {
		if a := EscapeHTML(cs.x); a != cs.a {
			t.Errorf("%q: %q != %q", cs.x, a, cs.a)
		}
		if b := builtin(cs.x); b != cs.a {
			t.Errorf("%q: %q != %q", cs.x, b, cs.a)
		}
	}
	for _, cs := range data {
		if x := EscapeHTML(template.HTML(cs.x)); x != cs.x {
			t.Errorf("HTML: %q != %q", x, cs.x)
		}
	}
}

func TestEscapeJSRegexp(t *testing.T) {
	// See html/template/js_test.go
	data := []struct {
		x, a, b string
	}{
		{"", `(?:)`, `(?:)`},
		{"foo", `foo`, `foo`},
		{"\u0000", `\u0000`, `\u0000`},
		{"\t", `\t`, `\t`},
		{"\n", `\n`, `\n`},
		{"\r", `\r`, `\r`},
		{"\u2028", `\u2028`, `\u2028`},
		{"\u2029", `\u2029`, `\u2029`},
		{"\\", `\\`, `\\`},
		{"\\n", `\\n`, `\\n`},
		{"foo\r\nbar", `foo\r\nbar`, `foo\r\nbar`},
		// Preserve attribute boundaries.
		{`"`, `\u0022`, `\u0022`},
		{`'`, `\u0027`, `\u0027`},
		// Allow embedding in HTML without further escaping.
		{`&amp;`, `\u0026amp\u003b`, `\u0026amp;`},
		// Prevent breaking out of text node and element boundaries.
		{"</script>", `\<\/script\>`, `\u003c\/script\u003e`},
		{"<![CDATA[", `\<\!\[CDATA\[`, `\u003c!\[CDATA\[`},
		{"]]>", `\]\]\>`, `\]\]\u003e`},
		// Escaping text spans.
		{"<!--", `\<\!\-\-`, `\u003c!\-\-`},
		{"-->", `\-\-\>`, `\-\-\u003e`},
		{"*", `\*`, `\*`},
		{"+", `\+`, `\u002b`},
		{"?", `\?`, `\?`},
		{"[](){}", `\[\]\(\)\{\}`, `\[\]\(\)\{\}`},
		{"$foo|x.y", `\$foo\|x\.y`, `\$foo\|x\.y`},
		{"x^y", `x\^y`, `x\^y`},
	}
	builtin := builtinEscaper("_html_template_jsregexpescaper")
	for _, cs := range data {
		if a := EscapeJSRegexp(cs.x); a != cs.a {
			t.Errorf("%q: %q != %q", cs.x, a, cs.a)
		}
		if b := builtin(cs.x); b != cs.b {
			t.Errorf("%q: %q != %q", cs.x, b, cs.b)
		}
	}
}

func TestEscapeJSStr(t *testing.T) {
	data := []struct {
		x, a, b string
	}{
		{"", "", ""},
		{"foo", "foo", "foo"},
		{"\000\u0000", `\u0000\u0000`, `\u0000\u0000`},
		{"foo\t\n\f\rbar", `foo\t\n\f\rbar`, `foo\t\n\f\rbar`},
		{`"'`, `\u0022\u0027`, `\u0022\u0027`},
		{"⌘", `\u2318`, "⌘"},
	}
	builtin := builtinEscaper("_html_template_jsstrescaper")
	for _, cs := range data {
		if a := EscapeJSStr(cs.x); a != cs.a {
			t.Errorf("%q: `%v` != `%v`", cs.x, a, cs.a)
		}
		if b := builtin(cs.x); b != cs.b {
			t.Errorf("%q: `%v` != `%v`", cs.x, b, cs.b)
		}
	}
	for _, cs := range data {
		if x := EscapeJSStr(template.JSStr(cs.x)); x != cs.x {
			t.Errorf("JSStr: %q != %q", x, cs.x)
		}
	}
}

func TestEscapeJSTmplLit(t *testing.T) {
	data := []struct {
		x, a, b string
	}{
		{"${foo}", `\u0024\u007bfoo\u007d`, `\u0024\u007bfoo\u007d`},
		{"`foo`", `\u0060foo\u0060`, `\u0060foo\u0060`},
		{
			"${alert(`foo`+\"bar\"+'baz')}",
			`\u0024\u007balert\u0028\u0060foo\u0060\u002b\u0022bar\u0022\u002b\u0027baz\u0027\u0029\u007d`,
			`\u0024\u007balert(\u0060foo\u0060\u002b\u0022bar\u0022\u002b\u0027baz\u0027)\u007d`,
		},
	}
	builtin := builtinEscaper("_html_template_jstmpllitescaper")
	for _, cs := range data {
		if a := EscapeJSTmplLit(cs.x); a != cs.a {
			t.Errorf("%q: `%v` != `%v`", cs.x, a, cs.a)
		}
		if b := builtin(cs.x); b != cs.b {
			t.Errorf("%q: `%v` != `%v`", cs.x, b, cs.b)
		}
	}
	for _, cs := range data {
		if a := EscapeJSTmplLit(template.JSStr(cs.x)); a != cs.x {
			t.Errorf("JSStr: `%v` != `%v`", a, cs.x)
		}
	}
}

func TestEscapeJS(t *testing.T) {
	failed := ";/* ERROR */null;"
	data := []struct {
		x  any
		js string
	}{
		{0, "0"},
		{42, "42"},
		{-42, "-42"},
		{float32(0.5), "0.5"},
		{float64(-0.5), "-0.5"},
		{"", `;`},
		{"/*", failed},
		{"*/", failed},
		{"foo", failed},
		{`"foo"`, `"foo"`},
		{"\r\n\u2028\u2029", failed},
		{`{"X":1,"Y":2}`, `{"X":1,"Y":2}`},
		{"[]", "[]"},
		{`[42, "foo", null]`, `[42, "foo", null]`},
		{`["<!--", "</script>", "-->"]`, `["<!--", "</script>", "-->"]`},
		{"<!--", failed},
		{"-->", failed},
		{"<![CDATA[", failed},
		{"]]>", failed},
		{"</script", failed},
		{"\U0001D11E", failed},
		{"null", "null"},
		{"(function () { return 'evil' })", failed},
		{"{ evil: () => 'evil' }", failed},
	}
	for _, cs := range data {
		if js := EscapeJS(nil, cs.x); js != cs.js {
			t.Errorf("%+v: `%v` != `%v`", cs.x, js, cs.js)
		}
	}
	var ew strings.Builder
	EscapeJS(&ew, "foo")
	if a, b := ew.String(), "invalid character 'o' in literal false (expecting 'a')\n"; a != b {
		t.Errorf("Error output: `%v` != `%v`", a, b)
	}
}

func TestEscapeRCData(t *testing.T) {
	data := []struct {
		x, a string
	}{
		{`/>`, `/&gt;`},
		{
			x: `O'Reilly: How are <i>you</i>?`,
			a: `O&#39;Reilly: How are &lt;i&gt;you&lt;/i&gt;?`,
		},
	}
	builtin := builtinEscaper("_html_template_rcdataescaper")
	for _, cs := range data {
		if a := EscapeRCData(cs.x); a != cs.a {
			t.Errorf("%q: `%v` != `%v`", cs.x, a, cs.a)
		}
		if b := builtin(cs.x); b != cs.a {
			t.Errorf("%q: `%v` != `%v`", cs.x, b, cs.a)
		}
	}
	for _, cs := range data {
		if a := EscapeRCData(template.HTML(cs.x)); a != cs.x {
			t.Errorf("HTML: `%v` != `%v`", a, cs.x)
		}
	}
}

func TestFilterAndEscapeSrcset(t *testing.T) {
	data := []struct {
		x, a string
	}{
		{"img.jpg", "img.jpg"},
		{"img.jpg 480w", "img.jpg 480w"},
		{"img.jpg 2x", "img.jpg 2x"},
		{" img.jpg", " img.jpg"},
		{" img.jpg\n", " img.jpg\n"},
		{" \n\fimg.jpg \t\r ", " \n\fimg.jpg \t\r "},
		{" \n\fimg.jpg \t2x\r ", " \n\fimg.jpg \t2x\r "},
		{" img.jpg 2x 480w 3x 640w", " img.jpg 2x 480w 3x 640w"},
		{"http://example.com/img.png", "http://example.com/img.png"},
		{"https://example.com/img.png", "https://example.com/img.png"},
		{" https://example.com/img.png", " https://example.com/img.png"},
		{"hTTps://example.com/img.png", "hTTps://example.com/img.png"},
		{"./path/to/img.png", "./path/to/img.png"},
		{"/path/to/img.png", "/path/to/img.png"},
		{"javascript:alert(1) 200w", "#ZgotmplZ"},
		{`O'Reilly: How are <i>you</i>?`, "#ZgotmplZ"},
	}
	builtin := builtinEscaper("_html_template_srcsetescaper")
	for _, cs := range data {
		if a := FilterAndEscapeSrcset(nil, cs.x); a != cs.a {
			t.Errorf("%q: `%v` != `%v`", cs.x, a, cs.a)
		}
		if b := builtin(cs.x); b != cs.a {
			t.Errorf("%q: `%v` != `%v`", cs.x, b, cs.a)
		}
	}
	for _, cs := range data {
		if a := FilterAndEscapeSrcset(nil, template.Srcset(cs.x)); a != cs.x {
			t.Errorf("Srcset: `%v` != `%v`", a, cs.x)
		}
	}
	var ew strings.Builder
	FilterAndEscapeSrcset(&ew, "javascript:alert(1)")
	if a, b := ew.String(), "url \"javascript:alert(1)\" is not safe\n"; a != b {
		t.Errorf("Error output: `%v` != `%v`", a, b)
	}
}

func TestEscapeURL(t *testing.T) {
	data := []struct {
		x, a, b string
	}{
		{"img.jpg", "img.jpg", "img.jpg"},
		{" a b c ", "+a+b+c+", "%20a%20b%20c%20"},
		{" a\nb\fc\td\re\n ", "+a%0Ab%0Cc%09d%0De%0A+", "%20a%0ab%0cc%09d%0de%0a%20"},
		{"http://example.com/img.png", "http%3A%2F%2Fexample.com%2Fimg.png", "http%3a%2f%2fexample.com%2fimg.png"},
		{"https://example.com/img.png", "https%3A%2F%2Fexample.com%2Fimg.png", "https%3a%2f%2fexample.com%2fimg.png"},
		{"./path/to/img.png", ".%2Fpath%2Fto%2Fimg.png", ".%2fpath%2fto%2fimg.png"},
		{"/path/to/img.png", "%2Fpath%2Fto%2Fimg.png", "%2fpath%2fto%2fimg.png"},
		{"javascript:alert(1)", "javascript%3Aalert%281%29", "javascript%3aalert%281%29"},
		{
			`O'Reilly: How are <i>you</i>?`,
			"O%27Reilly%3A+How+are+%3Ci%3Eyou%3C%2Fi%3E%3F",
			"O%27Reilly%3a%20How%20are%20%3ci%3eyou%3c%2fi%3e%3f",
		},
	}
	builtin := builtinEscaper("_html_template_urlescaper")
	for _, cs := range data {
		if a := EscapeURL(cs.x); a != cs.a {
			t.Errorf("%q: `%v` != `%v`", cs.x, a, cs.a)
		}
		if b := builtin(cs.x); b != cs.b {
			t.Errorf("%q: `%v` != `%v`", cs.x, b, cs.b)
		}
	}
	for _, cs := range data {
		if a := EscapeURL(template.URL(cs.x)); a != cs.x {
			t.Errorf("URL: `%v` != `%v`", a, cs.x)
		}
	}
}

func TestFilterURL(t *testing.T) {
	data := []struct {
		x, a, b string
	}{
		{"img.jpg", "img.jpg", "img.jpg"},
		{" a b c ", " a b c ", " a b c "},
		{" a\nb\fc\td\re\n ", " a\nb\fc\td\re\n ", " a\nb\fc\td\re\n "},
		{"http://example.com/img.png", "http://example.com/img.png", "http://example.com/img.png"},
		{"https://example.com/img.png", "https://example.com/img.png", "https://example.com/img.png"},
		{"  https://example.com/img.png  ", "  https://example.com/img.png  ", "#ZgotmplZ"},
		{"mailto:foo@example.com", "mailto:foo@example.com", "mailto:foo@example.com"},
		{"MailTo:foo@example.com", "MailTo:foo@example.com", "MailTo:foo@example.com"},
		{"ftp://foo/bar/baz.txt", "#ZgotmplZ", "#ZgotmplZ"},
		{"./path/to/img.png", "./path/to/img.png", "./path/to/img.png"},
		{"/path/to/img.png", "/path/to/img.png", "/path/to/img.png"},
		{"javascript:alert(1)", "#ZgotmplZ", "#ZgotmplZ"},
	}
	builtin := builtinEscaper("_html_template_urlfilter")
	for _, cs := range data {
		if a := FilterURL(nil, cs.x); a != cs.a {
			t.Errorf("%q: `%v` != `%v`", cs.x, a, cs.a)
		}
		if b := builtin(cs.x); b != cs.b {
			t.Errorf("%q: `%v` != `%v`", cs.x, b, cs.b)
		}
	}
	for _, cs := range data {
		if a := FilterURL(nil, template.URL(cs.x)); a != cs.x {
			t.Errorf("URL: `%v` != `%v`", a, cs.x)
		}
	}
}

func TestNormalizeURL(t *testing.T) {
	data := []struct {
		x, a string
	}{
		{"", ""},
		{
			"http://example.com:80/foo/bar?q=foo%20&bar=x+y#frag",
			"http://example.com:80/foo/bar?q=foo%20&bar=x+y#frag",
		},
		{" ", "%20"},
		{"%7c", "%7c"},
		{"%7C", "%7C"},
		{"%7C%7C", "%7C%7C"},
		{"%2", "%252"},
		{"%", "%25"},
		{"%z", "%25z"},
		{"/foo|bar/%5c\u1234", "/foo%7cbar/%5c%e1%88%b4"},
		{"⌘", "%e2%8c%98"},
		{"%0007C", "%0007C"},
	}
	builtin := builtinEscaper("_html_template_urlnormalizer")
	for _, cs := range data {
		if a := NormalizeURL(cs.x); a != cs.a {
			t.Errorf("%q: `%v` != `%v`", cs.x, a, cs.a)
		}
		if b := builtin(cs.x); b != cs.a {
			t.Errorf("%q: `%v` != `%v`", cs.x, b, cs.a)
		}
	}
	for _, cs := range data {
		if a := NormalizeURL(template.URL(cs.x)); a != cs.x {
			t.Errorf("URL: `%v` != `%v`", a, cs.x)
		}
	}
}

func TestBuiltinEscaper(t *testing.T) {
	if a, b := builtinEscaper("_html_template_cssvaluefilter")("+.33em"), "+.33em"; a != b {
		t.Errorf("%q != %q", a, b)
	}
}

func builtinEscaper(f string) func(any) string {
	txt := fmt.Sprintf("{{%s . | trick}}", f)
	tmpl := template.New(f).Funcs(template.FuncMap{
		// This dummy func needs to not fail the parsing, it will be redefined
		f: func(...any) string {
			return f
		},
		// Avoiding _html_template_htmlescaper
		"trick": func(v string) any {
			return template.HTML(v)
		},
	})
	tmpl = template.Must(tmpl.Parse(txt))
	return func(data any) string {
		var buf strings.Builder
		if err := tmpl.Execute(&buf, data); err != nil {
			panic(err)
		}
		return buf.String()
	}
}
