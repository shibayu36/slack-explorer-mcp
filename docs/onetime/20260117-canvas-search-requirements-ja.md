# Slack File Search & Canvas Content Retrieval - 要件定義書

## Context / Goal

### 背景
現在のslack-explorer-mcpは`search_messages`、`get_thread_replies`、`get_user_profiles`、`search_users_by_name`の4つのtoolを提供している。SlackのCanvasやPDFなどのファイルはナレッジ共有に使われているが、現状ファイル内の情報を検索・取得する手段がない。

### ゴール
Slackのファイルを汎用的に検索し、Canvas等の中身を取得できるMCP toolを追加することで、AIアシスタントがファイル内のナレッジにアクセスできるようにする。

## In scope

1. **search_files tool**: キーワード、ファイルタイプ、作成者、チャンネル、日付範囲などでファイルを検索
   - 内部的に`search.files` APIを使用
   - `types`パラメータで`canvases`、`pdfs`等を指定可能（複数指定可）
   - 既存の`search_messages`と同様のフィルタ条件をサポート

2. **get_canvas_content tool**: Canvas IDを指定して中身のテキストを取得
   - `url_private_download`からHTMLをダウンロード
   - サイズ制限なし（全文取得）

## Out of scope

- Canvasの作成・編集・削除機能
- Canvasのアクセス権限変更
- Canvas内の画像・添付ファイルの取得
- リアルタイム更新通知
- Canvasのバージョン履歴取得
- HTMLの高度な簡略化（見出し・リンクのみ抽出等、canvas_extractor.go）→ 次回スコープ
- Canvas以外のファイルタイプの中身取得（get_pdf_content等）→ 次回スコープ

## User Stories

1. **US1**: ユーザーとして、キーワードでCanvasを検索して、関連するドキュメントを見つけたい
2. **US2**: ユーザーとして、特定の人が作成したファイルを検索して、その人のナレッジを参照したい
3. **US3**: ユーザーとして、特定のチャンネルに関連するファイルを検索して、プロジェクト関連のドキュメントを見つけたい
4. **US4**: ユーザーとして、検索で見つけたCanvasの中身を読んで、詳細な情報を把握したい
5. **US5**: ユーザーとして、最近更新されたファイルを検索して、最新のナレッジを取得したい
6. **US6**: ユーザーとして、CanvasとPDFを同時に検索して、関連資料をまとめて見つけたい

## Acceptance Criteria (Gherkin)

### 主要ケース

```gherkin
Feature: ファイル検索機能

  Scenario: キーワードとタイプ指定でCanvasを検索する
    Given Slackトークンが設定されている
    When "プロジェクト計画"というキーワードとtypes=["canvases"]でsearch_filesを実行する
    Then Canvas一覧が返される
    And 各ファイルにはid、title、filetype、user（作成者）、channels、updated、permalinkが含まれる

  Scenario: 複数のファイルタイプを同時に検索する
    Given Slackトークンが設定されている
    When types=["canvases", "pdfs"]でsearch_filesを実行する
    Then CanvasとPDFの両方が検索結果に含まれる

  Scenario: 作成者でファイルをフィルタ検索する
    Given Slackトークンが設定されている
    When from_userに"U12345678"を指定してsearch_filesを実行する
    Then 指定ユーザーが作成したファイル一覧のみが返される

  Scenario: Canvasの中身を取得する
    Given Slackトークンが設定されている
    And Canvas ID "F12345678"が存在する
    When get_canvas_contentにcanvas_ids=["F12345678"]を指定して実行する
    Then Canvasのタイトルとコンテンツが返される
    And コンテンツはHTMLとして取得される
```

### 例外ケース

```gherkin
  Scenario: 認証エラー時の処理
    Given Slackトークンが設定されていない
    When search_filesを実行する
    Then "slack token not configured"エラーが返される

  Scenario: 存在しないCanvasの取得
    Given Slackトークンが設定されている
    When get_canvas_contentに存在しないcanvas_idsを指定して実行する
    Then "file not found"エラーが返される
```

## Technical Design

### Tool定義

#### search_files
| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| query | string | No | 検索キーワード |
| types | string[] | No | ファイルタイプでフィルタ（複数指定可） |
| in_channel | string | No | チャンネル名でフィルタ |
| from_user | string | No | 作成者のユーザーIDでフィルタ |
| with_user | string[] | No | スレッド/DMの相手ユーザーIDでフィルタ |
| before | string | No | この日付より前（YYYY-MM-DD） |
| after | string | No | この日付より後（YYYY-MM-DD） |
| on | string | No | この日付に作成（YYYY-MM-DD） |
| count | number | No | 取得件数（デフォルト: 20、最大: 100） |
| page | number | No | ページ番号（デフォルト: 1） |

**利用可能なファイルタイプ（types）:**
- `lists` - リスト
- `canvases` - キャンバス
- `documents` - ドキュメント（Google Docs等）
- `emails` - メール
- `images` - 画像
- `pdfs` - PDF
- `presentations` - プレゼンテーション
- `snippets` - スニペット
- `spreadsheets` - スプレッドシート
- `audio` - 音声
- `videos` - 動画

レスポンス例:
```json
{
  "ok": true,
  "query": "プロジェクト type:canvases",
  "pagination": {
    "total_count": 42,
    "page": 1,
    "page_count": 3,
    "per_page": 20,
    "first": 1,
    "last": 20
  },
  "files": [
    {
      "id": "F12345678",
      "title": "プロジェクト計画書",
      "filetype": "canvas",
      "user": "U12345678",
      "channels": ["C12345678"],
      "created": 1704067200,
      "updated": 1704153600,
      "permalink": "https://xxx.slack.com/files/U12345678/F12345678/xxx"
    }
  ]
}
```

#### get_canvas_content
| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| canvas_ids | string[] | Yes | CanvasのファイルID配列（Fから始まる） |

レスポンス例:
```json
{
  "canvases": [
    {
      "id": "F12345678",
      "title": "プロジェクト計画書",
      "content": "<h1>概要</h1><p>このプロジェクトは...</p><h2>スケジュール</h2>...",
      "user": "U12345678",
      "created": 1704067200,
      "updated": 1704153600,
      "permalink": "https://xxx.slack.com/files/U12345678/F12345678/xxx"
    }
  ]
}
```

### 実装アーキテクチャ

```
main.go
  └─ AddTool("search_files", handler.SearchFiles)
  └─ AddTool("get_canvas_content", handler.GetCanvasContent)

handler.go
  └─ SearchFiles() - パラメータ検証 → SearchFiles API呼び出し → レスポンス変換
  └─ GetCanvasContent() - url_private_downloadからダウンロード → レスポンス返却

slack_client.go
  └─ SearchFiles(query, params) - search.files API
  └─ GetFile(url, writer) - ファイルダウンロード（既存メソッド活用）
```

### 使用するSlack API

1. **search.files** - ファイル検索
   - 必要スコープ: `search:read`
   - クエリに`type:xxx`を付与してフィルタ

2. **HTTPダウンロード** - Canvas内容取得
   - `api.GetFile(url, writer)` を使用（Bearer認証自動）

## NFR（非機能要件）

### 性能
- 検索レスポンス: 3秒以内（Slack APIの制約に依存）
- コンテンツ取得: 5秒以内（HTMLダウンロード＋パース含む）

### コスト
- Slack APIのRate Limit（Tier 2: 20+ per minute）を考慮
- 大量のファイル取得時は適切にページネーション

### セキュリティ
- トークンはContext経由で管理（既存パターン踏襲）
- トークンをログに出力しない

### 監査
- エラー発生時のログ出力（既存パターン踏襲）

### 運用
- 既存のslack-goライブラリを活用

## 確認済み事項

1. **slack-goのsearch.files対応状況**: ✅ 対応済み
   - `api.SearchFiles(query, params)` で検索可能
   - `type:canvases` をクエリに付与してCanvas絞り込み
   - `api.GetFile(url, w)` でHTMLダウンロード可能（Bearer認証自動）

## Open Questions

なし（全て解決済み）

## 修正対象ファイル

- `main.go` - 新規tool定義追加
- `handler.go` - SearchFiles, GetCanvasContent実装
- `slack_client.go` - SearchFiles追加
- `handler_test.go` - テスト追加
- `slack_client_mock_test.go` - モック追加
- `README.md` / `README_ja.md` - ドキュメント更新

## 検証方法

1. **単体テスト**: handler_test.goで各toolのテスト
2. **手動テスト**: 実際のSlackワークスペースでファイル検索・Canvas取得を確認
3. **MCPクライアントテスト**: Claude Codeから実際にtoolを呼び出して動作確認

## 実装計画

実装時は https://github.com/shibayu36/slack-explorer-mcp/pull/26 のPoCのコード変更を参考にしてよい。

### フェーズ1: search_files tool（4 commits）

| # | commit message | 対象ファイル |
|---|----------------|-------------|
| 1-1 | Add SearchFiles method to SlackClient | slack_client.go, slack_client_mock_test.go |
| 1-2 | Add SearchFiles handler and tests | handler.go, handler_test.go |
| 1-3 | Add search_files tool definition | main.go |
| 1-4 | Add search_files documentation | README_ja.md, README.md |

### フェーズ2: get_canvas_content tool（5 commits）

| # | commit message | 対象ファイル |
|---|----------------|-------------|
| 2-1 | Add GetFileInfo method to SlackClient | slack_client.go, slack_client_mock_test.go |
| 2-2 | Add GetFile method to SlackClient | slack_client.go, slack_client_mock_test.go |
| 2-3 | Add GetCanvasContent handler and tests | handler.go, handler_test.go |
| 2-4 | Add get_canvas_content tool definition | main.go |
| 2-5 | Add get_canvas_content documentation | README_ja.md, README.md |
