package main

import (
	"testing"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_FindByDisplayName(t *testing.T) {
	t.Run("returns users when display name matches", func(t *testing.T) {
		mockClient := &SlackClientMock{}
		users := []slack.User{
			{
				ID: "U1234567",
				Profile: slack.UserProfile{
					DisplayName: "John Doe",
					RealName:    "John Doe",
					Email:       "john@example.com",
				},
			},
			{
				ID: "U2345678",
				Profile: slack.UserProfile{
					DisplayName: "Jane Smith",
					RealName:    "Jane Smith",
					Email:       "jane@example.com",
				},
			},
		}
		mockClient.On("GetUsers", t.Context(), []slack.GetUsersOption(nil)).Return(users, nil)

		repo := NewUserRepository(mockClient)

		result, err := repo.FindByDisplayName(t.Context(), "John Doe")

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "U1234567", result[0].ID)
		assert.Equal(t, "John Doe", result[0].Profile.DisplayName)
		mockClient.AssertExpectations(t)
	})

	t.Run("uses cache on subsequent calls without calling API", func(t *testing.T) {
		mockClient := &SlackClientMock{}
		users := []slack.User{
			{
				ID: "U1234567",
				Profile: slack.UserProfile{
					DisplayName: "John Doe",
					RealName:    "John Doe",
					Email:       "john@example.com",
				},
			},
		}
		mockClient.On("GetUsers", t.Context(), []slack.GetUsersOption(nil)).Return(users, nil).Once()

		repo := NewUserRepository(mockClient)

		// First call - should call API
		result1, err1 := repo.FindByDisplayName(t.Context(), "John Doe")
		assert.NoError(t, err1)
		assert.Len(t, result1, 1)

		// Second call - should use cache, not call API again
		result2, err2 := repo.FindByDisplayName(t.Context(), "John Doe")
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
					DisplayName: "John Doe",
					RealName:    "John Doe",
					Email:       "john@example.com",
				},
			},
		}
		mockClient.On("GetUsers", t.Context(), []slack.GetUsersOption(nil)).Return(users, nil)

		repo := NewUserRepository(mockClient)

		result, err := repo.FindByDisplayName(t.Context(), "Not Found User")

		assert.NoError(t, err)
		assert.Len(t, result, 0)
		mockClient.AssertExpectations(t)
	})
}
