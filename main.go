package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const Version = "0.8.1"

func main() {
	// Setup logging based on DEBUG environment variable
	if os.Getenv("DEBUG") == "1" {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
		slog.Debug("Debug logging enabled")
	}

	// Initialize handler
	handler := NewHandler()
	defer handler.Close()

	s := server.NewMCPServer(
		"slack-explorer-mcp",
		Version,
	)

	// Add search_messages tool
	s.AddTool(
		mcp.NewTool("search_messages",
			mcp.WithDescription(`Search for messages with specific criteria/filters. Use this when: 1) You need to find messages from a specific user, 2) You need messages from a specific date range, 3) You need to search by keywords, 4) You want to filter by channel. This tool is optimized for targeted searches.

Note: Response includes workspace_url, channel.id, and ts (timestamp) which can be used to construct Slack permalinks:
- Regular message (no thread_ts field): {workspace_url}/archives/{channel.id}/p{ts without dot}
- Thread reply (has thread_ts field): Same URL with ?thread_ts={thread_ts}&channel={channel.id}&message_ts={ts}`),
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
				mcp.Description("Search for messages during a specific time period (e.g., 'July', '2023')"),
			),
			mcp.WithBoolean("highlight",
				mcp.Description("Enable highlighting of search results (default: false)"),
				mcp.DefaultBool(false),
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
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithOpenWorldHintAnnotation(true),
		),
		handler.SearchMessages,
	)

	// Add get_thread_replies tool
	s.AddTool(
		mcp.NewTool("get_thread_replies",
			mcp.WithDescription(`Get all replies in a message thread

Note: To construct Slack permalinks from the response:
- Use workspace_url from search_messages tool response
- Thread reply URL: {workspace_url}/archives/{channel_id}/p{ts without dot}?thread_ts={thread_ts}&channel={channel_id}&message_ts={ts}
Where channel_id and thread_ts are the values provided as input parameters`),
			mcp.WithString("channel_id",
				mcp.Required(),
				mcp.Description("The ID of the channel containing the thread (e.g., 'C1234567')"),
			),
			mcp.WithString("thread_ts",
				mcp.Required(),
				mcp.Description("The timestamp of the parent message in format '1234567890.123456'"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Number of replies to retrieve (1-1000, default: 100)"),
			),
			mcp.WithString("cursor",
				mcp.Description("Pagination cursor for next page of results"),
			),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithOpenWorldHintAnnotation(true),
		),
		handler.GetThreadReplies,
	)

	// Add get_user_profiles tool
	s.AddTool(
		mcp.NewTool("get_user_profiles",
			mcp.WithDescription("Get multiple users profile information in bulk"),
			mcp.WithArray("user_ids",
				mcp.Required(),
				mcp.Items(
					map[string]interface{}{
						"type": "string",
					},
				),
				mcp.Description("Array of user IDs to retrieve profiles for (e.g., ['U1234567', 'U2345678']). Maximum 100 user IDs."),
			),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithOpenWorldHintAnnotation(true),
		),
		handler.GetUserProfiles,
	)

	// Add search_users_by_name tool
	s.AddTool(
		mcp.NewTool("search_users_by_name",
			mcp.WithDescription("Search users by display name"),
			mcp.WithString("display_name",
				mcp.Required(),
				mcp.Description("The display name to search for"),
			),
			mcp.WithBoolean("exact",
				mcp.Description("If true (default), performs exact match. If false, performs partial match"),
				mcp.DefaultBool(true),
			),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithOpenWorldHintAnnotation(true),
		),
		handler.SearchUsersByName,
	)

	// Add search_files tool
	s.AddTool(
		mcp.NewTool("search_files",
			mcp.WithDescription(`Search for files with specific criteria/filters. Use this when you need to find files (canvases, PDFs, images, etc.) in Slack.

Available file types for 'types' parameter:
- lists: Lists
- canvases: Canvases
- documents: Documents (Google Docs, etc.)
- emails: Emails
- images: Images
- pdfs: PDFs
- presentations: Presentations
- snippets: Snippets
- spreadsheets: Spreadsheets
- audio: Audio files
- videos: Video files`),
			mcp.WithString("query",
				mcp.Description("Basic search query text only. Do NOT include modifiers like 'from:', 'in:', 'type:', etc. - use the dedicated fields instead."),
			),
			mcp.WithArray("types",
				mcp.Items(
					map[string]interface{}{
						"type": "string",
					},
				),
				mcp.Description("File types to filter by (e.g., ['canvases', 'pdfs']). Multiple types can be specified."),
			),
			mcp.WithString("in_channel",
				mcp.Description("Search within a specific channel. Specify the channel name (e.g., 'general', 'random')."),
			),
			mcp.WithString("from_user",
				mcp.Description("Search for files from a specific user. Must be a Slack user ID (e.g., 'U1234567')."),
			),
			mcp.WithArray("with_user",
				mcp.Items(
					map[string]interface{}{
						"type": "string",
					},
				),
				mcp.Description("Search for files shared with specific users. Must be Slack user IDs (e.g., ['U1234567', 'U2345678'])."),
			),
			mcp.WithString("before",
				mcp.Description("Search for files before this date (YYYY-MM-DD)"),
			),
			mcp.WithString("after",
				mcp.Description("Search for files after this date (YYYY-MM-DD)"),
			),
			mcp.WithString("on",
				mcp.Description("Search for files on this specific date (YYYY-MM-DD)"),
			),
			mcp.WithNumber("count",
				mcp.Description("Number of results per page (1-100, default: 20)"),
			),
			mcp.WithNumber("page",
				mcp.Description("Page number of results (1-100, default: 1)"),
			),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithOpenWorldHintAnnotation(true),
		),
		handler.SearchFiles,
	)

	// Add get_canvas_content tool
	s.AddTool(
		mcp.NewTool("get_canvas_content",
			mcp.WithDescription("Get the content of Canvas files. Canvas IDs can be obtained from search_files results."),
			mcp.WithArray("canvas_ids",
				mcp.Required(),
				mcp.Items(
					map[string]interface{}{
						"type": "string",
					},
				),
				mcp.Description("Array of Canvas file IDs to retrieve content for (e.g., ['F12345678', 'F23456789']). IDs start with 'F'. Maximum 10 IDs."),
			),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithOpenWorldHintAnnotation(true),
		),
		handler.GetCanvasContent,
	)

	transport := os.Getenv("TRANSPORT")
	if transport == "" {
		transport = "stdio"
	}

	switch transport {
	case "stdio":
		if err := server.ServeStdio(s, server.WithStdioContextFunc(func(ctx context.Context) context.Context {
			ctx = WithSlackTokenFromEnv(ctx)

			// Add session ID from ClientSession
			if session := server.ClientSessionFromContext(ctx); session != nil {
				ctx = WithSessionID(ctx, SessionID(session.SessionID()))
			}

			return ctx
		})); err != nil {
			slog.Error("Failed to serve stdio", "error", err)
			os.Exit(1)
		}
	case "http":
		httpServer := server.NewStreamableHTTPServer(s,
			server.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
				ctx = WithSlackTokenFromHTTP(ctx, r)

				// Add session ID from ClientSession
				if session := server.ClientSessionFromContext(ctx); session != nil {
					ctx = WithSessionID(ctx, SessionID(session.SessionID()))
				}

				return ctx
			}),
		)
		host := os.Getenv("HTTP_HOST")
		if host == "" {
			host = "0.0.0.0"
		}
		port := os.Getenv("HTTP_PORT")
		if port == "" {
			port = "8080"
		}
		addr := host + ":" + port
		slog.Info("HTTP server listening", "address", addr)
		if err := httpServer.Start(addr); err != nil {
			slog.Error("Failed to serve http", "error", err)
			os.Exit(1)
		}
	default:
		slog.Error("Invalid transport type", "transport", transport, "valid_types", "stdio, http")
		os.Exit(1)
	}
}
