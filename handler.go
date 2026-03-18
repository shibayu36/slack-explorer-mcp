package main

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
)

// Error messages for user-facing errors
const (
	ErrSlackTokenNotConfigured = "Slack user token is not configured. Please set your Slack user token."
)

type ChannelInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SearchPagination struct {
	TotalCount int `json:"total_count"`
	Page       int `json:"page"`
	PageCount  int `json:"page_count"`
	PerPage    int `json:"per_page"`
	First      int `json:"first"`
	Last       int `json:"last"`
}

type AttachmentFieldInfo struct {
	Title string `json:"title,omitempty"`
	Value string `json:"value,omitempty"`
}

type AttachmentInfo struct {
	Title   string                `json:"title,omitempty"`
	Text    string                `json:"text,omitempty"`
	FromURL string                `json:"from_url,omitempty"`
	Fields  []AttachmentFieldInfo `json:"fields,omitempty"`
}

func convertAttachments(attachments []slack.Attachment) []AttachmentInfo {
	var result []AttachmentInfo
	for _, a := range attachments {
		info := AttachmentInfo{
			Title:   a.Title,
			Text:    a.Text,
			FromURL: a.FromURL,
		}
		for _, f := range a.Fields {
			if f.Title != "" || f.Value != "" {
				info.Fields = append(info.Fields, AttachmentFieldInfo{
					Title: f.Title,
					Value: f.Value,
				})
			}
		}
		result = append(result, info)
	}
	return result
}

// UserProfile represents a user profile result
type UserProfile struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name,omitempty"`
	RealName    string `json:"real_name,omitempty"`
	Email       string `json:"email,omitempty"`
	Error       string `json:"error,omitempty"`
}

// Handler struct implements the MCP handler
type Handler struct {
	getClient      func(ctx context.Context) (SlackClient, error)
	userRepository *UserRepository
}

// NewHandler creates a new handler with Slack client
func NewHandler() *Handler {
	return &Handler{
		getClient: func(ctx context.Context) (SlackClient, error) {
			token, err := SlackUserTokenFromContext(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get Slack token from context: %w", err)
			}
			return NewSlackClient(token), nil
		},
		userRepository: NewUserRepository(),
	}
}

// Close releases resources owned by Handler.
func (h *Handler) Close() {
	h.userRepository.Close()
}
