package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanvasHTMLStripper_Strip(t *testing.T) {
	stripper := NewCanvasHTMLStripper()

	t.Run("removes id and style attributes", func(t *testing.T) {
		input := `<h1 id='temp:C:ABC123' style=''>タイトル</h1><p id='temp:C:DEF456' class='line' style=''>本文</p>`
		result, err := stripper.Strip(input)
		require.NoError(t, err)
		assert.Contains(t, result, "<h1>タイトル</h1>")
		assert.Contains(t, result, "<p>本文</p>")
		assert.NotContains(t, result, "id=")
		assert.NotContains(t, result, "style=")
	})

	t.Run("preserves href attribute on links", func(t *testing.T) {
		input := `<a href="https://example.com/" id="temp:C:XXX">リンクテキスト</a>`
		result, err := stripper.Strip(input)
		require.NoError(t, err)
		assert.Contains(t, result, `href="https://example.com/"`)
		assert.Contains(t, result, "リンクテキスト")
		assert.NotContains(t, result, "id=")
	})

	t.Run("preserves data-section-style attribute", func(t *testing.T) {
		input := `<div data-section-style='7' class="" style=""><ul id='temp:C:AAA'><li>item</li></ul></div>`
		result, err := stripper.Strip(input)
		require.NoError(t, err)
		assert.Contains(t, result, `data-section-style="7"`)
	})

	t.Run("preserves checked class on li", func(t *testing.T) {
		input := `<li id='temp:C:XXX' class='checked' style=''>完了タスク</li>`
		result, err := stripper.Strip(input)
		require.NoError(t, err)
		assert.Contains(t, result, `class="checked"`)
		assert.Contains(t, result, "完了タスク")
	})

	t.Run("removes br elements", func(t *testing.T) {
		input := `<li>テキスト<br/></li>`
		result, err := stripper.Strip(input)
		require.NoError(t, err)
		assert.NotContains(t, result, "<br")
		assert.Contains(t, result, "テキスト")
	})

	t.Run("unwraps all span elements", func(t *testing.T) {
		input := `<ul><li id='temp:C:BBB' class='' style=''><span id='temp:C:BBB'>リスト項目</span></li></ul>`
		result, err := stripper.Strip(input)
		require.NoError(t, err)
		assert.Contains(t, result, "<li>リスト項目</li>")
		assert.NotContains(t, result, "<span")
	})

	t.Run("unwraps nested span with links", func(t *testing.T) {
		input := `<li><span>:miro: <a href="https://example.com">リンク</a></span></li>`
		result, err := stripper.Strip(input)
		require.NoError(t, err)
		assert.Contains(t, result, `:miro: <a href="https://example.com">リンク</a>`)
		assert.NotContains(t, result, "<span")
	})

	t.Run("converts Slack emoji to text format", func(t *testing.T) {
		input := `<control data-remapped="true" id="temp:C:XXX"><img src="https://..." alt="miro" data-is-slack style="...">:miro:</img></control>`
		result, err := stripper.Strip(input)
		require.NoError(t, err)
		assert.Contains(t, result, ":miro:")
		assert.NotContains(t, result, "<control")
		assert.NotContains(t, result, "<img")
	})

	t.Run("preserves embedded-file class", func(t *testing.T) {
		input := `<p class='embedded-file'>File ID: F0AAK0CBLKW File URL: https://xxx.slack.com/files/...</p>`
		result, err := stripper.Strip(input)
		require.NoError(t, err)
		assert.Contains(t, result, `class="embedded-file"`)
		assert.Contains(t, result, "File ID: F0AAK0CBLKW")
	})

	t.Run("preserves text formatting tags", func(t *testing.T) {
		input := `<p><b>太字</b><i>斜体</i><u>下線</u><del>取り消し</del></p>`
		result, err := stripper.Strip(input)
		require.NoError(t, err)
		assert.Contains(t, result, "<b>太字</b>")
		assert.Contains(t, result, "<i>斜体</i>")
		assert.Contains(t, result, "<u>下線</u>")
		assert.Contains(t, result, "<del>取り消し</del>")
	})

	t.Run("preserves table structure", func(t *testing.T) {
		input := `<table><tr><td><p>セル1</p></td><td><p>セル2</p></td></tr></table>`
		result, err := stripper.Strip(input)
		require.NoError(t, err)
		assert.Contains(t, result, "<table>")
		assert.Contains(t, result, "<tr>")
		assert.Contains(t, result, "<td>")
		assert.Contains(t, result, "セル1")
	})

	t.Run("complex example from requirements", func(t *testing.T) {
		input := `<h1 id='temp:C:ABC123'>参考リンク集</h1>
<div data-section-style='5' class="" style=""><ul id='temp:C:DEF456'><li id='temp:C:GHI789' class='' style='' value='1'><span id='temp:C:GHI789'><control data-remapped="true" id="temp:C:JKL012"><img src="https://slack-imgs.com/?..." alt="link" data-is-slack style="width: 18px; height: 18px">:link:</img></control> <a href="https://example.com/docs">ドキュメント</a></span>
<br/></li></ul></div>`

		result, err := stripper.Strip(input)
		require.NoError(t, err)

		// Check expected output
		assert.Contains(t, result, "<h1>参考リンク集</h1>")
		assert.Contains(t, result, `data-section-style="5"`)
		assert.Contains(t, result, ":link:")
		assert.Contains(t, result, `<a href="https://example.com/docs">ドキュメント</a>`)

		// Check removed elements
		assert.NotContains(t, result, "id=")
		assert.NotContains(t, result, ` style="`)
		assert.NotContains(t, result, "<br")
		assert.NotContains(t, result, "<control")
	})
}
