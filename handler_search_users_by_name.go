package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// SearchUsersByName searches for users by display name
func (h *Handler) SearchUsersByName(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := h.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(ErrSlackTokenNotConfigured), nil
	}

	displayName := request.GetString("display_name", "")
	if displayName == "" {
		return mcp.NewToolResultError("display_name is required"), nil
	}
	exact := request.GetBool("exact", true)

	users, err := h.userRepository.FindByDisplayName(ctx, client, displayName, exact)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Convert slack.User to UserProfile
	var profiles []UserProfile
	for _, user := range users {
		profiles = append(profiles, UserProfile{
			UserID:      user.ID,
			DisplayName: user.Profile.DisplayName,
			RealName:    user.Profile.RealName,
			Email:       user.Profile.Email,
		})
	}

	jsonData, err := json.Marshal(profiles)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}
