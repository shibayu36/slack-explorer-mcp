package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/slack-go/slack"
)

// SearchFilesResponse represents the output for search_files tool
type SearchFilesResponse struct {
	Pagination *SearchPagination `json:"pagination,omitempty"`
	Files      []FileInfo        `json:"files"`
}

type FileInfo struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Filetype  string   `json:"filetype"`
	User      string   `json:"user"`
	Channels  []string `json:"channels"`
	Created   int64    `json:"created"`
	Updated   int64    `json:"updated"`
	Permalink string   `json:"permalink"`
}

type buildSearchFilesParamsRequest struct {
	Query     string
	Types     []string
	InChannel string
	FromUser  string
	WithUsers []string
	Before    string
	After     string
	On        string
	Count     int
	Page      int
}

// SearchFiles handles the search_files tool call
func (h *Handler) SearchFiles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := h.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(ErrSlackTokenNotConfigured), nil
	}

	query, params, err := h.buildSearchFilesParams(buildSearchFilesParamsRequest{
		Query:     request.GetString("query", ""),
		Types:     request.GetStringSlice("types", []string{}),
		InChannel: request.GetString("in_channel", ""),
		FromUser:  request.GetString("from_user", ""),
		WithUsers: request.GetStringSlice("with_users", []string{}),
		Before:    request.GetString("before", ""),
		After:     request.GetString("after", ""),
		On:        request.GetString("on", ""),
		Count:     request.GetInt("count", 20),
		Page:      request.GetInt("page", 1),
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	searchResult, err := client.SearchFiles(query, params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	response := h.convertToFilesResponse(searchResult)

	jsonData, err := json.Marshal(response)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func (h *Handler) buildSearchFilesParams(request buildSearchFilesParamsRequest) (string, slack.SearchParameters, error) {
	var queryParts []string

	if request.Query != "" {
		modifierPattern := regexp.MustCompile(`\b(from|in|before|after|on|type):`)
		if modifierPattern.MatchString(request.Query) {
			return "", slack.SearchParameters{}, fmt.Errorf("query field cannot contain modifiers (from:, in:, type:, etc.). Please use the dedicated fields")
		}
		queryParts = append(queryParts, request.Query)
	}

	for _, t := range request.Types {
		if t != "" {
			queryParts = append(queryParts, fmt.Sprintf("type:%s", t))
		}
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

	for _, with := range request.WithUsers {
		if with != "" {
			if !strings.HasPrefix(with, "U") {
				return "", slack.SearchParameters{}, fmt.Errorf("invalid user ID format in with_users parameter: '%s'. Must start with 'U' (e.g., 'U1234567')", with)
			}
			queryParts = append(queryParts, fmt.Sprintf("with:<@%s>", with))
		}
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

	if request.Count < 1 || request.Count > 100 {
		return "", slack.SearchParameters{}, fmt.Errorf("count must be between 1 and 100, got %d", request.Count)
	}
	if request.Page < 1 || request.Page > 100 {
		return "", slack.SearchParameters{}, fmt.Errorf("page must be between 1 and 100, got %d", request.Page)
	}

	searchQuery := strings.Join(queryParts, " ")

	params := slack.SearchParameters{
		Sort:          "timestamp",
		SortDirection: "desc",
		Count:         request.Count,
		Page:          request.Page,
	}

	return searchQuery, params, nil
}

func (h *Handler) convertToFilesResponse(result *slack.SearchFiles) *SearchFilesResponse {
	response := &SearchFilesResponse{
		Files: make([]FileInfo, 0, len(result.Matches)),
	}

	for _, match := range result.Matches {
		file := FileInfo{
			ID:        match.ID,
			Title:     match.Title,
			Filetype:  match.Filetype,
			User:      match.User,
			Channels:  match.Channels,
			Created:   int64(match.Created),
			Updated:   int64(match.Timestamp),
			Permalink: match.Permalink,
		}
		response.Files = append(response.Files, file)
	}

	response.Pagination = &SearchPagination{
		TotalCount: result.Paging.Total,
		Page:       result.Paging.Page,
		PageCount:  result.Paging.Pages,
		PerPage:    result.Paging.Count,
		First:      1,
		Last:       result.Paging.Pages,
	}

	return response
}
