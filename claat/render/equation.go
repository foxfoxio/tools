package render

import (
	"bytes"
	"fmt"
	"github.com/googlecodelabs/tools/claat/types"
	qjskatex "github.com/graemephi/goldmark-qjs-katex"
	"github.com/yuin/goldmark"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	"strings"
)

func (hw *htmlWriter) equation(n *types.EquationNode) {
	md := goldmark.New(
		goldmark.WithExtensions(&qjskatex.Extension{}),
		goldmark.WithRendererOptions(
			gmhtml.WithUnsafe(),
		),
	)

	delimiter := "$"
	if n.DisplayMode {
		delimiter = "$$"
	}
	var buf bytes.Buffer
	in := []byte(fmt.Sprintf(`<e>%[1]s%[2]s%[1]s</e>`, delimiter, n.Equation))
	if err := md.Convert(in, &buf); err == nil {
		n := buf.String()
		n = strings.ReplaceAll(n, "<p>", "")
		n = strings.ReplaceAll(n, "</p>", "")
		hw.writeString(n)
	}
}
