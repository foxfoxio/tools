// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package render

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/googlecodelabs/tools/claat/types"
)

// TODO: render HTML using golang/x/net/html or template.

// HTML renders nodes as the markup for the target env.
func HTML(ctx Context, nodes ...types.Node) (htmlTemplate.HTML, error) {
	var buf bytes.Buffer
	if err := WriteHTML(&buf, ctx.Env, ctx.Format, nodes...); err != nil {
		return "", err
	}
	return htmlTemplate.HTML(buf.String()), nil
}

// WriteHTML does the same as HTML but outputs rendered markup to w.
func WriteHTML(w io.Writer, env string, fmt string, nodes ...types.Node) error {
	hw := htmlWriter{w: w, env: env, format: fmt}
	return hw.write(nodes...)
}

// ReplaceDoubleCurlyBracketsWithEntity replaces Double Curly Brackets with their charater entity.
func ReplaceDoubleCurlyBracketsWithEntity(s string) string {
	return strings.Replace(s, "{{", "&#123;&#123;", -1)
}

type htmlWriter struct {
	w      io.Writer // output writer
	env    string    // target environment
	format string    // target template
	err    error     // error during any writeXxx methods
}

func (hw *htmlWriter) matchEnv(v []string) bool {
	if len(v) == 0 || hw.env == "" {
		return true
	}
	i := sort.SearchStrings(v, hw.env)
	return i < len(v) && v[i] == hw.env
}

func (hw *htmlWriter) write(nodes ...types.Node) error {
	for _, n := range nodes {
		if !hw.matchEnv(n.Env()) {
			continue
		}
		switch n := n.(type) {
		case *types.TextNode:
			hw.text(n)
		case *types.ImageNode:
			hw.image(n)
		case *types.URLNode:
			hw.url(n)
		case *types.ButtonNode:
			hw.button(n)
		case *types.CodeNode:
			hw.code(n)
			hw.writeString("\n")
		case *types.ListNode:
			hw.list(n)
			hw.writeString("\n")
		case *types.ImportNode:
			if len(n.Content.Nodes) == 0 {
				break
			}
			hw.list(n.Content)
			hw.writeString("\n")
		case *types.ItemsListNode:
			hw.itemsList(n)
			hw.writeString("\n")
		case *types.GridNode:
			hw.grid(n)
			hw.writeString("\n")
		case *types.InfoboxNode:
			hw.infobox(n)
			hw.writeString("\n")
		case *types.SurveyNode:
			hw.survey(n)
			hw.writeString("\n")
		case *types.HeaderNode:
			hw.header(n)
			hw.writeString("\n")
		case *types.YouTubeNode:
			hw.youtube(n)
			hw.writeString("\n")
		case *types.IframeNode:
			hw.iframe(n)
			hw.writeString("\n")
		case *types.ScriptNode:
			hw.script(n)
			hw.writeString("\n")
		}
		if hw.err != nil {
			return hw.err
		}
	}
	return nil
}

// Writes a string to the htmlWriter unless a write error has occurred on this htmlWriter in the past.
// Will set a write error on this htmlWriter if the write fails.
func (hw *htmlWriter) writeString(s string) {
	if hw.err != nil {
		return
	}
	_, hw.err = hw.w.Write([]byte(s))
}

// Same as writeString, but with fmt.Sprintf arguments/semantics.
func (hw *htmlWriter) writeFmt(f string, a ...interface{}) {
	hw.writeString(fmt.Sprintf(f, a...))
}

func escape(s string) string {
	s = htmlTemplate.HTMLEscapeString(s)
	s = ReplaceDoubleCurlyBracketsWithEntity(s)
	return s
}

// Same as writeString, but performs HTML escaping and double curly bracket escaping.
func (hw *htmlWriter) writeEscape(s string) {
	hw.writeString(escape(s))
}

func (hw *htmlWriter) text(n *types.TextNode) {
	s := n.Value
	shouldEsc := true
	if n.Bold {
		hw.writeString("<strong>")
	}
	if n.Italic {
		hw.writeString("<em>")
	}
	if n.Code {
		hw.writeString("<code>")
		shouldEsc = false
	}
	if shouldEsc {
		s = htmlTemplate.HTMLEscapeString(n.Value)
	}
	s = ReplaceDoubleCurlyBracketsWithEntity(s)
	hw.writeString(strings.Replace(s, "\n", "<br>", -1))
	if n.Code {
		hw.writeString("</code>")
	}
	if n.Italic {
		hw.writeString("</em>")
	}
	if n.Bold {
		hw.writeString("</strong>")
	}
}

func (hw *htmlWriter) image(n *types.ImageNode) {
	hw.writeString("<img")
	if n.Alt != "" {
		hw.writeFmt(" alt=%q", n.Alt)
	}
	if n.Title != "" {
		hw.writeFmt(" title=%q", n.Title)
	}
	if n.Width > 0 {
		hw.writeFmt(` style="width: %.2fpx"`, n.Width)
	}
	hw.writeFmt(" src=%q>", n.Src)
}

func (hw *htmlWriter) url(n *types.URLNode) {
	hw.writeString("<a")
	if n.URL != "" {
		hw.writeFmt(" href=%q", n.URL)
	}
	if n.Name != "" {
		hw.writeFmt(" name=%q", escape(n.Name))
	}
	if n.Target != "" {
		hw.writeFmt(" target=%q", escape(n.Target))
	}
	hw.writeString(">")
	hw.write(n.Content.Nodes...)
	hw.writeString("</a>")
}

func (hw *htmlWriter) button(n *types.ButtonNode) {
	hw.writeString("<paper-button")
	if n.Colored {
		hw.writeString(` class="colored"`)
	}
	if n.Raised {
		hw.writeString(" raised")
	}
	hw.writeString(">")
	if n.Download {
		hw.writeString(`<iron-icon icon="file-download"></iron-icon>`)
	}
	hw.write(n.Content.Nodes...)
	hw.writeString("</paper-button>")
}

func (hw *htmlWriter) code(n *types.CodeNode) {
	hw.writeString("<pre>")
	if !n.Term {
		hw.writeString("<code")
		if n.Lang != "" {
			hw.writeFmt(" language=%q class=%q", n.Lang, n.Lang)
		}
		hw.writeString(">")
	}
	if hw.format == "devsite" {
		hw.writeString("{% verbatim %}")
	}
	hw.writeEscape(n.Value)
	if hw.format == "devsite" {
		hw.writeString("{% endverbatim %}")
	}
	if !n.Term {
		hw.writeString("</code>")
	}
	hw.writeString("</pre>")
}

func (hw *htmlWriter) list(n *types.ListNode) {
	wrap := n.Block() == true
	if wrap {
		if onlyImages(n.Nodes...) {
			hw.writeString(`<p class="image-container">`)
		} else {
			hw.writeString("<p>")
		}
	}
	hw.write(n.Nodes...)
	if wrap {
		hw.writeString("</p>")
	}
}

// Returns true if the list of Nodes contains only images or white spaces.
func onlyImages(nodes ...types.Node) bool {
	for _, n := range nodes {
		switch n := n.(type) {
		case *types.TextNode:
			if len(strings.TrimSpace(n.Value)) == 0 {
				continue
			}
			return false
		case *types.ImageNode:
			continue
		default:
			return false
		}
	}
	return true
}

func (hw *htmlWriter) itemsList(n *types.ItemsListNode) {
	tag := "ul"
	if n.Type() == types.NodeItemsList && (n.Start > 0 || n.ListType != "") {
		tag = "ol"
	}
	hw.writeString("<")
	hw.writeString(tag)
	switch n.Type() {
	case types.NodeItemsCheck:
		hw.writeString(` class="checklist"`)
	case types.NodeItemsFAQ:
		hw.writeString(` class="faq"`)
	default:
		if n.ListType != "" {
			hw.writeFmt(" type=%q", n.ListType)
		}
		if n.Start > 0 {
			hw.writeFmt(` start=%q`, strconv.Itoa(n.Start))
		}
	}
	hw.writeString(">\n")

	for _, i := range n.Items {
		hw.writeString("<li>")
		hw.write(i.Nodes...)
		hw.writeString("</li>\n")
	}

	hw.writeFmt("</%s>", tag)
}

func (hw *htmlWriter) grid(n *types.GridNode) {
	hw.writeString("<table>\n")
	for _, r := range n.Rows {
		hw.writeString("<tr>")
		for _, c := range r {
			hw.writeFmt(`<td colspan="%d" rowspan="%d">`, c.Colspan, c.Rowspan)
			hw.write(c.Content.Nodes...)
			hw.writeString("</td>")
		}
		hw.writeString("</tr>\n")
	}
	hw.writeString("</table>")
}

func (hw *htmlWriter) infobox(n *types.InfoboxNode) {
	hw.writeFmt("<aside class=%q>", escape(string(n.Kind)))
	hw.write(n.Content.Nodes...)
	hw.writeString("</aside>")
}

func (hw *htmlWriter) survey(n *types.SurveyNode) {
	hw.writeFmt("<google-codelab-survey survey-id=%q>\n", n.ID)
	for _, g := range n.Groups {
		hw.writeFmt("<h4>%s</h4>\n<paper-radio-group>\n", g.Name)
		for _, o := range g.Options {
			hw.writeFmt("<paper-radio-button>%s</paper-radio-button>\n", escape(o))
		}
		hw.writeString("</paper-radio-group>\n")
	}
	hw.writeString("</google-codelab-survey>")
}

func (hw *htmlWriter) header(n *types.HeaderNode) {
	tag := "h" + strconv.Itoa(n.Level)
	hw.writeString("<")
	hw.writeString(tag)
	switch n.Type() {
	case types.NodeHeaderCheck:
		hw.writeString(` class="checklist"`)
	case types.NodeHeaderFAQ:
		hw.writeString(` class="faq"`)

	}
	hw.writeString(` is-upgraded>`)
	hw.write(n.Content.Nodes...)
	hw.writeFmt("</%s>", tag)
}

func (hw *htmlWriter) youtube(n *types.YouTubeNode) {
	hw.writeFmt(`<iframe class="youtube-video" `+
		`src="https://www.youtube.com/embed/%s?rel=0" allow="accelerometer; `+
		`autoplay; encrypted-media; gyroscope; picture-in-picture" `+
		`allowfullscreen></iframe>`, n.VideoID)
}

func (hw *htmlWriter) iframe(n *types.IframeNode) {
	hw.writeFmt(`<iframe class="embedded-iframe" src=%q allowfullscreen></iframe>`, n.URL)
}

var (
	// textCleaner replaces "smart quotes" and other unicode runes
	// with their respective ascii versions.
	textCleaner = strings.NewReplacer(
		"\u2019", "'", "\u201C", `"`, "\u201D", `"`, "\u2026", "...",
		"\u00A0", " ", "\u0085", " ",
	)
)

func (hw *htmlWriter) script(n *types.ScriptNode) {
	fmt.Fprintln(os.Stdout, " ✨ script")
	hw.writeFmt(textCleaner.Replace(n.URL))
}
