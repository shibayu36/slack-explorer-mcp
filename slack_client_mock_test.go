package main

import (
	"context"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/mock"
)

type SlackClientMock struct{ mock.Mock }

func (m *SlackClientMock) SearchMessages(query string, params slack.SearchParameters) (*slack.SearchMessages, error) {
	args := m.Called(query, params)
	var res *slack.SearchMessages
	if v := args.Get(0); v != nil {
		res = v.(*slack.SearchMessages)
	}
	return res, args.Error(1)
}

func (m *SlackClientMock) SearchFiles(query string, params slack.SearchParameters) (*slack.SearchFiles, error) {
	args := m.Called(query, params)
	var res *slack.SearchFiles
	if v := args.Get(0); v != nil {
		res = v.(*slack.SearchFiles)
	}
	return res, args.Error(1)
}

func (m *SlackClientMock) GetConversationReplies(params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error) {
	args := m.Called(params)
	msgs, _ := args.Get(0).([]slack.Message)
	return msgs, args.Bool(1), args.String(2), args.Error(3)
}

func (m *SlackClientMock) GetUserProfile(userID string) (*slack.UserProfile, error) {
	args := m.Called(userID)
	var res *slack.UserProfile
	if v := args.Get(0); v != nil {
		res = v.(*slack.UserProfile)
	}
	return res, args.Error(1)
}

func (m *SlackClientMock) GetUsers(ctx context.Context, options ...slack.GetUsersOption) ([]slack.User, error) {
	args := m.Called(ctx, options)
	var res []slack.User
	if v := args.Get(0); v != nil {
		res = v.([]slack.User)
	}
	return res, args.Error(1)
}
