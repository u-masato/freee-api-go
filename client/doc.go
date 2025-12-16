// Package client はfreee APIのメインクライアントインターフェースを提供します。
//
// # 概要
//
// clientパッケージはfreee-api-go SDKを使用するための主要なエントリーポイントです。
// HTTP通信、認証、設定管理を処理する柔軟で設定可能なClient型を提供します。
//
// # 基本的な使い方
//
// デフォルト設定でクライアントを作成：
//
//	import "github.com/u-masato/freee-api-go/client"
//
//	c := client.NewClient()
//
// # 認証
//
// 最も一般的なユースケースは、OAuth2認証を使用してクライアントを作成することです：
//
//	import (
//	    "github.com/u-masato/freee-api-go/auth"
//	    "github.com/u-masato/freee-api-go/client"
//	)
//
//	// OAuth2を設定
//	config := auth.NewConfig(clientID, clientSecret, redirectURL, scopes)
//
//	// 認可コードをトークンに交換
//	token, err := config.Exchange(ctx, code)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 自動更新のためのTokenSourceを作成
//	tokenSource := config.TokenSource(ctx, token)
//
//	// TokenSourceを使用してクライアントを作成
//	c := client.NewClient(client.WithTokenSource(tokenSource))
//
// # 高度な設定
//
// 追加オプションでクライアントをカスタマイズ：
//
//	import (
//	    "time"
//	    "github.com/u-masato/freee-api-go/client"
//	    "github.com/u-masato/freee-api-go/transport"
//	)
//
//	// レート制限とリトライを備えたカスタムトランスポートを作成
//	t := transport.NewTransport(
//	    transport.WithRateLimit(10, 5),        // 10リクエスト/秒、バースト5
//	    transport.WithRetry(3, time.Second),   // 3回リトライ、指数バックオフ
//	    transport.WithLogging(logger),         // 構造化ロギング
//	)
//
//	// カスタムトランスポートを使用してOAuth2クライアントを作成
//	httpClient := &http.Client{
//	    Transport: &oauth2.Transport{
//	        Source: tokenSource,
//	        Base:   t,
//	    },
//	}
//
//	// すべてのカスタマイズを使用してクライアントを作成
//	c := client.NewClient(
//	    client.WithHTTPClient(httpClient),
//	    client.WithUserAgent("my-app/1.0.0"),
//	)
//
// # リクエストの実行
//
// クライアントのDoメソッドを使用して認証済みリクエストを実行：
//
//	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL()+"/api/1/companies", nil)
//	if err != nil {
//	    return err
//	}
//
//	resp, err := c.Do(req)
//	if err != nil {
//	    return err
//	}
//	defer resp.Body.Close()
//
//	// レスポンスを処理...
//
// # エラー処理
//
// このパッケージはfreee APIエラー用の構造化されたエラー型を提供します：
//
//   - [FreeeError]: ステータスコードとメッセージを含むAPIエラー
//   - [ErrBadRequest]: 400 Bad Requestエラー
//   - [ErrUnauthorized]: 401 Unauthorizedエラー
//   - [ErrForbidden]: 403 Forbiddenエラー
//   - [ErrNotFound]: 404 Not Foundエラー
//   - [ErrTooManyRequests]: 429 Too Many Requestsエラー
//
// エラーチェックの例：
//
//	resp, err := c.Do(req)
//	if err != nil {
//	    if client.IsUnauthorizedError(err) {
//	        // トークン期限切れ、再認証
//	    }
//	    if client.IsTooManyRequestsError(err) {
//	        // レート制限、待機してリトライ
//	    }
//	    return err
//	}
//
// # アーキテクチャ
//
// clientパッケージは、より高レベルのパッケージ（accounting/など）が
// 構築する基盤レイヤーとして設計されています。以下を処理します：
//
//   - HTTPクライアントの設定とライフサイクル
//   - ベースURL管理
//   - OAuth2 TokenSource統合
//   - User-Agentヘッダー管理
//   - 自動ヘッダー注入によるリクエスト実行
//
// 高レベルのパッケージは、基盤となるAPIの複雑さを隠し、
// 型安全でGoらしいインターフェースを提供するユーザーフレンドリーなファサードを提供します。
//
// # 設計原則
//
// clientパッケージは以下の設計原則に従います：
//
//  1. 柔軟性: 簡単なカスタマイズのための関数オプションパターン
//  2. 適切なデフォルト: 最小限の設定で動作
//  3. 組み合わせ可能性: authおよびtransportパッケージとクリーンに統合
//  4. テスト容易性: モックが容易（テストサーバーにはWithBaseURLを使用）
//  5. 明示性: 隠れたマジックのない明確で文書化された動作
//
// # スレッドセーフティ
//
// Clientは複数のゴルーチンによる並行使用に対して安全です。
// 基盤となるhttp.Clientは並行リクエストを安全に処理します。
//
// # コンテキスト処理
//
// Clientはコンテキストのキャンセルとタイムアウトを尊重します。
// WithTokenSourceを使用する場合、WithContextに提供されたコンテキスト
// （またはデフォルトでcontext.Background()）がトークン更新操作に使用されます。
// 個々のリクエストには、http.NewRequestWithContextを使用して
// リクエスト固有のコンテキストを提供してください。
package client
