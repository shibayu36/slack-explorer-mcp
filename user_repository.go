package main

import (
	"context"

	"github.com/slack-go/slack"
)

// UserRepository manages user information with caching
type UserRepository struct {
	client      SlackClient
	cachedUsers []slack.User
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(client SlackClient) *UserRepository {
	return &UserRepository{
		client: client,
	}
}

// FindByDisplayName searches for users by display name (exact match)
func (r *UserRepository) FindByDisplayName(ctx context.Context, displayName string) ([]slack.User, error) {
	// Load users if not cached yet
	if r.cachedUsers == nil {
		users, err := r.client.GetUsers(ctx)
		if err != nil {
			return nil, err
		}
		r.cachedUsers = users
	}

	// Search for users with matching display name
	var matches []slack.User
	for _, user := range r.cachedUsers {
		if user.Profile.DisplayName == displayName {
			matches = append(matches, user)
		}
	}

	return matches, nil
}
