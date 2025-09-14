package main

import (
	"context"
	"fmt"
	"os"
)

// slackUserTokenKey is the context key for Slack user token
type slackUserTokenKey struct{}

// SlackUserTokenFromContext retrieves Slack user token from context
func SlackUserTokenFromContext(ctx context.Context) (string, error) {
	token, ok := ctx.Value(slackUserTokenKey{}).(string)
	if !ok || token == "" {
		return "", fmt.Errorf("slack user token not found in context")
	}
	return token, nil
}

// WithSlackTokenFromEnv adds Slack token from environment variable to context
func WithSlackTokenFromEnv(ctx context.Context) context.Context {
	token := os.Getenv("SLACK_USER_TOKEN")
	if token == "" {
		// Return context as-is if no env var found
		return ctx
	}
	return withSlackUserToken(ctx, token)
}

func withSlackUserToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, slackUserTokenKey{}, token)
}
