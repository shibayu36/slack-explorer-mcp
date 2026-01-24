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
      <span id='temp:C:XXX'>Completed task</span><br/>
    </li>
    <li id='temp:C:YYY' class='' style=''>
      <span id='temp:C:YYY'>Incomplete task</span><br/>
    </li>
  </ul>
</div>`,
			expected: `<div data-section-style="7">
  <ul>
    <li class="checked">
      Completed task
    </li>
    <li>
      Incomplete task
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
		{
			name: "removes br and span elements but preserves text",
			input: `<ul id='temp:C:AAA'>
<li id='temp:C:BBB' class='parent' style=''><span id='temp:C:BBB'>List item 1</span><br/></li>
<li id='temp:C:CCC' class='' style=''><span id='temp:C:CCC'>List item 2</span><br/></li>
</ul>`,
			expected: `<ul>
<li>List item 1</li>
<li>List item 2</li>
</ul>`,
		},
		{
			name: "converts Slack emoji to text",
			input: `<p id='temp:C:XXX'>
<control data-remapped="true" id="temp:C:YYY">
<img src="https://example.com" alt="miro" data-is-slack style="width: 18px">:miro:</img>
</control> text</p>`,
			expected: `<p>
:miro: text</p>`,
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

func TestCanvasHTMLStripper_Strip_ComplexCanvasHTML(t *testing.T) {
	input := `<h1 id='temp:C:AAA'>Title</h1>

<p id='temp:C:BBB' class='line'>Body text</p>

<div data-section-style='5' class="" style="">
  <ul id='temp:C:CCC'>
    <li id='temp:C:DDD' class='parent' style='' value='1'>
      <span id='temp:C:DDD'><b>Item 1</b></span><br/>
    </li>
    <ul>
      <li id='temp:C:EEE' class='parent' style=''>
        <span id='temp:C:EEE'><i>Item 1-1</i></span><br/>
      </li>
      <ul>
        <li id='temp:C:FFF' class='' style=''>
          <span id='temp:C:FFF'>Item 1-1-1</span><br/>
        </li>
      </ul>
      <li id='temp:C:GGG' class='' style=''>
        <span id='temp:C:GGG'>Item 1-2</span><br/>
      </li>
    </ul>
    <li id='temp:C:HHH' class='parent' style=''>
      <span id='temp:C:HHH'>Item 2</span><br/>
    </li>
    <ul>
      <li id='temp:C:III' class='' style=''>
        <span id='temp:C:III'><u>Item 2-1</u></span><br/>
      </li>
    </ul>
  </ul>
</div>

<div data-section-style='7' class="" style="">
  <ul id='temp:C:JJJ'>
    <li id='temp:C:KKK' class='checked' style='' value='1'>
      <span id='temp:C:KKK'>Completed task</span><br/>
    </li>
    <li id='temp:C:LLL' class='' style=''>
      <span id='temp:C:LLL'><del>Incomplete task</del></span><br/>
    </li>
  </ul>
</div>

<table>
  <tr>
    <td><p id='temp:C:MMM' class='line'><b>Bold Column</b></p></td>
    <td><p id='temp:C:NNN' class='line'><i>Italic Column</i></p></td>
  </tr>
  <tr>
    <td><p id='temp:C:OOO' class='line'>Data 1</p></td>
    <td><p id='temp:C:PPP' class='line'><u>Data 2</u></p></td>
  </tr>
</table>

<blockquote id='temp:C:QQQ'>Quote line 1<br>Quote line 2</blockquote>

<pre id='temp:C:RRR' class='prettyprint'>Code line 1<br>Code line 2</pre>

<div data-section-style='5' class="" style="">
  <ul id='temp:C:SSS'>
    <li id='temp:C:TTT' class='' style='' value='1'>
      <span id='temp:C:TTT'><control data-remapped="true" id="temp:C:UUU">
          <img src="https://example.com" alt="link" data-is-slack style="width: 18px">:link:</img>
        </control>
        <a href="https://example.com/docs">Documentation</a></span><br/>
    </li>
  </ul>
</div>

<p class='embedded-file'>File ID: F0AAK0CBLKW File URL: https://example.slack.com/files/U123/F0AAK0CBLKW/image.jpg</p>`

	expected := `<h1>Title</h1>

<p>Body text</p>

<div data-section-style="5">
  <ul>
    <li>
      <b>Item 1</b>
    </li>
    <ul>
      <li>
        <i>Item 1-1</i>
      </li>
      <ul>
        <li>
          Item 1-1-1
        </li>
      </ul>
      <li>
        Item 1-2
      </li>
    </ul>
    <li>
      Item 2
    </li>
    <ul>
      <li>
        <u>Item 2-1</u>
      </li>
    </ul>
  </ul>
</div>

<div data-section-style="7">
  <ul>
    <li class="checked">
      Completed task
    </li>
    <li>
      <del>Incomplete task</del>
    </li>
  </ul>
</div>

<table>
  <tbody><tr>
    <td><p><b>Bold Column</b></p></td>
    <td><p><i>Italic Column</i></p></td>
  </tr>
  <tr>
    <td><p>Data 1</p></td>
    <td><p><u>Data 2</u></p></td>
  </tr>
</tbody></table>

<blockquote>Quote line 1Quote line 2</blockquote>

<pre>Code line 1Code line 2</pre>

<div data-section-style="5">
  <ul>
    <li>
      :link:
        <a href="https://example.com/docs">Documentation</a>
    </li>
  </ul>
</div>

<p class="embedded-file">File ID: F0AAK0CBLKW File URL: https://example.slack.com/files/U123/F0AAK0CBLKW/image.jpg</p>`

	stripper := NewCanvasHTMLStripper()
	result, err := stripper.Strip(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
