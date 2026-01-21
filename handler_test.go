package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_buildSearchFilesParams(t *testing.T) {
	handler := &Handler{}

	t.Run("query only with defaults", func(t *testing.T) {
		request := buildSearchFilesParamsRequest{
			Query: "hello world",
			Count: 20,
			Page:  1,
		}

		query, params, err := handler.buildSearchFilesParams(request)

		assert.NoError(t, err)
		assert.Equal(t, "hello world", query)
		assert.Equal(t, slack.SearchParameters{
			Sort:          "timestamp",
			SortDirection: "desc",
			Count:         20,
			Page:          1,
		}, params)
	})

	t.Run("all parameters specified", func(t *testing.T) {
		request := buildSearchFilesParamsRequest{
			Query:     "test file",
			Types:     []string{"canvases", "pdfs"},
			InChannel: "general",
			FromUser:  "U1234567",
			WithUsers: []string{"U2345678", "U3456789"},
			Before:    "2024-01-15",
			After:     "2024-01-01",
			On:        "2024-01-10",
			Count:     50,
			Page:      2,
		}

		query, params, err := handler.buildSearchFilesParams(request)

		assert.NoError(t, err)
		assert.Equal(t, "test file type:canvases type:pdfs in:general from:<@U1234567> with:<@U2345678> with:<@U3456789> before:2024-01-15 after:2024-01-01 on:2024-01-10", query)
		assert.Equal(t, slack.SearchParameters{
			Sort:          "timestamp",
			SortDirection: "desc",
			Count:         50,
			Page:          2,
		}, params)
	})

	t.Run("query with modifiers should error", func(t *testing.T) {
		request := buildSearchFilesParamsRequest{
			Query: "hello from:someone",
			Count: 20,
			Page:  1,
		}

		_, _, err := handler.buildSearchFilesParams(request)

		assert.Error(t, err)
		assert.Equal(t, "query field cannot contain modifiers (from:, in:, type:, etc.). Please use the dedicated fields", err.Error())
	})

	t.Run("invalid user ID format should error", func(t *testing.T) {
		testCases := []struct {
			name      string
			request   buildSearchFilesParamsRequest
			expectErr string
		}{
			{
				"invalid from_user",
				buildSearchFilesParamsRequest{FromUser: "invaliduser", Count: 20, Page: 1},
				"invalid user ID format. Must start with 'U' (e.g., 'U1234567')",
			},
			{
				"invalid with_users",
				buildSearchFilesParamsRequest{WithUsers: []string{"U1234567", "invaliduser"}, Count: 20, Page: 1},
				"invalid user ID format in with_users parameter: 'invaliduser'. Must start with 'U' (e.g., 'U1234567')",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, _, err := handler.buildSearchFilesParams(tc.request)

				assert.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			})
		}
	})

	t.Run("invalid date formats should error", func(t *testing.T) {
		testCases := []struct {
			name      string
			request   buildSearchFilesParamsRequest
			expectErr string
		}{
			{
				"invalid before",
				buildSearchFilesParamsRequest{Before: "2024/01/01", Count: 20, Page: 1},
				"before date must be in YYYY-MM-DD format",
			},
			{
				"invalid after",
				buildSearchFilesParamsRequest{After: "01-01-2024", Count: 20, Page: 1},
				"after date must be in YYYY-MM-DD format",
			},
			{
				"invalid on",
				buildSearchFilesParamsRequest{On: "2024-1-1", Count: 20, Page: 1},
				"on date must be in YYYY-MM-DD format",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, _, err := handler.buildSearchFilesParams(tc.request)

				assert.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			})
		}
	})

	t.Run("invalid count and page ranges should error", func(t *testing.T) {
		testCases := []struct {
			name      string
			request   buildSearchFilesParamsRequest
			expectErr string
		}{
			{
				"count too low",
				buildSearchFilesParamsRequest{Count: 0, Page: 1},
				"count must be between 1 and 100, got 0",
			},
			{
				"count too high",
				buildSearchFilesParamsRequest{Count: 101, Page: 1},
				"count must be between 1 and 100, got 101",
			},
			{
				"page too low",
				buildSearchFilesParamsRequest{Page: 0, Count: 20},
				"page must be between 1 and 100, got 0",
			},
			{
				"page too high",
				buildSearchFilesParamsRequest{Page: 101, Count: 20},
				"page must be between 1 and 100, got 101",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, _, err := handler.buildSearchFilesParams(tc.request)

				assert.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			})
		}
	})
}

func TestHandler_SearchFiles(t *testing.T) {
	t.Run("can search files with parameters", func(t *testing.T) {
		mockClient := &SlackClientMock{}

		mockResponse := &slack.SearchFiles{
			Matches: []slack.File{
				{
					ID:        "F12345678",
					Title:     "Project Plan",
					Filetype:  "canvas",
					User:      "U1234567",
					Channels:  []string{"C1234567"},
					Created:   slack.JSONTime(1704067200),
					Timestamp: slack.JSONTime(1704153600),
					Permalink: "https://workspace.slack.com/files/U1234567/F12345678/project_plan",
				},
				{
					ID:        "F23456789",
					Title:     "Report.pdf",
					Filetype:  "pdf",
					User:      "U2345678",
					Channels:  []string{"C1234567", "C2345678"},
					Created:   slack.JSONTime(1704067300),
					Timestamp: slack.JSONTime(1704153700),
					Permalink: "https://workspace.slack.com/files/U2345678/F23456789/report.pdf",
				},
			},
			Paging: slack.Paging{
				Count: 20,
				Total: 100,
				Page:  1,
				Pages: 5,
			},
		}

		expectedQuery := "project type:canvases in:general from:<@U1234567>"
		expectedParams := slack.SearchParameters{
			Sort:          "timestamp",
			SortDirection: "desc",
			Count:         20,
			Page:          1,
		}
		mockClient.On("SearchFiles", expectedQuery, expectedParams).Return(mockResponse, nil)

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
				Name: "search_files",
				Arguments: map[string]interface{}{
					"query":      "project",
					"types":      []interface{}{"canvases"},
					"in_channel": "general",
					"from_user":  "U1234567",
				},
			},
		}

		res, err := handler.SearchFiles(t.Context(), req)
		assert.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal([]byte(res.Content[0].(mcp.TextContent).Text), &response)
		assert.NoError(t, err)

		files := response["files"].([]interface{})
		assert.Equal(t, 2, len(files))

		firstFile := files[0].(map[string]interface{})
		assert.Equal(t, "F12345678", firstFile["id"])
		assert.Equal(t, "Project Plan", firstFile["title"])
		assert.Equal(t, "canvas", firstFile["filetype"])
		assert.Equal(t, "U1234567", firstFile["user"])
		assert.Equal(t, float64(1704067200), firstFile["created"])
		assert.Equal(t, float64(1704153600), firstFile["updated"])

		pagination := response["pagination"].(map[string]interface{})
		assert.Equal(t, float64(100), pagination["total_count"])
		assert.Equal(t, float64(1), pagination["page"])
		assert.Equal(t, float64(5), pagination["page_count"])
		assert.Equal(t, float64(20), pagination["per_page"])

		mockClient.AssertExpectations(t)
	})

	t.Run("returns empty when no files found", func(t *testing.T) {
		mockClient := &SlackClientMock{}

		expectedQuery := "nonexistent"
		expectedParams := slack.SearchParameters{
			Sort:          "timestamp",
			SortDirection: "desc",
			Count:         20,
			Page:          1,
		}

		mockResponse := &slack.SearchFiles{
			Matches: []slack.File{},
			Paging: slack.Paging{
				Count: 0,
				Total: 0,
				Page:  1,
				Pages: 0,
			},
		}

		mockClient.On("SearchFiles", expectedQuery, expectedParams).Return(mockResponse, nil)

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
				Name: "search_files",
				Arguments: map[string]interface{}{
					"query": "nonexistent",
				},
			},
		}

		res, err := handler.SearchFiles(t.Context(), req)
		assert.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal([]byte(res.Content[0].(mcp.TextContent).Text), &response)
		assert.NoError(t, err)

		files := response["files"].([]interface{})
		assert.Equal(t, 0, len(files))

		mockClient.AssertExpectations(t)
	})
}

func TestHandler_GetCanvasContent(t *testing.T) {
	t.Run("can get multiple canvases content", func(t *testing.T) {
		mockClient := &SlackClientMock{}
		mockClient.On("GetFileInfo", "F1234567").Return(&slack.File{
			ID:                 "F1234567",
			Title:              "Canvas 1",
			URLPrivateDownload: "https://files.slack.com/F1234567.html",
			Permalink:          "https://workspace.slack.com/files/U123/F1234567/canvas_1",
		}, nil)
		mockClient.On("GetFile", "https://files.slack.com/F1234567.html", mock.Anything).Run(func(args mock.Arguments) {
			w := args.Get(1).(io.Writer)
			w.Write([]byte("<html>Content 1</html>"))
		}).Return(nil)

		mockClient.On("GetFileInfo", "F2345678").Return(&slack.File{
			ID:                 "F2345678",
			Title:              "Canvas 2",
			URLPrivateDownload: "https://files.slack.com/F2345678.html",
			Permalink:          "https://workspace.slack.com/files/U123/F2345678/canvas_2",
		}, nil)
		mockClient.On("GetFile", "https://files.slack.com/F2345678.html", mock.Anything).Run(func(args mock.Arguments) {
			w := args.Get(1).(io.Writer)
			w.Write([]byte("<html>Content 2</html>"))
		}).Return(nil)

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
				Name: "get_canvas_content",
				Arguments: map[string]interface{}{
					"canvas_ids": []string{"F1234567", "F2345678"},
				},
			},
		}
		res, err := handler.GetCanvasContent(t.Context(), req)
		assert.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal([]byte(res.Content[0].(mcp.TextContent).Text), &response)
		assert.NoError(t, err)

		canvases := response["canvases"].([]interface{})
		assert.Equal(t, 2, len(canvases))

		canvas1 := canvases[0].(map[string]interface{})
		assert.Equal(t, "F1234567", canvas1["id"])
		assert.Equal(t, "Canvas 1", canvas1["title"])
		assert.Equal(t, "<html>Content 1</html>", canvas1["content"])
		assert.Equal(t, "https://workspace.slack.com/files/U123/F1234567/canvas_1", canvas1["permalink"])

		canvas2 := canvases[1].(map[string]interface{})
		assert.Equal(t, "F2345678", canvas2["id"])
		assert.Equal(t, "Canvas 2", canvas2["title"])
		assert.Equal(t, "<html>Content 2</html>", canvas2["content"])
		assert.Equal(t, "https://workspace.slack.com/files/U123/F2345678/canvas_2", canvas2["permalink"])

		mockClient.AssertExpectations(t)
	})

	t.Run("returns error when canvas_ids is empty", func(t *testing.T) {
		mockClient := &SlackClientMock{}

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
				Name: "get_canvas_content",
				Arguments: map[string]interface{}{
					"canvas_ids": []string{},
				},
			},
		}
		res, err := handler.GetCanvasContent(t.Context(), req)
		assert.NoError(t, err)
		assert.True(t, res.IsError)
		assert.Equal(t, "canvas_ids is required and cannot be empty", res.Content[0].(mcp.TextContent).Text)
	})

	t.Run("returns error when canvas_ids exceeds limit", func(t *testing.T) {
		mockClient := &SlackClientMock{}

		handler := &Handler{
			getClient: func(ctx context.Context) (SlackClient, error) {
				return mockClient, nil
			},
		}

		canvasIDs := make([]string, 21)
		for i := range canvasIDs {
			canvasIDs[i] = "F" + string(rune('A'+i))
		}

		req := mcp.CallToolRequest{
			Params: struct {
				Name      string    `json:"name"`
				Arguments any       `json:"arguments,omitempty"`
				Meta      *mcp.Meta `json:"_meta,omitempty"`
			}{
				Name: "get_canvas_content",
				Arguments: map[string]interface{}{
					"canvas_ids": canvasIDs,
				},
			},
		}
		res, err := handler.GetCanvasContent(t.Context(), req)
		assert.NoError(t, err)
		assert.True(t, res.IsError)
		assert.Equal(t, "canvas_ids cannot exceed 20 entries", res.Content[0].(mcp.TextContent).Text)
	})

	t.Run("returns success and multiple error types together", func(t *testing.T) {
		mockClient := &SlackClientMock{}

		// Success case
		mockClient.On("GetFileInfo", "F1234567").Return(&slack.File{
			ID:                 "F1234567",
			Title:              "Success Canvas",
			URLPrivateDownload: "https://files.slack.com/F1234567.html",
			Permalink:          "https://workspace.slack.com/files/U123/F1234567/success_canvas",
		}, nil)
		mockClient.On("GetFile", "https://files.slack.com/F1234567.html", mock.Anything).Run(func(args mock.Arguments) {
			w := args.Get(1).(io.Writer)
			w.Write([]byte("<html>Success</html>"))
		}).Return(nil)

		// Error: file has no download URL (both URLPrivateDownload and URLPrivate are empty)
		mockClient.On("GetFileInfo", "F2345678").Return(&slack.File{
			ID:                 "F2345678",
			Title:              "No URL Canvas",
			URLPrivateDownload: "",
			URLPrivate:         "",
		}, nil)

		// Error: download failed
		mockClient.On("GetFileInfo", "F3456789").Return(&slack.File{
			ID:                 "F3456789",
			Title:              "Download Fail Canvas",
			URLPrivateDownload: "https://files.slack.com/F3456789.html",
		}, nil)
		mockClient.On("GetFile", "https://files.slack.com/F3456789.html", mock.Anything).Return(errors.New("network error"))

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
				Name: "get_canvas_content",
				Arguments: map[string]interface{}{
					"canvas_ids": []string{"F1234567", "invalid123", "F2345678", "F3456789"},
				},
			},
		}
		res, err := handler.GetCanvasContent(t.Context(), req)
		assert.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal([]byte(res.Content[0].(mcp.TextContent).Text), &response)
		assert.NoError(t, err)

		canvases := response["canvases"].([]interface{})
		assert.Equal(t, 4, len(canvases))

		// Success
		canvas1 := canvases[0].(map[string]interface{})
		assert.Equal(t, "F1234567", canvas1["id"])
		assert.Equal(t, "Success Canvas", canvas1["title"])
		assert.Equal(t, "<html>Success</html>", canvas1["content"])
		assert.Equal(t, "https://workspace.slack.com/files/U123/F1234567/success_canvas", canvas1["permalink"])
		assert.Nil(t, canvas1["error"])

		// Invalid ID format
		canvas2 := canvases[1].(map[string]interface{})
		assert.Equal(t, "invalid123", canvas2["id"])
		assert.Equal(t, "invalid canvas ID format. Must start with 'F' (e.g., 'F1234567')", canvas2["error"])

		// No download URL
		canvas3 := canvases[2].(map[string]interface{})
		assert.Equal(t, "F2345678", canvas3["id"])
		assert.Equal(t, "file has no download URL", canvas3["error"])

		// Download failed
		canvas4 := canvases[3].(map[string]interface{})
		assert.Equal(t, "F3456789", canvas4["id"])
		assert.Equal(t, "failed to download file: network error", canvas4["error"])

		mockClient.AssertExpectations(t)
	})
}
