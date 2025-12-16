// Package accounting はfreee会計APIのユーザーフレンドリーなファサードを提供します。
//
// このパッケージは生成されたAPIクライアントをラップし、freee会計APIとの対話のための
// 高レベルでGoらしいインターフェースを提供します。一般的な操作を簡素化し、
// ページネーションを透過的に処理します。
//
// # 概要
//
// accountingパッケージは以下を提供します：
//
//   - サービスベースの構成 ([DealsService], [JournalsService] など)
//   - イテレータによる自動ページネーション
//   - キャンセルとタイムアウトのためのコンテキスト伝播
//   - 簡素化されたエラー処理
//
// # クイックスタート
//
// 設定済みのベースクライアントからaccountingクライアントを作成：
//
//	import (
//	    "github.com/u-masato/freee-api-go/auth"
//	    "github.com/u-masato/freee-api-go/client"
//	    "github.com/u-masato/freee-api-go/accounting"
//	)
//
//	// OAuth2を設定してトークンを取得
//	config := auth.NewConfig(clientID, clientSecret, redirectURL, scopes)
//	token, _ := config.Exchange(ctx, code)
//	tokenSource := config.TokenSource(ctx, token)
//
//	// ベースクライアントを作成
//	baseClient := client.NewClient(client.WithTokenSource(tokenSource))
//
//	// accountingファサードを作成
//	accountingClient, err := accounting.NewClient(baseClient)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # サービス
//
// accountingクライアントは様々なサービスへのアクセスを提供します：
//
//   - [DealsService]: 取引の管理 - 収入と支出
//   - [JournalsService]: 仕訳の管理 - 会計エントリ
//   - [WalletTxnService]: 口座明細の管理
//   - [TransfersService]: 振替の管理
//   - [PartnersService]: 取引先の管理
//
// 使用例：
//
//	// サービス固有の操作にアクセス
//	deals := accountingClient.Deals()
//	journals := accountingClient.Journals()
//	partners := accountingClient.Partners()
//
// # イテレータによるページネーション
//
// このパッケージは透過的なページネーションのための [Iterator] インターフェースを提供します。
// offset/limitパラメータを手動で処理する代わりに、イテレータを使用：
//
//	// ページネーション結果のイテレータを作成
//	iter := NewPager(ctx, fetchFunc, 50)  // 1ページあたり50件
//
//	// すべての結果を反復処理
//	for iter.Next() {
//	    item := iter.Value()
//	    // アイテムを処理...
//	}
//	if err := iter.Err(); err != nil {
//	    log.Fatal(err)
//	}
//
// イテレータは必要に応じて自動的に追加ページを取得し、
// すべてのページネーションロジックを内部で処理します。
//
// # エラー処理
//
// accountingパッケージは基盤となるAPIクライアントからエラーを返します。
// clientパッケージのエラーチェック関数を使用：
//
//	import "github.com/u-masato/freee-api-go/client"
//
//	deals, err := dealsService.List(ctx, companyID, nil)
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
// # 高度な使用方法
//
// ファサードでまだ公開されていない操作については、
// 基盤となる生成クライアントにアクセスできます：
//
//	// 高度な操作のために生成クライアントにアクセス
//	genClient := accountingClient.GenClient()
//
//	// 生成クライアントを直接使用
//	resp, err := genClient.GetCompaniesWithResponse(ctx)
//
// 注意：生成クライアントAPIはバージョン間で変更される可能性があります。
// 利用可能な場合はファサードメソッドを優先的に使用してください。
//
// # スレッドセーフティ
//
// [Client] およびすべてのサービス型は並行使用に対して安全です。
// サービスインスタンスは遅延初期化されキャッシュされ、
// スレッドセーフを保ちながら効率的なメモリ使用を実現します。
//
// # コンテキストの使用
//
// API呼び出しを行うすべてのメソッドは [context.Context] パラメータを受け入れます。
// コンテキストを使用して：
//
//   - 操作のタイムアウトを設定
//   - 長時間実行リクエストをキャンセル
//   - リクエストスコープの値を渡す
//
// タイムアウトの例：
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	deals, err := dealsService.List(ctx, companyID, nil)
package accounting

import (
	"github.com/muno/freee-api-go/client"
	"github.com/muno/freee-api-go/internal/gen"
)

// Client is the main facade for the freee Accounting API.
//
// It provides access to service-specific clients (Deals, Journals, etc.)
// and manages the underlying generated API client.
//
// Example:
//
//	client := client.NewClient(
//	    client.WithTokenSource(tokenSource),
//	)
//	accountingClient := accounting.NewClient(client)
//
//	// Use service-specific clients
//	deals := accountingClient.Deals()
//	journals := accountingClient.Journals()
type Client struct {
	// client is the base freee API client
	client *client.Client

	// genClient is the generated OpenAPI client with response handling
	genClient *gen.ClientWithResponses

	// Service clients (lazy initialization)
	deals     *DealsService
	journals  *JournalsService
	walletTxn *WalletTxnService
	transfers *TransfersService
	partners  *PartnersService
}

// NewClient creates a new accounting facade client.
//
// The provided client.Client should be configured with appropriate
// authentication (OAuth2 token source) and transport settings.
//
// Example:
//
//	baseClient := client.NewClient(
//	    client.WithTokenSource(tokenSource),
//	)
//	accountingClient := accounting.NewClient(baseClient)
func NewClient(c *client.Client) (*Client, error) {
	// Create the generated client with response handling
	genClient, err := gen.NewClientWithResponses(
		c.BaseURL(),
		gen.WithHTTPClient(c.HTTPClient()),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		client:    c,
		genClient: genClient,
	}, nil
}

// Deals returns the DealsService for managing deals (取引).
//
// The service is lazily initialized on first access.
//
// Example:
//
//	deals := accountingClient.Deals()
//	list, err := deals.List(ctx, companyID, nil)
func (c *Client) Deals() *DealsService {
	if c.deals == nil {
		c.deals = &DealsService{
			client:    c.client,
			genClient: c.genClient,
		}
	}
	return c.deals
}

// Journals returns the JournalsService for managing journals (仕訳).
//
// The service is lazily initialized on first access.
//
// Example:
//
//	journals := accountingClient.Journals()
//	list, err := journals.List(ctx, companyID, nil)
func (c *Client) Journals() *JournalsService {
	if c.journals == nil {
		c.journals = &JournalsService{
			client:    c.client,
			genClient: c.genClient,
		}
	}
	return c.journals
}

// WalletTxns returns the WalletTxnService for managing wallet transactions (口座明細).
//
// The service is lazily initialized on first access.
//
// Example:
//
//	walletTxns := accountingClient.WalletTxns()
//	list, err := walletTxns.List(ctx, companyID, nil)
func (c *Client) WalletTxns() *WalletTxnService {
	if c.walletTxn == nil {
		c.walletTxn = &WalletTxnService{
			client:    c.client,
			genClient: c.genClient,
		}
	}
	return c.walletTxn
}

// Transfers returns the TransfersService for managing transfers (取引（振替）).
//
// The service is lazily initialized on first access.
//
// Example:
//
//	transfers := accountingClient.Transfers()
//	list, err := transfers.List(ctx, companyID, nil)
func (c *Client) Transfers() *TransfersService {
	if c.transfers == nil {
		c.transfers = &TransfersService{
			client:    c.client,
			genClient: c.genClient,
		}
	}
	return c.transfers
}

// Partners returns the PartnersService for managing partners (取引先).
//
// The service is lazily initialized on first access.
//
// Example:
//
//	partners := accountingClient.Partners()
//	list, err := partners.List(ctx, companyID, nil)
func (c *Client) Partners() *PartnersService {
	if c.partners == nil {
		c.partners = &PartnersService{
			client:    c.client,
			genClient: c.genClient,
		}
	}
	return c.partners
}

// BaseClient returns the underlying base client.
//
// This can be useful for advanced use cases where direct access
// to the base client is needed.
func (c *Client) BaseClient() *client.Client {
	return c.client
}

// GenClient returns the underlying generated API client with response handling.
//
// This is intended for advanced use cases or when the facade
// doesn't yet provide a specific operation. Use with caution
// as this exposes the internal generated API.
func (c *Client) GenClient() *gen.ClientWithResponses {
	return c.genClient
}
