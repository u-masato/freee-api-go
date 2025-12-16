# client

公開APIを提供するパッケージ。

## 責務

- メインの `Client` 構造体
- クライアント設定オプション
- エラー型定義
- freee APIクライアントのエントリーポイント

## 使用例

```go
import "github.com/u-masato/freee-api-go/client"

c, err := client.NewClient(
    client.WithTokenSource(tokenSource),
)
```
