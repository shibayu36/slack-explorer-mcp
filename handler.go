package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/slack-go/slack"
)

// Error messages for user-facing errors
const (
	ErrSlackTokenNotConfigured = "Slack user token is not configured. Please set your Slack user token."
)

// SearchMessagesResponse represents the output for search_messages tool
type SearchMessagesResponse struct {
	WorkspaceURL string                 `json:"workspace_url"`
	Messages     *SearchMessagesMatches `json:"messages"`
}

type SearchMessagesMatches struct {
	Matches    []SearchMessage   `json:"matches"`
	Pagination *SearchPagination `json:"pagination,omitempty"`
}

type SearchMessage struct {
	User      string       `json:"user"`
	Text      string       `json:"text"`
	Timestamp string       `json:"ts"`
	Channel   *ChannelInfo `json:"channel,omitempty"`
	// Fill if the message is in a thread
	ThreadTs string `json:"thread_ts,omitempty"`
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

// SearchMessages handles the search_messages tool call
func (h *Handler) SearchMessages(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := h.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(ErrSlackTokenNotConfigured), nil
	}

	query, params, err := h.buildSearchParams(buildSearchParamsRequest{
		Query:     request.GetString("query", ""),
		InChannel: request.GetString("in_channel", ""),
		FromUser:  request.GetString("from_user", ""),
		With:      request.GetStringSlice("with", []string{}),
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

	searchResult, err := client.SearchMessages(query, params)
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
	With      []string
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

	for _, with := range request.With {
		if with != "" {
			if !strings.HasPrefix(with, "U") {
				return "", slack.SearchParameters{}, fmt.Errorf("invalid user ID format in with parameter: '%s'. Must start with 'U' (e.g., 'U1234567')", with)
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

// extractThreadTsFromPermalink extracts thread_ts from Slack permalink URL
func (h *Handler) extractThreadTsFromPermalink(permalink string) string {
	// Extract thread_ts from URL pattern like:
	// https://workspace.slack.com/archives/C123/p1234567890123456?thread_ts=1234567890.123456
	re := regexp.MustCompile(`[?&]thread_ts=([0-9.]+)`)
	matches := re.FindStringSubmatch(permalink)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractWorkspaceURLFromPermalink extracts workspace URL from Slack permalink
func (h *Handler) extractWorkspaceURLFromPermalink(permalink string) string {
	// Extract workspace URL from permalink pattern like:
	// https://workspace.slack.com/archives/C123/p1234567890123456
	// Returns: https://workspace.slack.com
	if permalink == "" {
		return ""
	}

	re := regexp.MustCompile(`^(https?://[^/]+\.slack\.com)`)
	matches := re.FindStringSubmatch(permalink)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// convertToSearchResponse converts Slack API response to our response format
func (h *Handler) convertToSearchResponse(result *slack.SearchMessages) *SearchMessagesResponse {
	response := &SearchMessagesResponse{
		WorkspaceURL: "",
		Messages: &SearchMessagesMatches{
			Matches: make([]SearchMessage, 0, len(result.Matches)),
		},
	}

	if len(result.Matches) > 0 {
		response.WorkspaceURL = h.extractWorkspaceURLFromPermalink(result.Matches[0].Permalink)
	}

	for _, match := range result.Matches {
		msg := SearchMessage{
			User:      match.User,
			Text:      match.Text,
			Timestamp: match.Timestamp,
			ThreadTs:  h.extractThreadTsFromPermalink(match.Permalink),
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

// GetThreadRepliesResponse represents the response structure for get_thread_replies
type GetThreadRepliesResponse struct {
	Messages   []ThreadMessage `json:"messages"`
	HasMore    bool            `json:"has_more"`
	NextCursor string          `json:"next_cursor,omitempty"`
}

type ThreadMessage struct {
	User       string     `json:"user"`
	Text       string     `json:"text"`
	Timestamp  string     `json:"ts"`
	ReplyCount int        `json:"reply_count,omitempty"`
	ReplyUsers []string   `json:"reply_users,omitempty"`
	Reactions  []Reaction `json:"reactions,omitempty"`
}

type Reaction struct {
	Name  string   `json:"name"`
	Count int      `json:"count"`
	Users []string `json:"users"`
}

// GetThreadReplies handles the get_thread_replies tool call
func (h *Handler) GetThreadReplies(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := h.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(ErrSlackTokenNotConfigured), nil
	}

	params, err := h.buildThreadRepliesParams(buildThreadRepliesRequest{
		ChannelID: request.GetString("channel_id", ""),
		ThreadTS:  request.GetString("thread_ts", ""),
		Limit:     request.GetInt("limit", 100),
		Cursor:    request.GetString("cursor", ""),
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	messages, hasMore, nextCursor, err := client.GetConversationReplies(params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	response := h.convertToThreadResponse(messages, hasMore, nextCursor)

	jsonData, err := json.Marshal(response)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type buildThreadRepliesRequest struct {
	ChannelID string
	ThreadTS  string
	Limit     int
	Cursor    string
}

func (h *Handler) buildThreadRepliesParams(request buildThreadRepliesRequest) (*slack.GetConversationRepliesParameters, error) {
	if request.ChannelID == "" {
		return nil, fmt.Errorf("channel_id is required")
	}
	if !strings.HasPrefix(request.ChannelID, "C") {
		return nil, fmt.Errorf("invalid channel ID format. Must start with 'C' (e.g., 'C1234567')")
	}

	if request.ThreadTS == "" {
		return nil, fmt.Errorf("thread_ts is required")
	}
	tsPattern := regexp.MustCompile(`^\d{10}\.\d{6}$`)
	if !tsPattern.MatchString(request.ThreadTS) {
		return nil, fmt.Errorf("thread_ts must be in format '1234567890.123456'")
	}

	if request.Limit < 1 || request.Limit > 1000 {
		return nil, fmt.Errorf("limit must be between 1 and 1000, got %d", request.Limit)
	}

	params := &slack.GetConversationRepliesParameters{
		ChannelID: request.ChannelID,
		Timestamp: request.ThreadTS,
		Limit:     request.Limit,
	}
	if request.Cursor != "" {
		params.Cursor = request.Cursor
	}

	return params, nil
}

func (h *Handler) convertToThreadResponse(messages []slack.Message, hasMore bool, nextCursor string) *GetThreadRepliesResponse {
	response := &GetThreadRepliesResponse{
		Messages: make([]ThreadMessage, 0, len(messages)),
		HasMore:  hasMore,
	}

	if nextCursor != "" {
		response.NextCursor = nextCursor
	}

	for _, msg := range messages {
		threadMsg := ThreadMessage{
			User:      msg.User,
			Text:      msg.Text,
			Timestamp: msg.Timestamp,
		}

		if msg.ReplyCount > 0 {
			threadMsg.ReplyCount = msg.ReplyCount
		}

		if len(msg.ReplyUsers) > 0 {
			threadMsg.ReplyUsers = msg.ReplyUsers
		}

		if len(msg.Reactions) > 0 {
			reactions := make([]Reaction, 0, len(msg.Reactions))
			for _, reaction := range msg.Reactions {
				reactions = append(reactions, Reaction{
					Name:  reaction.Name,
					Count: reaction.Count,
					Users: reaction.Users,
				})
			}
			threadMsg.Reactions = reactions
		}

		response.Messages = append(response.Messages, threadMsg)
	}

	return response
}

// UserProfile represents a user profile result
type UserProfile struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name,omitempty"`
	RealName    string `json:"real_name,omitempty"`
	Email       string `json:"email,omitempty"`
	Error       string `json:"error,omitempty"`
}

func (h *Handler) GetUserProfiles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := h.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(ErrSlackTokenNotConfigured), nil
	}

	userIDs := request.GetStringSlice("user_ids", []string{})

	if len(userIDs) == 0 {
		return mcp.NewToolResultError("user_ids is required and cannot be empty"), nil
	}
	if len(userIDs) > 100 {
		return mcp.NewToolResultError("user_ids cannot exceed 100 entries"), nil
	}

	var profiles []UserProfile

	for _, userID := range userIDs {
		profile := h.getUserProfile(client, userID)
		profiles = append(profiles, profile)
	}

	jsonData, err := json.Marshal(profiles)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func (h *Handler) getUserProfile(client SlackClient, userID string) UserProfile {
	if !strings.HasPrefix(userID, "U") {
		return UserProfile{
			UserID: userID,
			Error:  "invalid user ID format. Must start with 'U' (e.g., 'U1234567')",
		}
	}

	slackProfile, err := client.GetUserProfile(userID)
	if err != nil {
		return UserProfile{
			UserID: userID,
			Error:  err.Error(),
		}
	}

	return UserProfile{
		UserID:      userID,
		DisplayName: slackProfile.DisplayName,
		RealName:    slackProfile.RealName,
		Email:       slackProfile.Email,
	}
}

// SearchUsersByName searches for users by display name
func (h *Handler) SearchUsersByName(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := h.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(ErrSlackTokenNotConfigured), nil
	}

	displayName := request.GetString("display_name", "")
	if displayName == "" {
		return mcp.NewToolResultError("display_name is required"), nil
	}
	exact := request.GetBool("exact", true)

	users, err := h.userRepository.FindByDisplayName(ctx, client, displayName, exact)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Convert slack.User to UserProfile
	var profiles []UserProfile
	for _, user := range users {
		profiles = append(profiles, UserProfile{
			UserID:      user.ID,
			DisplayName: user.Profile.DisplayName,
			RealName:    user.Profile.RealName,
			Email:       user.Profile.Email,
		})
	}

	jsonData, err := json.Marshal(profiles)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// SearchFilesResponse represents the output for search_files tool
type SearchFilesResponse struct {
	OK         bool              `json:"ok"`
	Query      string            `json:"query"`
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
	WithUser  []string
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
		WithUser:  request.GetStringSlice("with_user", []string{}),
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

	response := h.convertToFilesResponse(query, searchResult)

	jsonData, err := json.Marshal(response)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func (h *Handler) buildSearchFilesParams(request buildSearchFilesParamsRequest) (string, slack.SearchParameters, error) {
	var queryParts []string

	if request.Query != "" {
		modifierPattern := regexp.MustCompile(`\b(from|in|before|after|on|type|ext):`)
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

	for _, with := range request.WithUser {
		if with != "" {
			if !strings.HasPrefix(with, "U") {
				return "", slack.SearchParameters{}, fmt.Errorf("invalid user ID format in with_user parameter: '%s'. Must start with 'U' (e.g., 'U1234567')", with)
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

func (h *Handler) convertToFilesResponse(query string, result *slack.SearchFiles) *SearchFilesResponse {
	response := &SearchFilesResponse{
		OK:    true,
		Query: query,
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

// GetCanvasContentResponse represents the output for get_canvas_content tool
type GetCanvasContentResponse struct {
	Canvases []CanvasContent `json:"canvases"`
}

type CanvasContent struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content,omitempty"`
	User      string `json:"user"`
	Created   int64  `json:"created"`
	Updated   int64  `json:"updated"`
	Permalink string `json:"permalink"`
	Error     string `json:"error,omitempty"`
}

// GetCanvasContent handles the get_canvas_content tool call
func (h *Handler) GetCanvasContent(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := h.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(ErrSlackTokenNotConfigured), nil
	}

	canvasIDs := request.GetStringSlice("canvas_ids", []string{})
	if len(canvasIDs) == 0 {
		return mcp.NewToolResultError("canvas_ids is required and cannot be empty"), nil
	}
	if len(canvasIDs) > 10 {
		return mcp.NewToolResultError("canvas_ids cannot exceed 10 entries"), nil
	}

	response := &GetCanvasContentResponse{
		Canvases: make([]CanvasContent, 0, len(canvasIDs)),
	}

	for _, canvasID := range canvasIDs {
		canvas := h.getCanvasContent(client, canvasID)
		response.Canvases = append(response.Canvases, canvas)
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func (h *Handler) getCanvasContent(client SlackClient, canvasID string) CanvasContent {
	if !strings.HasPrefix(canvasID, "F") {
		return CanvasContent{
			ID:    canvasID,
			Error: "invalid canvas ID format. Must start with 'F' (e.g., 'F12345678')",
		}
	}

	file, err := client.GetFileInfo(canvasID)
	if err != nil {
		return CanvasContent{
			ID:    canvasID,
			Error: err.Error(),
		}
	}

	if file.Filetype != "canvas" && file.Filetype != "quip" {
		return CanvasContent{
			ID:    canvasID,
			Error: fmt.Sprintf("file is not a canvas (filetype: %s)", file.Filetype),
		}
	}

	var buf bytes.Buffer
	if err := client.GetFile(file.URLPrivateDownload, &buf); err != nil {
		return CanvasContent{
			ID:    canvasID,
			Error: fmt.Sprintf("failed to download canvas content: %v", err),
		}
	}

	return CanvasContent{
		ID:        file.ID,
		Title:     file.Title,
		Content:   buf.String(),
		User:      file.User,
		Created:   int64(file.Created),
		Updated:   int64(file.Timestamp),
		Permalink: file.Permalink,
	}
}
