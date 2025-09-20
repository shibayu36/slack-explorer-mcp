package main

import (
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		t.Cleanup(repo.Close)

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
		t.Cleanup(repo.Close)

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
		t.Cleanup(repo.Close)

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
		t.Cleanup(repo.Close)

		// Create contexts with different session IDs
		ctx1 := WithSessionID(t.Context(), SessionID("session-1"))
		ctx2 := WithSessionID(t.Context(), SessionID("session-2"))

		// Mock API calls for each session
		mockClient.On("GetUsers", ctx1, mock.Anything).Return(users1, nil).Once()
		mockClient.On("GetUsers", ctx2, mock.Anything).Return(users2, nil).Once()

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
		assert.Len(t, repo.sessionCaches, 2)
		cache1 := repo.sessionCaches[SessionID("session-1")]
		if assert.NotNil(t, cache1) {
			assert.Equal(t, users1, cache1.users)
		}
		cache2 := repo.sessionCaches[SessionID("session-2")]
		if assert.NotNil(t, cache2) {
			assert.Equal(t, users2, cache2.users)
		}

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

	t.Run("refreshes cache after ttl expiry", func(t *testing.T) {
		mockClient := &SlackClientMock{}
		usersInitial := []slack.User{
			{
				ID:      "U1111111",
				Profile: slack.UserProfile{DisplayName: "initial"},
			},
		}
		usersRefreshed := []slack.User{
			{
				ID:      "U2222222",
				Profile: slack.UserProfile{DisplayName: "refreshed"},
			},
		}

		now := time.Now()

		repo := NewUserRepository()
		repo.now = func() time.Time { return now }
		t.Cleanup(repo.Close)

		ctx := WithSessionID(t.Context(), SessionID("session-refresh"))

		mockClient.On("GetUsers", ctx, mock.Anything).Return(usersInitial, nil).Once()

		// First call populates cache
		result1, err1 := repo.FindByDisplayName(ctx, mockClient, "initial", true)
		assert.NoError(t, err1)
		assert.Len(t, result1, 1)
		assert.Equal(t, "U1111111", result1[0].ID)

		mockClient.On("GetUsers", ctx, mock.Anything).Return(usersRefreshed, nil).Once()

		now = now.Add(cacheTTL)

		// Use cache yet
		result2, err2 := repo.FindByDisplayName(ctx, mockClient, "initial", true)
		assert.NoError(t, err2)
		assert.Len(t, result2, 1)
		assert.Equal(t, "U1111111", result2[0].ID)

		// Advance time beyond TTL and expect refreshed data
		now = now.Add(time.Second)

		result3, err3 := repo.FindByDisplayName(ctx, mockClient, "refreshed", true)
		assert.NoError(t, err3)
		assert.Len(t, result3, 1)
		assert.Equal(t, "U2222222", result3[0].ID)

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
		t.Cleanup(repo.Close)

		result, err := repo.FindByDisplayName(t.Context(), mockClient, "Not Found User", true)

		assert.NoError(t, err)
		assert.Len(t, result, 0)
		mockClient.AssertExpectations(t)
	})
}

func TestUserRepository_sweepExpiredCaches(t *testing.T) {
	t.Run("removes expired cache via sweeper", func(t *testing.T) {
		mockClient := &SlackClientMock{}
		users := []slack.User{
			{
				ID:      "U0000001",
				Profile: slack.UserProfile{DisplayName: "foo"},
			},
		}

		now := time.Now()

		repo := NewUserRepository()
		repo.now = func() time.Time { return now }
		t.Cleanup(repo.Close)

		ctx := WithSessionID(t.Context(), SessionID("session-clean"))

		mockClient.On("GetUsers", ctx, mock.Anything).Return(users, nil).Once()

		// Prime cache
		_, err := repo.FindByDisplayName(ctx, mockClient, "foo", true)
		assert.NoError(t, err)
		assert.Len(t, repo.sessionCaches, 1)
		assert.Equal(t, users, repo.sessionCaches[SessionID("session-clean")].users)
		assert.Equal(t, now, repo.sessionCaches[SessionID("session-clean")].cachedAt)

		// Advance time beyond TTL
		now = now.Add(cacheTTL + time.Second)

		// Run sweeper
		repo.sweepExpiredCaches()

		repo.mu.RLock()
		_, exists := repo.sessionCaches[SessionID("session-clean")]
		repo.mu.RUnlock()
		assert.False(t, exists)

		mockClient.AssertExpectations(t)
	})
}
