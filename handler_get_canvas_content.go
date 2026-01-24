package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// GetCanvasContentResponse represents the output for get_canvas_content tool
type GetCanvasContentResponse struct {
	Canvases []CanvasContent `json:"canvases"`
}

// CanvasContent represents a single canvas content result
type CanvasContent struct {
	ID        string `json:"id"`
	Title     string `json:"title,omitempty"`
	Content   string `json:"content,omitempty"`
	Permalink string `json:"permalink,omitempty"`
	Error     string `json:"error,omitempty"`
}

// GetCanvasContent retrieves content for multiple canvases
func (h *Handler) GetCanvasContent(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := h.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(ErrSlackTokenNotConfigured), nil
	}

	canvasIDs := request.GetStringSlice("canvas_ids", []string{})

	if len(canvasIDs) == 0 {
		return mcp.NewToolResultError("canvas_ids is required and cannot be empty"), nil
	}
	if len(canvasIDs) > 20 {
		return mcp.NewToolResultError("canvas_ids cannot exceed 20 entries"), nil
	}

	var canvases []CanvasContent

	for _, canvasID := range canvasIDs {
		canvas := h.getCanvasContent(client, canvasID)
		canvases = append(canvases, canvas)
	}

	response := GetCanvasContentResponse{
		Canvases: canvases,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func (h *Handler) getCanvasContent(client SlackClient, canvasID string) CanvasContent {
	if !strings.HasPrefix(canvasID, "F") {
		return CanvasContent{
			ID:    canvasID,
			Error: "invalid canvas ID format. Must start with 'F' (e.g., 'F1234567')",
		}
	}

	fileInfo, err := client.GetFileInfo(canvasID)
	if err != nil {
		return CanvasContent{
			ID:    canvasID,
			Error: err.Error(),
		}
	}

	downloadURL := fileInfo.URLPrivateDownload
	if downloadURL == "" {
		downloadURL = fileInfo.URLPrivate
	}
	if downloadURL == "" {
		return CanvasContent{
			ID:    canvasID,
			Error: "file has no download URL",
		}
	}

	var buf bytes.Buffer
	err = client.GetFile(downloadURL, &buf)
	if err != nil {
		return CanvasContent{
			ID:    canvasID,
			Error: fmt.Sprintf("failed to download file: %v", err),
		}
	}

	stripper := NewCanvasHTMLStripper()
	content, err := stripper.Strip(buf.String())
	if err != nil {
		return CanvasContent{
			ID:    canvasID,
			Error: fmt.Sprintf("failed to strip HTML: %v", err),
		}
	}

	return CanvasContent{
		ID:        canvasID,
		Title:     fileInfo.Title,
		Content:   content,
		Permalink: fileInfo.Permalink,
	}
}
