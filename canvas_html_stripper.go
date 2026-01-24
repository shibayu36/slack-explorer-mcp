package main

// CanvasHTMLStripper strips unnecessary HTML elements and attributes from Slack Canvas HTML.
type CanvasHTMLStripper struct {
}

// NewCanvasHTMLStripper creates a new CanvasHTMLStripper instance
func NewCanvasHTMLStripper() *CanvasHTMLStripper {
	return &CanvasHTMLStripper{}
}

// Strip removes unnecessary HTML elements and attributes from the input HTML.
// Currently returns the input unchanged (skeleton implementation for Commit 1).
func (s *CanvasHTMLStripper) Strip(html string) (string, error) {
	return html, nil
}
