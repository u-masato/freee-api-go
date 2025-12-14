# auth

OAuth2認証・認可機能を提供するパッケージ。

## 責務

- OAuth2 Authorization Code Grant フロー
- 認可URL生成
- アクセストークン取得
- リフレッシュトークン処理
- TokenSource実装

## 使用例

```go
import "github.com/muno/freee-api-go/auth"

config := auth.NewConfig(clientID, clientSecret, redirectURL, scopes)
authURL := config.AuthCodeURL("state")

// ユーザーが認可後、コールバックでcodeを取得
token, err := config.Exchange(ctx, code)
```
