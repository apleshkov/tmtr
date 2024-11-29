# tmtr

The CLI tool to generate static type-safe Go source code from [html/template](https://pkg.go.dev/html/template) or [text/template](https://pkg.go.dev/text/template).

Better with `go generate` (see <https://go.dev/blog/generate>).

## Benchmarks

```sh
$ cd ./bench
$ go test -bench .
goos: darwin
goarch: arm64
pkg: bench
cpu: Apple M3 Pro
# A lot of different escapers, see ./bench/lotsofesc.html.txt
BenchmarkGeneratedEscapers-11    	  202416	      5936 ns/op
BenchmarkTemplateEscapers-11     	   83209	     14402 ns/op
# Just `<div>{{.}}</div>`, see ./bench/basic.html
BenchmarkGeneratedBasic-11       	 6320578	       188.6 ns/op
BenchmarkTemplateBasic-11        	 2511336	       479.6 ns/op
PASS
ok  	bench	6.870s
```

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

## Security Notes

The tool uses the `html/template` escaping mechanism, so all the necessary [sanitizing functions](https://pkg.go.dev/html/template#hdr-Contexts) will be added.

But please keep in mind these sanitizing functions are *re-implemented with some changes*, cause the original ones are private to the `html/template` package. There's a [proposal](https://github.com/golang/go/issues/70375) to export them though.
