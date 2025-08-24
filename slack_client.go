package main

import (
	"fmt"

	"github.com/slack-go/slack"
)

// SlackClient is an interface for Slack API operations
type SlackClient interface {
	SearchMessages(query string, params slack.SearchParameters) (*slack.SearchMessages, error)
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
		default:
			return fmt.Errorf("slack API error: %s", slackErr.Err)
		}
	}

	return err
}
