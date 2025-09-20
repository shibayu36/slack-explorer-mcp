package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const Version = "0.5.0"

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
				mcp.Description("Search for messages during a specific time period (e.g., 'July', '2023')"),
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
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithOpenWorldHintAnnotation(true),
		),
		handler.SearchMessages,
	)

	// Add get_thread_replies tool
	s.AddTool(
		mcp.NewTool("get_thread_replies",
			mcp.WithDescription("Get all replies in a message thread"),
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
			),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithOpenWorldHintAnnotation(true),
		),
		handler.SearchUsersByName,
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
		slog.Info("HTTP server listening on :8080")
		if err := httpServer.Start(":8080"); err != nil {
			slog.Error("Failed to serve http", "error", err)
			os.Exit(1)
		}
	default:
		slog.Error("Invalid transport type", "transport", transport, "valid_types", "stdio, http")
		os.Exit(1)
	}
}
