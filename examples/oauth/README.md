# OAuth2 認証サンプル

freee APIのOAuth2 Authorization Code Grant フローの完全な実装例です。

このサンプルでは、以下の機能を実装しています:
- OAuth2設定の作成
- 認可URL生成
- ローカルHTTPサーバーでのコールバック受信
- 認可コードとアクセストークンの交換
- トークンのファイル保存・読み込み
- トークンの自動リフレッシュ

## 前提条件

1. **freee開発者アカウント**: [freee開発者ポータル](https://developer.freee.co.jp/)でアカウントを作成
2. **アプリケーション登録**: 開発者ポータルでアプリケーションを登録
3. **認証情報の取得**: Client IDとClient Secretを取得

## セットアップ

### 1. アプリケーション登録

freee開発者ポータルでアプリケーションを登録する際、以下の設定を行ってください:

- **リダイレクトURI**: `http://localhost:8080/callback`
- **スコープ**: 必要な権限を選択（例: `read`, `write`）

### 2. 環境変数の設定

取得した認証情報を環境変数に設定します:

```bash
export FREEE_CLIENT_ID="your-client-id-here"
export FREEE_CLIENT_SECRET="your-client-secret-here"
```

**セキュリティ注意**: 本番環境では、認証情報を環境変数や設定ファイルで管理し、ソースコードにハードコードしないでください。

### 3. 依存関係のインストール

```bash
go mod download
```

## 使い方

### 基本的な実行

```bash
cd examples/oauth
go run main.go
```

実行すると、以下のような出力が表示されます:

```
No existing token found. Starting OAuth2 flow...

Visit this URL to authorize the application:

https://accounts.secure.freee.co.jp/public_api/authorize?client_id=...

Waiting for authorization...
```

### 認可フロー

1. ターミナルに表示されたURLをブラウザで開く
2. freeeにログイン（まだログインしていない場合）
3. アプリケーションへのアクセスを許可
4. 自動的にコールバックが実行され、トークンが取得される

成功すると、以下のような出力が表示されます:

```
✓ Authorization received

Exchanging authorization code for access token...
✓ Access token obtained successfully
  Token Type: Bearer
  Access Token: eyJ0eXAiOiJKV1QiLCJ...
  Expires: 2025-12-14T15:30:00+09:00
  Refresh Token: def50200a1b2c3d4e5...

✓ Token saved to token.json

You can now use this token to make API requests.
```

### トークンの再利用

一度取得したトークンは `token.json` に保存されます。次回実行時には、保存されたトークンを再利用します:

```bash
go run main.go
```

既存のトークンがある場合:

```
✓ Loaded existing token from token.json
  Valid: true
  Expires in: 3599s

✓ Token is still valid. You can use it to make API requests.
  Access Token: eyJ0eXAiOiJKV1QiLCJ...
```

### トークンの自動リフレッシュ

トークンが期限切れの場合、リフレッシュトークンを使用して自動的に更新されます:

```
✓ Loaded existing token from token.json
  Valid: false
  Expires in: -3600s

✗ Token has expired or is invalid.
  Attempting to refresh token...
  ✓ Token refreshed successfully
  ✓ Saved refreshed token to token.json
  Access Token: eyJ0eXAiOiJKV1QiLCJ...
```

## ファイル構成

```
examples/oauth/
├── main.go          # メインアプリケーション
├── README.md        # このファイル
└── token.json       # 保存されたトークン（実行後に生成）
```

## コードの説明

### 主要な機能

#### 1. OAuth2設定の作成

```go
config := auth.NewConfig(
    clientID,
    clientSecret,
    redirectURL,
    []string{"read", "write"},
)
```

#### 2. 認可URL生成

```go
state, _ := generateRandomState() // CSRF保護用のランダム文字列
authURL := config.AuthCodeURL(state)
```

#### 3. コールバックサーバー

```go
server := &http.Server{
    Addr: ":8080",
    Handler: callbackHandler(state, codeChan, errorChan),
}
```

#### 4. トークン交換

```go
token, err := config.Exchange(context.Background(), code)
```

#### 5. トークン保存

```go
err := auth.SaveTokenToFile(token, "token.json")
```

#### 6. トークン読み込み

```go
token, err := auth.LoadTokenFromFile("token.json")
```

#### 7. トークンリフレッシュ

```go
ts := config.TokenSource(ctx, token)
newToken, err := ts.Token() // 自動的にリフレッシュ
```

## セキュリティ考慮事項

### CSRF保護

認可リクエストには `state` パラメータを使用してCSRF攻撃を防ぎます:

```go
state, err := generateRandomState()
authURL := config.AuthCodeURL(state)
```

コールバックでは、受信した `state` が元の値と一致することを確認します。

### トークンの保護

- `token.json` は制限されたパーミッション（`0600`）で保存されます
- トークンファイルを `.gitignore` に追加してバージョン管理から除外してください
- 本番環境では、より安全なストレージ（暗号化されたデータベースなど）を使用してください

## トラブルシューティング

### 環境変数が設定されていない

```
FREEE_CLIENT_ID and FREEE_CLIENT_SECRET must be set
```

**解決方法**: 環境変数を正しく設定してください。

### リダイレクトURIの不一致

```
redirect_uri_mismatch
```

**解決方法**: freee開発者ポータルで登録したリダイレクトURIと、コード内のリダイレクトURLが一致していることを確認してください。

### ポート8080が使用中

```
listen tcp :8080: bind: address already in use
```

**解決方法**:
1. 別のプロセスがポート8080を使用していないか確認
2. コード内の `callbackPort` を別のポートに変更

### タイムアウト

```
Authorization timeout (5 minutes)
```

**解決方法**: 5分以内に認可を完了してください。必要に応じて、コード内のタイムアウト時間を延長できます。

## 次のステップ

このサンプルでトークンを取得した後は:

1. **API呼び出し**: 取得したトークンを使用してfreee APIにリクエストを送信
2. **基本サンプル**: [../basic](../basic) で基本的なAPI呼び出しの例を確認
3. **高度なサンプル**: [../advanced](../advanced) でページングやエラーハンドリングの例を確認

## 参考資料

- [freee API ドキュメント](https://developer.freee.co.jp/docs)
- [OAuth 2.0 仕様](https://oauth.net/2/)
- [golang.org/x/oauth2 パッケージ](https://pkg.go.dev/golang.org/x/oauth2)
- [freee-api-go ドキュメント](../../README.md)

## ライセンス

このサンプルコードは、プロジェクト全体と同じMITライセンスの下で提供されています。
