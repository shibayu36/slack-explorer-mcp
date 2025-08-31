package main

import (
	"testing"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_FindByDisplayName(t *testing.T) {
	t.Run("returns users when display name matches exactly", func(t *testing.T) {
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
					DisplayName: "jane.s",
					RealName:    "Jane Marie Smith",
					Email:       "jane@example.com",
				},
			},
		}
		mockClient.On("GetUsers", t.Context(), []slack.GetUsersOption(nil)).Return(users, nil)

		repo := NewUserRepository(mockClient)

		result, err := repo.FindByDisplayName(t.Context(), "jdoe", true)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "U1234567", result[0].ID)
		assert.Equal(t, "jdoe", result[0].Profile.DisplayName)
		mockClient.AssertExpectations(t)
	})

	t.Run("returns users with partial match when exact is false", func(t *testing.T) {
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
					DisplayName: "anne.smith",
					RealName:    "Anne Elizabeth Smith",
					Email:       "anne@example.com",
				},
			},
			{
				ID: "U3456789",
				Profile: slack.UserProfile{
					DisplayName: "jane.johnson",
					RealName:    "Jane Marie Johnson",
					Email:       "jane@example.com",
				},
			},
		}
		mockClient.On("GetUsers", t.Context(), []slack.GetUsersOption(nil)).Return(users, nil)

		repo := NewUserRepository(mockClient)

		// Search for "john" should match john.doe and jane.johnson
		result, err := repo.FindByDisplayName(t.Context(), "john", false)

		assert.NoError(t, err)
		assert.Len(t, result, 2)

		foundUsers := make(map[string]bool)
		for _, user := range result {
			foundUsers[user.ID] = true
		}
		assert.True(t, foundUsers["U1234567"]) // john.doe
		assert.True(t, foundUsers["U3456789"]) // jane.johnson

		mockClient.AssertExpectations(t)
	})

	t.Run("uses cache on subsequent calls without calling API", func(t *testing.T) {
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
		mockClient.On("GetUsers", t.Context(), []slack.GetUsersOption(nil)).Return(users, nil).Once()

		repo := NewUserRepository(mockClient)

		// First call - should call API
		result1, err1 := repo.FindByDisplayName(t.Context(), "jdoe", true)
		assert.NoError(t, err1)
		assert.Len(t, result1, 1)

		// Second call - should use cache, not call API again
		result2, err2 := repo.FindByDisplayName(t.Context(), "jdoe", true)
		assert.NoError(t, err2)
		assert.Len(t, result2, 1)
		assert.Equal(t, result1[0].ID, result2[0].ID)

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

		repo := NewUserRepository(mockClient)

		result, err := repo.FindByDisplayName(t.Context(), "Not Found User", true)

		assert.NoError(t, err)
		assert.Len(t, result, 0)
		mockClient.AssertExpectations(t)
	})
}
