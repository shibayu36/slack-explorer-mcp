# query フィールドでの除外 modifier 許可 - 要件定義書

対象ブランチ: `allow-exclude-query`

## Context / Goal

### 背景
- `search_messages` / `search_files` ツールでは、`query` フィールドに modifier（`from:`, `in:`, `before:`, `after:`, `on:`, `during:`, `has:`, `is:`, `with:`, `type:`）が含まれていると validation でエラーになる。
- 包含条件は専用フィールド（`in_channel`, `from_user`, `with`, `has`, `hasmy`, `types`, `before`, `after`, `on`, `during`）で表現できるが、**除外条件**を表現する手段が現状ない。
- 「このチャンネルを除外して検索したい」「このユーザー以外」「PDF以外」のようなニーズを Slack 標準のシンタックス `-in:#channel` `-from:@user` で書きたいが、現在の validation 正規表現が `-` 付きの modifier にもマッチしてエラーになる。

### ゴール
- `query` フィールドで `-` プレフィックス付き modifier を許可し、Slack の除外検索シンタックスをそのまま利用できるようにする。
- `-` プレフィックスのない（=包含）modifier の混入は引き続きブロックし、専用フィールドへの誘導を維持する。

## In scope

1. **`search_messages` の `query` で `-` プレフィックス付き modifier を許可**
   - 対象 modifier: `from`, `in`, `before`, `after`, `on`, `during`, `has`, `is`, `with`
   - 「直前が空白 もしくは 行頭」にある `-` のみを除外プレフィックスとして扱う
2. **`search_files` の `query` で `-` プレフィックス付き modifier を許可**
   - 対象 modifier: `from`, `in`, `before`, `after`, `on`, `type`
   - 同上の境界条件
3. **専用フィールドと query 内除外指定の併用を許可**
   - 例: `in_channel="alpha"` + `query="-in:#beta"` を同時指定可
4. **包含 modifier の混入は引き続き validation エラーとする**
   - 既存挙動を維持
5. **ツール description（`mcp.WithString("query", ...)` の説明文）の更新**
   - 「`-` プレフィックス付きなら除外指定として `query` に書ける」旨を明示
6. **エラーメッセージ文面の更新**
   - 包含 modifier ブロック時のメッセージで、「除外（`-` プレフィックス付き）であれば `query` に書ける」旨を案内

## Out of scope

- 除外専用フィールド（例: `exclude_in_channel`, `exclude_from_user`）の新設
- 既存の modifier 一覧に新しい modifier を追加すること
- 単純な単語除外（例: `error -warning`）の挙動変更。現状でも validation を素通りして Slack 側に渡っている前提で、本要件の変更対象としない
- `search_messages` / `search_files` 以外のツールへの影響
- 単語の途中に現れる `-`（例: `foo-in:#a`）の除外指定としての解釈。これは除外プレフィックスとはみなさず、無印 modifier 扱いで validation エラーとする
- 個々の `-modifier:` 組み合わせ（例: `-is:thread`）が Slack 側で実際にサポートされるかの個別検証。validation 上はすべて通し、解釈は Slack API に委ねる

## User Stories

1. **検索ユーザーとして**、ノイズの多い特定チャンネルを除外して全体検索したい。**なぜなら**、`query` に `-in:#general` のような除外指定を書ければ、目的のメッセージにたどり着きやすくなるため。
2. **検索ユーザーとして**、bot の発言を除外して検索したい。**なぜなら**、`query` に `-from:@some-bot` を書ければ、人間の発言だけを抽出できるため。
3. **検索ユーザーとして**、`in_channel=#alpha` で対象を絞った上で、サブチャンネル `#beta` だけ除外したい。**なぜなら**、専用フィールドと `query` 内 `-in:#beta` を併用できれば、複雑な絞り込みを表現できるため。
4. **ファイル検索ユーザーとして**、PDF 以外のファイルを検索したい。**なぜなら**、`query` に `-type:pdf` を書ければ、ノイズを減らせるため。
5. **検索ユーザーとして**、誤って `from:U123` のような包含指定を `query` に書いてしまった場合は、これまで通りエラーで気付き、専用フィールドへ誘導されたい。**なぜなら**、誤用が黙って Slack に渡るより、明確なエラーで指摘される方が安全なため。
6. **AI アシスタント（ツール利用者）として**、`query` の説明文を読むだけで「`-` 付きなら除外として書ける」と判断したい。**なぜなら**、ツールの description が一次情報源であり、ここに書かれていない仕様は使われないため。

## Acceptance Criteria (Gherkin)

### 主要 3 本

**AC-1: `search_messages` で除外 modifier 単体が通る**
```gherkin
Given query パラメータに "-in:#general" だけ指定
When search_messages を呼ぶ
Then validation エラーにならない
And Slack API に渡されるクエリ文字列に "-in:#general" が含まれる
```

**AC-2: 専用フィールドと query 内除外を併用できる**
```gherkin
Given in_channel="alpha" と query="エラー -in:#beta" を同時に指定
When search_messages を呼ぶ
Then validation エラーにならない
And Slack API に渡されるクエリ文字列に "エラー", "-in:#beta", "in:alpha" が全て含まれる
```

**AC-3: `search_files` で除外 modifier が通る**
```gherkin
Given query パラメータに "報告書 -type:pdf" を指定
When search_files を呼ぶ
Then validation エラーにならない
And Slack API に渡されるクエリ文字列に "報告書", "-type:pdf" が含まれる
```

### 例外 2 本

**AC-4: 包含 modifier は引き続きブロック**
```gherkin
Given query パラメータに "from:U123" (除外プレフィックスなし) を指定
When search_messages を呼ぶ
Then validation エラーが返る
And エラーメッセージに「除外 (`-` プレフィックス付き) であれば query に書ける」旨の案内が含まれる
```

**AC-5: 単語の途中の `-` は除外指定とみなさない**
```gherkin
Given query パラメータに "foo-in:#a" を指定 (直前が空白でも行頭でもない)
When search_messages を呼ぶ
Then validation エラーが返る (無印 modifier 扱い)
```

## NFR

- **性能**: validation の正規表現を「除外プレフィックス考慮版」に差し替える程度。性能要件への影響なし。
- **コスト**: 追加の API コールやストレージなし。
- **セキュリティ**: `query` 文字列は従来から Slack API にそのまま渡している。`-` 付き modifier は Slack 側のサポート範囲であり、新たな injection 経路は発生しない（既存と同等）。
- **監査**: 特になし。
- **運用 / ドキュメント**:
  - 各ツールの `mcp.WithString("query", mcp.Description(...))` の文言を更新する。
  - エラーメッセージは、`-` 付きなら通る旨をユーザーが推測できる文面に整える。
  - `README.md` / `README_ja.md` で `query` の仕様に触れている箇所があれば追従する（更新時は README_ja.md を先に更新するプロジェクト規約に従う）。
- **互換性**:
  - 既存の「無印 modifier はエラー」テストケースは引き続き通る必要がある。
  - 専用フィールド経由の既存挙動は変更しない。
  - 単純な単語除外（例: `error -warning`）の現状挙動も変更しない。
