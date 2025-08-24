package main

import (
	"testing"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

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

func TestExtractThreadTsFromPermalink(t *testing.T) {
	tests := []struct {
		name      string
		permalink string
		expected  string
	}{
		{
			name:      "Valid permalink with thread_ts",
			permalink: "https://clustervr.slack.com/archives/C092E73H1A9/p1755857531080329?thread_ts=1755823401.732729",
			expected:  "1755823401.732729",
		},
		{
			name:      "Permalink without thread_ts",
			permalink: "https://clustervr.slack.com/archives/C092E73H1A9/p1755857531080329",
			expected:  "",
		},
		{
			name:      "Empty permalink",
			permalink: "",
			expected:  "",
		},
		{
			name:      "Permalink with thread_ts using & separator",
			permalink: "https://clustervr.slack.com/archives/C092E73H1A9/p1755857531080329?foo=bar&thread_ts=1755823401.732729",
			expected:  "1755823401.732729",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractThreadTsFromPermalink(tt.permalink)
			if result != tt.expected {
				t.Errorf("extractThreadTsFromPermalink(%q) = %q, expected %q", tt.permalink, result, tt.expected)
			}
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
