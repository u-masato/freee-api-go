# 高度な使用例

freee APIの高度な機能を使ったサンプルコード。

このサンプルでは、本番環境向けの設定やエラーハンドリングのベストプラクティスを紹介しています。

## 前提条件

1. OAuth2認証でアクセストークンを取得済み（`examples/oauth` 参照）
2. トークンが `../oauth/token.json` に保存されている

## 環境変数の設定

```bash
export FREEE_CLIENT_ID="your-client-id"
export FREEE_CLIENT_SECRET="your-client-secret"
export FREEE_COMPANY_ID="123456"
```

**注意**: `FREEE_CLIENT_ID` と `FREEE_CLIENT_SECRET` はトークンの自動更新に必要です。設定しない場合、トークン期限切れ時に手動での再認証が必要になります。

## 実行方法

```bash
cd examples/advanced
go run main.go
```

## 学べること

### 1. カスタムトランスポート設定

本番環境向けのHTTPトランスポート設定方法を学びます。

```go
transport := transport.NewTransport(
    // レート制限: 3リクエスト/秒、バースト5
    transport.WithRateLimit(3, 5),

    // リトライ: 3回、指数バックオフ
    transport.WithRetry(3, time.Second),

    // 構造化ロギング
    transport.WithLogging(logger),

    // カスタムUser-Agent
    transport.WithUserAgent("my-app/1.0"),
)
```

**機能:**
- **レート制限**: freee APIのレート制限に対応し、429エラーを防ぐ
- **自動リトライ**: 5xx系エラーや一時的な障害から自動復旧
- **構造化ログ**: 本番環境でのデバッグとモニタリング用

### 2. ページング処理（Iteratorパターン）

大量データを効率的に取得するIteratorパターンを学びます。

```go
iter := client.Deals().ListIter(ctx, companyID, opts)
for iter.Next() {
    deal := iter.Value()
    // 処理...
}
if err := iter.Err(); err != nil {
    log.Fatal(err)
}
```

**利点:**
- ページ管理を自動化
- メモリ効率が良い（一度に1ページのみ保持）
- 途中で処理を中断可能

### 3. トークン自動更新

OAuth2トークンの自動更新設定を学びます。

```go
config := auth.NewConfig(clientID, clientSecret, redirectURL, scopes)
tokenSource := config.TokenSource(ctx, token)
// tokenSourceは自動的にトークンを更新
```

**利点:**
- トークン期限切れを意識せずにAPIを利用可能
- リフレッシュトークンを使った自動更新
- バックグラウンドでの透過的な更新

### 4. 構造化ログ

本番環境向けの構造化ログ設定を学びます。

```go
// テキスト形式（開発用）
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

// JSON形式（本番用）
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
```

**セキュリティ:**
- `Authorization` ヘッダーは自動的にマスクされる
- `Cookie`、`X-Api-Key` などの機密情報も保護

### 5. エラーハンドリング

構造化されたエラーハンドリングのベストプラクティスを学びます。

```go
result, err := client.Deals().List(ctx, companyID, opts)
if err != nil {
    // エラーの種類に応じた処理
    switch {
    case errors.Is(err, context.Canceled):
        // コンテキストキャンセル
    case errors.Is(err, context.DeadlineExceeded):
        // タイムアウト
    default:
        // その他のエラー
    }
}
```

### 6. コンテキストキャンセレーション

リクエストのタイムアウトやキャンセル処理を学びます。

```go
// タイムアウト付きコンテキスト
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// キャンセル可能なコンテキスト
ctx, cancel := context.WithCancel(context.Background())
// 必要に応じて cancel() を呼び出し
```

**用途:**
- リクエストタイムアウトの実装
- グレースフルシャットダウン
- ユーザー操作による処理中断

## 本番環境向け推奨設定

```go
// 本番環境用トランスポート
prodTransport := transport.NewTransport(
    transport.WithRateLimit(3, 5),     // freee APIのレート制限に対応
    transport.WithRetry(3, time.Second), // 一時的なエラーから回復
    transport.WithLogging(logger),      // 監視・デバッグ用
    transport.WithUserAgent("your-app/version"),
)

// OAuth2と組み合わせ
httpClient := &oauth2.Transport{
    Source: tokenSource,
    Base:   prodTransport,
}
```

## 出力例

```
=== freee API Advanced Example ===

1. Loading OAuth2 token...
   Token loaded (expires: 2025-12-14T15:30:00+09:00)

2. Creating custom transport...
   Transport configured with:
   - Rate limit: 3 requests/second, burst 5
   - Retry: 3 attempts with exponential backoff
   - Structured logging enabled

3. Setting up OAuth2 with auto-refresh...
   Token source created with auto-refresh capability

4. Creating freee API client...
   Client created successfully

5. Demonstrating advanced features...

   A. Fetching deals with pagination (Iterator pattern)...
      Processing deals with iterator...
      [1] Deal ID: 123456, Amount: ¥50000, Date: 2024-01-15
      [2] Deal ID: 123457, Amount: ¥30000, Date: 2024-01-14
      ...
      Summary: Processed 25 deals, total amount: ¥1250000

   B. Demonstrating structured error handling...
      Expected error occurred: ...

   C. Using multiple accounting services...
      Partners: Found 3 partners
      Account Items: Found 3 items
      Tags: Found 5 tags

   D. Demonstrating context cancellation...
      Expected cancellation error: context canceled

=== Advanced example completed ===
```

## 関連サンプル

- [../oauth](../oauth) - OAuth2認証フロー
- [../basic](../basic) - 基本的なAPI呼び出し
- [../iterator](../iterator) - Iteratorパターンの詳細

## 参考資料

- [freee API ドキュメント](https://developer.freee.co.jp/docs)
- [freee-api-go Transport パッケージ](../../transport/README.md)
- [freee-api-go Auth パッケージ](../../auth/README.md)
