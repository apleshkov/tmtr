package funcs

import (
	"fmt"
	"html/template"
	"os"
	"runtime"
	"testing"
)

func TestAnd(t *testing.T) {
	testEq(t, true, And(true))       // last
	testEq(t, false, And(false))     // first encounter
	testEq(t, 1, And(4, 3, 2, 1))    // last
	testEq(t, "", And("2", "", "1")) // first encounter
}

func TestOr(t *testing.T) {
	testEq(t, true, Or(true))       // first encounter
	testEq(t, false, Or(false))     // last
	testEq(t, 0, Or(0, 0, 0, 0))    // last
	testEq(t, "1", Or("", "1", "")) // first encounter
}

// Check if all known escapers work
func TestFuncMap(t *testing.T) {
	s := `O'Reilly: How are <i>you</i>?`
	testEq(t, AttrEscaper(os.Stderr, s), "O&#39;Reilly: How are &lt;i&gt;you&lt;/i&gt;?")
	testEq(t, AttrEscaper(os.Stderr, template.HTML(s)), `O&#39;Reilly: How are you?`)
	testEq(t, CommentEscaper(os.Stderr, s), "")
	testEq(t, CSSEscaper(os.Stderr, s), "O\\27Reilly\\3a  How are \\3ci\\3eyou\\3c\\2fi\\3e?")
	testEq(t, CSSEscaper(os.Stderr, template.CSS(s)), "O\\27Reilly\\3a  How are \\3ci\\3eyou\\3c\\2fi\\3e?")
	testEq(t, CSSValueFilter(os.Stderr, s), "ZgotmplZ")
	testEq(t, CSSValueFilter(os.Stderr, template.CSS(s)), `O&#39;Reilly: How are &lt;i&gt;you&lt;/i&gt;?`)
	testEq(t, HTMLNameFilter(os.Stderr, s), "ZgotmplZ")
	testEq(t, HTMLNameFilter(os.Stderr, template.HTMLAttr(s)), `O&#39;Reilly: How are &lt;i&gt;you&lt;/i&gt;?`)
	testEq(t, HTMLEscaper(os.Stderr, s), "O&#39;Reilly: How are &lt;i&gt;you&lt;/i&gt;?")
	testEq(t, HTMLEscaper(os.Stderr, template.HTML(s)), s)
	testEq(t, JSRegexpEscaper(os.Stderr, s), "O\\u0027Reilly: How are \\u003ci\\u003eyou\\u003c\\/i\\u003e\\?")
	testEq(t, JSRegexpEscaper(os.Stderr, template.JS(s)), "O\\u0027Reilly: How are \\u003ci\\u003eyou\\u003c\\/i\\u003e\\?")
	testEq(t, JSStrEscaper(os.Stderr, s), "O\\u0027Reilly: How are \\u003ci\\u003eyou\\u003c\\/i\\u003e?")
	testEq(t, JSStrEscaper(os.Stderr, template.JS(s)), "O\\u0027Reilly: How are \\u003ci\\u003eyou\\u003c\\/i\\u003e?")
	testEq(t, JSTmplLitEscaper(os.Stderr, s), "O\\u0027Reilly: How are \\u003ci\\u003eyou\\u003c\\/i\\u003e?")
	testEq(t, JSTmplLitEscaper(os.Stderr, template.JS(s)), "O\\u0027Reilly: How are \\u003ci\\u003eyou\\u003c\\/i\\u003e?")
	testEq(t, JSValEscaper(os.Stderr, s), `&#34;O&#39;Reilly: How are \u003ci\u003eyou\u003c/i\u003e?&#34;`)
	testEq(t, JSValEscaper(os.Stderr, template.JS(s)), `O&#39;Reilly: How are &lt;i&gt;you&lt;/i&gt;?`)
	testEq(t, HTMLNospaceEscaper(os.Stderr, s), `O&amp;#39;Reilly:&amp;#32;How&amp;#32;are&amp;#32;&amp;lt;i&amp;gt;you&amp;lt;/i&amp;gt;?`)
	testEq(t, HTMLNospaceEscaper(os.Stderr, template.HTML(s)), `O&amp;#39;Reilly:&amp;#32;How&amp;#32;are&amp;#32;you?`)
	testEq(t, RCDataEscaper(os.Stderr, s), "O&#39;Reilly: How are &lt;i&gt;you&lt;/i&gt;?")
	testEq(t, RCDataEscaper(os.Stderr, template.HTML(s)), "O&#39;Reilly: How are &lt;i&gt;you&lt;/i&gt;?")
	testEq(t, SrcsetFilterAndEscaper(os.Stderr, s), "#ZgotmplZ")
	testEq(t, SrcsetFilterAndEscaper(os.Stderr, template.URL(s)), `O%27Reilly:%20How%20are%20%3ci%3eyou%3c/i%3e?`)
	testEq(t, SrcsetFilterAndEscaper(os.Stderr, template.Srcset(s)), `O&#39;Reilly: How are &lt;i&gt;you&lt;/i&gt;?`)
	testEq(t, URLEscaper(os.Stderr, s), "O%27Reilly%3a%20How%20are%20%3ci%3eyou%3c%2fi%3e%3f")
	testEq(t, URLEscaper(os.Stderr, template.URL(s)), `O%27Reilly:%20How%20are%20%3ci%3eyou%3c/i%3e?`)
	testEq(t, URLFilter(os.Stderr, s), "#ZgotmplZ")
	testEq(t, URLFilter(os.Stderr, template.URL(s)), `O&#39;Reilly: How are &lt;i&gt;you&lt;/i&gt;?`)
	testEq(t, URLNormalizer(os.Stderr, s), "O%27Reilly:%20How%20are%20%3ci%3eyou%3c/i%3e?")
	testEq(t, URLNormalizer(os.Stderr, template.URL(s)), "O%27Reilly:%20How%20are%20%3ci%3eyou%3c/i%3e?")
	testEq(t, EvalArgs(os.Stderr, s), `O&#39;Reilly: How are &lt;i&gt;you&lt;/i&gt;?`)
}

func testEq[T comparable](t *testing.T, a, b T) {
	if a != b {
		fmt.Fprintf(os.Stdout, "`%v` != `%v`\n", a, b)
		if _, file, line, ok := runtime.Caller(1); ok {
			fmt.Printf("[FAILED] %s:%d\n", file, line)
		}
		t.FailNow()
	}
}
