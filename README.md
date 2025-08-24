# Slack Explorer MCP Server

A Model Context Protocol (MCP) server specialized in **retrieving information** from Slack messages and threads. It provides tools to access messages that the authenticated user can view using a User Token (xoxp).

## Available Tools

- Message Search (`search_messages`)
  - Search Slack messages with advanced filtering options. You can search by channel, user, date range, and specific features (reactions, files, etc.).
  - Parameters
    - `query`: Basic search query (without modifiers)
    - `in_channel`: Filter by channel name (e.g., "general", "team-dev")
    - `from_user`: Search messages from specific user (User ID)
    - `with`: Search DMs/threads with specific users (array of User IDs)
    - `before`, `after`, `on`: Date range filtering (YYYY-MM-DD format)
    - `during`: Period specification (e.g., "July", "2023")
    - `has`: Messages containing specific features (emoji, "pin", "file", "link", "reaction")
    - `hasmy`: Messages where you reacted with specific emoji
    - `sort`: Sort method ("score" or "timestamp")
    - `count`: Number of results per page (1-100, default: 20)
    - `page`: Page number (1-100, default: 1)

- Thread Replies (`get_thread_replies`)
  - Get all replies in a message thread. Supports pagination for efficiently handling large numbers of replies.
  - Parameters
    - `channel_id`: Channel ID (required)
    - `thread_ts`: Parent message timestamp (required)
    - `limit`: Number of replies to retrieve (1-1000, default: 100)
    - `cursor`: Pagination cursor

- User Profiles (`get_user_profiles`)
  - Get profile information for multiple users in bulk. Retrieve display names, real names, email addresses, and other profile information by specifying a list of user IDs.
  - Parameters
    - `user_ids`: Array of user IDs (required, max 100)

## Setup

### Getting a Slack User Token

1. Create an app at [Slack API](https://api.slack.com/apps)
2. Add the following User Token Scopes in OAuth & Permissions:
   - `search:read` - For message search
   - `channels:read`, `channels:history` - For public channels
   - `groups:read`, `groups:history` - For private channels
   - `im:read`, `im:history` - For direct messages
   - `mpim:read`, `mpim:history` - For group direct messages
   - `users:read`, `users.profile:read` - For user information and profiles
3. Install the app to your workspace
4. Get the User OAuth Token (starts with xoxp-)

### MCP Server Configuration

1. Configure mcp.json

    ```json
    {
      "mcpServers": {
        "slack-explorer-mcp": {
          "command": "docker",
          "args": ["run", "-i", "--rm",
            "-e", "SLACK_USER_TOKEN=xoxp-your-token-here",
            "ghcr.io/shibayu36/slack-explorer-mcp:latest"
          ]
        }
      }
    }
    ```

    If you're using Claude Code:

    ```bash
    claude mcp add slack-explorer-mcp -- docker run -i --rm \
      -e SLACK_USER_TOKEN=xoxp-your-token-here \
      ghcr.io/shibayu36/slack-explorer-mcp:latest
    ```

2. Use the agent to perform Slack searches

    Examples:
    - "Search for meeting-related messages in the general channel from last week"
    - "Find messages from @john.doe about 'project'"
    - "Get all thread replies for this post"

## Usage

### Common Search Patterns

- **Search in a specific channel**
  ```
  Search for "release" messages in the general channel
  ```

- **Search messages from a specific user**
  ```
  Search for yesterday's messages from @john.doe
  ```

- **Search messages with reactions**
  ```
  Search for messages with :fire: reactions
  ```

- **Search messages you reacted to**
  ```
  Search for messages where you reacted with :eyes:
  ```

- **Search messages with file attachments**
  ```
  Search for messages with file attachments
  ```
