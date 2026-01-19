package main

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
)

// SlackClient is an interface for Slack API operations
type SlackClient interface {
	SearchMessages(query string, params slack.SearchParameters) (*slack.SearchMessages, error)
	SearchFiles(query string, params slack.SearchParameters) (*slack.SearchFiles, error)
	GetConversationReplies(params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error)
	GetUserProfile(userID string) (*slack.UserProfile, error)
	GetUsers(ctx context.Context, options ...slack.GetUsersOption) ([]slack.User, error)
	GetFileInfo(fileID string) (*slack.File, error)
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
	messages, err := c.client.SearchMessages(query, params)
	if err != nil {
		return nil, c.mapError(err)
	}
	return messages, nil
}

// SearchFiles searches for files in Slack workspace
func (c *slackClient) SearchFiles(query string, params slack.SearchParameters) (*slack.SearchFiles, error) {
	files, err := c.client.SearchFiles(query, params)
	if err != nil {
		return nil, c.mapError(err)
	}
	return files, nil
}

// GetConversationReplies retrieves replies to a message thread
func (c *slackClient) GetConversationReplies(params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error) {
	messages, hasMore, nextCursor, err := c.client.GetConversationReplies(params)
	if err != nil {
		return nil, false, "", c.mapError(err)
	}
	return messages, hasMore, nextCursor, nil
}

// GetUserProfile retrieves a user's profile information
func (c *slackClient) GetUserProfile(userID string) (*slack.UserProfile, error) {
	profile, err := c.client.GetUserProfile(&slack.GetUserProfileParameters{
		UserID: userID,
	})
	if err != nil {
		return nil, c.mapError(err)
	}
	return profile, nil
}

// GetUsers retrieves users from the workspace
// The slack-go library handles pagination internally when using GetUsersContext
func (c *slackClient) GetUsers(ctx context.Context, options ...slack.GetUsersOption) ([]slack.User, error) {
	users, err := c.client.GetUsersContext(ctx, options...)
	if err != nil {
		return nil, c.mapError(err)
	}
	return users, nil
}

// GetFileInfo retrieves file information by file ID
func (c *slackClient) GetFileInfo(fileID string) (*slack.File, error) {
	file, _, _, err := c.client.GetFileInfo(fileID, 0, 0)
	if err != nil {
		return nil, c.mapError(err)
	}
	return file, nil
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
		case "file_not_found":
			return fmt.Errorf("file not found: %s", slackErr.Err)
		default:
			return fmt.Errorf("slack API error: %s", slackErr.Err)
		}
	}

	return err
}
