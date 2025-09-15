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

		repo := NewUserRepository()

		result, err := repo.FindByDisplayName(t.Context(), mockClient, "jdoe", true)

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

		repo := NewUserRepository()

		// Search for "john" should match john.doe and jane.johnson
		result, err := repo.FindByDisplayName(t.Context(), mockClient, "john", false)

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

		repo := NewUserRepository()

		// First call - should call API
		result1, err1 := repo.FindByDisplayName(t.Context(), mockClient, "jdoe", true)
		assert.NoError(t, err1)
		assert.Len(t, result1, 1)

		// Second call - should use cache, not call API again
		result2, err2 := repo.FindByDisplayName(t.Context(), mockClient, "jdoe", true)
		assert.NoError(t, err2)
		assert.Len(t, result2, 1)
		assert.Equal(t, result1[0].ID, result2[0].ID)

		mockClient.AssertExpectations(t)
	})

	t.Run("uses cache with different session ID", func(t *testing.T) {
		mockClient := &SlackClientMock{}
		users1 := []slack.User{
			{
				ID: "U1111111",
				Profile: slack.UserProfile{
					DisplayName: "session1user",
					RealName:    "Session 1 User",
					Email:       "session1@example.com",
				},
			},
		}
		users2 := []slack.User{
			{
				ID: "U2222222",
				Profile: slack.UserProfile{
					DisplayName: "session2user",
					RealName:    "Session 2 User",
					Email:       "session2@example.com",
				},
			},
		}

		repo := NewUserRepository()

		// Create contexts with different session IDs
		ctx1 := WithSessionID(t.Context(), SessionID("session-1"))
		ctx2 := WithSessionID(t.Context(), SessionID("session-2"))

		// Mock API calls for each session
		mockClient.On("GetUsers", ctx1, []slack.GetUsersOption(nil)).Return(users1, nil).Once()
		mockClient.On("GetUsers", ctx2, []slack.GetUsersOption(nil)).Return(users2, nil).Once()

		// First call with session 1
		result1, err1 := repo.FindByDisplayName(ctx1, mockClient, "session1user", true)
		assert.NoError(t, err1)
		assert.Len(t, result1, 1)
		assert.Equal(t, "U1111111", result1[0].ID)

		// First call with session 2 - should call API because different session
		result2, err2 := repo.FindByDisplayName(ctx2, mockClient, "session2user", true)
		assert.NoError(t, err2)
		assert.Len(t, result2, 1)
		assert.Equal(t, "U2222222", result2[0].ID)

		// sessionCaches structure should be expected
		assert.Equal(t, map[SessionID]*SessionCache{
			"session-1": {
				users: users1,
			},
			"session-2": {
				users: users2,
			},
		}, repo.sessionCaches)

		// Second call with session 1 - should use cache
		result3, err3 := repo.FindByDisplayName(ctx1, mockClient, "session1user", true)
		assert.NoError(t, err3)
		assert.Len(t, result3, 1)
		assert.Equal(t, "U1111111", result3[0].ID)

		// Verify that session 1 context doesn't return session 2 data
		result4, err4 := repo.FindByDisplayName(ctx1, mockClient, "session2user", true)
		assert.NoError(t, err4)
		assert.Len(t, result4, 0) // Should not find session2user in session1 cache

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

		repo := NewUserRepository()

		result, err := repo.FindByDisplayName(t.Context(), mockClient, "Not Found User", true)

		assert.NoError(t, err)
		assert.Len(t, result, 0)
		mockClient.AssertExpectations(t)
	})
}
