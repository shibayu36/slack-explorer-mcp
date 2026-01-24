# Canvas HTML Strip - 要件定義書

## Context / Goal

### 背景
現在の`get_canvas_content`ツールはCanvasのHTMLをそのまま返却している。しかし、SlackのCanvasから取得されるHTMLには多くの冗長な情報（一時的なID、空のstyle属性、装飾用のclass、不要なbr要素など）が含まれており、LLMで処理する際のトークン効率が悪い。

例えば、シンプルなリンク集のCanvasでも、以下のような冗長なHTML構造になっている：
- すべての要素に`id='temp:C:XXX...'`のような一時ID
- 空の`style=""`属性
- 装飾用の`<span>`要素
- 意味のない`<br/>`要素
- 絵文字表示用の`<control>`と`<img>`タグ

### ゴール
`get_canvas_content`が返すHTMLから不要な情報を除去し、コンテンツの意味を保ったままトークン効率を向上させる。

## In scope

1. **不要な属性の除去**
   - `id`属性をすべて除去
   - `style`属性をすべて除去
   - `class`属性は基本的に除去（ただし`checked`など意味を持つものは保持）

2. **不要な要素の除去・変換**
   - `<br/>`要素の除去
   - すべての`<span>`要素の除去（中身のテキスト・子要素は保持）
   - Slack絵文字の`<control>`と`<img>`タグを`:emoji_name:`形式のテキストに変換

3. **保持する要素**
   - 構造を表すタグ（`h1`-`h3`, `p`, `ul`, `ol`, `li`, `table`, `tr`, `td`, `blockquote`, `pre`など）
   - テキスト装飾タグ（`b`, `i`, `u`, `del`など）
   - リンク（`<a>`タグとhref属性）
   - 添付ファイル情報（`class='embedded-file'`を持つ要素）
   - `data-section-style`属性（リストの種類を表す）
   - `checked`クラス（TODOアイテムの状態を表す）

4. **デフォルト動作の変更**
   - `get_canvas_content`はデフォルトでストリップ後のHTMLを返す

## Out of scope

- Markdown形式への変換
- Canvas以外のファイルタイプへの適用
- ストリップ前の生HTMLを取得するオプション（将来必要になれば追加）
- 画像ファイルの内容取得

## User Stories

1. **US1**: ユーザーとして、Canvasの内容を取得したときに、不要なID・style属性が除去されていてほしい。そうすればLLMでの処理時にトークンを節約できる。

2. **US2**: ユーザーとして、Canvasのリンク情報（`<a>`タグとURL）は保持されていてほしい。そうすれば参照先を辿れる。

3. **US3**: ユーザーとして、CanvasのTODOリストを取得したときに、チェック済みかどうかの状態（`checked`クラス）が分かるようにしてほしい。

4. **US4**: ユーザーとして、Canvas内のSlack絵文字が`:emoji_name:`形式で表示されてほしい。そうすれば何の絵文字か分かる。

5. **US5**: ユーザーとして、Canvas内の添付ファイル情報が保持されていてほしい。そうすればファイル参照として活用できる。

6. **US6**: ユーザーとして、Canvasのテーブル構造がそのまま保持されていてほしい。そうすれば表形式のデータを理解できる。

## Acceptance Criteria (Gherkin)

### 主要ケース

```gherkin
Feature: Canvas HTML Strip機能

  Scenario: 不要な属性が除去される
    Given Canvasに以下のHTMLが含まれている
      """
      <h1 id='temp:C:ABC123' style=''>タイトル</h1>
      <p id='temp:C:DEF456' class='line' style=''>本文</p>
      """
    When get_canvas_contentでCanvasを取得する
    Then 以下のようにストリップされたHTMLが返される
      """
      <h1>タイトル</h1>
      <p>本文</p>
      """

  Scenario: Slack絵文字がテキスト形式に変換される
    Given Canvasに以下のHTMLが含まれている
      """
      <control data-remapped="true" id="temp:C:XXX">
        <img src="https://..." alt="miro" data-is-slack style="...">:miro:</img>
      </control>
      """
    When get_canvas_contentでCanvasを取得する
    Then ":miro:"というテキストに変換されている

  Scenario: br要素とspan要素が除去される
    Given Canvasに以下のHTMLが含まれている
      """
      <ul id='temp:C:AAA'>
        <li id='temp:C:BBB' class='' style=''><span id='temp:C:BBB'>リスト項目1</span>
        <br/></li>
        <li id='temp:C:CCC' class='' style=''><span id='temp:C:CCC'>リスト項目2</span>
        <br/></li>
      </ul>
      """
    When get_canvas_contentでCanvasを取得する
    Then br要素とspan要素が除去され、テキストのみ保持される
      """
      <ul>
        <li>リスト項目1</li>
        <li>リスト項目2</li>
      </ul>
      """

  Scenario: チェックリストのchecked状態とdata-section-styleが保持される
    Given Canvasに以下のチェックリストHTMLが含まれている
      """
      <div data-section-style='7' class="" style="">
        <ul id='temp:C:AAA'>
          <li id='temp:C:XXX' class='checked' style=''>完了タスク</li>
          <li id='temp:C:YYY' class='' style=''>未完了タスク</li>
        </ul>
      </div>
      """
    When get_canvas_contentでCanvasを取得する
    Then 以下のようにdata-section-styleとcheckedクラスが保持される
      """
      <div data-section-style="7">
        <ul>
          <li class="checked">完了タスク</li>
          <li>未完了タスク</li>
        </ul>
      </div>
      """
```

### 例外ケース

```gherkin
  Scenario: 添付ファイル情報が保持される
    Given Canvasに添付ファイル情報が含まれている
      """
      <p class='embedded-file'>File ID: F0AAK0CBLKW File URL: https://xxx.slack.com/files/...</p>
      """
    When get_canvas_contentでCanvasを取得する
    Then 添付ファイル情報がそのまま保持されている

  Scenario: リンクのhref属性が保持される
    Given Canvasにリンクが含まれている
      """
      <a href="https://example.com/" id="temp:C:XXX">リンクテキスト</a>
      """
    When get_canvas_contentでCanvasを取得する
    Then href属性は保持され、id属性は除去される
      """
      <a href="https://example.com/">リンクテキスト</a>
      """
```

## NFR（非機能要件）

### 性能
- HTML stripの処理時間: 通常のCanvas（〜100KB）で100ms以内
- 既存の`get_canvas_content`のレスポンス時間に大きな影響を与えない

### コスト
- トークン削減効果: 元のHTMLに対して50%以上の削減を目標

### セキュリティ
- 既存のセキュリティ要件を維持（トークン管理など）

### 監査
- 既存のログ出力パターンを踏襲

### 運用
- 既存のエラーハンドリングパターンを踏襲

## Open Questions

現時点でOpen Questionsはなし。

### 解決済みの項目

1. **`data-section-style`の意味**: 5=箇条書きリスト、6=番号付きリスト、7=チェックリスト。この属性は保持する。

2. **`parent`クラスの扱い**: ネストした子リストを持つ`<li>`に付与されるマーカー。HTML構造から判断できる情報のため、除去してよい。

## 期待される変換例

### Before（現在の出力）
```html
<h1 id='temp:C:ABC123'>参考リンク集</h1>
<div data-section-style='5' class="" style=""><ul id='temp:C:DEF456'><li id='temp:C:GHI789' class='' style='' value='1'><span id='temp:C:GHI789'><control data-remapped="true" id="temp:C:JKL012"><img src="https://slack-imgs.com/?..." alt="link" data-is-slack style="width: 18px; height: 18px">:link:</img></control> <a href="https://example.com/docs">ドキュメント</a></span>
<br/></li></ul></div>
```

### After（期待する出力）
```html
<h1>参考リンク集</h1>
<div data-section-style="5"><ul><li>:link: <a href="https://example.com/docs">ドキュメント</a></li></ul></div>
```

## 実装計画

### 設計方針

#### 構成

```go
// canvas_html_stripper.go

type CanvasHTMLStripper struct {
    // 将来的に設定を持たせる可能性あり
}

func NewCanvasHTMLStripper() *CanvasHTMLStripper {
    return &CanvasHTMLStripper{}
}

func (s *CanvasHTMLStripper) Strip(html string) (string, error) {
    // HTML strip処理
}
```

#### 呼び出し側

```go
// handler_get_canvas_content.go

func (h *Handler) getCanvasContent(...) CanvasContent {
    // ...既存の処理...

    // Content取得後にstrip
    stripper := NewCanvasHTMLStripper()
    strippedContent, err := stripper.Strip(content)
    if err != nil {
        // エラー処理
    }

    return CanvasContent{
        Content: strippedContent,
        // ...
    }
}
```

#### 使用ライブラリ

- `golang.org/x/net/html`: Go標準のHTMLパーサー

### フェーズ1: Canvas HTML Strip機能の実装（1 PR）

#### Commit 1: Add CanvasHTMLStripper skeleton and integrate with handler
- `canvas_html_stripper.go` 新規作成
  - `CanvasHTMLStripper` struct定義
  - `NewCanvasHTMLStripper()` コンストラクタ
  - `Strip()` メソッド（空実装：入力をそのまま返す）
- `handler_get_canvas_content.go` への統合
  - `getCanvasContent()` 内でStripperを呼び出す
- この時点でエンドツーエンドで動作確認可能

#### Commit 2: Add attribute stripping to CanvasHTMLStripper
- `golang.org/x/net/html` 依存追加
- id, style属性の除去
- class属性の条件付き除去（`checked`, `embedded-file`は保持）
- `href`, `data-section-style`の保持
- `canvas_html_stripper_test.go` でユニットテスト

#### Commit 3: Implement element transformation for CanvasHTMLStripper
- br要素の除去
- span要素の全除去（中身は保持）
- Slack絵文字（`<control><img>`）の`:emoji:`変換
- 対応するユニットテスト追加
- 複合的なHTMLを入力として、全機能が正しく動作することを確認する統合テスト（`TestCanvasHTMLStripper_Strip_ComplexCanvasHTML`）
  - ネストしたリスト、テーブル内装飾、blockquote、pre、embedded-fileなどを含む

##### 実装上の注意点
- **ノード削除時のループ**: `for c := n.FirstChild; c != nil; c = c.NextSibling` でループ中にノードを削除すると`NextSibling`が壊れる。削除前に次のノードを保存しておく必要がある
  ```go
  for c := n.FirstChild; c != nil; {
      next := c.NextSibling  // 先に次のノードを保存
      // ノード削除処理
      c = next
  }
  ```
- **spanの中身移動**: spanの子要素を親に移動してからspan自体を削除する。順序としては (1) spanの子を全てspanの前に挿入 (2) spanを削除

