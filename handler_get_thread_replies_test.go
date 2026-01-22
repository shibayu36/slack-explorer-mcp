package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestHandler_GetThreadReplies(t *testing.T) {
	t.Run("can get thread replies with messages", func(t *testing.T) {
		mockClient := &SlackClientMock{}

		messages := []slack.Message{
			{
				Msg: slack.Msg{
					User:       "U1234567",
					Text:       "Original message",
					Timestamp:  "1234567890.123456",
					ReplyCount: 2,
					ReplyUsers: []string{"U2345678", "U3456789"},
				},
			},
			{
				Msg: slack.Msg{
					User:      "U2345678",
					Text:      "Reply message 1",
					Timestamp: "1234567891.123456",
				},
			},
			{
				Msg: slack.Msg{
					User:      "U3456789",
					Text:      "Reply message 2",
					Timestamp: "1234567892.123456",
					Reactions: []slack.ItemReaction{
						{
							Name:  "thumbsup",
							Count: 2,
							Users: []string{"U1234567", "U2345678"},
						},
					},
				},
			},
		}
		hasMore := false
		nextCursor := ""

		expectedParams := &slack.GetConversationRepliesParameters{
			ChannelID: "C1234567",
			Timestamp: "1234567890.123456",
			Limit:     50,
		}

		mockClient.On("GetConversationReplies", expectedParams).Return(messages, hasMore, nextCursor, nil)

		handler := &Handler{
			getClient: func(ctx context.Context) (SlackClient, error) {
				return mockClient, nil
			},
		}

		req := mcp.CallToolRequest{
			Params: struct {
				Name      string    `json:"name"`
				Arguments any       `json:"arguments,omitempty"`
				Meta      *mcp.Meta `json:"_meta,omitempty"`
			}{
				Name: "get_thread_replies",
				Arguments: map[string]interface{}{
					"channel_id": "C1234567",
					"thread_ts":  "1234567890.123456",
					"limit":      50,
				},
			},
		}

		res, err := handler.GetThreadReplies(t.Context(), req)
		assert.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal([]byte(res.Content[0].(mcp.TextContent).Text), &response)
		assert.NoError(t, err)

		assert.Contains(t, response, "messages")
		messages_response := response["messages"].([]interface{})
		assert.Equal(t, 3, len(messages_response))

		firstMsg := messages_response[0].(map[string]interface{})
		assert.Equal(t, "U1234567", firstMsg["user"])
		assert.Equal(t, "Original message", firstMsg["text"])
		assert.Equal(t, "1234567890.123456", firstMsg["ts"])
		assert.Equal(t, float64(2), firstMsg["reply_count"])
		assert.Equal(t, []interface{}{"U2345678", "U3456789"}, firstMsg["reply_users"])

		secondMsg := messages_response[1].(map[string]interface{})
		assert.Equal(t, "U2345678", secondMsg["user"])
		assert.Equal(t, "Reply message 1", secondMsg["text"])
		assert.Equal(t, "1234567891.123456", secondMsg["ts"])

		thirdMsg := messages_response[2].(map[string]interface{})
		assert.Equal(t, "U3456789", thirdMsg["user"])
		assert.Equal(t, "Reply message 2", thirdMsg["text"])
		assert.Contains(t, thirdMsg, "reactions")
		reactions := thirdMsg["reactions"].([]interface{})
		assert.Equal(t, 1, len(reactions))

		reaction := reactions[0].(map[string]interface{})
		assert.Equal(t, "thumbsup", reaction["name"])
		assert.Equal(t, float64(2), reaction["count"])

		assert.Equal(t, false, response["has_more"])
		assert.NotContains(t, response, "next_cursor")

		mockClient.AssertExpectations(t)
	})

	t.Run("returns empty when no replies found", func(t *testing.T) {
		mockClient := &SlackClientMock{}

		messages := []slack.Message{}
		hasMore := false
		nextCursor := ""

		expectedParams := &slack.GetConversationRepliesParameters{
			ChannelID: "C1234567",
			Timestamp: "1234567890.123456",
			Limit:     100,
		}

		mockClient.On("GetConversationReplies", expectedParams).Return(messages, hasMore, nextCursor, nil)

		handler := &Handler{
			getClient: func(ctx context.Context) (SlackClient, error) {
				return mockClient, nil
			},
		}

		req := mcp.CallToolRequest{
			Params: struct {
				Name      string    `json:"name"`
				Arguments any       `json:"arguments,omitempty"`
				Meta      *mcp.Meta `json:"_meta,omitempty"`
			}{
				Name: "get_thread_replies",
				Arguments: map[string]interface{}{
					"channel_id": "C1234567",
					"thread_ts":  "1234567890.123456",
				},
			},
		}

		res, err := handler.GetThreadReplies(t.Context(), req)
		assert.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal([]byte(res.Content[0].(mcp.TextContent).Text), &response)
		assert.NoError(t, err)

		assert.Contains(t, response, "messages")
		messages_response := response["messages"].([]interface{})
		assert.Equal(t, 0, len(messages_response))

		assert.Equal(t, false, response["has_more"])
		assert.NotContains(t, response, "next_cursor")

		mockClient.AssertExpectations(t)
	})
}

func TestHandler_buildThreadRepliesParams(t *testing.T) {
	handler := &Handler{}

	t.Run("required parameters only", func(t *testing.T) {
		request := buildThreadRepliesRequest{
			ChannelID: "C1234567",
			ThreadTS:  "1234567890.123456",
			Limit:     100,
		}

		params, err := handler.buildThreadRepliesParams(request)

		assert.NoError(t, err)
		assert.Equal(t, &slack.GetConversationRepliesParameters{
			ChannelID: "C1234567",
			Timestamp: "1234567890.123456",
			Limit:     100,
		}, params)
	})

	t.Run("all parameters specified", func(t *testing.T) {
		request := buildThreadRepliesRequest{
			ChannelID: "C1234567",
			ThreadTS:  "1234567890.123456",
			Limit:     50,
			Cursor:    "dXNlcjpVMDYxTkZUVDI=",
		}

		params, err := handler.buildThreadRepliesParams(request)

		assert.NoError(t, err)
		assert.Equal(t, &slack.GetConversationRepliesParameters{
			ChannelID: "C1234567",
			Timestamp: "1234567890.123456",
			Limit:     50,
			Cursor:    "dXNlcjpVMDYxTkZUVDI=",
		}, params)
	})

	t.Run("channel_id validation errors", func(t *testing.T) {
		testCases := []struct {
			name      string
			request   buildThreadRepliesRequest
			expectErr string
		}{
			{
				"empty channel_id",
				buildThreadRepliesRequest{ChannelID: "", ThreadTS: "1234567890.123456", Limit: 100},
				"channel_id is required",
			},
			{
				"invalid channel_id format",
				buildThreadRepliesRequest{ChannelID: "invalid123", ThreadTS: "1234567890.123456", Limit: 100},
				"invalid channel ID format. Must start with 'C' (e.g., 'C1234567')",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := handler.buildThreadRepliesParams(tc.request)

				assert.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			})
		}
	})

	t.Run("thread_ts validation errors", func(t *testing.T) {
		testCases := []struct {
			name      string
			request   buildThreadRepliesRequest
			expectErr string
		}{
			{
				"empty thread_ts",
				buildThreadRepliesRequest{ChannelID: "C1234567", ThreadTS: "", Limit: 100},
				"thread_ts is required",
			},
			{
				"invalid thread_ts format - missing dot",
				buildThreadRepliesRequest{ChannelID: "C1234567", ThreadTS: "1234567890123456", Limit: 100},
				"thread_ts must be in format '1234567890.123456'",
			},
			{
				"invalid thread_ts format - wrong digits",
				buildThreadRepliesRequest{ChannelID: "C1234567", ThreadTS: "123456789.12345", Limit: 100},
				"thread_ts must be in format '1234567890.123456'",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := handler.buildThreadRepliesParams(tc.request)

				assert.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			})
		}
	})

	t.Run("limit validation errors", func(t *testing.T) {
		testCases := []struct {
			name      string
			request   buildThreadRepliesRequest
			expectErr string
		}{
			{
				"limit too low",
				buildThreadRepliesRequest{ChannelID: "C1234567", ThreadTS: "1234567890.123456", Limit: 0},
				"limit must be between 1 and 1000, got 0",
			},
			{
				"limit too high",
				buildThreadRepliesRequest{ChannelID: "C1234567", ThreadTS: "1234567890.123456", Limit: 1001},
				"limit must be between 1 and 1000, got 1001",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := handler.buildThreadRepliesParams(tc.request)

				assert.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			})
		}
	})
}
