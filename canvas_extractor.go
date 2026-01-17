package main

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

// ExtractCanvasContent removes unnecessary HTML elements (script, style, etc.)
// from Canvas HTML to reduce token usage for LLM processing.
func ExtractCanvasContent(htmlContent string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	removeUnnecessaryElements(doc)

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// removeUnnecessaryElements recursively removes script, style, link, meta, and other unnecessary tags
func removeUnnecessaryElements(n *html.Node) {
	var toRemove []*html.Node

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			switch strings.ToLower(c.Data) {
			case "script", "style", "link", "meta", "noscript", "iframe":
				toRemove = append(toRemove, c)
				continue
			}
		}
		removeUnnecessaryElements(c)
	}

	for _, node := range toRemove {
		n.RemoveChild(node)
	}
}
