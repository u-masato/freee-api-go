# internal/testutil

テスト用のユーティリティ・ヘルパー関数。

## 責務

- モックHTTPサーバー
- テストフィクスチャ
- 共通テストヘルパー
- Golden fileパターン実装

## 使用例

```go
import "github.com/muno/freee-api-go/internal/testutil"

func TestSomething(t *testing.T) {
    server := testutil.NewMockServer(t)
    defer server.Close()

    // テストコード
}
```
