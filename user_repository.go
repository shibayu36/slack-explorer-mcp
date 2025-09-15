package main

import (
	"context"
	"strings"

	"github.com/slack-go/slack"
)

// UserRepository manages user information with caching
type UserRepository struct {
	cachedUsers []slack.User
}

// NewUserRepository creates a new UserRepository
func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// FindByDisplayName searches for users by display name
func (r *UserRepository) FindByDisplayName(
	ctx context.Context,
	client SlackClient,
	displayName string,
	exact bool,
) ([]slack.User, error) {
	// Load users if not cached yet
	if r.cachedUsers == nil {
		users, err := client.GetUsers(ctx)
		if err != nil {
			return nil, err
		}
		r.cachedUsers = users
	}

	// Search for users with matching display name
	var matches []slack.User
	for _, user := range r.cachedUsers {
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

	return matches, nil
}
