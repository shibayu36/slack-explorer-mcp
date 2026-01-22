package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func (h *Handler) GetUserProfiles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := h.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(ErrSlackTokenNotConfigured), nil
	}

	userIDs := request.GetStringSlice("user_ids", []string{})

	if len(userIDs) == 0 {
		return mcp.NewToolResultError("user_ids is required and cannot be empty"), nil
	}
	if len(userIDs) > 100 {
		return mcp.NewToolResultError("user_ids cannot exceed 100 entries"), nil
	}

	var profiles []UserProfile

	for _, userID := range userIDs {
		profile := h.getUserProfile(client, userID)
		profiles = append(profiles, profile)
	}

	jsonData, err := json.Marshal(profiles)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func (h *Handler) getUserProfile(client SlackClient, userID string) UserProfile {
	if !strings.HasPrefix(userID, "U") {
		return UserProfile{
			UserID: userID,
			Error:  "invalid user ID format. Must start with 'U' (e.g., 'U1234567')",
		}
	}

	slackProfile, err := client.GetUserProfile(userID)
	if err != nil {
		return UserProfile{
			UserID: userID,
			Error:  err.Error(),
		}
	}

	return UserProfile{
		UserID:      userID,
		DisplayName: slackProfile.DisplayName,
		RealName:    slackProfile.RealName,
		Email:       slackProfile.Email,
	}
}
