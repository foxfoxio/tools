module github.com/googlecodelabs/tools/claat

go 1.11

require (
	github.com/google/go-cmp v0.5.2
	github.com/graemephi/goldmark-qjs-katex v0.3.0
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/x1ddos/csslex v0.0.0-20160125172232-7894d8ab8bfe
	github.com/yuin/goldmark v1.2.1
	golang.org/x/net v0.0.0-20201016165138-7b1cca2348c0
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	gopkg.in/russross/blackfriday.v2 v2.0.1
)

replace gopkg.in/russross/blackfriday.v2 v2.0.1 => github.com/russross/blackfriday/v2 v2.0.1
