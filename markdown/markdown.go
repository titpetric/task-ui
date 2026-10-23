// Package markdown renders the markdown files runs leave as HTML, with
// blackfriday, the renderer the blog on incubator.to uses.
package markdown

import (
	"bytes"
	"html"
	"net/url"
	"path"
	"regexp"
	"strings"

	blackfriday "github.com/russross/blackfriday/v2"
)

// Options shape a Render.
type Options struct {
	// Prefix and Dir resolve a relative link or image: Dir is the
	// directory of the file, Prefix the URL its tree is at. Without a
	// Prefix, links stay as written.
	Prefix, Dir string
	// HardBreaks keeps a single line break a break, as a chat does.
	HardBreaks bool
	// Text rewrites the escaped text outside code and links, such as
	// marking @mentions; nil keeps it.
	Text func(escaped string) string
}

// HTML renders markdown as Render does, with links under prefix and relative paths against dir.
func HTML(src []byte, prefix, dir string) string {
	return Render(Options{Prefix: prefix, Dir: dir}, src)
}

// Render renders markdown as HTML, shaped by the options. Raw HTML in the
// source shows as text, links keep to safe schemes and open in a new view,
// and a relative link or image resolves against Dir, the directory of the
// file, and links under Prefix, the URL the files of the tree are at.
func Render(opts Options, src []byte) string {
	src = fenceBreaks(bytes.ReplaceAll(src, []byte("\r\n"), []byte("\n")))
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags | blackfriday.SkipHTML | blackfriday.Safelink |
			blackfriday.NofollowLinks | blackfriday.NoreferrerLinks | blackfriday.HrefTargetBlank,
	})
	// NoEmptyLineBeforeBlock starts a list, a quote or code right under a
	// line of text, as CommonMark and a chat do.
	extensions := blackfriday.CommonExtensions | blackfriday.AutoHeadingIDs | blackfriday.NoEmptyLineBeforeBlock
	if opts.HardBreaks {
		extensions |= blackfriday.HardLineBreak
	}
	root := blackfriday.New(blackfriday.WithExtensions(extensions)).Parse(src)
	root.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering && (node.Type == blackfriday.Link || node.Type == blackfriday.Image) {
			node.LinkData.Destination = []byte(resolve(opts.Prefix, opts.Dir, string(node.LinkData.Destination)))
		}
		return blackfriday.GoToNext
	})
	var out bytes.Buffer
	renderer.RenderHeader(&out, root)
	root.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		// Raw HTML shows escaped, as text: never markup, and
		// nothing written goes missing.
		switch node.Type {
		case blackfriday.HTMLSpan:
			out.WriteString(html.EscapeString(string(node.Literal)))
			return blackfriday.GoToNext
		case blackfriday.HTMLBlock:
			if strings.TrimSpace(string(node.Literal)) == listEnd {
				return blackfriday.GoToNext
			}
			out.WriteString("<p>" + html.EscapeString(string(node.Literal)) + "</p>\n")
			return blackfriday.GoToNext
		}
		if opts.Text != nil && node.Type == blackfriday.Text && !inLink(node) {
			out.WriteString(opts.Text(html.EscapeString(string(node.Literal))))
			return blackfriday.GoToNext
		}
		return renderer.RenderNode(&out, node, entering)
	})
	renderer.RenderFooter(&out, root)
	if opts.HardBreaks {
		// A kept break at the end of a list item is the item's own end.
		return strings.ReplaceAll(out.String(), "<br />\n</li>", "</li>")
	}
	return out.String()
}

// listEnd ends a list for blackfriday and renders as nothing.
const listEnd = "<!-- -->"

// listItem matches the first line of a list item.
var listItem = regexp.MustCompile(`^\s{0,3}([-*+]|\d{1,9}[.)])\s+\S`)

// fenceBreaks puts listEnd before a fence at the margin that follows a
// list, outside other fences: blackfriday, even with
// NoEmptyLineBeforeBlock, takes such a fence into the last item, where
// CommonMark ends the list.
func fenceBreaks(src []byte) []byte {
	lines := strings.Split(string(src), "\n")
	out := make([]string, 0, len(lines))
	fenced, inList := false, false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		margin := line != "" && line[0] != ' ' && line[0] != '\t'
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			if !fenced && inList && margin {
				out = append(out, "", listEnd, "")
				inList = false
			}
			fenced = !fenced
			out = append(out, line)
			continue
		}
		switch {
		case fenced:
		case listItem.MatchString(line):
			inList = true
		case margin && trimmed != "":
			inList = false
		}
		out = append(out, line)
	}
	return []byte(strings.Join(out, "\n"))
}

// inLink reports whether a node sits inside a link or an image.
func inLink(node *blackfriday.Node) bool {
	for p := node.Parent; p != nil; p = p.Parent {
		if p.Type == blackfriday.Link || p.Type == blackfriday.Image {
			return true
		}
	}
	return false
}

// resolve returns a relative link as the URL of the file it names in the
// tree, kept inside it; a link with a scheme, a host, an absolute path or
// only a fragment stays unchanged.
func resolve(prefix, dir, link string) string {
	u, err := url.Parse(link)
	if err != nil || prefix == "" || u.Scheme != "" || u.Host != "" || link == "" || strings.HasPrefix(link, "/") || strings.HasPrefix(link, "#") {
		return link
	}
	return strings.TrimSuffix(prefix, "/") + path.Clean("/"+path.Join(dir, u.Path)) + suffix(u)
}

// suffix is the query and the fragment of a link, as written.
func suffix(u *url.URL) string {
	out := ""
	if u.RawQuery != "" {
		out += "?" + u.RawQuery
	}
	if u.Fragment != "" {
		out += "#" + u.Fragment
	}
	return out
}
