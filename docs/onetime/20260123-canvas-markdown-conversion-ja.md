# Canvas HTML to Markdown 変換 - 要件定義書

## Context / Goal

### 背景
現在の`get_canvas_content` toolはCanvasの内容をHTMLとして返却している。しかし、SlackのCanvas HTMLには多数の`id`、`class`、`style`属性やSlack独自の`data-section-style`属性が含まれており、これらはLLMが内容を理解する上で不要である。結果として、トークン効率が悪く、コストが増大している。

実際のCanvas HTMLの例：
```html
<h1 id='temp:C:BVF3da2ce54278c4415a93435dc1'>タイトル</h1>
<div data-section-style='7' class="" style=""><ul id='BVF9CAMrAKK'>
<li id='BVF9CASJ1YB' class='checked' style='' value='1'>
<span id='BVF9CASJ1YB'>チェック済み項目</span>
</li></ul></div>
```

### ゴール
Canvas HTMLをMarkdownに変換することで、トークン効率を大幅に改善する（約70-80%削減）。これによりLLMがCanvasのナレッジにより効率的にアクセスできるようにする。

## In scope

1. **HTML→Markdown変換機能の実装**
   - `get_canvas_content`の出力を常にMarkdown形式で返す
   - 変換処理は内部で自動的に行う（パラメータ不要）

2. **対応する要素**
   - 見出し: `<h1>`, `<h2>`, `<h3>` → `#`, `##`, `###`
   - 段落: `<p>` → 空行区切りのテキスト
   - 箇条書きリスト: `data-section-style='5'` → `- `
   - 番号付きリスト: `data-section-style='6'` → `1. `
   - チェックリスト: `data-section-style='7'` → `- [ ]` / `- [x]`
   - ネストしたリスト: インデントで表現
   - テーブル: `<table>` → Markdownテーブル
   - 引用: `<blockquote>` → `> `
   - コードブロック: `<pre>` → ` ``` `
   - 太字: `<b>` → `**text**`
   - 斜体: `<i>` → `*text*`
   - 下線: `<u>` → `<u>text</u>`（Markdown非対応のためHTMLタグ維持）
   - 取り消し線: `<del>` → `~~text~~`
   - リンク: `<a href="...">` → `[text](url)`
   - 添付ファイル: `<p class='embedded-file'>` → `![File](url)`

3. **チェックリストの状態保持**
   - `class='checked'`のある項目 → `- [x]`
   - それ以外 → `- [ ]`

## Out of scope

- **インタラクティブウィジェットの取得**: プロフィールカード、ワークフローボタン等のSlack UI側でレンダリングされる要素（API経由では取得不可）
- **コメントの取得**: Canvas内の特定箇所に付けられたコメント（HTMLに含まれない）
- **リアクションの取得**: Canvas内の特定箇所に付けられたリアクション（HTMLに含まれない）
- **カラムレイアウトの再現**: 複数カラムのレイアウト情報（Markdownで表現困難）
- **formatパラメータの追加**: 出力形式の切り替え機能（常にMarkdown固定）
- **Canvas以外のファイルタイプの変換**: PDF等の他ファイルタイプへの適用

## User Stories

1. **US1**: ユーザーとして、Canvasの内容をMarkdownで取得して、トークン効率良くLLMに処理させたい
2. **US2**: ユーザーとして、チェックリストの完了状態を把握して、タスクの進捗を確認したい
3. **US3**: ユーザーとして、見出し構造を維持したまま内容を取得して、ドキュメントの構造を理解したい
4. **US4**: ユーザーとして、テーブルデータを構造化された形で取得して、情報を整理したい
5. **US5**: ユーザーとして、コードブロックを適切にフォーマットされた状態で取得して、技術情報を正確に把握したい

## Acceptance Criteria (Gherkin)

### 主要ケース

```gherkin
Feature: Canvas HTML to Markdown変換

  Scenario: 基本的なHTML要素をMarkdownに変換する
    Given Canvasが存在する
    And Canvasに見出し、段落、リストが含まれている
    When get_canvas_contentを実行する
    Then 見出しは#記法に変換される
    And 段落は空行区切りのテキストになる
    And リストは-または1.記法に変換される

  Scenario: チェックリストの状態を保持して変換する
    Given Canvasにチェックリストが含まれている
    And 一部の項目がチェック済みである
    When get_canvas_contentを実行する
    Then チェック済み項目は"- [x]"で表現される
    And 未チェック項目は"- [ ]"で表現される

  Scenario: テーブルをMarkdownテーブルに変換する
    Given Canvasにテーブルが含まれている
    When get_canvas_contentを実行する
    Then テーブルはMarkdownテーブル形式で出力される
    And ヘッダー行とデータ行が適切に区切られる
```

### 例外ケース

```gherkin
  Scenario: 不正なHTMLの処理
    Given CanvasのHTMLが不正な形式である
    When get_canvas_contentを実行する
    Then エラーにならずベストエフォートで変換される
    And 変換できない部分はそのまま出力される

  Scenario: 空のCanvasの処理
    Given Canvasが存在するが内容が空である
    When get_canvas_contentを実行する
    Then 空文字列が返される
```

## NFR（非機能要件）

### 性能
- 変換処理: 100ms以内（通常サイズのCanvas）
- 大規模Canvas（100KB以上）でも1秒以内

### コスト
- トークン数を70%以上削減
- 変換処理自体のCPU/メモリ使用量は軽微

### セキュリティ
- 既存の`get_canvas_content`と同様（トークン管理はContext経由）
- 変換処理で外部通信は行わない

### 運用
- 変換エラー時はログ出力（既存パターン踏襲）
- 変換失敗時もベストエフォートで結果を返す

## 制限事項

以下の要素はSlack API経由では取得できないため、変換対象外：

| 要素 | 理由 |
|------|------|
| プロフィールカード等のウィジェット | UI側でレンダリングされる |
| コメント | HTMLに含まれない |
| リアクション | HTMLに含まれない |
| カラムレイアウト情報 | HTMLに含まれない |

## Open Questions

なし（全て解決済み）

## 変換サンプル

### Before（HTML）: 約6,800文字
```html
<h1 id='temp:C:BVF3da2ce54278c4415a93435dc1'>新メンバーのオンボーディング</h1>
<p id='BVF9CAEgcSN' class='line'>ようこそ！</p>
<div data-section-style='7' class="" style=""><ul id='BVF9CAMrAKK'>
<li id='BVF9CASJ1YB' class='checked' style='' value='1'>
<span id='BVF9CASJ1YB'>完了したタスク</span></li>
<li id='BVF9CASJ1YC' class='' style='' value='2'>
<span id='BVF9CASJ1YC'>未完了のタスク</span></li>
</ul></div>
```

### After（Markdown）: 約1,600文字（76%削減）
```markdown
# 新メンバーのオンボーディング

ようこそ！

- [x] 完了したタスク
- [ ] 未完了のタスク
```

## Technical Design

### 実装方針

Goの`golang.org/x/net/html`パッケージを使用してHTMLをパースし、再帰的にMarkdownに変換する。

```
handler_get_canvas_content.go
  └─ getCanvasContent()
       └─ convertHTMLToMarkdown() - HTML→Markdown変換処理

canvas_converter.go (新規)
  └─ ConvertHTMLToMarkdown(html string) string
  └─ convertNode(node *html.Node) string - 再帰的な変換処理
```

### 変換ロジックの概要

1. HTMLをパースしてDOMツリーを構築
2. ルートから再帰的にノードを走査
3. 各ノードのタグ名に応じてMarkdown記法に変換
4. `data-section-style`属性でリストの種類を判定
5. `class='checked'`でチェック状態を判定
6. ネストレベルを追跡してインデントを適用

## 実装計画

### 設計方針

#### ファイル構成

```
handler_get_canvas_content.go
  └─ getCanvasContent()
       └─ buf.String() を ConvertHTMLToMarkdown() に渡す

canvas_converter.go (新規)
  └─ ConvertHTMLToMarkdown(html string) string
  └─ convertNode(node *html.Node, ctx *converterContext) string
```

#### 変換ロジックの概要

```go
// canvas_converter.go

type converterContext struct {
    listDepth    int      // ネストレベル
    listStyle    string   // リスト種別 ("5"=箇条書き, "6"=番号付き, "7"=チェック)
    listCounter  int      // 番号付きリストのカウンター
    inCodeBlock  bool     // コードブロック内かどうか
}

func ConvertHTMLToMarkdown(htmlContent string) string {
    // 1. HTMLをパース
    // 2. ルートから再帰的にノードを走査
    // 3. 各ノードに応じてMarkdown記法に変換
}

func convertNode(node *html.Node, ctx *converterContext) string {
    // タグごとの変換処理
    // 未知のタグでも子ノードを再帰処理してテキストを抽出
}
```

#### 未知要素の処理

未知のHTML要素に遭遇した場合でも、子ノードを再帰的に処理することでテキスト部分は必ず出力する（ベストエフォート）。

### フェーズ1: Canvas HTML to Markdown Converter の実装

Converterを独立したモジュールとして実装し、単体テストで品質を担保する。

#### Commit 1: Add basic element conversion (headings, paragraphs, text formatting, links)
- `canvas_converter.go` を新規作成
- `golang.org/x/net/html` パッケージを依存に追加
- `ConvertHTMLToMarkdown()` 関数と `convertNode()` 関数の基本構造を実装
- 対応要素:
  - 見出し: `<h1>`, `<h2>`, `<h3>` → `#`, `##`, `###`
  - 段落: `<p>` → 空行区切りのテキスト
  - 太字: `<b>` → `**text**`
  - 斜体: `<i>` → `*text*`
  - 下線: `<u>` → `<u>text</u>`（HTMLタグ維持）
  - 取り消し線: `<del>` → `~~text~~`
  - リンク: `<a href="...">` → `[text](url)`
- 未知要素のフォールバック処理（子ノードのテキストを抽出）
- `canvas_converter_test.go` にテストを追加

#### Commit 2: Add list conversion (bullet, numbered, checklist with nesting)
- 対応要素:
  - 箇条書きリスト: `data-section-style='5'` → `- `
  - 番号付きリスト: `data-section-style='6'` → `1. `
  - チェックリスト: `data-section-style='7'` → `- [ ]` / `- [x]`
  - ネストしたリスト: インデントで表現
- `converterContext` でネストレベルとリスト種別を追跡
- テストを追加

#### Commit 3: Add block element conversion (table, blockquote, code block, embedded file)
- 対応要素:
  - テーブル: `<table>` → Markdownテーブル
  - 引用: `<blockquote>` → `> `
  - コードブロック: `<pre>` → ` ``` `
  - 添付ファイル: `<p class='embedded-file'>` → `![File](url)`
- テストを追加

### フェーズ2: GetCanvasContent への統合

既存のAPIに変換機能を組み込む。

#### Commit 4: Integrate Markdown conversion into GetCanvasContent
- `handler_get_canvas_content.go` の `getCanvasContent()` を修正
- `buf.String()` の結果を `ConvertHTMLToMarkdown()` に渡す
- `handler_get_canvas_content_test.go` のテストを更新（HTMLではなくMarkdownが返ることを確認）

### フェーズ3: ドキュメント更新

READMEを更新してユーザーに変更を周知する。

#### Commit 5: Update get_canvas_content documentation
- `README_ja.md` を更新
  - `get_canvas_content` の説明を「HTMLコンテンツを取得」から「Markdownコンテンツを取得」に変更
  - 変換される要素の説明を追加
- `README.md` を更新（英語版）

## 参考資料

- 元の要件定義書: `docs/onetime/20260117-canvas-search-requirements-ja.md`
- 実際のCanvas HTML例: `https://clustervr.slack.com/docs/T076HLEP4/F0AAD9Q2W4E`
