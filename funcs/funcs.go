package funcs

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"strings"
	"unicode"
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

const filterFailsafe = "ZgotmplZ"

type valueType int

const (
	valueTypePlain valueType = iota
	valueTypeCSS
	valueTypeHTML
	valueTypeHTMLAttr
	valueTypeJS
	valueTypeJSStr
	valueTypeURL
	valueTypeSrcset
)

func stringify(args []any) (string, valueType) {
	if len(args) == 1 {
		switch s := args[0].(type) {
		case string:
			return s, valueTypePlain
		case template.CSS:
			return string(s), valueTypeCSS
		case template.HTML:
			return string(s), valueTypeHTML
		case template.HTMLAttr:
			return string(s), valueTypeHTMLAttr
		case template.JS:
			return string(s), valueTypeJS
		case template.JSStr:
			return string(s), valueTypeJSStr
		case template.URL:
			return string(s), valueTypeURL
		case template.Srcset:
			return string(s), valueTypeSrcset
		}
	}
	return fmt.Sprint(args...), valueTypePlain
}

// Use `template.HTML` or `template.HTMLAttr` to bypass.
func EscapeHTMLAttr(data ...any) string {
	s, t := stringify(data)
	if t == valueTypeHTML || t == valueTypeHTMLAttr {
		return s
	}
	return template.HTMLEscapeString(s)
}

// Use `template.HTMLAttr` to bypass.
func EscapeUnquotedHTMLAttr(data ...any) string {
	s, t := stringify(data)
	if t == valueTypeHTML || t == valueTypeHTMLAttr {
		return s
	}
	return fmt.Sprintf("%q", EscapeHTMLAttr(s))
}

func EscapeComment(...any) string {
	return ""
}

// Use `template.CSS` to bypass.
func EscapeCSS(data ...any) string {
	s, t := stringify(data)
	if t == valueTypeCSS {
		return s
	}
	var buf strings.Builder
	buf.Grow(len(s))
	for i := range len(s) {
		c := s[i]
		switch {
		case c == 0:
			continue
		case
			c == ' ',
			'0' <= c && c <= '9',
			'a' <= c && c <= 'z',
			'A' <= c && c <= 'Z':
			buf.WriteByte(c)
		default:
			fmt.Fprintf(&buf, "\\%06x", c)
		}
	}
	return buf.String()
}

// Use `template.CSS` to bypass.
func FilterCSS(data ...any) string {
	s, t := stringify(data)
	if t == valueTypeCSS || len(s) == 0 {
		return s
	}
	for _, r := range s {
		switch {
		case r == ' ',
			'0' <= r && r <= '9',
			'a' <= r && r <= 'z',
			'A' <= r && r <= 'Z':
		default:
			return filterFailsafe
		}
	}
	return s
}

// Use `template.HTML` to bypass.
func FilterHTMLTagContent(data ...any) string {
	s, t := stringify(data)
	if t == valueTypeHTML {
		return s
	}
	// Passing an empty string to smth like `<input checked {{.}}=...>`
	// leads to `<input checked =...>`, which could be harmful.
	if len(s) == 0 {
		return filterFailsafe
	}
	for _, r := range s {
		switch {
		case '0' <= r && r <= '9':
		case 'a' <= r && r <= 'z':
		case 'A' <= r && r <= 'Z':
		default:
			return filterFailsafe
		}
	}
	return s
}

// Use `template.HTML` to bypass
func EscapeHTML(data ...any) string {
	s, t := stringify(data)
	if t == valueTypeHTML {
		return s
	}
	return template.HTMLEscapeString(s)
}

var jsRegexpSpecials = map[rune]struct{}{
	'/':  {},
	'.':  {},
	'\\': {},
	'+':  {},
	'*':  {},
	'?':  {},
	'[':  {},
	'^':  {},
	']':  {},
	'$':  {},
	'(':  {},
	')':  {},
	'{':  {},
	'}':  {},
	'=':  {},
	'!':  {},
	'<':  {},
	'>':  {},
	'|':  {},
	':':  {},
	'-':  {},
	'#':  {},
}

var jsStrRepls = map[rune]string{
	'\t':     "\\t",
	'\n':     "\\n",
	'\r':     "\\r",
	'\f':     "\\f",
	'\u2028': "\\u2028",
	'\u2029': "\\u2029",
}

// No bypassing.
func EscapeJSRegexp(data ...any) string {
	s, _ := stringify(data)
	// Passing an empty string to smth like `/{{.}}/`
	// leads to `//`, which is invalid.
	if len(s) == 0 {
		return "(?:)"
	}
	var buf strings.Builder
	buf.Grow(len(s))
	for _, r := range s {
		switch {
		case
			'0' <= r && r <= '9',
			'a' <= r && r <= 'z',
			'A' <= r && r <= 'Z':
			buf.WriteRune(r)
		default:
			if _, ok := jsRegexpSpecials[r]; ok {
				buf.WriteByte('\\')
				buf.WriteRune(r)
				continue
			}
			if repl, ok := jsStrRepls[r]; ok {
				buf.WriteString(repl)
				continue
			}
			fmt.Fprintf(&buf, "\\u%04x", r)
		}
	}
	return buf.String()
}

// Use `template.JSStr` to bypass.
func EscapeJSStr(data ...any) string {
	s, t := stringify(data)
	if t == valueTypeJSStr {
		return s
	}
	var buf strings.Builder
	buf.Grow(len(s))
	for _, r := range s {
		switch {
		case
			r == ' ',
			'0' <= r && r <= '9',
			'a' <= r && r <= 'z',
			'A' <= r && r <= 'Z':
			buf.WriteRune(r)
		default:
			if repl, ok := jsStrRepls[r]; ok {
				buf.WriteString(repl)
				continue
			}
			fmt.Fprintf(&buf, "\\u%04x", r)
		}
	}
	return buf.String()
}

// Use `template.JSStr` to bypass.
func EscapeJSTmplLit(data ...any) string {
	return EscapeJSStr(data...)
}

// Checks an input string is valid JSON by unmarshalling it. If failed, then
// returns ";/* ERROR */null;" and writes an actual error to the optional `ew`
// writer. Also returns `;` if an input string is empty.
// Use `template.JS` to bypass.
func EscapeJS(ew io.Writer, data ...any) string {
	s, t := stringify(data)
	if t == valueTypeJS {
		return s
	}
	// For instance `x=y/{{.}}*z` shouldn't become `x=y/*z`
	if len(s) == 0 {
		return `;`
	}
	var js any
	if err := json.Unmarshal([]byte(s), &js); err != nil {
		if ew != nil {
			fmt.Fprintln(ew, err)
		}
		return ";/* ERROR */null;"
	}
	return s
}

// Use `template.HTML` to bypass.
func EscapeRCData(data ...any) string {
	return EscapeHTML(data...)
}

// https://infra.spec.whatwg.org/#ascii-whitespace
const asciiWhitespaces = " \t\n\f\r"

func fastURLScheme(s string) string {
	if b, _, ok := strings.Cut(s, ":"); ok {
		b = strings.TrimLeftFunc(b, unicode.IsSpace)
		return strings.ToLower(b)
	}
	return ""
}

// Use `template.Srcset` to bypass.
func FilterAndEscapeSrcset(ew io.Writer, data ...any) string {
	s, t := stringify(data)
	if t == valueTypeSrcset {
		return s
	}
	u := strings.Trim(s, asciiWhitespaces)
	if i := strings.IndexAny(s, asciiWhitespaces); i != -1 {
		u = s[:i]
	}
	switch fastURLScheme(s) {
	case "", "http", "https":
	default:
		if ew != nil {
			fmt.Fprintf(ew, "url \"%s\" is not safe\n", u)
		}
		return "#" + filterFailsafe
	}
	return s
}

// Use `template.URL` to bypass.
func EscapeURL(data ...any) string {
	s, t := stringify(data)
	if t == valueTypeURL {
		return s
	}
	return template.URLQueryEscaper(s)
}

// Use `template.URL` to bypass.
func FilterURL(ew io.Writer, data ...any) string {
	s, t := stringify(data)
	if t == valueTypeURL {
		return s
	}
	switch fastURLScheme(s) {
	case "", "http", "https", "mailto":
	default:
		if ew != nil {
			fmt.Fprintf(ew, "url \"%s\" is not safe\n", s)
		}
		return "#" + filterFailsafe
	}
	return s
}

// Normalizes an input so it can be embedded in double or single quotes.
// Use `template.URL` to bypass.
func NormalizeURL(data ...any) string {
	s, t := stringify(data)
	if t == valueTypeURL {
		return s
	}
	l := len(s)
	var buf strings.Builder
	buf.Grow(l)
	for i := 0; i < l; i += 1 {
		c := s[i]
		if c == '%' && i+2 < l {
			skip := true
			frag := s[i+1 : i+3]
			for _, r := range frag {
				ok := '0' <= r && r <= '9' || 'a' <= r && r <= 'f' || 'A' <= r && r <= 'F'
				if !ok {
					skip = false
					break
				}
			}
			if skip {
				io.WriteString(&buf, s[i:i+3])
				i += 2
				continue
			}
		}
		switch c {
		case
			'!', '#', '$', '&', '*', '+', ',', '/', ':', ';', '=', '?', '@', '[', ']',
			'-', '.', '_', '~':
			buf.WriteByte(c)
			continue
		}
		if '0' <= c && c <= '9' ||
			'a' <= c && c <= 'z' ||
			'A' <= c && c <= 'Z' {
			buf.WriteByte(c)
			continue
		}
		fmt.Fprintf(&buf, "%%%02x", c)
	}
	return buf.String()
}
