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
		assert.Equal(t, "Content 1", canvas1["content"]) // HTML stripped
		assert.Equal(t, "https://workspace.slack.com/files/U123/F1234567/canvas_1", canvas1["permalink"])

		canvas2 := canvases[1].(map[string]interface{})
		assert.Equal(t, "F2345678", canvas2["id"])
		assert.Equal(t, "Canvas 2", canvas2["title"])
		assert.Equal(t, "Content 2", canvas2["content"]) // HTML stripped
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
		assert.Equal(t, "Success", canvas1["content"]) // HTML stripped
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
