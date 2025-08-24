# Slack Explorer MCP Server

Slackのメッセージやスレッドなどの**情報の取得**に特化したModel Context Protocol (MCP) サーバーです。User Token（xoxp）を使用して、認可したユーザーがアクセス可能なメッセージを取得するツールを提供します。

## 提供するツール

- メッセージ検索 (`search_messages`)
  - 高度な検索フィルタ付きでSlackメッセージを検索します。チャンネル指定、ユーザー指定、日付範囲、特定の機能（リアクション、ファイル等）を含むメッセージの検索が可能です。
  - パラメータ
    - `query`: 基本検索クエリ（修飾子なし）
    - `in_channel`: チャンネル名での絞り込み（例: "general", "チーム-dev"）
    - `from_user`: 特定ユーザーのメッセージを検索（ユーザーID）
    - `with`: 特定ユーザーとのDM/スレッドを検索（ユーザーID配列）
    - `before`, `after`, `on`: 日付範囲指定（YYYY-MM-DD形式）
    - `during`: 期間指定（例: "July", "2023"）
    - `has`: 特定機能を含むメッセージ（絵文字、"pin", "file", "link", "reaction"）
    - `hasmy`: 自分がリアクションした絵文字を含むメッセージ
    - `sort`: ソート方法（"score" or "timestamp"）
    - `count`: ページあたりの結果数（1-100、デフォルト: 20）
    - `page`: ページ番号（1-100、デフォルト: 1）

- スレッド返信取得 (`get_thread_replies`)
  - 特定メッセージのスレッド返信一覧を取得します。ページネーション対応で大量の返信も効率的に取得できます。
  - パラメータ
    - `channel_id`: チャンネルID（必須）
    - `thread_ts`: 親メッセージのタイムスタンプ（必須）
    - `limit`: 取得する返信数（1-1000、デフォルト: 100）
    - `cursor`: ページネーション用カーソル

- ユーザープロフィール一括取得 (`get_user_profiles`)
  - 複数のユーザーのプロフィール情報を一括で取得します。ユーザーIDのリストを指定して、表示名、実名、メールアドレスなどの情報を取得できます。
  - パラメータ
    - `user_ids`: ユーザーID配列（必須、最大100個）

## セットアップ

### Slack User Tokenの取得

1. [Slack API](https://api.slack.com/apps)でアプリを作成
2. OAuth & Permissionsで以下のUser Token Scopesを追加：
   - `search:read` - メッセージ検索用
   - `channels:read`, `channels:history` - 公開チャンネル用
   - `groups:read`, `groups:history` - 非公開チャンネル用
   - `im:read`, `im:history` - DM用
   - `mpim:read`, `mpim:history` - グループDM用
   - `users:read`, `users.profile:read` - ユーザー情報・プロフィール取得用
3. ワークスペースにアプリをインストール
4. User OAuth Token（xoxp-で始まるトークン）を取得

### MCPサーバーの設定

1. mcp.jsonを設定

    ```json
    {
      "mcpServers": {
        "slack-explorer-mcp": {
          "command": "docker",
          "args": ["run", "-i", "--rm",
            "-e", "SLACK_USER_TOKEN=xoxp-your-token-here",
            "ghcr.io/shibayu36/slack-explorer-mcp:latest"
          ]
        }
      }
    }
    ```

    Claude Codeを使用している場合:

    ```bash
    claude mcp add slack-explorer-mcp -- docker run -i --rm \
      -e SLACK_USER_TOKEN=xoxp-your-token-here \
      ghcr.io/shibayu36/slack-explorer-mcp:latest
    ```

2. エージェントを利用してSlack検索を実行

    例:
    - 「generalチャンネルで先週の会議関連のメッセージを検索して」
    - 「@john.doeさんのメッセージで"プロジェクト"に関するものを探して」
    - 「この投稿のスレッド返信を全部取得して」

## 使い方

### よくある検索パターン

- **特定チャンネルでの検索**
  ```
  generalチャンネルで"リリース"に関するメッセージを検索
  ```

- **特定ユーザーのメッセージ検索**
  ```
  @john.doeさんの昨日のメッセージを検索
  ```

- **リアクション付きメッセージの検索**
  ```
  :fire:のリアクションがついているメッセージを検索
  ```

- **自分がリアクションしたメッセージ**
  ```
  自分が:eyes:のリアクションをつけたメッセージを検索
  ```

- **ファイル付きメッセージの検索**
  ```
  ファイルが添付されているメッセージを検索
  ```
