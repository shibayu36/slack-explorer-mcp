package main

import (
	"testing"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestConvertAttachments(t *testing.T) {
	t.Run("returns nil for nil input", func(t *testing.T) {
		result := convertAttachments(nil)
		assert.Nil(t, result)
	})

	t.Run("maps title, text, and from_url correctly", func(t *testing.T) {
		attachments := []slack.Attachment{
			{
				Title:   "Example Article Title",
				Text:    "This is a summary of the article",
				FromURL: "https://example.com/article/123",
			},
		}

		result := convertAttachments(attachments)

		assert.Len(t, result, 1)
		assert.Equal(t, "Example Article Title", result[0].Title)
		assert.Equal(t, "This is a summary of the article", result[0].Text)
		assert.Equal(t, "https://example.com/article/123", result[0].FromURL)
	})

	t.Run("includes attachment even when all fields are empty", func(t *testing.T) {
		attachments := []slack.Attachment{
			{
				Fallback: "[no preview available]",
			},
		}

		result := convertAttachments(attachments)

		assert.Len(t, result, 1)
		assert.Equal(t, AttachmentInfo{}, result[0])
	})

	t.Run("extracts fields title and value", func(t *testing.T) {
		attachments := []slack.Attachment{
			{
				Fields: []slack.AttachmentField{
					{Title: "", Value: "This is a tweet body via bot", Short: false},
				},
				Color:    "DDDDDD",
				Fallback: "<https://example.com/status/123>",
			},
		}

		result := convertAttachments(attachments)

		assert.Len(t, result, 1)
		assert.Len(t, result[0].Fields, 1)
		assert.Equal(t, "", result[0].Fields[0].Title)
		assert.Equal(t, "This is a tweet body via bot", result[0].Fields[0].Value)
	})

	t.Run("skips fields where both title and value are empty", func(t *testing.T) {
		attachments := []slack.Attachment{
			{
				Fields: []slack.AttachmentField{
					{Title: "Info", Value: "Some value", Short: false},
					{Title: "", Value: "", Short: false},
				},
			},
		}

		result := convertAttachments(attachments)

		assert.Len(t, result, 1)
		assert.Len(t, result[0].Fields, 1)
		assert.Equal(t, "Info", result[0].Fields[0].Title)
		assert.Equal(t, "Some value", result[0].Fields[0].Value)
	})

	t.Run("handles multiple attachments preserving order and excluding extra fields", func(t *testing.T) {
		attachments := []slack.Attachment{
			{
				Title:       "First link",
				Text:        "First description",
				FromURL:     "https://example.com/1",
				Color:       "ff0000",
				AuthorName:  "Author1",
				ServiceName: "YouTube",
			},
			{
				Title:   "Second link",
				FromURL: "https://example.com/2",
			},
		}

		result := convertAttachments(attachments)

		assert.Len(t, result, 2)
		assert.Equal(t, AttachmentInfo{
			Title:   "First link",
			Text:    "First description",
			FromURL: "https://example.com/1",
		}, result[0])
		assert.Equal(t, AttachmentInfo{
			Title:   "Second link",
			FromURL: "https://example.com/2",
		}, result[1])
	})
}
