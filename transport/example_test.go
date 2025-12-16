package transport_test

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/u-masato/freee-api-go/transport"
)

// ExampleNewTransport は基本的なトランスポート作成を示します。
func ExampleNewTransport() {
	// 複数の機能を持つトランスポートを作成
	t := transport.NewTransport(
		transport.WithRateLimit(10, 5),      // 10リクエスト/秒、バースト5
		transport.WithRetry(3, time.Second), // 3回リトライ、初期遅延1秒
		transport.WithUserAgent("my-app/1.0.0"),
	)

	// http.Clientで使用
	client := t.Client()
	fmt.Println("クライアント作成成功:", client != nil)
	// Output:
	// クライアント作成成功: true
}

// ExampleNewTransport_withLogging はロギング付きトランスポートを示します。
func ExampleNewTransport_withLogging() {
	// 構造化ロガーを作成
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// ロギング付きトランスポートを作成
	t := transport.NewTransport(
		transport.WithLogging(logger),
		transport.WithUserAgent("my-app/1.0.0"),
	)

	// リクエスト/レスポンスが自動的にログに記録される
	_ = t
	fmt.Println("ロギング付きトランスポート作成成功")
	// Output:
	// ロギング付きトランスポート作成成功
}

// ExampleNewRateLimitRoundTripper はレート制限の使い方を示します。
func ExampleNewRateLimitRoundTripper() {
	// 毎秒3リクエスト、バースト5を許可
	rt := transport.NewRateLimitRoundTripper(http.DefaultTransport, 3.0, 5)

	client := &http.Client{Transport: rt}
	fmt.Println("レート制限付きクライアント作成成功:", client != nil)
	// Output:
	// レート制限付きクライアント作成成功: true
}

// ExampleNewRetryRoundTripper はリトライロジックの使い方を示します。
func ExampleNewRetryRoundTripper() {
	// 最大3回リトライ、初期遅延500ミリ秒
	rt := transport.NewRetryRoundTripper(http.DefaultTransport, 3, 500*time.Millisecond)

	client := &http.Client{Transport: rt}
	fmt.Println("リトライ付きクライアント作成成功:", client != nil)
	// Output:
	// リトライ付きクライアント作成成功: true
}

// ExampleNewLoggingRoundTripper はロギングの使い方を示します。
func ExampleNewLoggingRoundTripper() {
	logger := slog.Default()

	// ロギングRoundTripperを作成
	rt := transport.NewLoggingRoundTripper(http.DefaultTransport, logger)

	client := &http.Client{Transport: rt}
	fmt.Println("ロギング付きクライアント作成成功:", client != nil)
	// Output:
	// ロギング付きクライアント作成成功: true
}

// ExampleNewUserAgentRoundTripper はUser-Agent管理を示します。
func ExampleNewUserAgentRoundTripper() {
	// カスタムUser-Agentを設定
	rt := transport.NewUserAgentRoundTripper(http.DefaultTransport, "my-freee-app/2.0.0")

	client := &http.Client{Transport: rt}
	fmt.Println("User-Agent付きクライアント作成成功:", client != nil)
	// Output:
	// User-Agent付きクライアント作成成功: true
}

// ExampleChainRoundTrippers は複数のRoundTripperをチェーンする方法を示します。
func ExampleChainRoundTrippers() {
	logger := slog.Default()

	// 複数のRoundTripperをチェーン
	// リクエストは左から右の順に処理される
	chain := transport.ChainRoundTrippers(
		transport.NewRateLimitRoundTripper(nil, 10.0, 5),
		transport.NewRetryRoundTripper(nil, 3, time.Second),
		transport.NewLoggingRoundTripper(nil, logger),
		http.DefaultTransport,
	)

	client := &http.Client{Transport: chain}
	fmt.Println("チェーン付きクライアント作成成功:", client != nil)
	// Output:
	// チェーン付きクライアント作成成功: true
}

// ExampleDefaultUserAgent はデフォルトUser-Agent文字列を示します。
func ExampleDefaultUserAgent() {
	ua := transport.DefaultUserAgent("1.0.0")
	fmt.Println("User-Agent:", ua)
	// Output:
	// User-Agent: freee-api-go/1.0.0 (+github.com/u-masato/freee-api-go)
}
