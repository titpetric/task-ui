// Package markdown renders the markdown files runs leave as HTML, with
// blackfriday, the renderer the blog on incubator.to uses.
package markdown

import (
	"bytes"
	"html"
	"net/url"
	"path"
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

// HTML renders markdown as HTML. Raw HTML in the source shows as text,
// links
// keep to safe schemes and open in a new view, and a relative link or
// image resolves against dir, the directory of the file, and links under
// prefix, the URL the files of the tree are at.
func HTML(src []byte, prefix, dir string) string {
	return Render(Options{Prefix: prefix, Dir: dir}, src)
}

// Render renders markdown as HTML, as HTML does, shaped by the options.
func Render(opts Options, src []byte) string {
	src = bytes.ReplaceAll(src, []byte("\r\n"), []byte("\n"))
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags | blackfriday.SkipHTML | blackfriday.Safelink |
			blackfriday.NofollowLinks | blackfriday.NoreferrerLinks | blackfriday.HrefTargetBlank,
	})
	extensions := blackfriday.CommonExtensions | blackfriday.AutoHeadingIDs
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
