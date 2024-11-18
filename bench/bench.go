package bench

type benchData struct {
	Title, URL string
	Items      []string
}

//go:generate tmtr -fn lotsofesc -type *benchData -in ./lotsofesc.html

func newBenchData() benchData {
	return benchData{
		Title: "Hello",
		URL:   "https://example.com",
		Items: []string{
			"foobar",
			"https://example.com",
			"javascript:alert('foo')",
			"html { background: url('example.com') }",
		},
	}
}
