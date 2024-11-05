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
)
