package gen

import "go/ast"

var (
	nilIdent             = ast.NewIdent("nil")
	writeIdent           = ast.NewIdent("Write")
	sprintIdent          = ast.NewIdent("Sprint")
	sprintfIdent         = ast.NewIdent("Sprintf")
	sprintlnIdent        = ast.NewIdent("Sprintln")
	andIdent             = ast.NewIdent("And")
	orIdent              = ast.NewIdent("Or")
	isTrueIdent          = ast.NewIdent("IsTrue")
	isNotTrueIdent       = ast.NewIdent("IsNotTrue")
	htmlEscaperIdent     = ast.NewIdent("HTMLEscaper")
	jsEscaperIdent       = ast.NewIdent("JSEscaper")
	urlQueryEscaperIdent = ast.NewIdent("URLQueryEscaper")
	maybeIdent           = ast.NewIdent("MayBe")

	escapeHTMLAttrIdent         = ast.NewIdent("EscapeHTMLAttr")
	escapeCommentIdent          = ast.NewIdent("EscapeComment")
	escapeCSSIdent              = ast.NewIdent("EscapeCSS")
	filterCSSIdent              = ast.NewIdent("FilterCSS")
	filterHTMLTagContentIdent   = ast.NewIdent("FilterHTMLTagContent")
	escapeHTMLIdent             = ast.NewIdent("EscapeHTML")
	escapeJSRegexpIdent         = ast.NewIdent("EscapeJSRegexp")
	escapeJSStrIdent            = ast.NewIdent("EscapeJSStr")
	escapeJSTmplLitIdent        = ast.NewIdent("EscapeJSTmplLit")
	escapeJSIdent               = ast.NewIdent("EscapeJS")
	escapeUnquotedHTMLAttrIdent = ast.NewIdent("EscapeUnquotedHTMLAttr")
	escapeRCDataIdent           = ast.NewIdent("EscapeRCData")
	filterAndEscapeSrcsetIdent  = ast.NewIdent("FilterAndEscapeSrcset")
	escapeURLIdent              = ast.NewIdent("EscapeURL")
	filterURLIdent              = ast.NewIdent("FilterURL")
	normalizeURLIdent           = ast.NewIdent("NormalizeURL")
)
