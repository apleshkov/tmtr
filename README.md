# tmtr

The tool to translate Go [text](https://pkg.go.dev/text/template) and [html](https://pkg.go.dev/html/template) templates to Go source code.

## Installation

Install the tool:
```sh
go install github.com/apleshkov/tmtr@latest
```

Get the runtime (it's used by generated code):
```sh
go get github.com/apleshkov/tmtr/funcs
```

## Usage

The recommended approach is using `go:generate`, so it's possible to run `go generate` for concrete files or folders.

```go
// main.go
package main

import (...)

type myData struct {
	title string
}
//go:generate tmtr -fn "RenderData" -type "myData" -in "./index.html"

func main() {
	// Run the generated function
	RenderData(os.Stdout, myData{title: "Hello"}, os.Stderr)
}
```

```html
<div>{{.title}}</div>
```

So running `go generate main.go` generates the `index.html.go` file:
```go
package main

import (
	io "io"
	tmtrfn "tmtr/fn"
)

// `errOutput` can be nil
func RenderData(output io.Writer, data myData, errOutput io.Writer) {
	tmtrfn.Write(output, "<div>", errOutput)
	tmtrfn.Write(output, tmtrfn.HTMLEscaper(errOutput, data.title), errOutput)
	tmtrfn.Write(output, "</div>", errOutput)
}
```

Run `tmtr -h` to see the full info.

## Limitations

The generator doesn't know if something is a field or a method/function, so you have to use the `call` builtin template function in case of ambiguity:
```go
type myData {
    Title string
}

func (d myData) GetTitle() string {
    return d.Title
}

func (d myData) PrefixedTitle(s string) string {
    return s + d.Title
}
```

```html
{{.Title}} <!-- OK: the `Title` field -->
{{call .GetTitle}} <!-- OK: `call` is neccessary, cause the function has no arguments -->
{{.PrefixedTitle "foo"}} <!-- OK: `call` is not neccesary due to the argument -->
```

## Handling function errors

The generator introduces the `maybe` template function, so you can handle errors:
```go
type myData { ... }

func (d myData) LoadText() (string, error) { ... }
```

```html
{{maybe .LoadText}} <!-- BTW `call` is not neccessary here -->
```

## Custom template functions

Use `-tplfn` to add them. Comma-separated values are also supported (e.g. `-tplfn "foo,bar"`).

HTML: `{{foo .}}{{bar . 1}}`

Running `tmtr -fn "RenderData" -type "myData" -in "./index.html" -tplfn "foo" -tplfn "bar"` generates:
```go
func RenderData(output io.Writer, data myData, errOutput io.Writer) {
	tmtrfn.Write(output, tmtrfn.HTMLEscaper(errOutput, foo(data)), errOutput)
	tmtrfn.Write(output, tmtrfn.HTMLEscaper(errOutput, bar(data, 1)), errOutput)
}
```

## Additional imports

Use `-import` to add them. Comma-separated values are also supported (e.g. `-import "net,net/http,unicode/utf16"`).

HTML: `{{strings.ToUpper path.Base .Title}}`

Running `tmtr -fn "RenderData" -type "myData" -in "./index.html" -import "path" -import "strings"` generates:
```go
import (
	io "io"
	path "path"
	strings "strings"
	tmtrfn "tmtr/fn"
)

func RenderData(output io.Writer, data myData, errOutput io.Writer) {
	tmtrfn.Write(output, tmtrfn.HTMLEscaper(errOutput, strings.ToUpper(path.Base(data.Title))), errOutput)
}
```

## Templates

HTML: `{{template "foo" .}}`

Running `tmtr -fn "RenderData" -type "myData" -in "./index.html"` generates:

```go
func RenderData(
	output io.Writer, 
	data myData, 
	// "foo" is the function argument now
	foo func(io.Writer, myData, io.Writer), 
    //                  ^^^^^^ uses the same data type by default
	errOutput io.Writer,
) {	
	foo(output, data, errOutput)
}
```

### Specifying type

Use `-tpl` to specify them. Comma-separated values are also supported (e.g. `-tpl "foo:string,bar:bool"`).

HTML: `{{template "foo" .Title}}`

Running `tmtr -fn "RenderData" -type "myData" -in "./index.html" -tpl "foo:string"` generates:

```go
func RenderData(
	output io.Writer, 
	data myData, 
	foo func(io.Writer, string, io.Writer), 
	//                  ^^^^^^ the specified type
	errOutput io.Writer,
) {
	foo(output, data.Title, errOutput)
}
```

### Without type & argument

HTML: `{{template "foo"}}`

Running `tmtr -fn "RenderData" -type "myData" -in "./index.html" -tpl "foo"` generates:

```go
func RenderData(output io.Writer, data myData, foo func(io.Writer, io.Writer), errOutput io.Writer) {
	foo(output, errOutput)
}
```
