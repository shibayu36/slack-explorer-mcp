package main

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

// CanvasHTMLStripper strips unnecessary attributes and elements from Canvas HTML
// to reduce token usage while preserving meaningful content.
type CanvasHTMLStripper struct{}

// NewCanvasHTMLStripper creates a new CanvasHTMLStripper instance.
func NewCanvasHTMLStripper() *CanvasHTMLStripper {
	return &CanvasHTMLStripper{}
}

// Strip removes unnecessary attributes and elements from Canvas HTML.
func (s *CanvasHTMLStripper) Strip(htmlContent string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	s.processNode(doc)

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return "", err
	}

	result := buf.String()
	result = s.cleanupRenderedHTML(result)

	return result, nil
}

// processNode recursively processes HTML nodes.
func (s *CanvasHTMLStripper) processNode(n *html.Node) {
	if n.Type == html.ElementNode {
		// Handle Slack emoji: <control><img alt="emoji" data-is-slack>:emoji:</img></control>
		if n.Data == "control" {
			if emojiText := s.extractSlackEmoji(n); emojiText != "" {
				// Replace control node with text node
				textNode := &html.Node{
					Type: html.TextNode,
					Data: emojiText,
				}
				n.Parent.InsertBefore(textNode, n)
				n.Parent.RemoveChild(n)
				return
			}
		}

		// Remove br elements
		if n.Data == "br" {
			if n.Parent != nil {
				n.Parent.RemoveChild(n)
			}
			return
		}

		// Unwrap all span elements (move children to parent, remove span)
		if n.Data == "span" {
			s.unwrapElement(n)
			return
		}

		// Strip attributes
		s.stripAttributes(n)
	}

	// Process children (collect first to avoid modification during iteration)
	var children []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		children = append(children, c)
	}
	for _, c := range children {
		s.processNode(c)
	}
}

// stripAttributes removes unnecessary attributes from an element.
func (s *CanvasHTMLStripper) stripAttributes(n *html.Node) {
	var newAttrs []html.Attribute
	for _, attr := range n.Attr {
		switch attr.Key {
		case "id", "style", "value":
			// Remove these attributes
			continue
		case "class":
			// Keep only "checked" and "embedded-file" classes
			if attr.Val == "checked" || attr.Val == "embedded-file" {
				newAttrs = append(newAttrs, attr)
			}
			// Otherwise remove class attribute
		case "href", "data-section-style":
			// Always keep these attributes
			newAttrs = append(newAttrs, attr)
		default:
			// Remove other attributes (src, alt, data-remapped, data-is-slack, etc.)
			continue
		}
	}
	n.Attr = newAttrs
}

// extractSlackEmoji extracts emoji text from a Slack emoji control element.
// Returns empty string if not a Slack emoji.
func (s *CanvasHTMLStripper) extractSlackEmoji(controlNode *html.Node) string {
	// Look for <img> with data-is-slack attribute inside control
	for c := controlNode.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "img" {
			hasDataIsSlack := false
			altText := ""
			for _, attr := range c.Attr {
				if attr.Key == "data-is-slack" {
					hasDataIsSlack = true
				}
				if attr.Key == "alt" {
					altText = attr.Val
				}
			}
			if hasDataIsSlack && altText != "" {
				return ":" + altText + ":"
			}
		}
	}
	return ""
}

// unwrapElement moves all children of an element to its parent and removes the element.
func (s *CanvasHTMLStripper) unwrapElement(n *html.Node) {
	if n.Parent == nil {
		return
	}

	// Collect children first
	var children []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		children = append(children, c)
	}

	// Move children to parent (before the current node)
	for _, c := range children {
		n.RemoveChild(c)
		n.Parent.InsertBefore(c, n)
	}

	// Remove the now-empty element
	n.Parent.RemoveChild(n)

	// Process moved children
	for _, c := range children {
		s.processNode(c)
	}
}

// cleanupRenderedHTML removes the html/head/body wrapper added by html.Parse.
func (s *CanvasHTMLStripper) cleanupRenderedHTML(rendered string) string {
	// html.Parse wraps content in <html><head></head><body>...</body></html>
	// We need to extract just the body content

	// Find <body> and </body>
	bodyStart := strings.Index(rendered, "<body>")
	bodyEnd := strings.LastIndex(rendered, "</body>")

	if bodyStart != -1 && bodyEnd != -1 {
		bodyStart += len("<body>")
		return rendered[bodyStart:bodyEnd]
	}

	return rendered
}
