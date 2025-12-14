# freee-api-go

[![CI](https://github.com/u-masato/freee-api-go/actions/workflows/ci.yml/badge.svg)](https://github.com/u-masato/freee-api-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/u-masato/freee-api-go.svg)](https://pkg.go.dev/github.com/u-masato/freee-api-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/u-masato/freee-api-go)](https://goreportcard.com/report/github.com/u-masato/freee-api-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

freee（フリー株式会社）が提供する会計API用のGo言語クライアントライブラリ。

## 特徴

- 🔐 **OAuth2認証**: freee APIのOAuth2フローを完全サポート
- 🛡️ **型安全**: OpenAPIスキーマから生成された型安全なクライアント
- ⚡ **自動リトライ**: レート制限・エラー時の自動リトライ機能
- 📄 **ページング**: 大量データの取得を透過的に処理
- 🧪 **テスト容易**: モックサーバーによるテスト支援
- 📚 **充実したドキュメント**: GoDocとサンプルコード完備

## インストール

```bash
go get github.com/u-masato/freee-api-go
```

**必要要件**: Go 1.21以上

## クイックスタート

### 1. OAuth2認証

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/u-masato/freee-api-go/auth"
)

func main() {
    config := auth.NewConfig(
        "YOUR_CLIENT_ID",
        "YOUR_CLIENT_SECRET",
        "http://localhost:8080/callback",
        []string{"read", "write"},
    )

    // 認可URL生成
    authURL := config.AuthCodeURL("random-state-string")
    fmt.Printf("Visit this URL to authorize: %s\n", authURL)

    // ユーザーが認可後、コールバックでcodeを取得
    code := "AUTHORIZATION_CODE_FROM_CALLBACK"

    // アクセストークン取得
    ctx := context.Background()
    token, err := config.Exchange(ctx, code)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Access Token: %s\n", token.AccessToken)
}
```

### 2. 会計APIの利用

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
    // OAuth2トークンソース
    token := &oauth2.Token{
        AccessToken: "YOUR_ACCESS_TOKEN",
    }
    tokenSource := oauth2.StaticTokenSource(token)

    // クライアント作成
    c, err := client.NewClient(
        client.WithTokenSource(tokenSource),
    )
    if err != nil {
        log.Fatal(err)
    }

    // 会計クライアント
    ac := accounting.NewClient(c)

    // 取引一覧取得
    ctx := context.Background()
    deals, err := ac.Deals.List(ctx, &accounting.DealsListOptions{
        CompanyID: 123456,
        Limit:     100,
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, deal := range deals {
        fmt.Printf("Deal: %d - %s\n", deal.ID, deal.IssueDate)
    }
}
```

## 主要機能

### OAuth2認証

freee APIのOAuth2 Authorization Code Grantフローを完全サポート。

- 認可URL生成
- アクセストークン取得
- リフレッシュトークン自動更新
- TokenSource実装

詳細: [examples/oauth](examples/oauth)

### 会計API

会計APIの主要なリソースへのアクセスを提供。

- **取引（Deals）**: 作成、読取、更新、削除
- **仕訳（Journals）**: 振替伝票の管理
- **取引先（Partners）**: 取引先マスタ
- **その他**: 勘定科目、品目、部門など

### HTTP Transport

共通のHTTP処理を自動化:

- **レート制限**: freee API制限に準拠した自動制御
- **リトライ**: エラー時の指数バックオフ
- **ロギング**: 構造化ログ（機密情報マスキング）
- **タイムアウト**: コンテキストベースの制御

## アーキテクチャ

```
利用者コード
    ↓
Facade (accounting/*)
    ↓
Generated Client (internal/gen)
    ↓
Transport (http.Client)
    ↓
freee API
```

### パッケージ構成

- `client/` - メインクライアントと設定
- `auth/` - OAuth2認証
- `accounting/` - 会計API Facade
- `transport/` - HTTP共通処理
- `internal/gen/` - OpenAPI生成コード（非公開）
- `examples/` - サンプルコード

詳細: [PLAN.md](PLAN.md)

## サンプル

- [OAuth2認証](examples/oauth) - 認証フローの完全な例
- [基本的な使い方](examples/basic) - シンプルなAPI呼び出し
- [高度な使い方](examples/advanced) - ページング、エラーハンドリング

## 開発

### ビルド

```bash
# 依存関係のダウンロード
go mod download

# ビルド
go build ./...

# テスト
go test ./...

# Lint
golangci-lint run
```

### OpenAPIからのコード生成

```bash
# oapi-codegenのインストール
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# コード生成
go generate ./tools
```

## ドキュメント

- [計画書（PLAN.md）](PLAN.md) - プロジェクト全体の設計・方針
- [実装TODO（TODO.md）](TODO.md) - 実装タスク一覧
- [GoDoc](https://pkg.go.dev/github.com/u-masato/freee-api-go) - API リファレンス

## コントリビューション

コントリビューションを歓迎します。

1. このリポジトリをフォーク
2. フィーチャーブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

詳細: CONTRIBUTING.md（準備中）

## ライセンス

MIT License - 詳細は [LICENSE](LICENSE) を参照してください。

## クレジット

このプロジェクトは freee株式会社が提供する [freee API](https://developer.freee.co.jp/) を利用しています。

## 免責事項

本ライブラリは非公式のクライアントライブラリであり、freee株式会社とは関係ありません。
freee APIの利用には freee の利用規約が適用されます。

---

**開発状況**: 🚧 開発中（Phase 1完了、Phase 2以降実装予定）

最新の進捗: [TODO.md](TODO.md)
