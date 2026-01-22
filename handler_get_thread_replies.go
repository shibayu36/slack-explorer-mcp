package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/slack-go/slack"
)

// GetThreadRepliesResponse represents the response structure for get_thread_replies
type GetThreadRepliesResponse struct {
	Messages   []ThreadMessage `json:"messages"`
	HasMore    bool            `json:"has_more"`
	NextCursor string          `json:"next_cursor,omitempty"`
}

type ThreadMessage struct {
	User       string     `json:"user"`
	Text       string     `json:"text"`
	Timestamp  string     `json:"ts"`
	ReplyCount int        `json:"reply_count,omitempty"`
	ReplyUsers []string   `json:"reply_users,omitempty"`
	Reactions  []Reaction `json:"reactions,omitempty"`
}

type Reaction struct {
	Name  string   `json:"name"`
	Count int      `json:"count"`
	Users []string `json:"users"`
}

// GetThreadReplies handles the get_thread_replies tool call
func (h *Handler) GetThreadReplies(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := h.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(ErrSlackTokenNotConfigured), nil
	}

	params, err := h.buildThreadRepliesParams(buildThreadRepliesRequest{
		ChannelID: request.GetString("channel_id", ""),
		ThreadTS:  request.GetString("thread_ts", ""),
		Limit:     request.GetInt("limit", 100),
		Cursor:    request.GetString("cursor", ""),
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	messages, hasMore, nextCursor, err := client.GetConversationReplies(params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	response := h.convertToThreadResponse(messages, hasMore, nextCursor)

	jsonData, err := json.Marshal(response)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type buildThreadRepliesRequest struct {
	ChannelID string
	ThreadTS  string
	Limit     int
	Cursor    string
}

func (h *Handler) buildThreadRepliesParams(request buildThreadRepliesRequest) (*slack.GetConversationRepliesParameters, error) {
	if request.ChannelID == "" {
		return nil, fmt.Errorf("channel_id is required")
	}
	if !strings.HasPrefix(request.ChannelID, "C") {
		return nil, fmt.Errorf("invalid channel ID format. Must start with 'C' (e.g., 'C1234567')")
	}

	if request.ThreadTS == "" {
		return nil, fmt.Errorf("thread_ts is required")
	}
	tsPattern := regexp.MustCompile(`^\d{10}\.\d{6}$`)
	if !tsPattern.MatchString(request.ThreadTS) {
		return nil, fmt.Errorf("thread_ts must be in format '1234567890.123456'")
	}

	if request.Limit < 1 || request.Limit > 1000 {
		return nil, fmt.Errorf("limit must be between 1 and 1000, got %d", request.Limit)
	}

	params := &slack.GetConversationRepliesParameters{
		ChannelID: request.ChannelID,
		Timestamp: request.ThreadTS,
		Limit:     request.Limit,
	}
	if request.Cursor != "" {
		params.Cursor = request.Cursor
	}

	return params, nil
}

func (h *Handler) convertToThreadResponse(messages []slack.Message, hasMore bool, nextCursor string) *GetThreadRepliesResponse {
	response := &GetThreadRepliesResponse{
		Messages: make([]ThreadMessage, 0, len(messages)),
		HasMore:  hasMore,
	}

	if nextCursor != "" {
		response.NextCursor = nextCursor
	}

	for _, msg := range messages {
		threadMsg := ThreadMessage{
			User:      msg.User,
			Text:      msg.Text,
			Timestamp: msg.Timestamp,
		}

		if msg.ReplyCount > 0 {
			threadMsg.ReplyCount = msg.ReplyCount
		}

		if len(msg.ReplyUsers) > 0 {
			threadMsg.ReplyUsers = msg.ReplyUsers
		}

		if len(msg.Reactions) > 0 {
			reactions := make([]Reaction, 0, len(msg.Reactions))
			for _, reaction := range msg.Reactions {
				reactions = append(reactions, Reaction{
					Name:  reaction.Name,
					Count: reaction.Count,
					Users: reaction.Users,
				})
			}
			threadMsg.Reactions = reactions
		}

		response.Messages = append(response.Messages, threadMsg)
	}

	return response
}
