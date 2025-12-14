# transport

HTTP通信の共通処理を提供するパッケージ。

## 責務

- カスタムRoundTripper実装
- レート制限（Rate Limiting）
- リトライロジック
- ロギング（構造化ログ）
- User-Agent付与
- その他横断的関心事

## 実装

複数のRoundTripperをチェーン化し、リクエスト/レスポンスを順次処理:

```
Request → RateLimit → Retry → Logging → UserAgent → HTTP
Response ← RateLimit ← Retry ← Logging ← UserAgent ← HTTP
```
