package main

import (
	"context"
	"encoding/json"
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
