package auth_test

import (
	"context"
	"fmt"
	"time"

	"github.com/muno/freee-api-go/auth"
	"golang.org/x/oauth2"
)

// ExampleNewConfig は OAuth2設定の基本的な作成方法を示します。
func ExampleNewConfig() {
	// OAuth2設定を作成
	config := auth.NewConfig(
		"your-client-id",
		"your-client-secret",
		"http://localhost:8080/callback",
		[]string{"read", "write"},
	)

	// 認可URLを生成（stateはCSRF保護用のランダム文字列）
	state := "random-state-string"
	authURL := config.AuthCodeURL(state)

	fmt.Println("認可URLにリダイレクト:", authURL)
}

// ExampleConfig_Exchange は認可コードをトークンに交換する方法を示します。
func ExampleConfig_Exchange() {
	config := auth.NewConfig(
		"your-client-id",
		"your-client-secret",
		"http://localhost:8080/callback",
		[]string{"read", "write"},
	)

	ctx := context.Background()
	code := "authorization-code-from-callback"

	// 認可コードをトークンに交換
	token, err := config.Exchange(ctx, code)
	if err != nil {
		fmt.Println("トークン交換エラー:", err)
		return
	}

	fmt.Println("アクセストークン取得成功")
	fmt.Println("有効期限:", token.Expiry)
}

// ExampleIsTokenValid はトークンの有効性チェックを示します。
func ExampleIsTokenValid() {
	// 有効なトークン
	validToken := &oauth2.Token{
		AccessToken: "valid-access-token",
		Expiry:      time.Now().Add(1 * time.Hour),
	}

	// 期限切れのトークン
	expiredToken := &oauth2.Token{
		AccessToken: "expired-access-token",
		Expiry:      time.Now().Add(-1 * time.Hour),
	}

	fmt.Println("有効なトークン:", auth.IsTokenValid(validToken))
	fmt.Println("期限切れトークン:", auth.IsTokenValid(expiredToken))
	// Output:
	// 有効なトークン: true
	// 期限切れトークン: false
}

// ExampleGetTokenInfo はトークン情報の取得方法を示します。
func ExampleGetTokenInfo() {
	token := &oauth2.Token{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		Expiry:       time.Now().Add(30 * time.Minute),
	}

	info := auth.GetTokenInfo(token)

	fmt.Println("有効:", info.Valid)
	fmt.Println("リフレッシュトークンあり:", info.HasRefreshToken)
	fmt.Println("更新が必要:", info.NeedsRefresh)
}

// ExampleNewCachedTokenSource はキャッシュ付きTokenSourceの使い方を示します。
func ExampleNewCachedTokenSource() {
	config := auth.NewConfig(
		"your-client-id",
		"your-client-secret",
		"http://localhost:8080/callback",
		[]string{"read", "write"},
	)

	ctx := context.Background()

	// 既存のトークンを読み込み（初回はnil）
	token, _ := auth.LoadTokenFromFile("token.json")

	// キャッシュ付きTokenSourceを作成
	// トークンが更新されると自動的にファイルに保存
	tokenSource := auth.NewCachedTokenSource(
		config.TokenSource(ctx, token),
		token,
		"token.json",
	)

	// トークンを取得（必要に応じて自動更新）
	newToken, err := tokenSource.Token()
	if err != nil {
		fmt.Println("トークン取得エラー:", err)
		return
	}

	fmt.Println("トークン取得成功:", newToken.AccessToken[:10]+"...")
}

// ExampleStaticTokenSource は固定トークンのTokenSourceを示します。
func ExampleStaticTokenSource() {
	// テストや長寿命トークン用のStaticTokenSource
	token := &oauth2.Token{
		AccessToken: "static-access-token",
	}

	tokenSource := auth.StaticTokenSource(token)

	// 常に同じトークンを返す
	t, _ := tokenSource.Token()
	fmt.Println("トークン:", t.AccessToken)
	// Output:
	// トークン: static-access-token
}
