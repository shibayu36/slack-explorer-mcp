package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestHandler_SearchMessages(t *testing.T) {
	t.Run("can search messages with parameters", func(t *testing.T) {
		mockClient := &SlackClientMock{}

		mockResponse := &slack.SearchMessages{
			Matches: []slack.SearchMessage{
				{
					Type:      "message",
					User:      "U1234567",
					Username:  "john",
					Text:      "This is a test message",
					Timestamp: "1234567890.123456",
					Permalink: "https://workspace.slack.com/archives/C1234567/p1234567890123456",
					Channel: slack.CtxChannel{
						ID:   "C1234567",
						Name: "general",
					},
				},
				{
					Type:      "message",
					User:      "U2345678",
					Username:  "jane",
					Text:      "Another test message",
					Timestamp: "1234567891.123456",
					Permalink: "https://workspace.slack.com/archives/C1234567/p1234567891123456?thread_ts=1234567890.123456",
					Channel: slack.CtxChannel{
						ID:   "C1234567",
						Name: "general",
					},
				},
			},
			Paging: slack.Paging{
				Count: 50,
				Total: 100,
				Page:  2,
				Pages: 2,
			},
			Total: 100,
		}

		expectedQuery := "test message in:general from:<@U1234567>"
		expectedParams := slack.SearchParameters{
			Sort:          "timestamp",
			SortDirection: "asc",
			Highlight:     true,
			Count:         50,
			Page:          2,
		}
		mockClient.On("SearchMessages", expectedQuery, expectedParams).Return(mockResponse, nil)

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
				Name: "search_messages",
				Arguments: map[string]interface{}{
					"query":      "test message",
					"in_channel": "general",
					"from_user":  "U1234567",
					"highlight":  true,
					"sort":       "timestamp",
					"sort_dir":   "asc",
					"count":      50,
					"page":       2,
				},
			},
		}

		res, err := handler.SearchMessages(t.Context(), req)
		assert.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal([]byte(res.Content[0].(mcp.TextContent).Text), &response)
		assert.NoError(t, err)

		assert.Equal(t, "https://workspace.slack.com", response["workspace_url"])
		assert.Contains(t, response, "messages")
		messages := response["messages"].(map[string]interface{})

		assert.Contains(t, messages, "matches")
		matches := messages["matches"].([]interface{})
		assert.Equal(t, 2, len(matches))

		firstMsg := matches[0].(map[string]interface{})
		assert.Equal(t, "U1234567", firstMsg["user"])
		assert.Equal(t, "This is a test message", firstMsg["text"])
		assert.Equal(t, "1234567890.123456", firstMsg["ts"])
		assert.Nil(t, firstMsg["thread_ts"])

		channel1 := firstMsg["channel"].(map[string]interface{})
		assert.Equal(t, "C1234567", channel1["id"])
		assert.Equal(t, "general", channel1["name"])

		secondMsg := matches[1].(map[string]interface{})
		assert.Equal(t, "U2345678", secondMsg["user"])
		assert.Equal(t, "Another test message", secondMsg["text"])
		assert.Equal(t, "1234567891.123456", secondMsg["ts"])
		assert.Equal(t, "1234567890.123456", secondMsg["thread_ts"])

		assert.Contains(t, messages, "pagination")
		pagination := messages["pagination"].(map[string]interface{})
		assert.Equal(t, float64(100), pagination["total_count"])
		assert.Equal(t, float64(2), pagination["page"])
		assert.Equal(t, float64(2), pagination["page_count"])
		assert.Equal(t, float64(50), pagination["per_page"])

		mockClient.AssertExpectations(t)
	})

	t.Run("returns empty when no messages found", func(t *testing.T) {
		mockClient := &SlackClientMock{}

		expectedQuery := "nonexistent query"
		expectedParams := slack.SearchParameters{
			Sort:          "score",
			SortDirection: "desc",
			Highlight:     false,
			Count:         20,
			Page:          1,
		}

		mockResponse := &slack.SearchMessages{
			Matches: []slack.SearchMessage{},
			Paging: slack.Paging{
				Count: 0,
				Total: 0,
				Page:  1,
				Pages: 0,
			},
			Total: 0,
		}

		mockClient.On("SearchMessages", expectedQuery, expectedParams).Return(mockResponse, nil)

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
				Name: "search_messages",
				Arguments: map[string]interface{}{
					"query": "nonexistent query",
				},
			},
		}

		res, err := handler.SearchMessages(t.Context(), req)
		assert.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal([]byte(res.Content[0].(mcp.TextContent).Text), &response)
		assert.NoError(t, err)

		assert.Equal(t, "", response["workspace_url"])
		assert.Contains(t, response, "messages")
		messages := response["messages"].(map[string]interface{})
		matches := messages["matches"].([]interface{})
		assert.Equal(t, 0, len(matches))

		mockClient.AssertExpectations(t)
	})
}

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

func TestHandler_buildSearchParams(t *testing.T) {
	handler := &Handler{}

	t.Run("query only with defaults", func(t *testing.T) {
		request := buildSearchParamsRequest{
			Query:     "hello world",
			Highlight: false,
			Sort:      "score",
			SortDir:   "desc",
			Count:     20,
			Page:      1,
		}

		query, params, err := handler.buildSearchParams(request)

		assert.NoError(t, err)
		assert.Equal(t, "hello world", query)
		assert.Equal(t, slack.SearchParameters{
			Sort:          "score",
			SortDirection: "desc",
			Highlight:     false,
			Count:         20,
			Page:          1,
		}, params)
	})

	t.Run("all parameters specified", func(t *testing.T) {
		request := buildSearchParamsRequest{
			Query:     "test message",
			InChannel: "general",
			FromUser:  "U1234567",
			With:      []string{"U2345678", "U3456789"},
			Before:    "2024-01-15",
			After:     "2024-01-01",
			On:        "2024-01-10",
			During:    "January",
			Has:       []string{":eyes:", "pin", "file"},
			HasMy:     []string{":fire:", ":thumbsup:"},
			Highlight: true,
			Sort:      "timestamp",
			SortDir:   "asc",
			Count:     50,
			Page:      2,
		}

		query, params, err := handler.buildSearchParams(request)

		assert.NoError(t, err)
		assert.Equal(t, "test message in:general from:<@U1234567> with:<@U2345678> with:<@U3456789> before:2024-01-15 after:2024-01-01 on:2024-01-10 during:January has::eyes: has:pin has:file hasmy::fire: hasmy::thumbsup:", query)
		assert.Equal(t, slack.SearchParameters{
			Sort:          "timestamp",
			SortDirection: "asc",
			Highlight:     true,
			Count:         50,
			Page:          2,
		}, params)
	})

	t.Run("query with modifiers should error", func(t *testing.T) {
		request := buildSearchParamsRequest{
			Query:   "hello from:someone",
			Sort:    "score",
			SortDir: "desc",
			Count:   20,
			Page:    1,
		}

		_, _, err := handler.buildSearchParams(request)

		assert.Error(t, err)
		assert.Equal(t, "query field cannot contain modifiers (from:, in:, etc.). Please use the dedicated fields", err.Error())
	})

	t.Run("invalid user ID format should error", func(t *testing.T) {
		testCases := []struct {
			name      string
			request   buildSearchParamsRequest
			expectErr string
		}{
			{
				"invalid from_user",
				buildSearchParamsRequest{FromUser: "invaliduser", Sort: "score", SortDir: "desc", Count: 20, Page: 1},
				"invalid user ID format. Must start with 'U' (e.g., 'U1234567')",
			},
			{
				"invalid with user",
				buildSearchParamsRequest{With: []string{"U1234567", "invaliduser"}, Sort: "score", SortDir: "desc", Count: 20, Page: 1},
				"invalid user ID format in with parameter: 'invaliduser'. Must start with 'U' (e.g., 'U1234567')",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, _, err := handler.buildSearchParams(tc.request)

				assert.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			})
		}
	})

	t.Run("invalid date formats should error", func(t *testing.T) {
		testCases := []struct {
			name      string
			request   buildSearchParamsRequest
			expectErr string
		}{
			{
				"invalid before",
				buildSearchParamsRequest{Before: "2024/01/01", Sort: "score", SortDir: "desc", Count: 20, Page: 1},
				"before date must be in YYYY-MM-DD format",
			},
			{
				"invalid after",
				buildSearchParamsRequest{After: "01-01-2024", Sort: "score", SortDir: "desc", Count: 20, Page: 1},
				"after date must be in YYYY-MM-DD format",
			},
			{
				"invalid on",
				buildSearchParamsRequest{On: "2024-1-1", Sort: "score", SortDir: "desc", Count: 20, Page: 1},
				"on date must be in YYYY-MM-DD format",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, _, err := handler.buildSearchParams(tc.request)

				assert.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			})
		}
	})

	t.Run("invalid count and page ranges should error", func(t *testing.T) {
		testCases := []struct {
			name      string
			request   buildSearchParamsRequest
			expectErr string
		}{
			{
				"count too low",
				buildSearchParamsRequest{Count: 0, Sort: "score", SortDir: "desc", Page: 1},
				"count must be between 1 and 100, got 0",
			},
			{
				"count too high",
				buildSearchParamsRequest{Count: 101, Sort: "score", SortDir: "desc", Page: 1},
				"count must be between 1 and 100, got 101",
			},
			{
				"page too low",
				buildSearchParamsRequest{Page: 0, Sort: "score", SortDir: "desc", Count: 20},
				"page must be between 1 and 100, got 0",
			},
			{
				"page too high",
				buildSearchParamsRequest{Page: 101, Sort: "score", SortDir: "desc", Count: 20},
				"page must be between 1 and 100, got 101",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, _, err := handler.buildSearchParams(tc.request)

				assert.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			})
		}
	})

	t.Run("invalid sort and sort_dir should error", func(t *testing.T) {
		testCases := []struct {
			name      string
			request   buildSearchParamsRequest
			expectErr string
		}{
			{
				"invalid sort",
				buildSearchParamsRequest{Sort: "invalid", SortDir: "desc", Count: 20, Page: 1},
				"sort must be 'score' or 'timestamp', got 'invalid'",
			},
			{
				"invalid sort_dir",
				buildSearchParamsRequest{Sort: "score", SortDir: "invalid", Count: 20, Page: 1},
				"sort_dir must be 'asc' or 'desc', got 'invalid'",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, _, err := handler.buildSearchParams(tc.request)

				assert.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			})
		}
	})
}

func TestHandler_ExtractThreadTsFromPermalink(t *testing.T) {
	handler := &Handler{}
	tests := []struct {
		name      string
		permalink string
		expected  string
	}{
		{
			name:      "Valid permalink with thread_ts",
			permalink: "https://workspace1.slack.com/archives/C092E73H1A9/p1755857531080329?thread_ts=1755823401.732729",
			expected:  "1755823401.732729",
		},
		{
			name:      "Permalink without thread_ts",
			permalink: "https://workspace1.slack.com/archives/C092E73H1A9/p1755857531080329",
			expected:  "",
		},
		{
			name:      "Empty permalink",
			permalink: "",
			expected:  "",
		},
		{
			name:      "Permalink with thread_ts using & separator",
			permalink: "https://workspace1.slack.com/archives/C092E73H1A9/p1755857531080329?foo=bar&thread_ts=1755823401.732729",
			expected:  "1755823401.732729",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.extractThreadTsFromPermalink(tt.permalink)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHandler_ExtractWorkspaceURLFromPermalink(t *testing.T) {
	handler := &Handler{}
	tests := []struct {
		name      string
		permalink string
		expected  string
	}{
		{
			name:      "Valid Slack URL with query parameters",
			permalink: "https://workspace.slack.com/archives/C092E73H1A9/p1755857531080329?thread_ts=1755823401.732729",
			expected:  "https://workspace.slack.com",
		},
		{
			name:      "Valid Slack URL with subdomain",
			permalink: "https://my-company.slack.com/archives/C092E73H1A9/p1755857531080329",
			expected:  "https://my-company.slack.com",
		},
		{
			name:      "Empty permalink",
			permalink: "",
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.extractWorkspaceURLFromPermalink(tt.permalink)
			assert.Equal(t, tt.expected, result)
		})
	}
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

func TestHandler_GetUserProfiles(t *testing.T) {
	t.Run("can retrieve profiles for multiple users", func(t *testing.T) {
		mockClient := &SlackClientMock{}
		mockClient.On("GetUserProfile", "U1234567").Return(&slack.UserProfile{
			DisplayName: "john",
			RealName:    "John Doe",
			Email:       "john@example.com",
		}, nil)
		mockClient.On("GetUserProfile", "U2345678").Return(&slack.UserProfile{
			DisplayName: "jane",
			RealName:    "Jane Doe",
			Email:       "jane@example.com",
		}, nil)

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
				Name: "get_user_profiles",
				Arguments: map[string]interface{}{
					"user_ids": []string{"U1234567", "U2345678"},
				},
			},
		}
		res, err := handler.GetUserProfiles(t.Context(), req)
		assert.NoError(t, err)

		var profiles []map[string]interface{}
		err = json.Unmarshal([]byte(res.Content[0].(mcp.TextContent).Text), &profiles)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(profiles))

		assert.Equal(t, "U1234567", profiles[0]["user_id"])
		assert.Equal(t, "john", profiles[0]["display_name"])
		assert.Equal(t, "John Doe", profiles[0]["real_name"])
		assert.Equal(t, "john@example.com", profiles[0]["email"])

		assert.Equal(t, "U2345678", profiles[1]["user_id"])
		assert.Equal(t, "jane", profiles[1]["display_name"])
		assert.Equal(t, "Jane Doe", profiles[1]["real_name"])
		assert.Equal(t, "jane@example.com", profiles[1]["email"])
	})

	t.Run("can retrieve other user profiles even if one user fails", func(t *testing.T) {
		mockClient := &SlackClientMock{}
		mockClient.On("GetUserProfile", "U1234567").Return(&slack.UserProfile{
			DisplayName: "john",
			RealName:    "John Doe",
			Email:       "john@example.com",
		}, nil)
		mockClient.On("GetUserProfile", "U2345678").Return(nil, errors.New("user not found"))

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
				Name: "get_user_profiles",
				Arguments: map[string]interface{}{
					"user_ids": []string{"U1234567", "U2345678"},
				},
			},
		}
		res, err := handler.GetUserProfiles(t.Context(), req)
		assert.NoError(t, err)

		var profiles []map[string]interface{}
		err = json.Unmarshal([]byte(res.Content[0].(mcp.TextContent).Text), &profiles)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(profiles))

		assert.Equal(t, "U1234567", profiles[0]["user_id"])
		assert.Equal(t, "john", profiles[0]["display_name"])
		assert.Equal(t, "John Doe", profiles[0]["real_name"])
		assert.Equal(t, "john@example.com", profiles[0]["email"])

		assert.Equal(t, "U2345678", profiles[1]["user_id"])
		assert.Equal(t, "user not found", profiles[1]["error"])
	})
}

func TestHandler_SearchUsersByName(t *testing.T) {
	t.Run("can search with exact match", func(t *testing.T) {
		mockClient := &SlackClientMock{}
		users := []slack.User{
			{
				ID: "U1234567",
				Profile: slack.UserProfile{
					DisplayName: "jdoe",
					RealName:    "John David Doe",
					Email:       "john@example.com",
				},
			},
			{
				ID: "U2345678",
				Profile: slack.UserProfile{
					DisplayName: "jane.j",
					RealName:    "Jane Marie Johnson",
					Email:       "jane@example.com",
				},
			},
		}
		mockClient.On("GetUsers", t.Context(), []slack.GetUsersOption(nil)).Return(users, nil)

		userRepo := NewUserRepository()
		handler := &Handler{
			getClient: func(ctx context.Context) (SlackClient, error) {
				return mockClient, nil
			},
			userRepository: userRepo,
		}

		req := mcp.CallToolRequest{
			Params: struct {
				Name      string    `json:"name"`
				Arguments any       `json:"arguments,omitempty"`
				Meta      *mcp.Meta `json:"_meta,omitempty"`
			}{
				Name: "search_users_by_name",
				Arguments: map[string]interface{}{
					"display_name": "jdoe",
					"exact":        true,
				},
			},
		}

		res, err := handler.SearchUsersByName(t.Context(), req)
		assert.NoError(t, err)

		var profiles []map[string]interface{}
		err = json.Unmarshal([]byte(res.Content[0].(mcp.TextContent).Text), &profiles)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(profiles))

		assert.Equal(t, "U1234567", profiles[0]["user_id"])
		assert.Equal(t, "jdoe", profiles[0]["display_name"])
		assert.Equal(t, "John David Doe", profiles[0]["real_name"])
		assert.Equal(t, "john@example.com", profiles[0]["email"])

		mockClient.AssertExpectations(t)
	})

	t.Run("can search with partial match", func(t *testing.T) {
		mockClient := &SlackClientMock{}
		users := []slack.User{
			{
				ID: "U1234567",
				Profile: slack.UserProfile{
					DisplayName: "john.doe",
					RealName:    "John David Doe",
					Email:       "john@example.com",
				},
			},
			{
				ID: "U2345678",
				Profile: slack.UserProfile{
					DisplayName: "jane.johnson",
					RealName:    "Jane Marie Johnson",
					Email:       "jane@example.com",
				},
			},
			{
				ID: "U3456789",
				Profile: slack.UserProfile{
					DisplayName: "anne.smith",
					RealName:    "Anne Elizabeth Smith",
					Email:       "anne@example.com",
				},
			},
		}
		mockClient.On("GetUsers", t.Context(), []slack.GetUsersOption(nil)).Return(users, nil)

		userRepo := NewUserRepository()
		handler := &Handler{
			getClient: func(ctx context.Context) (SlackClient, error) {
				return mockClient, nil
			},
			userRepository: userRepo,
		}

		req := mcp.CallToolRequest{
			Params: struct {
				Name      string    `json:"name"`
				Arguments any       `json:"arguments,omitempty"`
				Meta      *mcp.Meta `json:"_meta,omitempty"`
			}{
				Name: "search_users_by_name",
				Arguments: map[string]interface{}{
					"display_name": "john",
					"exact":        false,
				},
			},
		}

		res, err := handler.SearchUsersByName(t.Context(), req)
		assert.NoError(t, err)

		var profiles []map[string]interface{}
		err = json.Unmarshal([]byte(res.Content[0].(mcp.TextContent).Text), &profiles)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(profiles))

		assert.Equal(t, "U1234567", profiles[0]["user_id"])
		assert.Equal(t, "john.doe", profiles[0]["display_name"])

		assert.Equal(t, "U2345678", profiles[1]["user_id"])
		assert.Equal(t, "jane.johnson", profiles[1]["display_name"])

		mockClient.AssertExpectations(t)
	})

	t.Run("returns empty array when no matches found", func(t *testing.T) {
		mockClient := &SlackClientMock{}
		users := []slack.User{
			{
				ID: "U1234567",
				Profile: slack.UserProfile{
					DisplayName: "jdoe",
					RealName:    "John David Doe",
					Email:       "john@example.com",
				},
			},
		}
		mockClient.On("GetUsers", t.Context(), []slack.GetUsersOption(nil)).Return(users, nil)

		userRepo := NewUserRepository()
		handler := &Handler{
			getClient: func(ctx context.Context) (SlackClient, error) {
				return mockClient, nil
			},
			userRepository: userRepo,
		}

		req := mcp.CallToolRequest{
			Params: struct {
				Name      string    `json:"name"`
				Arguments any       `json:"arguments,omitempty"`
				Meta      *mcp.Meta `json:"_meta,omitempty"`
			}{
				Name: "search_users_by_name",
				Arguments: map[string]interface{}{
					"display_name": "NonExistent",
					"exact":        true,
				},
			},
		}

		res, err := handler.SearchUsersByName(t.Context(), req)
		assert.NoError(t, err)

		var profiles []map[string]interface{}
		err = json.Unmarshal([]byte(res.Content[0].(mcp.TextContent).Text), &profiles)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(profiles))

		mockClient.AssertExpectations(t)
	})
}
