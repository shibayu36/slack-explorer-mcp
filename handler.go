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
	query, params, err := h.buildSearchParams(request)
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

// buildSearchParams validates parameters, applies defaults, and builds search query and parameters
func (h *Handler) buildSearchParams(request mcp.CallToolRequest) (string, slack.SearchParameters, error) {
	var queryParts []string

	// Extract parameters with defaults
	query := request.GetString("query", "")
	inChannel := request.GetString("in_channel", "")
	fromUser := request.GetString("from_user", "")
	before := request.GetString("before", "")
	after := request.GetString("after", "")
	on := request.GetString("on", "")
	during := request.GetString("during", "")
	highlight := request.GetBool("highlight", false)
	sort := request.GetString("sort", "score")
	sortDir := request.GetString("sort_dir", "desc")
	count := request.GetInt("count", 20)
	page := request.GetInt("page", 1)

	// Prevent modifiers in query field to enforce use of dedicated parameter fields
	if query != "" {
		modifierPattern := regexp.MustCompile(`\b(from|in|before|after|on|during|has|is|with):`)
		if modifierPattern.MatchString(query) {
			return "", slack.SearchParameters{}, fmt.Errorf("query field cannot contain modifiers (from:, in:, etc.). Please use the dedicated fields")
		}
		queryParts = append(queryParts, query)
	}

	if inChannel != "" {
		queryParts = append(queryParts, fmt.Sprintf("in:%s", inChannel))
	}

	if fromUser != "" {
		if !strings.HasPrefix(fromUser, "U") {
			return "", slack.SearchParameters{}, fmt.Errorf("invalid user ID format. Must start with 'U' (e.g., 'U1234567')")
		}
		queryParts = append(queryParts, fmt.Sprintf("from:<@%s>", fromUser))
	}

	datePattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if before != "" {
		if !datePattern.MatchString(before) {
			return "", slack.SearchParameters{}, fmt.Errorf("before date must be in YYYY-MM-DD format")
		}
		queryParts = append(queryParts, fmt.Sprintf("before:%s", before))
	}
	if after != "" {
		if !datePattern.MatchString(after) {
			return "", slack.SearchParameters{}, fmt.Errorf("after date must be in YYYY-MM-DD format")
		}
		queryParts = append(queryParts, fmt.Sprintf("after:%s", after))
	}
	if on != "" {
		if !datePattern.MatchString(on) {
			return "", slack.SearchParameters{}, fmt.Errorf("on date must be in YYYY-MM-DD format")
		}
		queryParts = append(queryParts, fmt.Sprintf("on:%s", on))
	}
	if during != "" {
		queryParts = append(queryParts, fmt.Sprintf("during:%s", during))
	}

	if count < 1 || count > 100 {
		return "", slack.SearchParameters{}, fmt.Errorf("count must be between 1 and 100, got %d", count)
	}
	if page < 1 || page > 100 {
		return "", slack.SearchParameters{}, fmt.Errorf("page must be between 1 and 100, got %d", page)
	}
	if sort != "score" && sort != "timestamp" {
		return "", slack.SearchParameters{}, fmt.Errorf("sort must be 'score' or 'timestamp', got '%s'", sort)
	}
	if sortDir != "asc" && sortDir != "desc" {
		return "", slack.SearchParameters{}, fmt.Errorf("sort_dir must be 'asc' or 'desc', got '%s'", sortDir)
	}

	searchQuery := strings.Join(queryParts, " ")

	params := slack.SearchParameters{
		Sort:          sort,
		SortDirection: sortDir,
		Highlight:     highlight,
		Count:         count,
		Page:          page,
	}

	return searchQuery, params, nil
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
