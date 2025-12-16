# freee-api-go への貢献

freee-api-go への貢献に興味をお持ちいただきありがとうございます！このドキュメントでは、本プロジェクトへの貢献に関するガイドラインと手順を説明します。

## 目次

- [開発環境のセットアップ](#開発環境のセットアップ)
- [プルリクエストの作成手順](#プルリクエストの作成手順)
- [コーディング規約](#コーディング規約)
- [テスト要件](#テスト要件)
- [ドキュメント要件](#ドキュメント要件)

## 開発環境のセットアップ

### 前提条件

- Go 1.21 以降
- Git
- golangci-lint（リンター用）
- Make（オプション、Makefileコマンド使用時）

### セットアップ手順

1. **フォークとクローン**
   ```bash
   git clone https://github.com/YOUR_USERNAME/freee-api-go.git
   cd freee-api-go
   ```

2. **依存関係のインストール**
   ```bash
   go mod download
   ```

3. **開発ツールのインストール**
   ```bash
   # golangci-lint のインストール
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

   # oapi-codegen のインストール（コード生成用）
   go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
   ```

4. **セットアップの確認**
   ```bash
   # 全パッケージをビルド
   make build
   # または
   go build ./...

   # テストを実行
   make test
   # または
   go test ./...

   # リンターを実行
   make lint
   # または
   golangci-lint run
   ```

### 利用可能な Make コマンド

```bash
make help          # 利用可能なコマンドを表示
make build         # 全パッケージをビルド
make test          # 全テストを実行
make coverage      # テストカバレッジレポートを生成
make lint          # リンターを実行
make generate      # OpenAPI 仕様からコードを生成
make update-openapi # OpenAPI 仕様を更新
make clean         # ビルド成果物を削除
```

## プルリクエストの作成手順

### 作業開始前に

1. **既存の Issue を確認**: あなたのアイデアやバグが既に報告されていないか、既存の Issue を検索してください
2. **Issue を作成**: 新機能や大きな変更の場合は、まず Issue を作成してアプローチについて議論してください
3. **Issue を担当**: 作業中であることを他の人に知らせるため、Issue にコメントしてください

### プルリクエストの作成

1. **ブランチを作成**
   ```bash
   git checkout -b feature/your-feature-name
   # または
   git checkout -b fix/your-bug-fix
   ```

2. **変更を実装**
   - [コーディング規約](#コーディング規約)に従ってコードを書く
   - 新機能にはテストを追加
   - 必要に応じてドキュメントを更新

3. **テストとリンターを実行**
   ```bash
   # テストを実行
   make test

   # リンターを実行
   make lint

   # テストカバレッジを確認
   make coverage
   ```

4. **変更をコミット**
   ```bash
   git add .
   git commit -m "変更の簡潔な説明

   詳細な説明:
   - 変更点 1
   - 変更点 2

   Fixes #123"
   ```

5. **プッシュして PR を作成**
   ```bash
   git push origin feature/your-feature-name
   ```
   その後、GitHub でプルリクエストを作成してください。

### PR の要件

- 全てのテストがパスすること
- コードカバレッジが低下しないこと（目標: 80%以上）
- リンターがエラーなしでパスすること
- ユーザー向けの変更にはドキュメントの更新が必要
- PR の説明には変更内容とその目的を明確に記載

## コーディング規約

### Go スタイルガイド

公式の [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) と [Effective Go](https://go.dev/doc/effective_go) ガイドラインに従います。

**重要なポイント**:

- フォーマットには `gofmt` と `goimports` を使用
- Go の命名規則に従う:
  - `MixedCaps` または `mixedCaps` を使用（アンダースコアは使わない）
  - 頭字語は全て大文字にする（例: `APIClient`、`ApiClient` ではない）
- 関数は小さく、焦点を絞ったものにする
- 意味のある変数名を使用（短いスコープ以外では1文字の変数名は避ける）
- エクスポートされた関数と型にはコメントを追加

### パッケージ構成

```
freee-api-go/
├── auth/           # OAuth2 認証（公開）
├── transport/      # HTTP トランスポートミドルウェア（公開）
├── client/         # メインクライアント設定（公開）
├── accounting/     # 高レベル Facade（公開）
├── internal/       # 内部パッケージ（変更の可能性あり）
│   ├── gen/        # OpenAPI 生成コード
│   └── testutil/   # テストユーティリティ
├── examples/       # サンプルコード
└── tools/          # コード生成スクリプト
```

**ガイドライン**:
- 公開パッケージは API の安定性を維持する必要がある（セマンティックバージョニング）
- 変更の可能性がある実装の詳細には `internal/` を使用
- 生成されたコードは `internal/gen/` に配置し、リンターから除外

### エラーハンドリング

コンテキストを含むカスタムエラー型を使用:

```go
type MyError struct {
    Op  string  // 操作名
    Err error   // 元のエラー
}

func (e *MyError) Error() string {
    return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *MyError) Unwrap() error {
    return e.Err
}
```

エラーには常にコンテキストを提供:
```go
if err != nil {
    return fmt.Errorf("リクエストの処理に失敗しました: %w", err)
}
```

### Context の使用

以下の場合は常に `context.Context` を最初のパラメータとして受け取る:
- ネットワーク操作
- データベース操作
- 長時間実行される操作
- API 呼び出し

```go
func DoSomething(ctx context.Context, param string) error {
    // 実装
}
```

## テスト要件

### テストカバレッジ

- 目標: 80%以上のコードカバレッジ
- 全ての公開関数にテストが必要
- 成功ケースとエラーケースの両方をテスト

### テストパターン

**テーブル駆動テストとサブテストを使用**:

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "有効な入力",
            input: "test",
            want:  "result",
        },
        {
            name:    "無効な入力",
            input:   "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Feature(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Feature() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("Feature() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

**HTTP モッキングには httptest.Server を使用**:

```go
func TestHTTPClient(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // リクエストを検証
        // モックレスポンスを返す
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ok"}`))
    }))
    defer server.Close()

    // 実装をテスト
}
```

### テストの実行

```bash
# 全テストを実行
go test ./...

# 詳細出力で実行
go test -v ./...

# カバレッジ付きで実行
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# 特定のパッケージを実行
go test -v ./auth/...

# 単一のテストを実行
go test -v -run TestName ./package/
```

## ドキュメント要件

### コードドキュメント

**エクスポートされた関数と型にはコメントが必要**:

```go
// NewClient は指定されたオプションで新しい freee API クライアントを作成します。
// 設定が無効な場合はエラーを返します。
func NewClient(opts ...Option) (*Client, error) {
    // 実装
}
```

**パッケージドキュメント**（`doc.go` またはパッケージファイル内）:

```go
// Package auth は freee API 用の OAuth2 認証を提供します。
//
// このパッケージは PKCE を使用した OAuth2 認可コードフローを実装し、
// 自動リフレッシュ機能を持つトークン管理を提供します。
//
// 使用例:
//
//     config := auth.NewConfig(clientID, clientSecret, redirectURL, scopes)
//     authURL := config.AuthCodeURL(state)
//     // ... ユーザーを authURL にリダイレクト ...
//     token, err := config.Exchange(ctx, code)
//
package auth
```

### ユーザー向けドキュメント

ユーザー向けの変更には、以下を更新:

- **README.md**: 主要な機能や使用方法の変更
- **Examples**: `examples/` にコード例を追加または更新
- **CLAUDE.md**: アーキテクチャの変更時にプロジェクト手順を更新
- **TODO.md**: 完了したタスクをマークし、進捗を更新

### サンプルコード

全てのサンプルは以下の要件を満たす必要があります:
- 完全で実行可能であること
- エラーハンドリングを含むこと
- 重要なステップを説明するコメントがあること
- セットアップ手順を記載した README.md を含むこと

## Issue と PR テンプレート

Issue やプルリクエストを作成する際は、提供されているテンプレートを使用してください:

- バグ報告: `.github/ISSUE_TEMPLATE/bug_report.md`
- 機能要望: `.github/ISSUE_TEMPLATE/feature_request.md`
- プルリクエスト: `.github/PULL_REQUEST_TEMPLATE.md`

## 質問がありますか？

質問がある場合:

1. [README.md](README.md)、[PLAN.md](PLAN.md)、[TODO.md](TODO.md) を確認
2. 既存の Issue を検索
3. question ラベルを付けて新しい Issue を作成

## 行動規範

全てのやり取りにおいて、敬意を持ち建設的であってください。私たちは皆、一緒に有用なものを作るためにここにいます。

## ライセンス

freee-api-go への貢献により、あなたの貢献が MIT ライセンスの下でライセンスされることに同意したものとみなされます。
