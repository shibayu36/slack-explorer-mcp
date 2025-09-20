package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

// slackUserTokenKey is the context key for Slack user token
type slackUserTokenKey struct{}

// sessionIDKey is the context key for session ID
type sessionIDKey struct{}

// SessionID represents a unique session identifier
type SessionID string

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

// WithSlackTokenFromHTTP adds Slack token from HTTP header to context
func WithSlackTokenFromHTTP(ctx context.Context, r *http.Request) context.Context {
	token := r.Header.Get("X-Slack-User-Token")
	if token == "" {
		// Return context as-is if no token found in header
		return ctx
	}
	return withSlackUserToken(ctx, token)
}

func withSlackUserToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, slackUserTokenKey{}, token)
}

// WithSessionID adds session ID to context
func WithSessionID(ctx context.Context, sessionID SessionID) context.Context {
	return context.WithValue(ctx, sessionIDKey{}, sessionID)
}

// SessionIDFromContext retrieves session ID from context
func SessionIDFromContext(ctx context.Context) SessionID {
	if sessionID, ok := ctx.Value(sessionIDKey{}).(SessionID); ok {
		return sessionID
	}
	return "default"
}
