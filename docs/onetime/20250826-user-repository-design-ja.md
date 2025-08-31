# UserRepository 設計ドキュメント

## 概要
`search_users_by_name` ツールの実装に向けて、ユーザー情報の取得とキャッシュを管理する `UserRepository` を実装する。

## 設計方針
- シンプルに始める（YAGNI原則）
- 必要になったら拡張する
- 早期に動くものを作る

## 実装設計

### ファイル構成
```
user_repository.go  # UserRepository の実装
```

### 構造体定義
```go
type UserRepository struct {
    client *slack.Client
    cache  []slack.User
    cached bool
}
```

### メソッド設計

#### NewUserRepository
```go
func NewUserRepository(client *slack.Client) *UserRepository
```
- Slack クライアントを受け取って UserRepository を初期化

#### FindByDisplayName
```go
func (r *UserRepository) FindByDisplayName(ctx context.Context, displayName string) ([]slack.User, error)
```
- display_name で完全一致検索
- 初回呼び出し時に `users.list` でユーザー一覧を取得してキャッシュ
- 2回目以降はキャッシュから検索

#### loadUsers (private)
```go
func (r *UserRepository) loadUsers(ctx context.Context) error
```
- Slack API の `users.list` を呼び出し
- ページネーション対応で全ユーザーを取得
- 取得したユーザーを `cache` フィールドに保存
- `cached` フラグを true に設定

## キャッシュ戦略
- **有効期限なし**: プロセスが生きている限りキャッシュを保持
- **並列性考慮なし**: mutex なし（シングルスレッドアクセス前提）
- **更新なし**: キャッシュの更新機能は実装しない

## エラーハンドリング
- Slack API エラーはそのまま返す
- ユーザーが見つからない場合は空配列を返す（エラーではない）

## 実装の簡略化
- インターフェースは定義しない（必要になったら抽出）
- GetAllUsers は公開しない（内部実装のみ）
- 並列アクセス対策（mutex）は実装しない

## 将来の拡張ポイント
この設計が破綻する可能性があるケース：
1. **並列アクセスが必要になった場合**: mutex の追加が必要
2. **キャッシュ更新が必要になった場合**: 有効期限やリフレッシュ機能の追加
3. **メモリ使用量が問題になった場合**: 数万ユーザー規模でメモリ制約が出る可能性
4. **real_name 検索が必要になった場合**: FindByRealName メソッドの追加（既存設計で対応可能）

## 実装順序（段階的アプローチ）

### 実装方針
エンドツーエンドで最小動作版から始め、段階的に機能を追加していく。
各ステップを1コミットずつ作成し、常に動作する状態を保つ。

### コミット計画

#### コミット1: MCPツール定義とスタブHandler
- **ファイル**: `main.go`, `handler.go`
- **内容**:
  - main.go に `search_users_by_name` ツール定義を追加
  - handler.go に `SearchUsersByName` メソッドを追加（"not implemented" エラーを返す）
- **効果**: MCPツールとして認識され、呼び出し可能になる

#### コミット2: SlackClient に users.list 実装
- **ファイル**: `slack_client.go`
- **内容**:
  - `SlackClient` インターフェースに `ListUsers` メソッドを追加
  - 実装でページネーション対応（cursor を使った全ユーザー取得）
- **効果**: Slack API との通信部分が完成

#### コミット3: UserRepository 実装
- **ファイル**: `user_repository.go` (新規)
- **内容**:
  - `UserRepository` 構造体の実装
  - SlackClient を使った `loadUsers` 実装（private）
  - キャッシュ機能付き `FindByDisplayName` 実装
- **効果**: ユーザー検索ロジックが完成

#### コミット4: Handler統合
- **ファイル**: `handler.go`, `main.go`
- **内容**:
  - Handler に UserRepository フィールドを追加
  - `SearchUsersByName` を本実装に切り替え
  - main.go で Handler 初期化時に UserRepository を注入
- **効果**: 全機能が統合され、実際に動作する

### この方針のメリット
- 各コミットが独立して意味を持つ
- git bisect でのデバッグが容易
- コードレビューがしやすい
- 責務が明確に分離される
- 問題があった場合の切り分けが簡単