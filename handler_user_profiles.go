package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// UserProfile represents a user profile result
type UserProfile struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name,omitempty"`
	RealName    string `json:"real_name,omitempty"`
	Email       string `json:"email,omitempty"`
	Error       string `json:"error,omitempty"`
}

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
