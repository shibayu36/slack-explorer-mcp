# Attachment（unfurl）情報の取得対応 - 要件定義書

## Context / Goal

### 背景
Slack Explorer MCPの `search_messages` と `get_thread_replies` は現在、メッセージの基本情報（テキスト、ユーザー、タイムスタンプ等）のみを返している。URL unfurl情報（Slack APIレスポンスの`attachments`フィールド）が完全にドロップされているため、AIアシスタントがメッセージに含まれるリンク先の情報（ツイート内容、YouTube動画タイトル等）を把握できない。

### ゴール
`search_messages` と `get_thread_replies` で attachments から unfurl情報を取得できるようにし、AIアシスタントがメッセージの文脈をより正確に理解できるようにする。トークン効率のため、返すフィールドは `title`, `text`, `from_url` の3つに絞る。

## In scope

1. **search_messages のレスポンスに attachments 情報を追加**
   - 各メッセージに `attachments` フィールドを新規追加
   - 各 attachment から `title`, `text`, `from_url` のみを抽出して返す

2. **get_thread_replies のレスポンスに attachments 情報を追加**
   - 同上

## Out of scope

- `blocks`（Block Kit情報）の対応
- `attachments` 内の `blocks`（Figma等のapp unfurl）の対応
- 画像・動画・サムネイル等のメディア情報（`image_url`, `thumb_url`, `video_html`等）
- `attachments` のその他フィールド（`service_name`, `author_name`, `fallback`, `color`等）
- 新しいMCPツールの追加

## User Stories

1. **AIアシスタントとして**、search_messagesの結果からリンク先のタイトルと概要テキストを取得したい。**なぜなら**、メッセージに貼られたURLの中身（ツイート内容、記事タイトル等）を把握してユーザーの質問に正確に答えるため。

2. **AIアシスタントとして**、get_thread_repliesの結果からリンク先のタイトルと概要テキストを取得したい。**なぜなら**、スレッド内で共有されたリンクの文脈を理解するため。

3. **AIアシスタントとして**、unfurl情報から元URLを取得したい。**なぜなら**、情報の出典をユーザーに提示するため。

4. **AIアシスタントとして**、attachmentsが存在しないメッセージでは余分な情報が返されないようにしたい。**なぜなら**、コンテキストウィンドウを効率的に使うため。

## Acceptance Criteria

### 主要シナリオ

```gherkin
Scenario: search_messagesでURL付きメッセージのunfurl情報が取得できる
  Given SlackにX/Twitterリンクを含むメッセージが存在する
  When search_messagesでそのメッセージを検索する
  Then レスポンスの各メッセージにattachmentsフィールドが含まれている
  And 各attachmentにtitle, text, from_urlが含まれている
  And それ以外のフィールド（service_name, thumb_url等）は含まれていない
```

```gherkin
Scenario: get_thread_repliesでスレッド内メッセージのunfurl情報が取得できる
  Given スレッド内にunfurlされたリンクを含む返信が存在する
  When get_thread_repliesでそのスレッドを取得する
  Then 返信メッセージのattachmentsフィールドにtitle, text, from_urlが含まれている
```

```gherkin
Scenario: 1つのメッセージに複数のリンクがある場合
  Given メッセージに複数のURLが含まれそれぞれunfurlされている
  When search_messagesでそのメッセージを検索する
  Then attachmentsフィールドに複数のattachmentが含まれている
  And 各attachmentにそれぞれのリンク先情報が含まれている
```

### 例外シナリオ

```gherkin
Scenario: attachmentsが存在しないメッセージ
  Given Slackに通常のテキストのみのメッセージが存在する
  When search_messagesでそのメッセージを検索する
  Then attachmentsフィールドは空またはnullである
  And 既存のフィールド（text, user, ts等）は正常に返される
```

```gherkin
Scenario: unfurlされなかったURL付きメッセージ
  Given メッセージにURLが含まれるがSlackがunfurlしなかった
  When search_messagesでそのメッセージを検索する
  Then attachmentsフィールドは空またはnullである
  And テキスト内のURLは引き続きtextフィールドで確認できる
```

## NFR（非機能要件）

- **性能**: Slack APIからすでに取得済みのデータから抽出するだけなので、応答時間への影響なし
- **トークン効率**: フィールドを3つ（`title`, `text`, `from_url`）に限定し、トークン消費を最小化
- **後方互換性**: 既存のレスポンスフィールド（user, text, ts, channel, thread_ts等）は一切変更しない。attachmentsは新規フィールドとして追加
- **セキュリティ**: 既存のSlack User Token権限範囲内で動作。新たな権限スコープは不要

## Open Questions

（なし）

## メモ: 将来のblocks対応時の参考情報

attachments内の`blocks`フィールドにテキスト情報を持つケースがある（Figma, Notion/JIRA, Sentry等のapp unfurl）。将来blocks対応する際は、以下のblock typeからテキストを抽出するのが有効。

### テキストを持つblock type（公式ドキュメントより）

| type | テキストの場所 |
|---|---|
| `section` | `text.text`（mrkdwn/plain_text）、`fields[].text` |
| `context` | `elements[]`内のtext object |
| `header` | `text.text`（plain_textのみ、最大150文字） |
| `rich_text` | `elements[].elements[].text`（ネストが深い） |
| `markdown` | `text`（文字列直接、最大12,000文字） |

### テキストを持たないblock type

`actions`, `divider`, `image`, `video`, `file`, `input`

### 実データでの出現パターン

- **Figma**: `section`（ファイル名+URL）+ `image`（サムネイル）
- **Notion/JIRA**: `section`（コメント内容）+ `actions`（ボタン）
- **Sentry**: `section`（エラー名・メッセージ）+ `context`（スタックトレース・統計）+ `actions`（ボタン）

### 参考リンク

- https://docs.slack.dev/reference/block-kit/blocks/
- https://docs.slack.dev/reference/block-kit/blocks/section-block/
- https://docs.slack.dev/reference/block-kit/blocks/context-block/
- https://docs.slack.dev/reference/block-kit/blocks/header-block/
- https://docs.slack.dev/reference/block-kit/blocks/rich-text-block/
- https://docs.slack.dev/reference/block-kit/blocks/markdown-block/
