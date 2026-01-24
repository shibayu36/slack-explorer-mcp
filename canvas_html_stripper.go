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

// processNode recursively processes all nodes in the tree, filtering attributes and transforming elements
func (s *CanvasHTMLStripper) processNode(n *html.Node) {
	if n.Type == html.ElementNode {
		n.Attr = s.filterAttributes(n.Attr)
	}

	// Save next sibling before processing, as node may be removed during iteration
	for c := n.FirstChild; c != nil; {
		next := c.NextSibling
		s.transformChild(n, c)
		c = next
	}
}

// transformChild handles the transformation of a child element from parent's perspective
func (s *CanvasHTMLStripper) transformChild(parent, child *html.Node) {
	if child.Type == html.ElementNode {
		switch child.Data {
		case "br":
			parent.RemoveChild(child)
		case "span":
			s.unwrapElement(parent, child)
		case "control":
			if s.isSlackEmoji(child) {
				s.convertSlackEmoji(parent, child)
			} else {
				s.processNode(child)
			}
		default:
			s.processNode(child)
		}
	} else {
		s.processNode(child)
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

// unwrapElement moves all children of target to its parent, then removes target
func (s *CanvasHTMLStripper) unwrapElement(parent, target *html.Node) {
	// Move all children of target to before target, then transform them
	for child := target.FirstChild; child != nil; {
		nextChild := child.NextSibling
		target.RemoveChild(child)
		parent.InsertBefore(child, target)
		s.transformChild(parent, child)
		child = nextChild
	}
	// Remove the now-empty target element
	parent.RemoveChild(target)
}

// isSlackEmoji checks if the control element contains a Slack emoji img tag
func (s *CanvasHTMLStripper) isSlackEmoji(n *html.Node) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "img" {
			for _, attr := range c.Attr {
				if attr.Key == "data-is-slack" {
					return true
				}
			}
		}
	}
	return false
}

// convertSlackEmoji converts a Slack emoji control element to plain text
func (s *CanvasHTMLStripper) convertSlackEmoji(parent, control *html.Node) {
	var emojiText string
	// The emoji text is a sibling of img inside control, not a child of img
	// HTML parser treats <img> as a void element, so :emoji: becomes a text node sibling
	for c := control.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			text := strings.TrimSpace(c.Data)
			if text != "" {
				emojiText = text
				break
			}
		}
	}

	if emojiText != "" {
		textNode := &html.Node{
			Type: html.TextNode,
			Data: emojiText,
		}
		parent.InsertBefore(textNode, control)
	}
	parent.RemoveChild(control)
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
