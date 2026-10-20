// Package markdown renders the markdown files runs leave as HTML, with
// blackfriday, the renderer the blog on incubator.to uses.
package markdown

import (
	"bytes"
	"net/url"
	"path"
	"strings"

	blackfriday "github.com/russross/blackfriday/v2"
)

// HTML renders markdown as HTML. It drops raw HTML in the source, links
// keep to safe schemes and open in a new view, and a relative link or
// image resolves against dir, the directory of the file, and links under
// prefix, the URL the files of the tree are at.
func HTML(src []byte, prefix, dir string) string {
	src = bytes.ReplaceAll(src, []byte("\r\n"), []byte("\n"))
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags | blackfriday.SkipHTML | blackfriday.Safelink |
			blackfriday.NofollowLinks | blackfriday.NoreferrerLinks | blackfriday.HrefTargetBlank,
	})
	parser := blackfriday.New(blackfriday.WithExtensions(blackfriday.CommonExtensions | blackfriday.AutoHeadingIDs))
	root := parser.Parse(src)
	root.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering && (node.Type == blackfriday.Link || node.Type == blackfriday.Image) {
			node.LinkData.Destination = []byte(resolve(prefix, dir, string(node.LinkData.Destination)))
		}
		return blackfriday.GoToNext
	})
	var out bytes.Buffer
	renderer.RenderHeader(&out, root)
	root.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		return renderer.RenderNode(&out, node, entering)
	})
	renderer.RenderFooter(&out, root)
	return out.String()
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
