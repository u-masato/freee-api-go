// Package auth はfreee APIのOAuth2認証機能を提供します。
//
// このパッケージはfreee APIへのアクセスに必要なOAuth2 Authorization Code Grant
// フローを実装しています。ユーザー認証とアクセストークン管理の完全なソリューションを提供します。
//
// # 概要
//
// freee APIはOAuth2を使用して認証を行います。このパッケージは以下を提供します：
//
//   - OAuth2設定管理 ([Config])
//   - 認可URL生成
//   - アクセストークンの取得と更新
//   - トークンの検証と永続化
//   - 高度なユースケース向けのカスタムTokenSource
//
// # OAuth2フロー
//
// このパッケージを使用した典型的なOAuth2フロー：
//
//  1. アプリケーション認証情報でConfigを作成
//  2. 認可URLを生成してユーザーをリダイレクト
//  3. コールバックを処理してコードをトークンに交換
//  4. 自動トークン更新のためにTokenSourceを使用
//
// 使用例：
//
//	// ステップ1: OAuth2設定の作成
//	config := auth.NewConfig(
//	    "your-client-id",
//	    "your-client-secret",
//	    "http://localhost:8080/callback",
//	    []string{"read", "write"},
//	)
//
//	// ステップ2: 認可URLの生成
//	state := generateRandomState()  // CSRF保護を実装
//	authURL := config.AuthCodeURL(state)
//	// ユーザーをauthURLにリダイレクト...
//
//	// ステップ3: コールバックの処理（HTTPハンドラ内）
//	code := r.URL.Query().Get("code")
//	token, err := config.Exchange(ctx, code)
//	if err != nil {
//	    // エラー処理
//	}
//
//	// ステップ4: 自動更新のためのTokenSource作成
//	tokenSource := config.TokenSource(ctx, token)
//
// # トークン管理
//
// このパッケージはトークン管理のためのユーティリティを提供します：
//
//   - [IsTokenValid]: トークンが有効で期限切れでないかチェック
//   - [NeedsRefresh]: トークンの更新が必要かチェック
//   - [SaveTokenToFile]: トークンをJSONファイルに保存（0600権限）
//   - [LoadTokenFromFile]: JSONファイルからトークンを読み込み
//   - [GetTokenInfo]: トークンの状態に関する詳細情報を取得
//
// # TokenSource
//
// TokenSourceは自動トークン管理を提供します：
//
//   - [CachedTokenSource]: メモリキャッシュとオプションのファイル永続化
//   - [ReuseTokenSourceWithCallback]: トークン更新時にコールバックを呼び出し
//   - [StaticTokenSource]: 固定トークンを返す（テスト用）
//
// # エラー処理
//
// このパッケージは認証エラー用の構造化されたエラー型を提供します：
//
//   - [AuthError]: コンテキスト付きの認証エラーをラップ
//   - [ErrInvalidToken]: トークンが無効または期限切れ
//   - [ErrInvalidRefreshToken]: リフレッシュトークンが無効
//   - [ErrStateMismatch]: OAuth2 stateパラメータの不一致（CSRF）
//
// # セキュリティ考慮事項
//
// このパッケージを使用する際の注意点：
//
//   - 本番環境ではリダイレクトURLにHTTPSを使用
//   - stateパラメータを使用してCSRF保護を実装
//   - トークンを安全に保存（このパッケージは0600ファイル権限を使用）
//   - アクセストークンをログや出力に露出させない
//
// # freee APIエンドポイント
//
// このパッケージは以下のfreee OAuth2エンドポイントを使用します：
//
//   - 認可: https://accounts.secure.freee.co.jp/public_api/authorize
//   - トークン: https://accounts.secure.freee.co.jp/public_api/token
//
// テスト目的では、[NewConfigWithEndpoint] を使用してカスタムエンドポイント
// （例：モックOAuth2サーバー）を指定できます。
//
// # スレッドセーフティ
//
// このパッケージのすべてのTokenSourceは並行使用に対して安全です。
// [CachedTokenSource] と [ReuseTokenSourceWithCallback] は内部同期を使用して
// スレッドセーフなトークンアクセスと更新を保証します。
package auth

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
)

const (
	// AuthURL is the freee OAuth2 authorization endpoint.
	AuthURL = "https://accounts.secure.freee.co.jp/public_api/authorize"

	// TokenURL is the freee OAuth2 token endpoint.
	TokenURL = "https://accounts.secure.freee.co.jp/public_api/token"
)

// Config holds OAuth2 configuration for the freee API.
//
// Example usage:
//
//	config := auth.NewConfig(
//	    "your-client-id",
//	    "your-client-secret",
//	    "http://localhost:8080/callback",
//	    []string{"read", "write"},
//	)
type Config struct {
	oauth2Config *oauth2.Config
}

// NewConfig creates a new OAuth2 configuration for the freee API.
//
// Parameters:
//   - clientID: OAuth2 client ID obtained from freee developer portal
//   - clientSecret: OAuth2 client secret obtained from freee developer portal
//   - redirectURL: Callback URL registered in freee application settings
//   - scopes: List of permission scopes (e.g., []string{"read", "write"})
//
// Returns a Config instance that can be used to generate authorization URLs
// and exchange authorization codes for access tokens.
func NewConfig(clientID, clientSecret, redirectURL string, scopes []string) *Config {
	return &Config{
		oauth2Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  AuthURL,
				TokenURL: TokenURL,
			},
		},
	}
}

// AuthCodeURL generates the authorization URL for the OAuth2 flow.
//
// The state parameter should be a random string to prevent CSRF attacks.
// The user should be redirected to this URL to authorize the application.
//
// Example:
//
//	state := generateRandomState() // Your CSRF token generation
//	url := config.AuthCodeURL(state)
//	// Redirect user to url
func (c *Config) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return c.oauth2Config.AuthCodeURL(state, opts...)
}

// Exchange exchanges an authorization code for an access token.
//
// This should be called in the OAuth2 callback handler after the user
// has authorized the application. The code parameter is obtained from
// the callback URL query parameter.
//
// Example:
//
//	token, err := config.Exchange(ctx, code)
//	if err != nil {
//	    // Handle error
//	}
//	// Use token.AccessToken to make API requests
func (c *Config) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return c.oauth2Config.Exchange(ctx, code, opts...)
}

// TokenSource creates a TokenSource that automatically refreshes tokens.
//
// The returned TokenSource will automatically refresh the access token
// when it expires, using the refresh token if available.
//
// Example:
//
//	ts := config.TokenSource(ctx, token)
//	client := oauth2.NewClient(ctx, ts)
//	// Use client for HTTP requests
func (c *Config) TokenSource(ctx context.Context, token *oauth2.Token) oauth2.TokenSource {
	return c.oauth2Config.TokenSource(ctx, token)
}

// Client creates an HTTP client using the provided token.
//
// The returned client will automatically include the access token in
// the Authorization header and refresh it when necessary.
//
// Example:
//
//	httpClient := config.Client(ctx, token)
//	resp, err := httpClient.Get("https://api.freee.co.jp/...")
func (c *Config) Client(ctx context.Context, token *oauth2.Token) *http.Client {
	return c.oauth2Config.Client(ctx, token)
}

// NewConfigWithEndpoint creates a new OAuth2 configuration with custom endpoints.
//
// This is primarily useful for testing with a mock OAuth2 server.
//
// Parameters:
//   - clientID: OAuth2 client ID
//   - clientSecret: OAuth2 client secret
//   - redirectURL: Callback URL
//   - scopes: List of permission scopes
//   - authURL: Custom authorization endpoint URL
//   - tokenURL: Custom token endpoint URL
func NewConfigWithEndpoint(clientID, clientSecret, redirectURL string, scopes []string, authURL, tokenURL string) *Config {
	return &Config{
		oauth2Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  authURL,
				TokenURL: tokenURL,
			},
		},
	}
}
