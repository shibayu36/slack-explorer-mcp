package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanvasHTMLStripper_Strip(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "removes unnecessary attributes",
			input: `
<h1 id='temp:C:ABC123'>Title</h1>

<p id='temp:C:DEF456' class='line'>Body text here.</p>`,
			expected: `<h1>Title</h1>

<p>Body text here.</p>`,
		},
		{
			name: "preserves checklist with checked class and data-section-style",
			input: `
<div data-section-style='7' class="" style="">
  <ul id='temp:C:AAA'>
    <li id='temp:C:XXX' class='checked' style='' value='1'>
      <span id='temp:C:XXX'>Completed task</span>
      <br/>
    </li>
    <li id='temp:C:YYY' class='' style=''>
      <span id='temp:C:YYY'>Incomplete task</span>
      <br/>
    </li>
  </ul>
</div>`,
			expected: `<div data-section-style="7">
  <ul>
    <li class="checked">
      <span>Completed task</span>
      <br/>
    </li>
    <li>
      <span>Incomplete task</span>
      <br/>
    </li>
  </ul>
</div>`,
		},
		{
			name: "preserves embedded-file class",
			input: `
<p class='embedded-file'>File ID: F0AAK0CBLKW File URL: https://example.slack.com/files/U123/F0AAK0CBLKW/image.jpg</p>`,
			expected: `<p class="embedded-file">File ID: F0AAK0CBLKW File URL: https://example.slack.com/files/U123/F0AAK0CBLKW/image.jpg</p>`,
		},
		{
			name: "preserves href attribute on anchor tags",
			input: `
<p id='temp:C:ABC' class='line'>
  <a href="https://example.com/">Link text</a>
</p>`,
			expected: `<p>
  <a href="https://example.com/">Link text</a>
</p>`,
		},
		{
			name:     "returns empty string for empty input",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stripper := NewCanvasHTMLStripper()
			result, err := stripper.Strip(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
