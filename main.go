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
			mcp.WithDescription("Search for messages in Slack"),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("The search query string"),
			),
		),
		handler.SearchMessages,
	)

	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
		log.Fatalf("Failed to serve: %v", err)
	}
}
