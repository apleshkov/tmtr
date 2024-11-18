package bench

import (
	"html/template"
	"io"
	"testing"
)

func BenchmarkGeneratedEscapers(b *testing.B) {
	d := newBenchData()
	for range b.N {
		lotsofesc(io.Discard, &d, nil)
	}
}

func BenchmarkBuiltinEscapers(b *testing.B) {
	t := template.Must(template.ParseFiles("./lotsofesc.html"))
	b.ResetTimer()
	d := newBenchData()
	for range b.N {
		if err := t.Execute(io.Discard, &d); err != nil {
			b.Fatal(err)
		}
	}
}
