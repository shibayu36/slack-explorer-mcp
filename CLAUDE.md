# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
This is a Slack Explorer MCP (Model Context Protocol) server written in Go that provides tools for searching and exploring Slack messages. The server implements MCP protocol to expose Slack functionality as tools that can be used by AI assistants.

## Commands

### Build
```bash
go build -o slack-explorer-mcp
```

### Run
```bash
# Requires SLACK_USER_TOKEN environment variable
SLACK_USER_TOKEN=xoxp-your-token ./slack-explorer-mcp
```

### Dependencies
```bash
go mod download  # Download dependencies
go mod tidy      # Clean up dependencies
```

### Testing
```bash
go test ./...    # Run all tests
```

## Architecture

### Core Components
- `main.go`: Entry point that initializes the MCP server and registers tools
- `handler.go`: Contains the Handler struct that implements the business logic for each MCP tool

### MCP Tools Implementation Status
According to `docs/onetime/20250823-requirements-ja.md`, the following tools are planned (implementation priority order):
1. `search_messages` - Currently has a placeholder implementation in handler.go
2. `get_thread_replies` - Not yet implemented
3. `get_channel_history` - Not yet implemented
4. `list_channels` - Not yet implemented

### Key Dependencies
- `github.com/mark3labs/mcp-go`: MCP protocol implementation for Go
- The project uses MCP v0.38.0 based on go.mod

## Important Context

### Slack API Integration
- Uses User Token (xoxp) for authentication (stored in `SLACK_USER_TOKEN` environment variable)
- Required Slack OAuth scopes:
  - Public channels: `channels:read`, `channels:history`
  - Private channels: `groups:read`, `groups:history`
  - DMs: `im:read`, `im:history`, `mpim:read`, `mpim:history`
  - Users: `users:read`
  - Search: `search:read`

### Current Implementation State
- The project structure is set up but the actual Slack API integration is not yet implemented
- `handler.go` contains a TODO for implementing actual Slack search functionality
- No Slack SDK is currently imported (will need to add `github.com/slack-go/slack` or similar)

## Development Guidelines

### Adding New MCP Tools
1. Define the tool in `main.go` using `mcp.NewTool()`
2. Implement the handler method in `handler.go`
3. Follow the existing pattern: extract parameters, perform Slack API call, return formatted result

### Error Handling
- Use `mcp.NewToolResultError()` for error responses
- Map Slack API errors appropriately (auth errors, not found, rate limits, etc.)

### Pagination
- Use Slack's `next_cursor` mechanism for pagination
- Pass through cursors transparently to maintain state

## Code Guidelines

### Language
- Use English for all comments and documentation.
- Use Japanese for all comments and documentation when the code is written in Japanese.

### Appropriate Use of Comments
**⚠️ Important**: Keep comments to the absolute minimum. Do not comment on things that are immediately obvious from reading the code.

**Limited cases where comments should be written**:
1. **Public API documentation**: For documenting public methods and interfaces
2. **Background context**: When there's important background information not evident from the code
3. **Complex logic explanation**: When complex code requires explanation to understand the intended behavior
4. **Large code section separation**: For 100+ line code blocks where sectioning with explanatory comments is beneficial

## Documentation Guidelines

### Documentation Update Process
- Always update README_ja.md first when adding or updating documentation
- Then update README.md with English translation using simple, clear English that non-native speakers can easily understand

## Creating Pull Request
Use `.github/pull_request_template.md` to create pull request.
