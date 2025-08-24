package main

import (
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const Version = "0.1.0"

func main() {
	// Initialize handler
	handler := NewHandler()

	s := server.NewMCPServer(
		"slack-explorer-mcp",
		Version,
	)

	// Add search_messages tool
	s.AddTool(
		mcp.NewTool("search_messages",
			mcp.WithDescription("Search for messages with specific criteria/filters. Use this when: 1) You need to find messages from a specific user, 2) You need messages from a specific date range, 3) You need to search by keywords, 4) You want to filter by channel. This tool is optimized for targeted searches."),
			mcp.WithString("query",
				mcp.Description("Basic search query text only. Do NOT include modifiers like 'from:', 'in:', etc. - use the dedicated fields instead."),
			),
			mcp.WithString("in_channel",
				mcp.Description("Search within a specific channel. Specify the channel name (e.g., 'general', 'random', 'チーム-dev')."),
			),
			mcp.WithString("from_user",
				mcp.Description("Search for messages from a specific user. Must be a Slack user ID (e.g., 'U1234567')."),
			),
			mcp.WithArray("with",
				mcp.Items(
					map[string]interface{}{
						"type": "string",
					},
				),
				mcp.Description("Search for messages in threads and direct messages with specific users. Must be Slack user IDs (e.g., ['U1234567', 'U2345678']). Multiple users can be specified."),
			),
			mcp.WithString("before",
				mcp.Description("Search for messages before this date (YYYY-MM-DD)"),
			),
			mcp.WithString("after",
				mcp.Description("Search for messages after this date (YYYY-MM-DD)"),
			),
			mcp.WithString("on",
				mcp.Description("Search for messages on this specific date (YYYY-MM-DD)"),
			),
			mcp.WithString("during",
				mcp.Description("Search for messages during a specific time period (e.g., 'July', '2023', 'last week')"),
			),
			mcp.WithBoolean("highlight",
				mcp.Description("Enable highlighting of search results (default: false)"),
			),
			mcp.WithString("sort",
				mcp.Description("Search result sort method: 'score' or 'timestamp' (default: 'score')"),
			),
			mcp.WithString("sort_dir",
				mcp.Description("Sort direction: 'asc' or 'desc' (default: 'desc')"),
			),
			mcp.WithNumber("count",
				mcp.Description("Number of results per page (1-100, default: 20)"),
			),
			mcp.WithNumber("page",
				mcp.Description("Page number of results (1-100, default: 1)"),
			),
			mcp.WithArray("has",
				mcp.Items(
					map[string]interface{}{
						"type": "string",
					},
				),
				mcp.Description("Search for messages containing specific features. Supported values: emoji reactions (\":eyes:\", \":fire:\"), \"pin\" (pinned messages), \"file\" (messages with file attachments), \"link\" (messages with links), \"reaction\" (messages with any reactions). When specifying emoji reactions, they must be wrapped with colons (e.g., \":eyes:\"). Multiple values can be specified."),
			),
			mcp.WithArray("hasmy",
				mcp.Items(
					map[string]interface{}{
						"type": "string",
					},
				),
				mcp.Description("Search for messages where the authenticated user has specific emoji reactions. Only emoji codes are supported (e.g., [\":eyes:\", \":fire:\"]). Emoji codes must be wrapped with colons (e.g., \":eyes:\"). Multiple emoji reactions can be specified."),
			),
		),
		handler.SearchMessages,
	)

	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
		log.Fatalf("Failed to serve: %v", err)
	}
}
