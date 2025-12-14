# 基本的な使用例

freee APIを使った基本的な操作のサンプルコード。

## 実行方法

1. OAuth2認証でアクセストークンを取得（`examples/oauth` 参照）
2. 環境変数を設定:

```bash
export FREEE_ACCESS_TOKEN="your-access-token"
export FREEE_COMPANY_ID="123456"
```

3. サンプル実行:

```bash
go run main.go
```

## 学べること

- freee APIクライアントの基本的な使い方
- 取引データの取得
- エラーハンドリング
- コンテキストの使い方
