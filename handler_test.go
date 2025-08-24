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
