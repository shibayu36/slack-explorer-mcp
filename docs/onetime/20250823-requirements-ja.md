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

## MCPツール（APIサーフェスの名前のみ）
- `list_channels` — List channels in the workspace with pagination
- `get_channel_history` — Get recent messages from a channel
- `get_thread_replies` — Get all replies in a message thread
- `search_messages` — Search for messages in the workspace with powerful filters

※ 入力・出力の詳細は**TBD**。後で詰める。

---

## Slack API依存とスコープ
- 使用メソッド（想定）：`conversations.list`, `conversations.history`, `conversations.replies`, `users.list`（表示名解決用）, `search.messages`（検索）
- 必要スコープ（User Token想定）
  - 公開：`channels:read`, `channels:history`
  - 非公開：`groups:read`, `groups:history`
  - DM/マルチDM：`im:read`, `im:history`, `mpim:read`, `mpim:history`
  - ユーザー：`users:read`
  - 検索：`search:read`
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
1) `search_messages`
2) `get_thread_replies`
3) `get_channel_history`
4) `list_channels`
