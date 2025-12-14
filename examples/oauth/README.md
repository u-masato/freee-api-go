# OAuth2 認証サンプル

freee APIのOAuth2認証フローを実装したサンプルコード。

## 実行方法

1. freee開発者ポータルでアプリケーション登録
2. Client ID と Client Secret を取得
3. 環境変数を設定:

```bash
export FREEE_CLIENT_ID="your-client-id"
export FREEE_CLIENT_SECRET="your-client-secret"
export FREEE_REDIRECT_URL="http://localhost:8080/callback"
```

4. サンプル実行:

```bash
go run main.go
```

5. ブラウザで `http://localhost:8080` にアクセス
6. freeeにログインして認可
7. トークンが取得されてコンソールに表示されます

## 学べること

- OAuth2 Authorization Code Grant フロー
- 認可URL生成
- コールバックサーバーの実装
- アクセストークン取得
