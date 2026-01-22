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
