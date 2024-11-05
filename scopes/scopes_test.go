package scopes

import (
	"testing"
	"text/template"
	"text/template/parse"

	"github.com/apleshkov/tmtr/util"
)

func TestRootScope(t *testing.T) {
	s := NewRootScope(
		parseNode("{{$x := 0}} {{$y = 1}}"),
	)
	util.TestAssert(t, s.ns.has("data"))
	util.TestAssert(t, s.Dot().Name == "data")
	util.TestAssert(t, Uniq(s, "data").Name == "data_")
	util.TestAssert(t, s.ns.has("x"))
	util.TestAssert(t, Uniq(s, "x").Name == "x_")
	util.TestAssert(t, Uniq(s, "x").Name == "x__")
	util.TestAssert(t, !s.ns.has("y"))
}

func TestRangeScope1(t *testing.T) {
	root := parseNode(`{{range .}}{{end}}`)
	s := NewRangeScope(NewRootScope(root), root.Nodes[0].(*parse.RangeNode))
	util.TestAssert(t, s.ns.has("data"))
	util.TestAssert(t, s.ns.has("list"))
	util.TestAssert(t, s.Key().Name == "_")
	util.TestAssert(t, s.Value().Name == "elem")
	util.TestAssert(t, s.Dot().Name == "elem")
}

func TestRangeScope2(t *testing.T) {
	root := parseNode(`{{range $v := .}}{{end}}`)
	s := NewRangeScope(NewRootScope(root), root.Nodes[0].(*parse.RangeNode))
	util.TestAssert(t, s.ns.has("data"))
	util.TestAssert(t, s.ns.has("list"))
	util.TestAssert(t, s.Key().Name == "_")
	util.TestAssert(t, s.Value().Name == "v")
	util.TestAssert(t, s.Dot().Name == "v")
}

func TestRangeScope3(t *testing.T) {
	root := parseNode(`{{range $i, $value := .}}{{end}}`)
	s := NewRangeScope(NewRootScope(root), root.Nodes[0].(*parse.RangeNode))
	util.TestAssert(t, s.ns.has("data"))
	util.TestAssert(t, s.ns.has("list"))
	util.TestAssert(t, s.Key().Name == "i")
	util.TestAssert(t, s.Value().Name == "value")
	util.TestAssert(t, s.Dot().Name == "value")
}

func TestListScope(t *testing.T) {
	root := parseNode(`{{range .}} {{$a := 0}} {{else}} {{$b := 1}} {{end}}`)
	rn := root.Nodes[0].(*parse.RangeNode)
	s := NewRangeScope(NewRootScope(root), rn)
	util.TestAssert(t, s.ns.has("data"))
	util.TestAssert(t, s.ns.has("list"))
	util.TestAssert(t, s.ns.has("elem"))
	util.TestAssert(t, s.ns.has("a"))
	util.TestAssert(t, !s.ns.has("b"))
	els := s.ElseScope
	util.TestAssert(t, els.ns.has("data"))
	util.TestAssert(t, !els.ns.has("list"))
	util.TestAssert(t, !els.ns.has("elem"))
	util.TestAssert(t, !els.ns.has("a"))
	util.TestAssert(t, els.ns.has("b"))
}

func TestIfScope(t *testing.T) {
	root := parseNode(`{{if $a := .}} {{$b := $a}} {{else}} {{$c := $a}} {{end}}`)
	ifn := root.Nodes[0].(*parse.IfNode)
	ifs := NewIfScope(NewRootScope(root), ifn)
	ts := ifs.ThenScope
	util.TestAssert(t, ts.ns.has("a"))
	util.TestAssert(t, ts.ns.has("b"))
	util.TestAssert(t, !ts.ns.has("c"))
	util.TestAssert(t, ts.Dot().Name == "data")
	es := ifs.ElseScope
	util.TestAssert(t, es.ns.has("a"))
	util.TestAssert(t, !es.ns.has("b"))
	util.TestAssert(t, es.ns.has("c"))
	util.TestAssert(t, es.Dot().Name == "data")
}

func TestWithScope1(t *testing.T) {
	root := parseNode(`{{with .Foo}}{{.}}{{end}}`)
	n := root.Nodes[0].(*parse.WithNode)
	ws := NewWithScope(NewRootScope(root), n)
	util.TestAssert(t, ws.Dot().Name == "with")
}

func TestWithScope2(t *testing.T) {
	root := parseNode(`{{with $v := .Foo}}{{.}}{{end}}`)
	n := root.Nodes[0].(*parse.WithNode)
	ws := NewWithScope(NewRootScope(root), n)
	util.TestAssert(t, ws.Dot().Name == "v")
}

func parseNode(text string) *parse.ListNode {
	t := template.Must(template.New("test").Parse(text))
	return t.Root
}
