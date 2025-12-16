package client_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/muno/freee-api-go/client"
	"github.com/muno/freee-api-go/transport"
	"golang.org/x/oauth2"
)

// ExampleNewClient は基本的なクライアント作成を示します。
func ExampleNewClient() {
	// デフォルト設定でクライアントを作成
	c := client.NewClient()

	fmt.Println("BaseURL:", c.BaseURL())
	fmt.Println("UserAgent:", c.UserAgent())
	// Output:
	// BaseURL: https://api.freee.co.jp
	// UserAgent: freee-api-go/dev (+github.com/u-masato/freee-api-go)
}

// ExampleNewClient_withOptions はオプション付きクライアント作成を示します。
func ExampleNewClient_withOptions() {
	// カスタムオプションでクライアントを作成
	c := client.NewClient(
		client.WithUserAgent("my-app/1.0.0"),
	)

	fmt.Println("UserAgent:", c.UserAgent())
	// Output:
	// UserAgent: my-app/1.0.0
}

// ExampleNewClient_withTransport はカスタムトランスポートの使い方を示します。
func ExampleNewClient_withTransport() {
	// レート制限とリトライを備えたトランスポートを作成
	t := transport.NewTransport(
		transport.WithRateLimit(10, 5),
		transport.WithRetry(3, time.Second),
	)

	// カスタムHTTPクライアントを作成
	httpClient := t.Client()

	// クライアントを作成
	c := client.NewClient(
		client.WithHTTPClient(httpClient),
		client.WithUserAgent("my-app/1.0.0"),
	)

	fmt.Println("クライアント作成成功:", c != nil)
	// Output:
	// クライアント作成成功: true
}

// ExampleNewClient_withTokenSource はTokenSourceの使い方を示します。
func ExampleNewClient_withTokenSource() {
	// テスト用の静的トークンソース
	token := &oauth2.Token{
		AccessToken: "test-access-token",
		Expiry:      time.Now().Add(1 * time.Hour),
	}
	tokenSource := oauth2.StaticTokenSource(token)

	// TokenSourceを使用してクライアントを作成
	c := client.NewClient(
		client.WithTokenSource(tokenSource),
	)

	fmt.Println("クライアント作成成功:", c != nil)
	// Output:
	// クライアント作成成功: true
}

// ExampleClient_Do はリクエスト実行を示します。
func ExampleClient_Do() {
	c := client.NewClient()

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL()+"/api/1/users/me", nil)
	if err != nil {
		fmt.Println("リクエスト作成エラー:", err)
		return
	}

	// リクエストを実行（認証なしの場合は401エラーが予想される）
	resp, err := c.Do(req)
	if err != nil {
		fmt.Println("リクエストエラー")
		return
	}
	defer resp.Body.Close()

	fmt.Println("ステータスコード:", resp.StatusCode)
}

// ExampleIsUnauthorizedError はエラーチェックの使い方を示します。
func ExampleIsUnauthorizedError() {
	// FreeeErrorを作成（通常はParseErrorResponseから取得）
	err := &client.FreeeError{
		StatusCode: 401,
		Message:    "認証が必要です",
	}

	// エラーチェック
	if client.IsUnauthorizedError(err) {
		fmt.Println("認証エラー: 再認証が必要です")
	}
	// Output:
	// 認証エラー: 再認証が必要です
}

// ExampleIsTooManyRequestsError はレート制限エラーチェックを示します。
func ExampleIsTooManyRequestsError() {
	err := &client.FreeeError{
		StatusCode: 429,
		Message:    "リクエストが多すぎます",
	}

	if client.IsTooManyRequestsError(err) {
		fmt.Println("レート制限エラー: しばらく待ってからリトライしてください")
	}
	// Output:
	// レート制限エラー: しばらく待ってからリトライしてください
}

// ExampleFreeeError_GetMessages はエラーメッセージの取得を示します。
func ExampleFreeeError_GetMessages() {
	err := &client.FreeeError{
		StatusCode: 400,
		Errors: []client.ErrorDetail{
			{
				Type:     client.ErrorTypeValidation,
				Messages: []string{"会社IDは必須です", "金額は0より大きい必要があります"},
			},
		},
	}

	messages := err.GetMessages()
	for _, msg := range messages {
		fmt.Println("エラー:", msg)
	}
	// Output:
	// エラー: 会社IDは必須です
	// エラー: 金額は0より大きい必要があります
}

// ExampleIsFreeeError はFreeeErrorの判定を示します。
func ExampleIsFreeeError() {
	var err error = &client.FreeeError{
		StatusCode: 404,
		Message:    "リソースが見つかりません",
	}

	if client.IsFreeeError(err) {
		fmt.Println("freee APIエラーです")
	}

	// 通常のエラーの場合
	normalErr := errors.New("通常のエラー")
	if !client.IsFreeeError(normalErr) {
		fmt.Println("freee APIエラーではありません")
	}
	// Output:
	// freee APIエラーです
	// freee APIエラーではありません
}
