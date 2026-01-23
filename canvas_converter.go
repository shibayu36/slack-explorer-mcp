package main

import (
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// converterContext tracks state during HTML to Markdown conversion
type converterContext struct {
	listDepth   int    // Nesting level for lists
	listStyle   string // List type: "5"=bullet, "6"=numbered, "7"=checklist
	listCounter int    // Counter for numbered lists
}

// ConvertHTMLToMarkdown converts Slack Canvas HTML to Markdown
func ConvertHTMLToMarkdown(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		// If parsing fails, return original content
		return htmlContent
	}

	ctx := &converterContext{}
	result := convertNode(doc, ctx)

	// Clean up excessive newlines
	result = cleanupMarkdown(result)

	return result
}

// convertNode recursively converts HTML nodes to Markdown
func convertNode(node *html.Node, ctx *converterContext) string {
	if node == nil {
		return ""
	}

	var result strings.Builder

	switch node.Type {
	case html.TextNode:
		text := node.Data
		// Skip zero-width spaces and whitespace-only nodes in certain contexts
		if strings.TrimSpace(text) == "" || text == "\u200b" {
			return ""
		}
		return text

	case html.ElementNode:
		result.WriteString(convertElement(node, ctx))

	case html.DocumentNode:
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			result.WriteString(convertNode(child, ctx))
		}
	}

	return result.String()
}

// convertElement handles specific HTML elements
func convertElement(node *html.Node, ctx *converterContext) string {
	switch node.Data {
	case "h1":
		return "# " + getChildrenText(node, ctx) + "\n\n"
	case "h2":
		return "## " + getChildrenText(node, ctx) + "\n\n"
	case "h3":
		return "### " + getChildrenText(node, ctx) + "\n\n"

	case "p":
		class := getAttr(node, "class")
		if strings.Contains(class, "embedded-file") {
			return convertEmbeddedFile(node) + "\n\n"
		}
		if strings.Contains(class, "embedded-link") {
			return convertEmbeddedLink(node) + "\n\n"
		}
		text := getChildrenText(node, ctx)
		if text == "" {
			return ""
		}
		return text + "\n\n"

	case "div":
		sectionStyle := getAttr(node, "data-section-style")
		if sectionStyle != "" {
			return convertList(node, ctx, sectionStyle)
		}
		return getChildrenText(node, ctx)

	case "ul":
		return convertUL(node, ctx)

	case "li":
		return convertLI(node, ctx)

	case "table":
		return convertTable(node, ctx) + "\n\n"

	case "blockquote":
		return convertBlockquote(node, ctx) + "\n\n"

	case "pre":
		return convertCodeBlock(node) + "\n\n"

	case "b", "strong":
		return "**" + getChildrenText(node, ctx) + "**"

	case "i", "em":
		return "*" + getChildrenText(node, ctx) + "*"

	case "u":
		return "<u>" + getChildrenText(node, ctx) + "</u>"

	case "del", "s", "strike":
		return "~~" + getChildrenText(node, ctx) + "~~"

	case "a":
		href := getAttr(node, "href")
		text := getChildrenText(node, ctx)
		if href != "" {
			return "[" + text + "](" + href + ")"
		}
		return text

	case "br":
		return "\n"

	case "span":
		return getChildrenText(node, ctx)

	case "html", "head", "body":
		return getChildrenText(node, ctx)

	default:
		// Unknown elements: extract text from children (best effort)
		return getChildrenText(node, ctx)
	}
}

// getChildrenText gets text content from all children
func getChildrenText(node *html.Node, ctx *converterContext) string {
	var result strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		result.WriteString(convertNode(child, ctx))
	}
	return result.String()
}

// getAttr retrieves an attribute value from a node
func getAttr(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

// convertList handles list containers with data-section-style
func convertList(node *html.Node, ctx *converterContext, style string) string {
	oldStyle := ctx.listStyle
	ctx.listStyle = style
	ctx.listCounter = 0

	result := getChildrenText(node, ctx)

	ctx.listStyle = oldStyle
	return result
}

// convertUL handles unordered list elements
func convertUL(node *html.Node, ctx *converterContext) string {
	var result strings.Builder

	// Check if this is a nested ul (not inside div with data-section-style)
	parent := node.Parent
	isNested := false
	if parent != nil && parent.Data == "li" {
		isNested = true
	}
	if parent != nil && parent.Data == "ul" {
		isNested = true
	}

	if isNested {
		ctx.listDepth++
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		result.WriteString(convertNode(child, ctx))
	}

	if isNested {
		ctx.listDepth--
	}

	return result.String()
}

// convertLI handles list item elements
func convertLI(node *html.Node, ctx *converterContext) string {
	var result strings.Builder

	indent := strings.Repeat("  ", ctx.listDepth)
	ctx.listCounter++

	// Get text content (excluding nested ul)
	var textContent strings.Builder
	var nestedContent strings.Builder

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "ul" {
			// Handle nested list
			ctx.listDepth++
			nestedContent.WriteString(convertNode(child, ctx))
			ctx.listDepth--
		} else {
			textContent.WriteString(convertNode(child, ctx))
		}
	}

	text := strings.TrimSpace(textContent.String())

	switch ctx.listStyle {
	case "5": // Bullet list
		result.WriteString(indent + "- " + text + "\n")
	case "6": // Numbered list
		result.WriteString(indent + "1. " + text + "\n")
	case "7": // Checklist
		checked := strings.Contains(getAttr(node, "class"), "checked")
		if checked {
			result.WriteString(indent + "- [x] " + text + "\n")
		} else {
			result.WriteString(indent + "- [ ] " + text + "\n")
		}
	default:
		result.WriteString(indent + "- " + text + "\n")
	}

	result.WriteString(nestedContent.String())

	return result.String()
}

// convertTable handles table elements
func convertTable(node *html.Node, ctx *converterContext) string {
	var rows [][]string

	// Extract all rows
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode {
			switch child.Data {
			case "tr":
				row := extractTableRow(child, ctx)
				if len(row) > 0 {
					rows = append(rows, row)
				}
			case "thead", "tbody", "tfoot":
				for grandchild := child.FirstChild; grandchild != nil; grandchild = grandchild.NextSibling {
					if grandchild.Type == html.ElementNode && grandchild.Data == "tr" {
						row := extractTableRow(grandchild, ctx)
						if len(row) > 0 {
							rows = append(rows, row)
						}
					}
				}
			}
		}
	}

	if len(rows) == 0 {
		return ""
	}

	var result strings.Builder

	// Calculate column count
	colCount := 0
	for _, row := range rows {
		if len(row) > colCount {
			colCount = len(row)
		}
	}

	// Output header row
	result.WriteString("| " + strings.Join(padRow(rows[0], colCount), " | ") + " |\n")
	result.WriteString("|" + strings.Repeat(" --- |", colCount) + "\n")

	// Output data rows
	for i := 1; i < len(rows); i++ {
		result.WriteString("| " + strings.Join(padRow(rows[i], colCount), " | ") + " |\n")
	}

	return result.String()
}

// extractTableRow extracts cell content from a table row
func extractTableRow(tr *html.Node, ctx *converterContext) []string {
	var cells []string
	for child := tr.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && (child.Data == "td" || child.Data == "th") {
			text := strings.TrimSpace(getChildrenText(child, ctx))
			cells = append(cells, text)
		}
	}
	return cells
}

// padRow ensures row has correct number of columns
func padRow(row []string, colCount int) []string {
	for len(row) < colCount {
		row = append(row, "")
	}
	return row
}

// convertBlockquote handles blockquote elements
func convertBlockquote(node *html.Node, ctx *converterContext) string {
	content := getChildrenText(node, ctx)
	lines := strings.Split(content, "\n")
	var result strings.Builder
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result.WriteString("> " + line)
			if i < len(lines)-1 {
				result.WriteString("\n")
			}
		}
	}
	return result.String()
}

// convertCodeBlock handles pre elements
func convertCodeBlock(node *html.Node) string {
	// Get raw text content preserving line breaks
	content := getCodeBlockText(node)
	return "```\n" + content + "\n```"
}

// getCodeBlockText extracts text from code block, preserving br as newlines
func getCodeBlockText(node *html.Node) string {
	var result strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		switch child.Type {
		case html.TextNode:
			result.WriteString(child.Data)
		case html.ElementNode:
			if child.Data == "br" {
				result.WriteString("\n")
			} else {
				result.WriteString(getCodeBlockText(child))
			}
		}
	}
	return result.String()
}

// convertEmbeddedFile handles embedded file elements
func convertEmbeddedFile(node *html.Node) string {
	text := getChildrenText(node, &converterContext{})
	// Extract file URL from text like "File ID: F123 File URL: https://..."
	if strings.Contains(text, "File URL:") {
		parts := strings.Split(text, "File URL:")
		if len(parts) > 1 {
			url := strings.TrimSpace(parts[1])
			return "![File](" + url + ")"
		}
	}
	return text
}

// convertEmbeddedLink handles embedded link elements
func convertEmbeddedLink(node *html.Node) string {
	text := getChildrenText(node, &converterContext{})
	// Extract link URL from text like "Link URL: https://..."
	if strings.Contains(text, "Link URL:") {
		parts := strings.Split(text, "Link URL:")
		if len(parts) > 1 {
			url := strings.TrimSpace(parts[1])
			return "[Link](" + url + ")"
		}
	}
	return text
}

// cleanupMarkdown removes excessive newlines and whitespace
func cleanupMarkdown(content string) string {
	// Remove more than 2 consecutive newlines
	re := regexp.MustCompile(`\n{3,}`)
	content = re.ReplaceAllString(content, "\n\n")

	// Trim leading/trailing whitespace
	content = strings.TrimSpace(content)

	return content
}
