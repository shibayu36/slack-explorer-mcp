# UserRepository Expire/TTL 設計（2025-09-18）

## 目的とスコープ
- `UserRepository` のセッション単位キャッシュに有効期限（expire/TTL）を導入し、期限切れ時の再取得と不要エントリの掃除を行う。
- 既存の公開API（特に `FindByDisplayName`）のシグネチャは変更しない。
- 最小限の複雑さで導入し、将来の拡張（singleflight 等）を阻害しない。

## 採用方針（結論）
- 固定TTL（アクセスで延長しない）。
- バックグラウンドスイーパー（定期掃除 goroutine + `Close()`）。
- スタンピード対策は初期は導入しない（必要になれば追加）。
- TTL は構造体フィールドに持たず、パッケージレベルの定数として `user_repository.go` に定義する。
- テストのため `now func() time.Time` を差し替え可能にする。
- デフォルト TTL: 30分。スイープ間隔: 5分。

## データ構造の変更
```go
// 既存
type SessionCache struct {
    users []slack.User
}

// 変更後
type SessionCache struct {
    users     []slack.User
    fetchedAt time.Time
}

// 既存
type UserRepository struct {
    sessionCaches map[SessionID]*SessionCache
    mu            sync.RWMutex
}

// 変更後
type UserRepository struct {
    sessionCaches map[SessionID]*SessionCache
    mu            sync.RWMutex

    stopCh chan struct{}  // スイーパー停止用
    wg     sync.WaitGroup // スイーパー終了待ち

    // --- test helpers ---
    now func() time.Time // 既定は time.Now、テストで差し替え
}

// 定数（提案）
const (
    userRepositoryTTL       = 30 * time.Minute
    userCacheSweepInterval  = 5 * time.Minute
)
```

## 公開APIの変更
- なし（`FindByDisplayName` の引数・戻り値はそのまま）。
- 追加: `func (r *UserRepository) Close()`（スイーパー停止用）。
- `Handler` に `Close()` を追加し、内部で `userRepository.Close()` を呼ぶ。
- `main.go` で `handler := NewHandler(); defer handler.Close()` を追加。

## 動作仕様（読み取り時のTTL判定）
- `FindByDisplayName` 呼び出し時：
  1. `SessionID` をキーに `sessionCaches` から参照（`RLock`）。
  2. エントリが存在し、かつ `now() - fetchedAt <= ttl` ならキャッシュ使用。
  3. それ以外は `RUnlock` → Slack API (`GetUsers`) を呼ぶ（ロック外）。
  4. 取得成功後 `Lock` で `users` と `fetchedAt = now()` を更新。
  5. メモリ上の `users` を対象にマッチングして返す。
- TTL は陽に 0 に設定しない想定（常に 0 より大きい）。

## スイーパー仕様（バックグラウンド掃除）
- `NewUserRepository()` で `go r.sweeper()` を起動。`time.NewTicker(userCacheSweepInterval)` を使用。
- tick 毎に `Lock` し、`sessionCaches` を全走査して期限切れ (`now() - fetchedAt > ttl`) のセッションを削除。
- `Close()` で `stopCh` を閉じ、`wg.Wait()` でスイーパー終了を待つ。
- 多重 `Close()` は安全（`sync.Once` もしくは `select` で保護）。

## 同時実行設計
- 読み取りは `RLock`、更新は `Lock`。
- API呼び出しはロック外で行い、ブロッキング時間を短縮。
- TTL境界で複数ゴルーチンが同時に再取得を行う可能性は許容（初期はスタンピード対策なし）。
- 競合例：スイーパーが削除 → 次の読み取り時に再取得され整合性が回復。

## エラーハンドリング／レート制限
- APIエラー時はそのままエラーを返す（既存方針）。
- 期限切れで再取得中にエラーが出た場合：キャッシュ更新は行わずエラーを返す。
- レート制限が問題になったら TTL を延長 or スタンピード対策を導入（将来対応）。

## 設定とテスト容易性
- TTL は `user_repository.go` に定義（定数）。まずは外部設定にせずシンプルに維持。
- `now` をフィールド化し、テストでは `repo.now = func() time.Time { return fixed }` のように差し替え可能（テストは同一パッケージのため可）。
- スイーパーの統合テストは困難なため、スイープ条件判定をヘルパー関数に切り出し、単体テスト可能にすることを検討（任意）。

## 破綻シナリオと将来拡張
- スタンピード（TTL境界での同時再取得）が顕著：`x/sync/singleflight` をセッションキーで導入。
- メモリ増大（セッション増加・長寿命）：スイーパーに最終アクセス時刻と上限/LRUを追加、もしくはスイープ間隔短縮。
- 「より新鮮なデータ」が強く求められる：TTL短縮、手動リフレッシュAPIの追加（例：`Refresh(sessionID)`）。
- スイーパーのライフサイクル管理漏れ：`Handler.Close()` 経由で必ず `UserRepository.Close()` を呼ぶ運用徹底。

## 段階的実装手順（小さく進める）
1. 型拡張・定数追加：`SessionCache.fetchedAt`、`UserRepository.stopCh/wg/now`、定数2つ。
2. 読み取り時TTL判定：`expired()` ヘルパー導入、`FindByDisplayName` に適用。既存テストが通ることを確認。
3. スイーパー実装：`sweeper()` と `Close()` を追加、goroutine 起動と停止を組み込む。
4. ハンドラ/エントリ統合：`Handler.Close()` を追加し、`main.go` に `defer handler.Close()`。
5. テスト追加：TTL前後、TTLゼロ/負値、（可能なら）スイープ条件判定の単体テスト。

## 互換性と移行
- 既存シグネチャ・挙動は基本互換。TTL導入により長時間起動でのユーザー情報が自動更新されるようになる。
- 新規の `Close()` 呼び出しを追加（`main.go` で `defer` ）。

## 受け入れ条件（Acceptance Criteria）
- TTL内の2回目以降の検索で Slack API が追加で呼ばれない。
- TTL経過後の最初の検索で Slack API が呼ばれ、キャッシュが更新される。
- スイーパー起動中に期限切れエントリが一定時間内（スイープ間隔以内）に削除される。
- `handler.Close()` 実行でスイーパーが停止し、プロセス終了時に goroutine リークがない。

## 参考実装スケッチ（抜粋）
```go
func NewUserRepository() *UserRepository {
    r := &UserRepository{
        sessionCaches: make(map[SessionID]*SessionCache),
        stopCh:        make(chan struct{}),

        // test helpers
        now: time.Now,
    }
    r.wg.Add(1)
    go r.sweeper()
    return r
}

func (r *UserRepository) expired(t time.Time) bool {
    if userRepositoryTTL <= 0 {
        return false
    }
    return r.now().Sub(t) > userRepositoryTTL
}

func (r *UserRepository) Close() {
    // once/do 保護は実装時に検討
    select {
    case <-r.stopCh:
        // already closed
    default:
        close(r.stopCh)
    }
    r.wg.Wait()
}

func (r *UserRepository) sweeper() {
    defer r.wg.Done()
    ticker := time.NewTicker(userCacheSweepInterval)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            now := r.now()
            r.mu.Lock()
            for sid, entry := range r.sessionCaches {
                if entry != nil && now.Sub(entry.fetchedAt) > userRepositoryTTL {
                    delete(r.sessionCaches, sid)
                }
            }
            r.mu.Unlock()
        case <-r.stopCh:
            return
        }
    }
}
```

## 開放課題
- 将来的な singleflight 導入の閾値（APIレート/平均同時数）をどこで判断するか。
- スイープのメトリクス（削除数、所要時間）の観測基盤導入（必要なら）。
- ユーザー一覧サイズが大きい場合のメモリ影響評価（必要なら分割・圧縮を検討）。
```
