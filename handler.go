package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/slack-go/slack"
)

// SearchMessagesRequest represents the input parameters for search_messages tool
type SearchMessagesRequest struct {
	Query     string `json:"query,omitempty"`
	InChannel string `json:"in_channel,omitempty"`
	FromUser  string `json:"from_user,omitempty"`
	Before    string `json:"before,omitempty"`
	After     string `json:"after,omitempty"`
	On        string `json:"on,omitempty"`
	During    string `json:"during,omitempty"`
	Highlight bool   `json:"highlight,omitempty"`
	Sort      string `json:"sort,omitempty"`
	SortDir   string `json:"sort_dir,omitempty"`
	Count     int    `json:"count,omitempty"`
	Page      int    `json:"page,omitempty"`
}

// SearchMessagesResponse represents the output for search_messages tool
type SearchMessagesResponse struct {
	OK       bool                   `json:"ok"`
	Messages *SearchMessagesMatches `json:"messages,omitempty"`
	Error    string                 `json:"error,omitempty"`
}

type SearchMessagesMatches struct {
	Matches    []SearchMessage   `json:"matches"`
	Pagination *SearchPagination `json:"pagination,omitempty"`
}

type SearchMessage struct {
	Type      string       `json:"type"`
	User      string       `json:"user,omitempty"`
	Text      string       `json:"text"`
	Timestamp string       `json:"ts"`
	Channel   *ChannelInfo `json:"channel,omitempty"`
	Permalink string       `json:"permalink,omitempty"`
}

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

// Handler struct implements the MCP handler
type Handler struct {
	slackClient SlackClient
}

// NewHandler creates a new handler with Slack client
func NewHandler() *Handler {
	userToken := os.Getenv("SLACK_USER_TOKEN")
	if userToken == "" {
		panic("SLACK_USER_TOKEN environment variable is not set")
	}

	return &Handler{
		slackClient: NewSlackClient(userToken),
	}
}

// SearchMessages handles the search_messages tool call
func (h *Handler) SearchMessages(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	req := SearchMessagesRequest{
		Query:     request.GetString("query", ""),
		InChannel: request.GetString("in_channel", ""),
		FromUser:  request.GetString("from_user", ""),
		Before:    request.GetString("before", ""),
		After:     request.GetString("after", ""),
		On:        request.GetString("on", ""),
		During:    request.GetString("during", ""),
		Highlight: request.GetBool("highlight", false),
		Sort:      request.GetString("sort", "score"),
		SortDir:   request.GetString("sort_dir", "desc"),
		Count:     request.GetInt("count", 20),
		Page:      request.GetInt("page", 1),
	}

	if err := h.validateSearchRequest(&req); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	searchResult, err := h.slackClient.SearchMessages(h.buildSearchQuery(req), slack.SearchParameters{
		Sort:          req.Sort,
		SortDirection: req.SortDir,
		Highlight:     req.Highlight,
		Count:         req.Count,
		Page:          req.Page,
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	response := h.convertToSearchResponse(searchResult)

	jsonData, err := json.Marshal(response)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// validateSearchRequest validates the search request parameters
func (h *Handler) validateSearchRequest(req *SearchMessagesRequest) error {
	// Prevent modifiers in query field to enforce use of dedicated parameter fields
	if req.Query != "" {
		modifierPattern := regexp.MustCompile(`\b(from|in|before|after|on|during|has|is|with):`)
		if modifierPattern.MatchString(req.Query) {
			return fmt.Errorf("query field cannot contain modifiers (from:, in:, etc.). Please use the dedicated fields")
		}
	}

	if req.InChannel != "" && !strings.HasPrefix(req.InChannel, "C") {
		return fmt.Errorf("invalid channel ID format. Must start with 'C' (e.g., 'C1234567')")
	}

	if req.FromUser != "" && !strings.HasPrefix(req.FromUser, "U") {
		return fmt.Errorf("invalid user ID format. Must start with 'U' (e.g., 'U1234567')")
	}

	datePattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if req.Before != "" && !datePattern.MatchString(req.Before) {
		return fmt.Errorf("before date must be in YYYY-MM-DD format")
	}
	if req.After != "" && !datePattern.MatchString(req.After) {
		return fmt.Errorf("after date must be in YYYY-MM-DD format")
	}
	if req.On != "" && !datePattern.MatchString(req.On) {
		return fmt.Errorf("on date must be in YYYY-MM-DD format")
	}

	if req.Count < 1 || req.Count > 100 {
		req.Count = 20
	}

	if req.Page < 1 || req.Page > 100 {
		req.Page = 1
	}

	if req.Sort != "score" && req.Sort != "timestamp" {
		req.Sort = "score"
	}

	if req.SortDir != "asc" && req.SortDir != "desc" {
		req.SortDir = "desc"
	}

	return nil
}

// buildSearchQuery builds the Slack search query from request parameters
func (h *Handler) buildSearchQuery(req SearchMessagesRequest) string {
	var queryParts []string

	if req.Query != "" {
		queryParts = append(queryParts, req.Query)
	}

	if req.InChannel != "" {
		queryParts = append(queryParts, fmt.Sprintf("in:%s", req.InChannel))
	}

	if req.FromUser != "" {
		queryParts = append(queryParts, fmt.Sprintf("from:<@%s>", req.FromUser))
	}

	if req.Before != "" {
		queryParts = append(queryParts, fmt.Sprintf("before:%s", req.Before))
	}
	if req.After != "" {
		queryParts = append(queryParts, fmt.Sprintf("after:%s", req.After))
	}
	if req.On != "" {
		queryParts = append(queryParts, fmt.Sprintf("on:%s", req.On))
	}
	if req.During != "" {
		queryParts = append(queryParts, fmt.Sprintf("during:%s", req.During))
	}

	return strings.Join(queryParts, " ")
}

// convertToSearchResponse converts Slack API response to our response format
func (h *Handler) convertToSearchResponse(result *slack.SearchMessages) *SearchMessagesResponse {
	response := &SearchMessagesResponse{
		OK: true,
		Messages: &SearchMessagesMatches{
			Matches: make([]SearchMessage, 0, len(result.Matches)),
		},
	}

	for _, match := range result.Matches {
		msg := SearchMessage{
			Type:      match.Type,
			User:      match.User,
			Text:      match.Text,
			Timestamp: match.Timestamp,
			Permalink: match.Permalink,
		}

		if match.Channel.ID != "" {
			msg.Channel = &ChannelInfo{
				ID:   match.Channel.ID,
				Name: match.Channel.Name,
			}
		}

		response.Messages.Matches = append(response.Messages.Matches, msg)
	}

	response.Messages.Pagination = &SearchPagination{
		TotalCount: result.Paging.Total,
		Page:       result.Paging.Page,
		PageCount:  result.Paging.Pages,
		PerPage:    result.Paging.Count,
		First:      1,
		Last:       result.Paging.Pages,
	}

	return response
}
