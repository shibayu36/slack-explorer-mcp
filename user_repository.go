package main

import (
	"context"
	"strings"
	"sync"

	"github.com/slack-go/slack"
)

// SessionCache holds cached users for a specific session
type SessionCache struct {
	users []slack.User
}

// UserRepository manages user information with session-based caching
type UserRepository struct {
	sessionCaches map[SessionID]*SessionCache
	mu            sync.RWMutex
}

// NewUserRepository creates a new UserRepository
func NewUserRepository() *UserRepository {
	return &UserRepository{
		sessionCaches: make(map[SessionID]*SessionCache),
	}
}

// FindByDisplayName searches for users by display name
func (r *UserRepository) FindByDisplayName(
	ctx context.Context,
	client SlackClient,
	displayName string,
	exact bool,
) ([]slack.User, error) {
	sessionID := SessionIDFromContext(ctx)

	r.mu.RLock()
	cache, exists := r.sessionCaches[sessionID]
	r.mu.RUnlock()

	if exists && cache != nil {
		return r.searchInUsers(cache.users, displayName, exact), nil
	}

	users, err := client.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.sessionCaches[sessionID] = &SessionCache{
		users: users,
	}
	r.mu.Unlock()

	return r.searchInUsers(users, displayName, exact), nil
}

func (r *UserRepository) searchInUsers(users []slack.User, displayName string, exact bool) []slack.User {
	var matches []slack.User
	for _, user := range users {
		if exact {
			if user.Profile.DisplayName == displayName {
				matches = append(matches, user)
			}
		} else {
			if strings.Contains(user.Profile.DisplayName, displayName) {
				matches = append(matches, user)
			}
		}
	}
	return matches
}
