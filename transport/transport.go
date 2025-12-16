// Package transport はfreee APIクライアント用のHTTPトランスポートミドルウェアを提供します。
//
// このパッケージは、レート制限、リトライロジック、ロギング、User-Agent管理などの機能を
// HTTPリクエストに追加する、組み合わせ可能なRoundTripperラッパーを実装しています。
// ミドルウェアパターンに従い、機能を柔軟に階層化して組み合わせることができます。
//
// # 概要
//
// transportパッケージは以下を提供します：
//
//   - [RateLimitRoundTripper]: トークンバケットによるレート制限
//   - [RetryRoundTripper]: 指数バックオフによる自動リトライ
//   - [LoggingRoundTripper]: 構造化されたリクエスト/レスポンスロギング
//   - [UserAgentRoundTripper]: User-Agentヘッダー管理
//   - [Transport]: 関数オプションによる組み合わせ可能なトランスポート
//
// # クイックスタート
//
// 複数の機能を持つトランスポートを作成：
//
//	import (
//	    "log/slog"
//	    "time"
//	    "github.com/u-masato/freee-api-go/transport"
//	)
//
//	// レート制限、リトライ、ロギングを備えたトランスポートを作成
//	t := transport.NewTransport(
//	    transport.WithRateLimit(10, 5),        // 10リクエスト/秒、バースト5
//	    transport.WithRetry(3, time.Second),   // 3回リトライ、初期遅延1秒
//	    transport.WithLogging(slog.Default()), // 構造化ロギング
//	    transport.WithUserAgent("my-app/1.0"), // カスタムUser-Agent
//	)
//
//	// http.Clientで使用
//	client := &http.Client{Transport: t}
//
// # レート制限
//
// [RateLimitRoundTripper] はトークンバケットアルゴリズムを使用してリクエストを制限します：
//
//	// 毎秒10リクエスト、バースト5を許可
//	rt := transport.NewRateLimitRoundTripper(http.DefaultTransport, 10.0, 5)
//
// レートリミッターはコンテキストのキャンセルを尊重するため、
// レート制限トークンを待っている間にリクエストをキャンセルできます。
//
// # リトライロジック
//
// [RetryRoundTripper] は指数バックオフで失敗したリクエストを自動的にリトライします：
//
//	// 最大3回リトライ、初期遅延1秒
//	rt := transport.NewRetryRoundTripper(http.DefaultTransport, 3, time.Second)
//
// 以下のステータスコードでリトライが試行されます：
//   - 429 Too Many Requests
//   - 500 Internal Server Error
//   - 502 Bad Gateway
//   - 503 Service Unavailable
//   - 504 Gateway Timeout
//
// バックオフ遅延はリトライごとに2倍になり（1秒、2秒、4秒...）、最大30秒まで増加します。
//
// # ロギング
//
// [LoggingRoundTripper] は [log/slog] を使用して構造化ロギングを提供します：
//
//	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
//	rt := transport.NewLoggingRoundTripper(http.DefaultTransport, logger)
//
// ログに記録される情報：
//   - リクエストメソッド、URL、ヘッダー（機密データはマスク）
//   - レスポンスステータスコードと所要時間
//   - エラー（発生した場合）
//
// # 機密データの保護
//
// ロギングミドルウェアは自動的に機密ヘッダーをマスクします：
//   - Authorization
//   - Cookie / Set-Cookie
//   - X-Api-Key / Api-Key
//
// これらの値はログ出力で "[REDACTED]" に置き換えられます。
//
// # User-Agent管理
//
// [UserAgentRoundTripper] はすべてのリクエストにUser-Agentを含めることを保証します：
//
//	rt := transport.NewUserAgentRoundTripper(http.DefaultTransport, "my-app/1.0.0")
//
// リクエストに既にUser-Agentがある場合、カスタム値が追加されます。
//
// # OAuth2との統合
//
// 認証済みリクエストの場合、oauth2.Transportと組み合わせます：
//
//	import "golang.org/x/oauth2"
//
//	// レート制限とリトライを備えたトランスポートを作成
//	baseTransport := transport.NewTransport(
//	    transport.WithRateLimit(10, 5),
//	    transport.WithRetry(3, time.Second),
//	)
//
//	// 自動トークン処理のためにOAuth2でラップ
//	oauthTransport := &oauth2.Transport{
//	    Source: tokenSource,
//	    Base:   baseTransport,
//	}
//
//	client := &http.Client{Transport: oauthTransport}
//
// # SetBaseパターン
//
// 各RoundTripperは柔軟なチェーンのためにSetBaseメソッドを実装しています：
//
//	rt := transport.NewRetryRoundTripper(nil, 3, time.Second)
//	rt.SetBase(customTransport)  // 構築後にベースを設定
//
// このパターンは [ChainRoundTrippers] とオプション関数によって
// トランスポートチェーンを構築するために内部的に使用されます。
//
// # スレッドセーフティ
//
// このパッケージのすべてのRoundTripper実装は並行使用に対して安全です。
// [RateLimitRoundTripper] は内部で同期されるgolang.org/x/time/rateリミッターを使用します。
//
// # ベストプラクティス
//
// freee APIとの本番使用には、以下を推奨します：
//
//	t := transport.NewTransport(
//	    transport.WithRateLimit(3, 5),         // freee APIにはレート制限があります
//	    transport.WithRetry(3, time.Second),   // 一時的なエラーを処理
//	    transport.WithLogging(logger),         // デバッグ用
//	    transport.WithUserAgent("your-app/version"),
//	)
package transport

import (
	"net/http"
)

// Transport is a configurable HTTP transport for the freee API client.
type Transport struct {
	base http.RoundTripper
}

// NewTransport creates a new Transport with the given options.
// If no base transport is provided via options, http.DefaultTransport is used.
func NewTransport(opts ...Option) *Transport {
	t := &Transport{
		base: http.DefaultTransport,
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// RoundTrip implements the http.RoundTripper interface.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.base.RoundTrip(req)
}

// Client creates an http.Client using this transport.
func (t *Transport) Client() *http.Client {
	return &http.Client{
		Transport: t,
	}
}

// ChainRoundTrippers creates a RoundTripper that chains multiple RoundTrippers together.
// The RoundTrippers are applied in the order they are provided.
// The last RoundTripper in the chain should be the actual transport.
func ChainRoundTrippers(roundTrippers ...http.RoundTripper) http.RoundTripper {
	if len(roundTrippers) == 0 {
		return http.DefaultTransport
	}

	if len(roundTrippers) == 1 {
		return roundTrippers[0]
	}

	// Build chain from right to left
	result := roundTrippers[len(roundTrippers)-1]
	for i := len(roundTrippers) - 2; i >= 0; i-- {
		result = wrapRoundTripper(roundTrippers[i], result)
	}

	return result
}

// wrapRoundTripper wraps an outer RoundTripper around an inner one.
type roundTripperWrapper struct {
	outer http.RoundTripper
	inner http.RoundTripper
}

func wrapRoundTripper(outer, inner http.RoundTripper) http.RoundTripper {
	// If outer is a wrapper type that can accept a base, inject inner
	if w, ok := outer.(interface{ SetBase(http.RoundTripper) }); ok {
		w.SetBase(inner)
		return outer
	}
	// Otherwise, create a simple wrapper
	return &roundTripperWrapper{outer: outer, inner: inner}
}

func (w *roundTripperWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	// This is a fallback; in practice, each RoundTripper should handle its own wrapping
	return w.inner.RoundTrip(req)
}
