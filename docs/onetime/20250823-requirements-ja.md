# slack-explorer-mcp / requirements.md

## 概要
- 目的：Slackのメッセージ検索、チャンネル/DM履歴取得、スレッド返信取得を**Model Context Protocol (MCP)** サーバとして提供する。
- 方針：**User Token（xoxp）を基本**。認可ユーザーが閲覧可能な範囲のみを扱う。
- 先送り：**入出力スキーマ詳細はTBD**。まずは機能要件と制約を固める。

---

## スコープ
### In（MVP）
- `list_channels`：ワークスペースのチャンネル一覧（公開/非公開/DM/マルチDMを対象にできる）
- `get_channel_history`：チャンネル/DMの最近メッセージ取得（ページネーション対応）
- `get_thread_replies`：特定メッセージのスレッド返信一覧取得
- `search_messages`：強力なフィルタ付きメッセージ検索（Slack検索APIを前提）
- ページネーション（Slackの`next_cursor`準拠）
- 権限・可視性：**ユーザー視点**でアクセス制御（認可ユーザーが見えるもののみ）

### Out（非MVP）
- 投稿/編集/削除など書き込み系
- 監査・法務向けDiscovery/Audit API連携
- Enterprise Grid横断
- 恒久的な全文アーカイブ/インデックス作成（後フェーズで検討）
- Slackディープリンク生成（当面**非対応**）

---

## 非機能要件
- 認証/認可
  - **User Token 基本**（xoxp）。必要スコープは下記参照
  - 可視性は**認可ユーザーが参加/閲覧可能な会話**に限定
- レート制限
  - 初期は考慮しない。429は**そのままエラー返却**（自動リトライ無し）
- 設定/運用
  - 環境変数でトークン注入（例：`SLACK_USER_TOKEN`）
  - ローカル（Mac）で起動可能。将来Docker化

---

## MCPツール（APIサーフェス）

### 実装済み（✅）
- **`search_messages`** — Search for messages in the workspace with powerful filters
  - 高度な検索フィルタ（with, has, hasmy）対応
  - ページネーション完全対応
  - エラーハンドリング実装済み
- **`get_thread_replies`** — Get all replies in a message thread
  - ページネーション対応
  - エラーハンドリング実装済み

### 未実装（🚧）
- `list_channels` — List channels in the workspace with pagination
- `get_channel_history` — Get recent messages from a channel
- `get_user_profiles` — Get multiple users profile information in bulk

### search_messages スキーマ詳細

#### 入力パラメータ
```json
{
  "query": "string (optional)",        // 基本検索クエリ（修飾子なし）
  "in_channel": "string (optional)",   // チャンネル名（例: "general", "random", "チーム-dev"）
  "from_user": "string (optional)",    // ユーザーID（例: "U1234567"）
  "with": "array (optional)",          // 特定ユーザーとのDM/スレッドを検索。ユーザーID配列（例: ["U1234567", "U2345678"]）
  "before": "string (optional)",       // YYYY-MM-DD形式
  "after": "string (optional)",        // YYYY-MM-DD形式  
  "on": "string (optional)",           // YYYY-MM-DD形式
  "during": "string (optional)",       // 期間指定（例: "July", "2023", "last week"）
  "has": "array (optional)",           // 特定機能を含むメッセージ検索。対応値: 絵文字（":eyes:", ":fire:"）, "pin", "file", "link", "reaction"
  "hasmy": "array (optional)",         // 認証ユーザーが特定絵文字リアクションしたメッセージ検索。絵文字のみ（例: [":eyes:", ":fire:"]）
  "highlight": "boolean (optional)",   // 検索結果のハイライト（デフォルト: false）
  "sort": "string (optional)",         // ソート方法: "score" | "timestamp"（デフォルト: "score"）
  "sort_dir": "string (optional)",     // ソート順: "asc" | "desc"（デフォルト: "desc"）
  "count": "number (optional)",        // ページあたりの結果数（1-100、デフォルト: 20）
  "page": "number (optional)"          // ページ番号（1-100、デフォルト: 1）
}
```

#### 出力形式
```json
{
  "messages": {
    "matches": [
      {
        "user": "U1234567",
        "text": "メッセージ本文",
        "ts": "1234567890.123456",
        "thread_ts": "1234567890.123456",  // スレッドの場合のみ。メインメッセージの場合は空文字
        "channel": {
          "id": "C1234567",
          "name": "channel-name"
        }
      }
    ],
    "pagination": {
      "total_count": 100,
      "page": 1,
      "page_count": 5,
      "per_page": 20,
      "first": 1,
      "last": 5
    }
  }
}
```

#### 実装上の注意点
- `query`フィールドには修飾子（from:, in:等）を含めない
- チャンネル指定は`in_channel`でチャンネル名を受け取り、そのままSlack検索APIに渡す
- ユーザー指定は`from_user`でユーザーIDを受け取り、`<@USER_ID>`形式でクエリに追加
- `with`パラメータはユーザーID配列を受け取り、`with:<@USER_ID>`形式でクエリに追加
- `has`パラメータは機能指定を受け取り、`has:VALUE`形式でクエリに追加
- `hasmy`パラメータは絵文字コードを受け取り、`hasmy:EMOJI`形式でクエリに追加
- `thread_ts`は Slack のパーマリンクURLから抽出して設定
- User Token（xoxp）を使用し、`search.messages` APIを呼び出す

### get_thread_replies スキーマ詳細

#### 入力パラメータ
```json
{
  "channel_id": "string (required)",    // チャンネルID（例: "C1234567"）
  "thread_ts": "string (required)",     // 親メッセージのタイムスタンプ（例: "1234567890.123456"）
  "limit": "number (optional)",         // 取得する返信数（1-1000、デフォルト: 100）
  "cursor": "string (optional)"         // ページネーション用カーソル
}
```

#### 出力形式
```json
{
  "messages": [
    {
      "user": "U1234567",
      "text": "メッセージ本文",
      "ts": "1234567890.123456",
      "reply_count": 5,                    // 返信数（親メッセージの場合のみ）
      "reply_users": ["U1234567", "U2345678"], // 返信したユーザー（親メッセージの場合のみ）
      "reactions": [                       // リアクション情報（オプション）
        {
          "name": "thumbsup",
          "count": 2,
          "users": ["U1234567", "U2345678"]
        }
      ]
    }
  ],
  "has_more": false,                      // さらに返信があるか
  "response_metadata": {
    "next_cursor": "string"                // 次ページ用カーソル（has_more=trueの場合）
  }
}
```

#### 実装上の注意点
- `channel_id`はSlack APIの`conversations.replies`に必要なチャンネルID
- `thread_ts`は親メッセージのタイムスタンプ（形式: `1234567890.123456`）
- 最初のメッセージは親メッセージ自体を含む（Slack APIの仕様）
- User Token（xoxp）を使用し、`conversations.replies` APIを呼び出す
- ページネーションは`cursor`と`limit`で制御
- `has_more`がtrueの場合、`response_metadata.next_cursor`を次回リクエストの`cursor`に使用

### get_user_profiles スキーマ詳細

#### 入力パラメータ
```json
{
  "user_ids": "array (required)"         // ユーザーID配列（例: ["U1234567", "U2345678"]）最大100個
}
```

#### 出力形式
```json
[
  {
    "user_id": "U1234567",
    "display_name": "John Doe",
    "real_name": "John Doe",
    "email": "john@example.com"
  },
  {
    "user_id": "U2345678",
    "error": "user_not_found"              // エラーの場合はerrorフィールドのみ
  }
]
```

#### 実装上の注意点
- `user_ids`は必須パラメータで、最大100個のユーザーIDを受け取る
- 各ユーザーIDは`U`で始まる形式（例: `U1234567`）である必要がある
- Slack APIの`users.profile.get`を順次呼び出し、個別のエラーも含めて結果を返す
- User Token（xoxp）を使用し、`users.profile.get` APIを呼び出す
- 一部のユーザーでエラーが発生しても、他のユーザーの情報は正常に返す
- 必要スコープ: `users:read`, `users.profile:read`

※ 他のツールの入力・出力詳細は後続実装時に定義

---

## Slack API依存とスコープ
- **使用SDK**: `github.com/slack-go/slack` - 導入済み ✅
- 使用メソッド（想定）：`conversations.list`, `conversations.history`, `conversations.replies`, `users.list`（表示名解決用）, `search.messages`（検索）
  - **実装済み**: `search.messages` ✅
- 必要スコープ（User Token想定）
  - 公開：`channels:read`, `channels:history`
  - 非公開：`groups:read`, `groups:history`
  - DM/マルチDM：`im:read`, `im:history`, `mpim:read`, `mpim:history`
  - ユーザー：`users:read`
  - 検索：`search:read` ✅
- 注意点
  - プライベート/DMは**ユーザー参加済み**でなければ取得不可
  - ワークスペース単位の前提（Grid横断はしない）

---

## データモデルの扱い（TBD方針）
- 共通メッセージ表現（ID/チャンネル/ユーザー/本文/タイムスタンプ/スレッド情報/リアクション/ファイルメタ等）を**薄く正規化**
- タイムスタンプはSlack互換（文字列`ts`）で扱う方針
- 添付ファイルは**メタ情報のみ**（URL実体は扱わない）

---

## エラー方針
- 原則：Slackエラーを**薄くマップ**して返却（人間可読メッセージ付き）
  - 認証系：`not_authed`/`invalid_auth`/`missing_scope` → 認可エラーとして返却
  - 見つからない系：`channel_not_found`/`user_not_found`
  - レート制限：HTTP 429 をそのまま返却（`Retry-After`値をメッセージに添付するのみ）
  - パラメータ不備：バリデーションエラー

---

## ページネーション/並び順
- ページネーション：Slackの`next_cursor`をそのまま透過
- 並び順：
  - `get_channel_history`：当面Slackの取得順に従う（後で昇順正規化は検討）
  - `get_thread_replies`：時系列昇順を基本

---

## マイルストン/実装順

### 完了済み（✅）
1) **`search_messages`** - 完全実装済み
   - Slack SDK（`github.com/slack-go/slack`）導入完了
   - 高度な検索フィルタリング機能（with, has, hasmy）実装
   - パラメータバリデーション実装
   - Slackエラーマッピング実装
   - ページネーション対応

### 未実装（🚧）
2) `get_thread_replies` - 未実装
3) `get_channel_history` - 未実装  
4) `list_channels` - 未実装

### 技術的進展
- **認証・エラーハンドリング**: 完了済み
- **SlackClient インターフェース設計**: 完了済み
- **MCPサーバー基盤**: 完了済み
