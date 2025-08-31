package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/slack-go/slack"
)

// SlackClient is an interface for Slack API operations
type SlackClient interface {
	SearchMessages(query string, params slack.SearchParameters) (*slack.SearchMessages, error)
	GetConversationReplies(params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error)
	GetUserProfile(userID string) (*slack.UserProfile, error)
	GetUsers(ctx context.Context, options ...slack.GetUsersOption) ([]slack.User, error)
}

type slackClient struct {
	client *slack.Client
}

// NewSlackClient creates a new Slack API client with user token
func NewSlackClient(userToken string) SlackClient {
	return &slackClient{
		client: slack.New(userToken),
	}
}

// SearchMessages searches for messages in Slack workspace
func (c *slackClient) SearchMessages(query string, params slack.SearchParameters) (*slack.SearchMessages, error) {
	slog.Debug("Calling Slack API: SearchMessages",
		"query", query,
		"params", params,
	)
	messages, err := c.client.SearchMessages(query, params)
	if err != nil {
		slog.Debug("SearchMessages failed", "error", err)
		return nil, c.mapError(err)
	}
	slog.Debug("SearchMessages completed",
		"total_count", messages.Total,
	)
	return messages, nil
}

// GetConversationReplies retrieves replies to a message thread
func (c *slackClient) GetConversationReplies(params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error) {
	slog.Debug("Calling Slack API: GetConversationReplies",
		"channel_id", params.ChannelID,
		"thread_ts", params.Timestamp,
		"limit", params.Limit,
		"cursor", params.Cursor,
	)
	messages, hasMore, nextCursor, err := c.client.GetConversationReplies(params)
	if err != nil {
		slog.Debug("GetConversationReplies failed", "error", err)
		return nil, false, "", c.mapError(err)
	}
	slog.Debug("GetConversationReplies completed",
		"messages_count", len(messages),
		"has_more", hasMore,
		"next_cursor", nextCursor,
	)
	return messages, hasMore, nextCursor, nil
}

// GetUserProfile retrieves a user's profile information
func (c *slackClient) GetUserProfile(userID string) (*slack.UserProfile, error) {
	slog.Debug("Calling Slack API: GetUserProfile",
		"user_id", userID,
	)
	profile, err := c.client.GetUserProfile(&slack.GetUserProfileParameters{
		UserID: userID,
	})
	if err != nil {
		slog.Debug("GetUserProfile failed", "error", err)
		return nil, c.mapError(err)
	}
	slog.Debug("GetUserProfile completed",
		"user_id", userID,
		"display_name", profile.DisplayName,
		"real_name", profile.RealName,
	)
	return profile, nil
}

// GetUsers retrieves users from the workspace
// The slack-go library handles pagination internally when using GetUsersContext
func (c *slackClient) GetUsers(ctx context.Context, options ...slack.GetUsersOption) ([]slack.User, error) {
	slog.Debug("Calling Slack API: GetUsers",
		"options_count", len(options),
	)
	users, err := c.client.GetUsersContext(ctx, options...)
	if err != nil {
		slog.Debug("GetUsers failed", "error", err)
		return nil, c.mapError(err)
	}
	slog.Debug("GetUsers completed",
		"users_count", len(users),
	)
	return users, nil
}

func (c *slackClient) mapError(err error) error {
	if rateLimitErr, ok := err.(*slack.RateLimitedError); ok {
		return fmt.Errorf("rate limited: retry after %d seconds", rateLimitErr.RetryAfter)
	}

	if slackErr, ok := err.(slack.SlackErrorResponse); ok {
		switch slackErr.Err {
		case "not_authed", "invalid_auth":
			return fmt.Errorf("authentication failed: %s", slackErr.Err)
		case "missing_scope":
			return fmt.Errorf("missing required scope: %s", slackErr.Err)
		case "channel_not_found":
			return fmt.Errorf("channel not found: %s", slackErr.Err)
		case "user_not_found":
			return fmt.Errorf("user not found: %s", slackErr.Err)
		case "thread_not_found":
			return fmt.Errorf("thread not found: %s", slackErr.Err)
		default:
			return fmt.Errorf("slack API error: %s", slackErr.Err)
		}
	}

	return err
}
