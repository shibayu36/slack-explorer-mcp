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

// SearchMessagesResponse represents the output for search_messages tool
type SearchMessagesResponse struct {
	Messages *SearchMessagesMatches `json:"messages"`
}

type SearchMessagesMatches struct {
	Matches    []SearchMessage   `json:"matches"`
	Pagination *SearchPagination `json:"pagination,omitempty"`
}

type SearchMessage struct {
	User      string       `json:"user"`
	Username  string       `json:"username"`
	Text      string       `json:"text"`
	Timestamp string       `json:"ts"`
	Channel   *ChannelInfo `json:"channel,omitempty"`
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
	query, params, err := h.buildSearchParams(buildSearchParamsRequest{
		Query:     request.GetString("query", ""),
		InChannel: request.GetString("in_channel", ""),
		FromUser:  request.GetString("from_user", ""),
		Before:    request.GetString("before", ""),
		After:     request.GetString("after", ""),
		On:        request.GetString("on", ""),
		During:    request.GetString("during", ""),
		Has:       request.GetStringSlice("has", []string{}),
		HasMy:     request.GetStringSlice("hasmy", []string{}),
		Highlight: request.GetBool("highlight", false),
		Sort:      request.GetString("sort", "score"),
		SortDir:   request.GetString("sort_dir", "desc"),
		Count:     request.GetInt("count", 20),
		Page:      request.GetInt("page", 1),
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	searchResult, err := h.slackClient.SearchMessages(query, params)
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

type buildSearchParamsRequest struct {
	Query     string
	InChannel string
	FromUser  string
	Before    string
	After     string
	On        string
	During    string
	Has       []string
	HasMy     []string
	Highlight bool
	Sort      string
	SortDir   string
	Count     int
	Page      int
}

// buildSearchParams validates parameters, applies defaults, and builds search query and parameters
func (h *Handler) buildSearchParams(request buildSearchParamsRequest) (string, slack.SearchParameters, error) {
	var queryParts []string

	// Prevent modifiers in query field to enforce use of dedicated parameter fields
	if request.Query != "" {
		modifierPattern := regexp.MustCompile(`\b(from|in|before|after|on|during|has|is|with):`)
		if modifierPattern.MatchString(request.Query) {
			return "", slack.SearchParameters{}, fmt.Errorf("query field cannot contain modifiers (from:, in:, etc.). Please use the dedicated fields")
		}
		queryParts = append(queryParts, request.Query)
	}

	if request.InChannel != "" {
		queryParts = append(queryParts, fmt.Sprintf("in:%s", request.InChannel))
	}

	if request.FromUser != "" {
		if !strings.HasPrefix(request.FromUser, "U") {
			return "", slack.SearchParameters{}, fmt.Errorf("invalid user ID format. Must start with 'U' (e.g., 'U1234567')")
		}
		queryParts = append(queryParts, fmt.Sprintf("from:<@%s>", request.FromUser))
	}

	datePattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if request.Before != "" {
		if !datePattern.MatchString(request.Before) {
			return "", slack.SearchParameters{}, fmt.Errorf("before date must be in YYYY-MM-DD format")
		}
		queryParts = append(queryParts, fmt.Sprintf("before:%s", request.Before))
	}
	if request.After != "" {
		if !datePattern.MatchString(request.After) {
			return "", slack.SearchParameters{}, fmt.Errorf("after date must be in YYYY-MM-DD format")
		}
		queryParts = append(queryParts, fmt.Sprintf("after:%s", request.After))
	}
	if request.On != "" {
		if !datePattern.MatchString(request.On) {
			return "", slack.SearchParameters{}, fmt.Errorf("on date must be in YYYY-MM-DD format")
		}
		queryParts = append(queryParts, fmt.Sprintf("on:%s", request.On))
	}
	if request.During != "" {
		queryParts = append(queryParts, fmt.Sprintf("during:%s", request.During))
	}

	for _, has := range request.Has {
		if has != "" {
			queryParts = append(queryParts, fmt.Sprintf("has:%s", has))
		}
	}

	for _, hasmy := range request.HasMy {
		if hasmy != "" {
			queryParts = append(queryParts, fmt.Sprintf("hasmy:%s", hasmy))
		}
	}

	if request.Count < 1 || request.Count > 100 {
		return "", slack.SearchParameters{}, fmt.Errorf("count must be between 1 and 100, got %d", request.Count)
	}
	if request.Page < 1 || request.Page > 100 {
		return "", slack.SearchParameters{}, fmt.Errorf("page must be between 1 and 100, got %d", request.Page)
	}
	if request.Sort != "score" && request.Sort != "timestamp" {
		return "", slack.SearchParameters{}, fmt.Errorf("sort must be 'score' or 'timestamp', got '%s'", request.Sort)
	}
	if request.SortDir != "asc" && request.SortDir != "desc" {
		return "", slack.SearchParameters{}, fmt.Errorf("sort_dir must be 'asc' or 'desc', got '%s'", request.SortDir)
	}

	searchQuery := strings.Join(queryParts, " ")

	params := slack.SearchParameters{
		Sort:          request.Sort,
		SortDirection: request.SortDir,
		Highlight:     request.Highlight,
		Count:         request.Count,
		Page:          request.Page,
	}

	return searchQuery, params, nil
}

// convertToSearchResponse converts Slack API response to our response format
func (h *Handler) convertToSearchResponse(result *slack.SearchMessages) *SearchMessagesResponse {
	response := &SearchMessagesResponse{
		Messages: &SearchMessagesMatches{
			Matches: make([]SearchMessage, 0, len(result.Matches)),
		},
	}

	for _, match := range result.Matches {
		msg := SearchMessage{
			User:      match.User,
			Username:  match.Username,
			Text:      match.Text,
			Timestamp: match.Timestamp,
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
