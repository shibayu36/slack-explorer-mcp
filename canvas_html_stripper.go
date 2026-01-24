package main

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

// Whitelist of attributes to preserve
var preservedAttributes = map[string]bool{
	"href": true,
	// data-section-style indicates list type: 5=bullet, 6=numbered, 7=checklist
	"data-section-style": true,
}

// Whitelist of class values to preserve
var preservedClassValues = map[string]bool{
	// checked indicates a completed TODO item in checklists
	"checked": true,
	// embedded-file indicates an attached file reference
	"embedded-file": true,
}

// CanvasHTMLStripper strips unnecessary HTML elements and attributes from Slack Canvas HTML.
type CanvasHTMLStripper struct {
}

// NewCanvasHTMLStripper creates a new CanvasHTMLStripper instance
func NewCanvasHTMLStripper() *CanvasHTMLStripper {
	return &CanvasHTMLStripper{}
}

// Strip removes unnecessary HTML elements and attributes from the input HTML.
func (s *CanvasHTMLStripper) Strip(htmlStr string) (string, error) {
	if htmlStr == "" {
		return "", nil
	}

	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return "", err
	}

	s.processNode(doc)

	// html.Parse automatically wraps content in <html><head><body>, so we extract only the body's children
	body := s.findBody(doc)
	if body == nil {
		return "", nil
	}

	var buf bytes.Buffer
	for c := body.FirstChild; c != nil; c = c.NextSibling {
		if err := html.Render(&buf, c); err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}

// processNode recursively processes all nodes in the tree, filtering attributes
func (s *CanvasHTMLStripper) processNode(n *html.Node) {
	if n.Type == html.ElementNode {
		n.Attr = s.filterAttributes(n.Attr)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		s.processNode(c)
	}
}

// filterAttributes filters the attributes, keeping only whitelisted ones
func (s *CanvasHTMLStripper) filterAttributes(attrs []html.Attribute) []html.Attribute {
	result := make([]html.Attribute, 0, len(attrs))

	for _, attr := range attrs {
		if preservedAttributes[attr.Key] {
			result = append(result, attr)
			continue
		}

		if attr.Key == "class" {
			if filteredClass := s.filterClassValue(attr.Val); filteredClass != "" {
				result = append(result, html.Attribute{Key: "class", Val: filteredClass})
			}
		}
	}

	return result
}

// filterClassValue filters class values, keeping only whitelisted ones
func (s *CanvasHTMLStripper) filterClassValue(classVal string) string {
	classes := strings.Fields(classVal)
	preserved := make([]string, 0, len(classes))

	for _, class := range classes {
		if preservedClassValues[class] {
			preserved = append(preserved, class)
		}
	}

	return strings.Join(preserved, " ")
}

// findBody finds the body element in the parsed HTML tree
func (s *CanvasHTMLStripper) findBody(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "body" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if body := s.findBody(c); body != nil {
			return body
		}
	}
	return nil
}
