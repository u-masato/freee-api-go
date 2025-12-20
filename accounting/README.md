# accounting

freee 会計API用のFacadeパッケージ。

## 責務

- 会計API操作の簡易化
- サービスごとのFacade提供
  - 取引（Deals）
  - 仕訳（Journals）
  - 事業所（Companies）
  - 口座（Walletables）
  - 取引先（Partners）
  - その他会計リソース
- ページング処理の隠蔽

## 使用例

```go
import "github.com/u-masato/freee-api-go/accounting"

ac := accounting.NewClient(baseClient)

// 取引一覧取得
deals, err := ac.Deals.List(ctx, &accounting.DealsListOptions{
    CompanyID: 123456,
})

// 事業所一覧取得
companies, err := ac.Companies().List(ctx)

// 口座一覧取得
walletables, err := ac.Walletables().List(ctx, 123456, nil)
```
