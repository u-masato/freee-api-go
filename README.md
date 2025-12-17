# freee-api-go

[![CI](https://github.com/u-masato/freee-api-go/actions/workflows/ci.yml/badge.svg)](https://github.com/u-masato/freee-api-go/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/u-masato/freee-api-go/branch/main/graph/badge.svg)](https://codecov.io/gh/u-masato/freee-api-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/u-masato/freee-api-go.svg)](https://pkg.go.dev/github.com/u-masato/freee-api-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/u-masato/freee-api-go)](https://goreportcard.com/report/github.com/u-masato/freee-api-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://golang.org/)

freee（フリー株式会社）が提供する会計API用のGo言語クライアントライブラリ（SDK）です。

**freee会計API**へ簡単・安全にアクセスするための機能を提供します。OAuth2認証、自動リトライ、レート制限、ページング処理などを透過的に処理し、開発者が本来のビジネスロジックに集中できるよう設計されています。

## 特徴

| 機能 | 説明 |
|------|------|
| **OAuth2認証** | freee APIのOAuth2フローを完全サポート。トークン取得・リフレッシュを自動化 |
| **型安全** | OpenAPIスキーマから生成された型安全なクライアント。コンパイル時にエラーを検出 |
| **自動リトライ** | レート制限（429）やサーバーエラー（5xx）時の指数バックオフリトライ |
| **レート制限** | freee APIのレート制限に準拠した自動制御。API呼び出しを最適化 |
| **透過的ページング** | 大量データの取得をIteratorパターンで簡単に処理 |
| **構造化ロギング** | slogベースのロギング。機密情報は自動マスキング |
| **テスト容易** | モックサーバーによるテスト支援。実際のAPIを呼び出さずにテスト可能 |

## 目次

- [インストール](#インストール)
- [クイックスタート](#クイックスタート)
- [認証フロー](#認証フロー)
- [使用例](#使用例)
  - [基本的な使い方](#基本的な使い方)
  - [エラーハンドリング](#エラーハンドリング)
  - [ページング処理](#ページング処理)
  - [カスタム設定](#カスタム設定)
- [アーキテクチャ](#アーキテクチャ)
- [トラブルシューティング](#トラブルシューティング)
- [FAQ](#faq)
- [開発](#開発)
- [ライセンス](#ライセンス)

## インストール

### 必要要件

- **Go 1.24以上**
- **freee開発者アカウント**（[freee開発者ポータル](https://developer.freee.co.jp/)で作成）

### インストール方法

```bash
go get github.com/u-masato/freee-api-go
```

### 依存関係

本ライブラリは以下の外部依存関係を持ちます：

- `golang.org/x/oauth2` - OAuth2認証
- `golang.org/x/time/rate` - レート制限
- `github.com/oapi-codegen/runtime` - OpenAPI生成コードのランタイム

## クイックスタート

### 1. freee開発者ポータルでアプリケーションを登録

1. [freee開発者ポータル](https://developer.freee.co.jp/)にログイン
2. 「アプリ管理」→「新しいアプリを作成」
3. リダイレクトURLに `http://localhost:8080/callback` を設定
4. 必要なスコープを選択（例：`read`, `write`）
5. Client IDとClient Secretを取得

### 2. 環境変数を設定

```bash
export FREEE_CLIENT_ID="your-client-id"
export FREEE_CLIENT_SECRET="your-client-secret"
```

### 3. 認証してAPIを呼び出す

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/u-masato/freee-api-go/accounting"
    "github.com/u-masato/freee-api-go/auth"
    "github.com/u-masato/freee-api-go/client"
)

func main() {
    ctx := context.Background()

    // OAuth2設定
    config := auth.NewConfig(
        os.Getenv("FREEE_CLIENT_ID"),
        os.Getenv("FREEE_CLIENT_SECRET"),
        "http://localhost:8080/callback",
        []string{"read", "write"},
    )

    // 保存済みトークンを読み込み（または新規取得）
    token, err := auth.LoadTokenFromFile("token.json")
    if err != nil {
        // トークンがない場合はOAuth2フローを実行
        // 詳細は examples/oauth を参照
        log.Fatal("Token not found. Run OAuth2 flow first.")
    }

    // TokenSourceを作成（自動リフレッシュ対応）
    tokenSource := config.TokenSource(ctx, token)

    // クライアントを作成
    c := client.NewClient(
        client.WithTokenSource(tokenSource),
    )

    // 会計クライアントを作成
    ac := accounting.NewClient(c)

    // 取引一覧を取得
    result, err := ac.Deals.List(ctx, 123456, nil) // 123456 = 会社ID
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("取引数: %d\n", result.TotalCount)
    for _, deal := range result.Deals {
        fmt.Printf("  取引ID: %d, 発生日: %s\n", deal.Id, deal.IssueDate)
    }
}
```

## 認証フロー

freee APIはOAuth 2.0 Authorization Code Grantフローを使用します。

### フロー図

```
┌──────────┐     1. 認可URL生成      ┌──────────────┐
│          │ ───────────────────────>│              │
│  あなたの │     2. ブラウザで認可    │    freee     │
│  アプリ   │<───────────────────────│   認可サーバー │
│          │     3. 認可コード        │              │
│          │ ───────────────────────>│              │
│          │     4. アクセストークン   │              │
└──────────┘<───────────────────────└──────────────┘
      │                                    │
      │     5. APIリクエスト               │
      │ ───────────────────────────────────┘
      │     (Authorization: Bearer token)
      │
      ▼
┌──────────────┐
│   freee      │
│   API        │
└──────────────┘
```

### ステップ詳細

#### Step 1: 認可URL生成

```go
config := auth.NewConfig(clientID, clientSecret, redirectURL, scopes)

// CSRF保護のためのstateパラメータを生成
state := generateRandomState()

// 認可URLを生成
authURL := config.AuthCodeURL(state)
fmt.Println("このURLをブラウザで開いてください:", authURL)
```

#### Step 2-3: ユーザー認可とコールバック

ユーザーがfreeeにログインし、アプリケーションを認可すると、リダイレクトURLに認可コードが付与されます。

```go
// コールバックを受け取るHTTPサーバーを起動
http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
    // stateパラメータを検証（CSRF保護）
    if r.URL.Query().Get("state") != expectedState {
        http.Error(w, "Invalid state", http.StatusBadRequest)
        return
    }

    // 認可コードを取得
    code := r.URL.Query().Get("code")
    // ...
})
```

#### Step 4: アクセストークン取得

```go
// 認可コードをアクセストークンに交換
token, err := config.Exchange(ctx, code)
if err != nil {
    log.Fatal(err)
}

// トークンを保存（次回起動時に再利用）
err = auth.SaveTokenToFile(token, "token.json")
```

#### Step 5: APIリクエスト

```go
// TokenSourceを使用するとトークンが自動的にリフレッシュされます
tokenSource := config.TokenSource(ctx, token)
c := client.NewClient(client.WithTokenSource(tokenSource))
```

完全な認証フローの実装例は [examples/oauth](examples/oauth) を参照してください。

## 使用例

### 基本的な使い方

#### 取引の取得

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/u-masato/freee-api-go/accounting"
    "github.com/u-masato/freee-api-go/client"
    "golang.org/x/oauth2"
)

func main() {
    ctx := context.Background()

    // 既存のトークンを使用
    token := &oauth2.Token{
        AccessToken:  "your-access-token",
        RefreshToken: "your-refresh-token",
        TokenType:    "Bearer",
    }
    tokenSource := oauth2.StaticTokenSource(token)

    // クライアント作成
    c := client.NewClient(client.WithTokenSource(tokenSource))
    ac := accounting.NewClient(c)

    // 取引一覧を取得（フィルタなし）
    result, err := ac.Deals.List(ctx, 123456, nil)
    if err != nil {
        log.Fatal(err)
    }

    for _, deal := range result.Deals {
        fmt.Printf("ID: %d, 日付: %s, 金額: %d\n",
            deal.Id, deal.IssueDate, deal.Amount)
    }
}
```

#### 取引の作成

```go
// 新しい取引を作成
params := gen.DealCreateParams{
    CompanyId: 123456,
    IssueDate: "2024-01-15",
    Type:      gen.DealCreateParamsTypeExpense, // 支出
    Details: []gen.DealCreateParamsDetails{
        {
            AccountItemId: 12345, // 勘定科目ID
            TaxCode:       108,   // 税区分
            Amount:        10000, // 金額
        },
    },
}

deal, err := ac.Deals.Create(ctx, params)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("作成された取引ID: %d\n", deal.Deal.Id)
```

#### 条件付き検索

```go
// 支出のみ、特定期間でフィルタ
startDate := "2024-01-01"
endDate := "2024-01-31"
dealType := "expense"
limit := int64(50)

opts := &accounting.ListDealsOptions{
    Type:           &dealType,
    StartIssueDate: &startDate,
    EndIssueDate:   &endDate,
    Limit:          &limit,
}

result, err := ac.Deals.List(ctx, companyID, opts)
```

### エラーハンドリング

本ライブラリは構造化されたエラー型を提供します。

```go
import (
    "errors"
    "github.com/u-masato/freee-api-go/client"
)

result, err := ac.Deals.List(ctx, companyID, nil)
if err != nil {
    // 具体的なエラー種別をチェック
    switch {
    case client.IsUnauthorizedError(err):
        // トークンが無効または期限切れ
        log.Println("認証エラー: トークンを再取得してください")

    case client.IsTooManyRequestsError(err):
        // レート制限に達した
        log.Println("レート制限: しばらく待ってから再試行してください")

    case client.IsNotFoundError(err):
        // リソースが見つからない
        log.Println("指定されたリソースが見つかりません")

    case client.IsBadRequestError(err):
        // リクエストパラメータが不正
        var freeeErr *client.FreeeError
        if errors.As(err, &freeeErr) {
            // 詳細なエラーメッセージを取得
            for _, msg := range freeeErr.GetMessages() {
                log.Printf("エラー: %s\n", msg)
            }
        }

    default:
        log.Printf("予期しないエラー: %v\n", err)
    }
    return
}
```

#### バリデーションエラーの処理

```go
var freeeErr *client.FreeeError
if errors.As(err, &freeeErr) && freeeErr.HasValidationError() {
    fmt.Println("バリデーションエラー:")
    for _, errDetail := range freeeErr.Errors {
        if errDetail.Type == client.ErrorTypeValidation {
            for _, msg := range errDetail.Messages {
                fmt.Printf("  - %s\n", msg)
            }
        }
    }
}
```

### ページング処理

大量のデータを取得する場合、Iteratorパターンを使用して透過的にページングを処理できます。

#### Iteratorを使用したページング

```go
// 全ての支出取引を取得（自動ページング）
dealType := "expense"
opts := &accounting.ListDealsOptions{
    Type: &dealType,
}

iter := ac.Deals.ListIter(ctx, companyID, opts)
for iter.Next() {
    deal := iter.Value()
    fmt.Printf("取引ID: %d, 金額: %d\n", deal.Id, deal.Amount)
}

// エラーチェック（必須）
if err := iter.Err(); err != nil {
    log.Fatal(err)
}
```

#### 手動ページング

```go
offset := int64(0)
limit := int64(100)
totalProcessed := 0

for {
    opts := &accounting.ListDealsOptions{
        Offset: &offset,
        Limit:  &limit,
    }

    result, err := ac.Deals.List(ctx, companyID, opts)
    if err != nil {
        log.Fatal(err)
    }

    for _, deal := range result.Deals {
        // 処理...
        totalProcessed++
    }

    // 全件取得完了をチェック
    if int64(totalProcessed) >= result.TotalCount {
        break
    }

    offset += limit
}
```

### カスタム設定

#### レート制限とリトライの設定

```go
import (
    "time"
    "github.com/u-masato/freee-api-go/transport"
)

// カスタムTransportを作成
customTransport := transport.NewTransport(
    transport.WithRateLimit(5, 3),           // 5リクエスト/秒、バースト3
    transport.WithRetry(3, time.Second),     // 最大3回リトライ、初期待機1秒
    transport.WithUserAgent("my-app/1.0.0"), // カスタムUser-Agent
)

httpClient := &http.Client{
    Transport: customTransport,
    Timeout:   30 * time.Second,
}

c := client.NewClient(
    client.WithHTTPClient(httpClient),
    client.WithTokenSource(tokenSource),
)
```

#### ロギングの有効化

```go
import (
    "log/slog"
    "os"
)

// 構造化ロガーを作成
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

customTransport := transport.NewTransport(
    transport.WithLogging(logger),
    transport.WithRateLimit(10, 5),
)
```

ロギングでは機密情報（Authorization、Cookie、API-Key）が自動的にマスキングされます。

#### コンテキストの使用

```go
// タイムアウト付きコンテキスト
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// キャンセル可能なコンテキスト
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// 別のgoroutineからキャンセル可能
go func() {
    time.Sleep(5 * time.Second)
    cancel() // リクエストをキャンセル
}()

result, err := ac.Deals.List(ctx, companyID, nil)
```

## アーキテクチャ

本ライブラリは階層化されたアーキテクチャを採用しています：

```
┌─────────────────────────────────────────────────────────────┐
│                     利用者コード                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  Facade (accounting/*)                       │
│  - 使いやすいAPI                                             │
│  - ページング処理の隠蔽                                       │
│  - エラーハンドリングの統一                                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│               Generated Client (internal/gen)                │
│  - OpenAPIから自動生成                                        │
│  - 型安全なリクエスト/レスポンス                               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  Transport (transport/*)                     │
│  - レート制限                                                │
│  - 自動リトライ                                              │
│  - ロギング                                                  │
│  - User-Agent管理                                            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Auth (auth/*)                            │
│  - OAuth2認証                                                │
│  - トークン管理                                              │
│  - 自動リフレッシュ                                           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                       freee API                              │
└─────────────────────────────────────────────────────────────┘
```

### パッケージ構成

| パッケージ | 説明 |
|-----------|------|
| `client/` | メインクライアントと設定オプション |
| `auth/` | OAuth2認証、トークン管理 |
| `accounting/` | 会計APIのFacade（取引、仕訳、取引先など） |
| `transport/` | HTTP共通処理（リトライ、レート制限、ロギング） |
| `internal/gen/` | OpenAPI生成コード（非公開） |
| `examples/` | サンプルコード |

## トラブルシューティング

### 認証関連

#### 「unauthorized」エラーが発生する

**原因**: アクセストークンが無効または期限切れ

**解決方法**:
1. トークンのリフレッシュを試みる
2. それでも失敗する場合はOAuth2フローを再実行
3. freee開発者ポータルでアプリの設定を確認

```go
// TokenSourceを使用して自動リフレッシュ
tokenSource := config.TokenSource(ctx, token)
newToken, err := tokenSource.Token()
if err != nil {
    // リフレッシュ失敗：再認証が必要
    log.Println("トークンリフレッシュ失敗。再認証してください。")
}
```

#### 「redirect_uri_mismatch」エラー

**原因**: freee開発者ポータルに登録したリダイレクトURIとコード内のURLが一致しない

**解決方法**:
1. freee開発者ポータルで登録したリダイレクトURIを確認
2. コード内の `redirectURL` と完全に一致させる（末尾のスラッシュも含む）

### レート制限

#### 「too_many_requests」エラー（HTTP 429）

**原因**: freee APIのレート制限に達した

**解決方法**:
1. `transport.WithRateLimit()` でリクエスト頻度を調整
2. 指数バックオフリトライを有効化

```go
// レート制限を厳しく設定
transport := transport.NewTransport(
    transport.WithRateLimit(3, 1),       // 3リクエスト/秒、バースト1
    transport.WithRetry(5, 2*time.Second), // 最大5回リトライ
)
```

### 接続関連

#### タイムアウトエラー

**原因**: ネットワーク遅延またはAPIサーバーの応答遅延

**解決方法**:
```go
// タイムアウトを延長
httpClient := &http.Client{
    Timeout: 60 * time.Second,
}
c := client.NewClient(client.WithHTTPClient(httpClient))
```

#### 「context canceled」エラー

**原因**: コンテキストがキャンセルされた（タイムアウトまたは明示的キャンセル）

**解決方法**:
1. コンテキストのタイムアウト設定を確認
2. 長時間実行が予想される処理には十分なタイムアウトを設定

### データ関連

#### 「bad_request」エラー（HTTP 400）

**原因**: リクエストパラメータが不正

**解決方法**:
```go
var freeeErr *client.FreeeError
if errors.As(err, &freeeErr) {
    // 詳細なエラーメッセージを確認
    fmt.Println("エラー詳細:")
    for _, msg := range freeeErr.GetMessages() {
        fmt.Printf("  %s\n", msg)
    }
}
```

## FAQ

### 一般的な質問

#### Q: このライブラリは公式ですか？

**A**: いいえ、非公式のコミュニティ製クライアントライブラリです。freee株式会社とは関係ありません。

#### Q: どのfreee APIに対応していますか？

**A**: 現在、**freee会計API**のみ対応しています。請求書APIやHR APIは今後の対応予定です。

#### Q: Go以外の言語で使えますか？

**A**: このライブラリはGo専用です。他の言語については[freee公式SDK](https://developer.freee.co.jp/tutorials/sdk)をご確認ください。

### 認証

#### Q: トークンはどこに保存すべきですか？

**A**:
- **開発環境**: ファイルに保存（`auth.SaveTokenToFile()`）で十分です
- **本番環境**: 暗号化されたデータベース、シークレット管理サービス（AWS Secrets Manager、HashiCorp Vaultなど）を推奨

#### Q: リフレッシュトークンの有効期限は？

**A**: freee APIのリフレッシュトークンは無期限ですが、長期間使用しないと無効になる可能性があります。定期的な使用をお勧めします。

#### Q: 複数のfreeeアカウントに対応できますか？

**A**: はい。アカウントごとに別々のトークンを管理し、それぞれのTokenSourceを作成してください。

### パフォーマンス

#### Q: 大量のデータを効率的に取得するには？

**A**:
1. `ListIter()` を使用した自動ページング
2. 適切なフィルタ条件を設定して取得件数を減らす
3. 必要に応じて並行処理を実装

```go
// 並行処理の例（複数会社のデータ取得）
var wg sync.WaitGroup
for _, companyID := range companyIDs {
    wg.Add(1)
    go func(id int64) {
        defer wg.Done()
        result, _ := ac.Deals.List(ctx, id, nil)
        // 処理...
    }(companyID)
}
wg.Wait()
```

#### Q: 推奨されるレート制限設定は？

**A**: freee APIは通常、**1分あたり300リクエスト**程度の制限があります。安全のため、5リクエスト/秒程度を推奨します：

```go
transport.WithRateLimit(5, 3) // 5リクエスト/秒、バースト3
```

### エラー処理

#### Q: 一時的なエラーと永続的なエラーを区別するには？

**A**:
- **一時的（リトライ可能）**: 429 Too Many Requests、5xx Server Errors
- **永続的（修正が必要）**: 400 Bad Request、401 Unauthorized、403 Forbidden、404 Not Found

```go
// リトライ可能なエラーかチェック
if client.IsTooManyRequestsError(err) || client.IsInternalServerError(err) {
    // リトライ
}
```

## 開発

### ビルドとテスト

```bash
# 依存関係のダウンロード
go mod download

# ビルド
go build ./...

# テスト実行
go test ./...

# カバレッジレポート
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
go tool cover -html=coverage.txt -o coverage.html

# Lint
golangci-lint run
```

### Makefile

```bash
make help      # 利用可能なコマンド一覧
make build     # ビルド
make test      # テスト
make lint      # Lint
make coverage  # カバレッジレポート生成
make generate  # OpenAPIからコード生成
```

### コード生成

```bash
# OpenAPI仕様の更新（最新版をダウンロード）
make update-openapi

# コード生成
make generate
```

**OpenAPI仕様**:
- **ソース**: [freee/freee-api-schema](https://github.com/freee/freee-api-schema)
- **バージョン**: OpenAPI 3.0.1
- **ファイル**: `api/openapi.json` (約1.6MB)

## サンプルコード

より詳細なサンプルは `examples/` ディレクトリを参照してください：

- [OAuth2認証](examples/oauth) - 認証フローの完全な実装例
- [基本的な使い方](examples/basic) - シンプルなAPI呼び出し
- [高度な使い方](examples/advanced) - ページング、エラーハンドリング、並行処理

## ドキュメント

- [GoDoc](https://pkg.go.dev/github.com/u-masato/freee-api-go) - APIリファレンス
- [PLAN.md](PLAN.md) - プロジェクト設計書
- [TODO.md](TODO.md) - 実装進捗

## コントリビューション

コントリビューションを歓迎します！

1. このリポジトリをフォーク
2. フィーチャーブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

詳細は [CONTRIBUTING.md](CONTRIBUTING.md)（準備中）を参照してください。

## ライセンス

MIT License - 詳細は [LICENSE](LICENSE) を参照してください。

## クレジット

- [freee株式会社](https://www.freee.co.jp/) - freee APIの提供
- [freee開発者ポータル](https://developer.freee.co.jp/) - APIドキュメント

## 免責事項

本ライブラリは非公式のクライアントライブラリであり、freee株式会社とは関係ありません。freee APIの利用にはfreeeの利用規約が適用されます。

---

**開発状況**: Phase 5進行中（Facade API実装中）

最新の進捗は [TODO.md](TODO.md) を参照してください。
