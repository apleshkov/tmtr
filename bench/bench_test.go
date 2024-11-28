package bench

import (
	"html/template"
	"io"
	"testing"
)

const orly = `O'Reilly: How are <i>you</i>?`

//go:generate tmtr -fn lotsofesc -type string -in ./lotsofesc.html.txt -mode html

func BenchmarkGeneratedEscapers(b *testing.B) {
	for range b.N {
		lotsofesc(io.Discard, orly, nil)
	}
}

func BenchmarkTemplateEscapers(b *testing.B) {
	t := template.Must(template.ParseFiles("./lotsofesc.html.txt"))
	for range b.N {
		if err := t.Execute(io.Discard, orly); err != nil {
			b.Fatal(err)
		}
	}
}

//go:generate tmtr -fn basic -type string -in ./basic.html

func BenchmarkGeneratedBasic(b *testing.B) {
	for range b.N {
		basic(io.Discard, orly, nil)
	}
}

func BenchmarkTemplateBasic(b *testing.B) {
	t := template.Must(template.ParseFiles("./basic.html"))
	for range b.N {
		if err := t.Execute(io.Discard, orly); err != nil {
			b.Fatal(err)
		}
	}
}
