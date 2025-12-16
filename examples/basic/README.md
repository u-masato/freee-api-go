# 基本的な使用例

freee APIを使った基本的な操作のサンプルコード。

このサンプルでは、freee-api-goの基本的な使い方を学ぶことができます。

## 前提条件

1. OAuth2認証でアクセストークンを取得済み（`examples/oauth` 参照）
2. freeeの事業所ID（FREEE_COMPANY_ID）を取得済み

## 環境変数の設定

```bash
export FREEE_ACCESS_TOKEN="your-access-token"
export FREEE_COMPANY_ID="123456"
```

**アクセストークンの取得方法:**

```bash
cd ../oauth
go run main.go
# 表示されたURLにアクセスして認証を完了
# token.json からアクセストークンを確認
```

## 実行方法

```bash
cd examples/basic
go run main.go
```

## 学べること

### 1. freee APIクライアントの作成

```go
// トークンソースを作成
token := &oauth2.Token{
    AccessToken: accessToken,
}
tokenSource := oauth2.StaticTokenSource(token)

// クライアントを作成
c := client.NewClient(
    client.WithTokenSource(tokenSource),
    client.WithContext(ctx),
)

// accountingクライアントを作成
accountingClient, err := accounting.NewClient(c)
```

### 2. 取引データの取得（List）

```go
// オプションで取得件数を制限
limit := int64(5)
opts := &accounting.ListDealsOptions{
    Limit: &limit,
}

// 取引一覧を取得
result, err := accountingClient.Deals().List(ctx, companyID, opts)
if err != nil {
    return fmt.Errorf("failed to list deals: %w", err)
}

// 結果を処理
for _, deal := range result.Deals {
    fmt.Printf("Deal ID: %d, Amount: %d\n", deal.Id, deal.Amount)
}
```

### 3. フィルタリング

```go
// 経費（expense）のみを取得
dealType := "expense"
opts := &accounting.ListDealsOptions{
    Type:  &dealType,
    Limit: &limit,
}

result, err := accountingClient.Deals().List(ctx, companyID, opts)
```

### 4. 単一データの取得（Get）

```go
// 特定の取引を取得
resp, err := accountingClient.Deals().Get(ctx, companyID, dealID, nil)
if err != nil {
    return fmt.Errorf("failed to get deal: %w", err)
}

deal := &resp.Deal
fmt.Printf("Deal ID: %d\n", deal.Id)
```

### 5. エラーハンドリング

```go
result, err := accountingClient.Deals().List(ctx, companyID, opts)
if err != nil {
    // エラーをログに出力
    log.Printf("Error listing deals: %v", err)
    return
}
```

### 6. コンテキストの使い方

```go
// タイムアウト付きコンテキストを作成
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()  // 必ずキャンセルを呼び出す

// コンテキストを使用してAPIを呼び出し
result, err := accountingClient.Deals().List(ctx, companyID, opts)
```

## 出力例

```
=== freee API Basic Example ===
1. Fetching recent deals...
   Found 5 deals (total: 150)

   Deal #1:
     ID: 123456
     Issue Date: 2024-01-15
     Type: expense
     Status: settled
     Partner ID: 789
     Amount: 50000

   Deal #2:
     ID: 123455
     Issue Date: 2024-01-14
     Type: income
     Status: settled
     Amount: 100000

   ...

2. Fetching expense deals only...
   Found 3 expense deals

   Expense Deal #1:
     ID: 123456
     Issue Date: 2024-01-15
     Partner ID: 789
     Amount: 50000

   ...

3. Fetching a specific deal...
   Deal Details:
     ID: 123456
     Issue Date: 2024-01-15
     Type: expense
     Status: settled
     Partner ID: 789
     Amount: 50000
     Due Date: 2024-01-31
     Details: 2 item(s)
       Detail #1:
         Account Item ID: 100
         Amount: 30000
         Entry Side: debit
         Tax Code: 21
       Detail #2:
         Account Item ID: 200
         Amount: 20000
         Entry Side: debit
         Tax Code: 21

=== Example completed successfully ===
```

## 次のステップ

このサンプルで基本を学んだら:

1. **高度なサンプル**: [../advanced](../advanced) でページング、リトライ、レート制限を学ぶ
2. **Iteratorパターン**: [../iterator](../iterator) で大量データの効率的な取得方法を学ぶ
3. **他のAPIエンドポイント**: `Partners()`, `AccountItems()`, `Tags()` など他のサービスを試す

## 関連サンプル

- [../oauth](../oauth) - OAuth2認証フロー
- [../advanced](../advanced) - 高度な機能（レート制限、リトライ、ログ）
- [../iterator](../iterator) - ページング処理（Iteratorパターン）

## 参考資料

- [freee API ドキュメント](https://developer.freee.co.jp/docs)
- [freee API リファレンス（取引）](https://developer.freee.co.jp/reference/accounting/reference#get-deals)
