# freee-api-go 実装TODO

本ドキュメントは、PLAN.mdで定義した実装フェーズを具体的なタスクレベルに落とし込んだものです。

## 📋 進行状況サマリー

| フェーズ | ステータス | 進捗 |
|---------|----------|------|
| Phase 1: プロジェクト基盤 | ✅ Completed | 7/7 |
| Phase 2: OAuth2認証 | ✅ Completed | 7/7 |
| Phase 3: HTTP Transport層 | ✅ Completed | 7/7 |
| Phase 4: Generated API Client | ✅ Completed | 7/7 |
| Phase 5: Accounting Facade | ✅ Completed | 8/8 |
| Phase 6: ドキュメント・サンプル | ✅ Completed | 6/6 |
| Phase 7: 拡張・改善 | 🔄 In Progress | 1/5 |

**凡例**: 🔲 未着手 | 🔄 進行中 | ✅ 完了

**最終更新**: 2025-12-17
**現在のフェーズ**: Phase 5, 6 完了 / Phase 7 進行中

---

## Phase 1: プロジェクト基盤（Foundation） ✅

**目標**: 開発環境・ビルド基盤の構築

**ステータス**: ✅ 完了（2025-12-14）

### 1.1 リポジトリ初期化 ✅

- [x] `go.mod` 初期化（`go mod init github.com/u-masato/freee-api-go`）
- [x] `.gitignore` 作成（Go標準 + IDE設定）
- [x] `LICENSE` 作成（MIT License）
- [x] `.editorconfig` 作成（コーディングスタイル統一）

**コミット**: `5fc95ca` - Initialize repository with foundational files

### 1.2 ディレクトリ構造作成 ✅

```bash
mkdir -p {client,auth,accounting,transport,internal/{gen,testutil},examples/{oauth,basic,advanced},tools,api}
```

- [x] 上記ディレクトリ構造を作成
- [x] 各ディレクトリに `README.md` を配置
- [x] パッケージ構成ドキュメントを各 README.md に記載

**コミット**: `4ec4e3a` - Create project directory structure with documentation

### 1.3 GitHub Actions CI/CD設定 ✅

- [x] `.github/workflows/ci.yml` 作成
  - Lint ジョブ（golangci-lint）
  - Test ジョブ（go test -race -coverprofile）
  - Build ジョブ（マルチOS: Linux, macOS, Windows）
- [x] `.github/workflows/release.yml` 作成（タグプッシュ時の自動リリース）
- [x] `.github/dependabot.yml` 作成（依存関係自動更新）

### 1.4 golangci-lint設定 ✅

- [x] `.golangci.yml` 作成
  - 有効化: gofmt, govet, staticcheck, errcheck, gosec, etc.
  - 除外設定: internal/gen/*（生成コード）
- [ ] ローカル実行確認（`golangci-lint run`） ※コードが無いため次フェーズで確認

### 1.5 OpenAPI仕様ファイル取得 ✅

- [x] freee公式リポジトリから会計API OpenAPI v3仕様をダウンロード
- [x] `api/openapi.json` として保存 (1.6MB)
- [x] バージョン情報を README.md に記載
- [x] `tools/update-openapi.sh` スクリプト作成（自動更新用）

**コミット**: 予定（Issue #8）
**ソース**: https://github.com/freee/freee-api-schema

### 1.6 oapi-codegen セットアップ ⏭️

- [ ] `go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest`
- [ ] `tools/generate.go` 作成（`//go:generate` ディレクティブ）
- [ ] `oapi-codegen.yaml` 設定ファイル作成
  - 出力先: `internal/gen/`
  - パッケージ名: `gen`
  - 生成オプション設定
- [ ] 初回生成実行（`go generate ./tools`）

**注**: Phase 4で実施予定（コード生成フェーズ）

### 1.7 README.md基本構造 ✅

- [x] プロジェクト概要
- [x] インストール方法
- [x] クイックスタート（簡易サンプル）
- [x] ドキュメントリンク
- [x] ライセンス表記
- [x] バッジ追加（CI Status, Go Version, License）

**コミット**: `68e9127` - Add CI/CD configuration and comprehensive README

### Phase 1 成果物

✅ **完了条件達成**: プロジェクト基盤が整い、Phase 2に進む準備完了

**作成ファイル**: 18ファイル
- 設定ファイル: 6個（go.mod, .gitignore, LICENSE, .editorconfig, .golangci.yml, dependabot.yml）
- CI/CD: 2個（ci.yml, release.yml）
- ドキュメント: 10個（README.md × 10）

**コミット数**: 3
**次のフェーズ**: Phase 2 - OAuth2認証

---

## Phase 2: OAuth2認証（Authentication） ✅

**目標**: freee OAuth2フロー実装

**ステータス**: ✅ 完了（2025-12-14）

### 2.1 auth/ パッケージ構造設計 ✅

- [x] `auth/config.go` 作成（OAuth2設定構造体）
  - ClientID, ClientSecret, RedirectURL, Scopes
- [x] `auth/token.go` 作成（トークン管理）
- [x] `auth/errors.go` 作成（認証エラー型）
- [x] `auth/tokensource.go` 作成（TokenSource拡張実装）

**コミット**: `eb04dfe` - Add OAuth2 authentication core files

### 2.2 認可URL生成機能 ✅

- [x] `auth.NewConfig()` 実装
- [x] `auth.Config.AuthCodeURL(state string)` 実装
  - oauth2.Config を利用
  - state パラメータ対応

**コミット**: `eb04dfe` - Add OAuth2 authentication core files

### 2.3 アクセストークン取得 ✅

- [x] `auth.Config.Exchange(ctx, code)` 実装
  - 認可コードからトークン取得
  - コンテキスト対応
  - エラーハンドリング

**コミット**: `eb04dfe` - Add OAuth2 authentication core files

### 2.4 リフレッシュトークン処理 ✅

- [x] `auth.Config.TokenSource(ctx, token)` 実装
  - トークン更新ロジック（oauth2パッケージ利用）
  - 有効期限チェック

**コミット**: `eb04dfe` - Add OAuth2 authentication core files

### 2.5 TokenSource実装 ✅

- [x] `CachedTokenSource` 実装（キャッシュ機能）
  - ファイル保存機能
  - メモリキャッシュ
- [x] `ReuseTokenSourceWithCallback` 実装
  - コールバック機能付きTokenSource
- [x] `oauth2.TokenSource` 互換

**コミット**: `2fed110` - Add OAuth2 TokenSource and comprehensive tests

### 2.6 ユニットテスト（モック） ✅

- [x] `auth/config_test.go` 作成
  - OAuth2設定テスト
  - 認可URLテスト
  - トークン交換テスト
- [x] `auth/auth_test.go` 作成
  - httptest.Server でモックOAuth2サーバー
  - トークン管理テスト
  - エラー処理テスト
  - 正常系・異常系テスト
- [x] カバレッジ確認（23テスト全て成功）

**コミット**: `2fed110` - Add OAuth2 TokenSource and comprehensive tests

### 2.7 examples/oauth/ サンプル作成 ✅

- [x] `examples/oauth/main.go` 作成
  - 認可URL生成
  - コールバックサーバー起動（ポート8080）
  - トークン取得・表示
  - CSRF保護（state パラメータ）
  - トークンの自動保存/読み込み
  - トークンリフレッシュ機能
- [x] `examples/oauth/README.md` 作成（使い方ガイド）
  - セットアップ手順
  - 使い方詳細
  - セキュリティ考慮事項
  - トラブルシューティング

**コミット**: `c10f030` - Add OAuth2 example application and documentation

### Phase 2 成果物

✅ **完了条件達成**: OAuth2フローが動作し、トークン取得・リフレッシュが可能

**作成ファイル**: 7ファイル
- コアファイル: 4個（config.go, errors.go, token.go, tokensource.go）
- テストファイル: 2個（config_test.go, auth_test.go）
- サンプル: 1個（examples/oauth/main.go）
- ドキュメント: 1個（examples/oauth/README.md更新）

**テスト**: 23テスト全て成功
**コミット数**: 3
**次のフェーズ**: Phase 3 - HTTP Transport層

---

## Phase 3: HTTP Transport層（Transport） ✅

**目標**: 共通HTTP処理の実装

**ステータス**: ✅ 完了（2025-12-14）

### 3.1 transport/ パッケージ設計 ✅

- [x] `transport/transport.go` 作成（基本構造）
- [x] `transport/options.go` 作成（設定オプション）
- [x] RoundTripperチェーン機能実装

**コミット**: `00ccadf` - Add HTTP Transport layer implementation

### 3.2 カスタムRoundTripper実装 ✅

- [x] `ChainRoundTrippers` 実装（複数RoundTripperチェーン）
- [x] ベースRoundTripper（http.DefaultTransport）
- [x] SetBase メソッドによる柔軟な構成

**コミット**: `00ccadf` - Add HTTP Transport layer implementation

### 3.3 レート制限（rate.Limiter統合） ✅

- [x] `transport/ratelimit.go` 作成
- [x] `RateLimitRoundTripper` 実装
  - `golang.org/x/time/rate` 利用
  - リクエスト前にWait
  - コンテキストキャンセル対応
- [x] レート制限テスト作成（4テスト成功）

**コミット**: `00ccadf`, `294ec64` - Add HTTP Transport layer + dependency

### 3.4 リトライロジック ✅

- [x] `transport/retry.go` 作成
- [x] `RetryRoundTripper` 実装
  - エクスポネンシャルバックオフ
  - リトライ条件設定（5xx, 429）
  - 最大リトライ回数設定
  - 最大遅延30秒のキャップ
- [x] リトライテスト作成（10テスト成功）

**コミット**: `00ccadf` - Add HTTP Transport layer implementation

### 3.5 ロギング（構造化ログ） ✅

- [x] `transport/logging.go` 作成
- [x] `LoggingRoundTripper` 実装
  - リクエスト/レスポンスログ
  - シークレットマスキング（Authorization, Cookie, API-Key）
  - slog（Go 1.21+）利用
  - 構造化ログ出力
- [x] ロギングテスト作成（7テスト成功）

**コミット**: `00ccadf` - Add HTTP Transport layer implementation

### 3.6 User-Agent付与 ✅

- [x] `transport/useragent.go` 作成
- [x] `UserAgentRoundTripper` 実装
  - カスタムUser-Agent設定
  - 既存User-Agentへの追加
  - DefaultUserAgent ヘルパー関数
- [x] User-Agentテスト作成（7テスト成功）

**コミット**: `00ccadf` - Add HTTP Transport layer implementation

### 3.7 ユニットテスト ✅

- [x] 各RoundTripperのテスト作成
  - transport_test.go（4テスト）
  - ratelimit_test.go（4テスト）
  - retry_test.go（10テスト）
  - logging_test.go（7テスト）
  - useragent_test.go（7テスト）
- [x] httptest.Server でエンドポイントモック
- [x] レート制限・リトライ動作検証
- [x] 全42テスト成功

**コミット**: `00ccadf` - Add HTTP Transport layer implementation

### Phase 3 成果物

✅ **完了条件達成**: Transport層が統合され、堅牢なHTTP通信が可能

**作成ファイル**: 11ファイル
- 実装ファイル: 5個（transport.go, options.go, ratelimit.go, retry.go, logging.go, useragent.go）
- テストファイル: 5個（各_test.go）

**テスト**: 42テスト全て成功
**コミット数**: 2
- `00ccadf` - Transport層実装
- `294ec64` - 依存関係追加

**次のフェーズ**: Phase 4 - Generated API Client

---

## Phase 4: Generated API Client（Code Generation） ✅

**目標**: OpenAPIからクライアント生成

**ステータス**: ✅ 完了（2025-12-14）

### 4.1 oapi-codegenテンプレート設定 ✅

- [x] `oapi-codegen.yaml` 詳細設定
  - models: true
  - client: true
  - types: true
  - skip-prune: false
  - always-prefix-enum-values: true
  - embedded-spec: false
- [x] 設定ファイルのドキュメント化（コメント追加）
- [x] CLAUDE.md に設定内容を記載

**コミット**: `0b677af` (PR #32)

### 4.2 OpenAPI仕様ファイル取得 ✅

- [x] freee公式リポジトリから会計API OpenAPI v3仕様をダウンロード
- [x] `api/openapi.json` として保存 (1.6MB, OpenAPI 3.0.1, API v1.0)
- [x] バージョン情報を README.md に記載
- [x] `tools/update-openapi.sh` スクリプト作成（自動更新用）

**コミット**: 予定（Issue #8）
**ソース**: https://github.com/freee/freee-api-schema

### 4.3 internal/gen/ コード生成 ✅

- [x] `oapi-codegen` 実行（~46,000行生成）
- [x] 生成コードレビュー
  - 構造体定義確認
  - メソッドシグネチャ確認
- [x] 生成コードを `.gitignore` から除外（バージョン管理対象）
- [x] 問題のある5エンドポイントを除外（参照深度制限による）

**コミット**: 予定（Issue #9）
**除外エンドポイント**: account_items, items, sections, partners, segment_tags の upsert_by_code

### 4.4 生成コードの検証 ✅

- [x] 型安全性確認
- [x] ビルド成功確認
- [x] 依存関係追加（oapi-codegen/runtime）

### 4.5 エラー型定義（freee APIエラー） ✅

- [x] `client/error.go` 作成
- [x] `FreeeError` 構造体定義
  - HTTPステータスコード
  - エラーメッセージ
  - freee APIエラーコード
  - エラー詳細配列
- [x] `Error()` メソッド実装
- [x] `ParseErrorResponse()` 関数実装
- [x] ヘルパー関数実装（IsBadRequestError など）
- [x] 完全なユニットテスト作成

**コミット**: 予定（Issue #10）

### 4.6 基本的なAPI呼び出しテスト ✅

- [x] httptest.Server で freee API モック
- [x] 生成クライアントで呼び出しテスト
- [x] レスポンスデシリアライズ確認
- [x] エラーレスポンステスト
- [x] `internal/gen/client_test.go` 作成

**コミット**: 予定（Issue #11）

### 4.7 生成スクリプト整備（tools/） ✅

- [x] `tools/update-openapi.sh` 作成（完了 - Issue #8）
- [x] `tools/generate.go` 作成（`//go:generate` ディレクティブ）
- [x] Makefile 作成（`make generate`, `make test`, `make lint` など）

**コミット**: 予定（Issue #12）

**Phase 4 完了条件**: ✅ 生成コードで freee API を呼び出せること

---

## Phase 5: Accounting Facade（User-Facing API） ✅

**目標**: 使いやすいFacade API提供

**ステータス**: ✅ 完了（2025-12-17）

### 5.1 client/ パッケージ設計（Client構造体） ✅

- [x] `client/client.go` 作成
- [x] `Client` 構造体定義
  - HTTPClient
  - BaseURL
  - TokenSource
  - UserAgent
  - Context
- [x] `NewClient(opts ...Option)` 実装
- [x] `Option` パターン実装
  - WithHTTPClient
  - WithBaseURL
  - WithTokenSource
  - WithUserAgent
  - WithContext
- [x] `Do(req)` メソッド実装
- [x] 包括的なユニットテスト作成

**コミット**: `23685a8` (PR #35) - Implement Phase 5.1: Client structure and options pattern

### 5.2 accounting/ Facade設計 ✅

- [x] `accounting/client.go` 作成
- [x] `AccountingClient` 構造体定義
- [x] サービスごとのサブクライアント設計
  - `DealsService` - 取引
  - `JournalsService` - 仕訳
  - `WalletTxnService` - 口座明細
  - `TransfersService` - 取引（振替）
- [x] `accounting/services.go` 作成（サービス構造体定義）
- [x] 遅延初期化（Lazy initialization）実装
- [x] ClientWithResponses 統合（自動レスポンス解析）
- [x] 包括的なユニットテスト作成（8テスト成功）

**コミット**: `8136d35` (PR #37) - Implement Phase 5.2: Design Accounting Facade architecture

### 5.3 取引（Deals）API実装 ✅

- [x] `accounting/deals.go` 作成
- [x] `DealsService.List(ctx, opts)` 実装
  - 柔軟なフィルタリングオプション
  - ページネーション対応（offset/limit）
  - ListDealsOptions 型定義
- [x] `DealsService.Get(ctx, companyID, id, opts)` 実装
  - Accruals 表示制御
  - GetDealOptions 型定義
- [x] `DealsService.Create(ctx, params)` 実装
  - DealCreateParams 使用
  - 適切なエラーハンドリング
- [x] `DealsService.Update(ctx, id, params)` 実装
  - DealUpdateParams 使用
  - 部分更新対応
- [x] `DealsService.Delete(ctx, companyID, id)` 実装
  - 適切なステータスコード処理
- [x] `accounting/deals_test.go` 作成
  - 13テスト全て成功
  - httptest.Server でモック
  - 全CRUD操作の検証

**コミット**: `3aa77c7` (PR #38) - Implement Phase 5.3: Deals API implementation

### 5.4 仕訳（Journals）API実装 ✅

- [x] `accounting/journals.go` 作成
- [x] `JournalsService.Download(ctx, opts)` 実装（仕訳帳ダウンロード）
- [x] `JournalsService.ListManualJournals(ctx, opts)` 実装（振替伝票一覧）
- [x] `JournalsService.GetManualJournal(ctx, id)` 実装
- [x] `JournalsService.CreateManualJournal(ctx, params)` 実装
- [x] `JournalsService.UpdateManualJournal(ctx, id, params)` 実装
- [x] `JournalsService.DeleteManualJournal(ctx, id)` 実装
- [x] `accounting/journals_test.go` 作成

**コミット**: PR #51 - Implement Phase 5.4: Journals API

### 5.5 取引先（Partners）API実装 ✅

- [x] `accounting/partners.go` 作成
- [x] `PartnersService.List(ctx, opts)` 実装
- [x] `PartnersService.Get(ctx, id)` 実装
- [x] `PartnersService.Create(ctx, params)` 実装
- [x] `PartnersService.Update(ctx, id, params)` 実装
- [x] `PartnersService.Delete(ctx, id)` 実装
- [x] `accounting/partners_test.go` 作成

**コミット**: PR #52 - Implement Phase 5.5: Partners API

### 5.6 ページング実装（Iterator/Pager） ✅

- [x] `accounting/pager.go` 作成
- [x] `Iterator[T]` ジェネリックインターフェース定義
- [x] `Next()`, `Value()`, `Err()` メソッド
- [x] `PageFetcher[T]` 型定義
- [x] `PaginatedIterator[T]` 実装（自動ページフェッチ）
- [x] `accounting/pager_test.go` 作成

**コミット**: PR #53 - Implement pagination with Iterator pattern

### 5.7 ユニットテスト ✅

- [x] 各サービスのテスト作成
  - `client_test.go` - クライアント構造体テスト
  - `deals_test.go` - 取引APIテスト
  - `journals_test.go` - 仕訳APIテスト
  - `partners_test.go` - 取引先APIテスト
  - `accountitems_test.go` - 勘定科目テスト
  - `items_test.go` - 品目テスト
  - `sections_test.go` - 部門テスト
  - `tags_test.go` - タグテスト
  - `transfers_test.go` - 振替テスト
  - `wallettxns_test.go` - 口座明細テスト
  - `pager_test.go` - ページングテスト
- [x] httptest.Server でモック
- [x] ページング動作検証

### 5.8 統合テスト（E2E with mock） ✅

- [x] `tests/integration/` ディレクトリ作成
- [x] エンドツーエンドシナリオテスト
  - `auth_test.go` - 認証フローテスト
  - `deals_test.go` - 取引E2Eテスト
  - `error_handling_test.go` - エラーハンドリングテスト
  - `pagination_test.go` - ページングE2Eテスト
  - `golden_test.go` - Goldenファイルテスト
- [x] Golden file パターンでレスポンス管理
  - `golden/golden.go` - Goldenファイルユーティリティ
- [x] `mockserver/server.go` - モックサーバー実装

**コミット**: PR #53, #54, #55 - Add E2E integration tests

### Phase 5 成果物

✅ **完了条件達成**: Facade経由で会計APIを利用できる

**作成ファイル**: 23ファイル
- 実装ファイル: 11個（client.go, deals.go, journals.go, partners.go, accountitems.go, items.go, sections.go, tags.go, transfers.go, wallettxns.go, pager.go）
- テストファイル: 11個（各_test.go）
- 統合テスト: 7ファイル

**次のフェーズ**: Phase 6 - ドキュメント・サンプル

---

## Phase 6: ドキュメント・サンプル（Documentation） ✅

**目標**: ユーザー向けドキュメント整備

**ステータス**: ✅ 完了（2025-12-17）

### 6.1 GoDoc コメント充実 ✅

- [x] すべての公開型・関数にコメント追加
- [x] パッケージレベルのdoc.go作成（`client/doc.go`）
- [x] サンプルコード埋め込み（Example関数）
- [x] `go doc` で確認
- [x] 日本語ドキュメント化

**コミット**: PR #56 - GoDoc enhancement with Japanese documentation

### 6.2 README.md完全版 ✅

- [x] プロジェクト説明充実（日本語化）
- [x] インストール手順詳細化
- [x] 認証フロー説明
- [x] コードサンプル複数パターン
- [x] トラブルシューティング
- [x] FAQ

### 6.3 examples/ 複数パターン ✅

- [x] `examples/basic/main.go` 作成（基本的な取引取得）
- [x] `examples/advanced/main.go` 作成（複数サービス、エラーハンドリング）
- [x] `examples/iterator/main.go` 作成（Iteratorパターンによるページング）
- [x] `examples/oauth/main.go` 改善
- [x] 各exampleにREADME.md追加
  - `examples/basic/README.md`
  - `examples/advanced/README.md`
  - `examples/iterator/README.md`
  - `examples/oauth/README.md`

**コミット**: PR #58 - Create multiple example applications

### 6.4 CONTRIBUTING.md ✅

- [x] コントリビューションガイドライン作成
- [x] 開発環境セットアップ手順
- [x] プルリクエストプロセス
- [x] コーディング規約
- [x] テスト要件
- [x] ドキュメント要件

**コミット**: PR #49 - Add security policy and contribution guidelines

### 6.5 セキュリティポリシー（SECURITY.md） ✅

- [x] セキュリティ脆弱性報告方法
- [x] サポート対象バージョン
- [x] セキュリティベストプラクティス
- [x] GitHub Security Advisories 設定

**コミット**: PR #49 - Add security policy and contribution guidelines

### 6.6 pkg.go.dev 公開準備 ✅

- [x] `godoc` でローカル確認
- [x] pkg.go.dev 公開準備
  - [x] go.mod モジュールパスを正しいリポジトリURLに修正
  - [x] 全ソースファイルのimportパスを更新
- [x] README.md にpkg.go.devリンク・バッジ（既存）
- [x] Go Report Card バッジ（既存）
- [x] ビルド・テスト確認

**コミット**: PR #59 - pkg.go.dev preparation

### Phase 6 成果物

✅ **完了条件達成**: ドキュメント完備

**作成ファイル**:
- `CONTRIBUTING.md` - コントリビューションガイドライン
- `SECURITY.md` - セキュリティポリシー
- `client/doc.go` - パッケージドキュメント
- `examples/basic/` - 基本サンプル
- `examples/advanced/` - 応用サンプル
- `examples/iterator/` - ページングサンプル
- 各README.md

**次のステップ**: v0.1.0 リリースタグ作成

**次のフェーズ**: Phase 7 - 拡張・改善

---

## Phase 7: 拡張・改善（Enhancement） 🔄

**目標**: フィードバック反映・機能拡充

**ステータス**: 🔄 進行中（2025-12-17）

### 7.1 パフォーマンス最適化

- [ ] プロファイリング実施（pprof）
- [ ] 不要なアロケーション削減
- [ ] コネクションプーリング最適化
- [ ] ベンチマークテスト追加

### 7.2 より多くの会計API対応 ✅

- [x] 勘定科目（AccountItems）- `AccountItemsService` 実装（CRUD + テスト）
- [x] 品目（Items）- `ItemsService` 実装（CRUD + ページネーション + テスト）
- [x] 部門（Sections）- `SectionsService` 実装（CRUD + テスト）
- [x] タグ（Tags）- `TagsService` 実装（CRUD + ページネーション + テスト）
- [x] 振替（Transfers）- `TransfersService` 実装（CRUD + テスト）
- [x] 口座明細（WalletTxns）- `WalletTxnService` 実装（List, Get + テスト）
- [ ] その他エンドポイント

**コミット**: PR #57 - Extend support for more accounting API endpoints

### 7.3 キャッシュ機能（オプション）

- [ ] `cache/` パッケージ設計
- [ ] メモリキャッシュ実装
- [ ] TTL設定
- [ ] キャッシュ無効化API

### 7.4 メトリクス収集

- [ ] Prometheus メトリクス対応
- [ ] リクエスト数・レイテンシ計測
- [ ] エラー率計測

### 7.5 コミュニティフィードバック対応

- [ ] GitHub Issue 対応
- [ ] プルリクエストレビュー
- [ ] 機能リクエスト検討
- [ ] バグ修正

**Phase 7 完了条件**: v0.2.0以降のリリース

---

## 🎯 即座に着手すべきタスク（Quick Wins）

### ✅ Phase 1 完了（2025-12-14）

1. ✅ git init
2. ✅ go mod init
3. ✅ .gitignore 作成
4. ✅ LICENSE 作成
5. ✅ ディレクトリ構造作成
6. ✅ CI/CD 設定
7. ✅ README.md 基本構造

### ✅ Phase 2 完了（2025-12-14）

1. ✅ `auth/config.go` 作成（OAuth2設定）
2. ✅ `auth/errors.go` 作成（エラー型）
3. ✅ `auth/token.go` 作成（トークン管理）
4. ✅ `auth/tokensource.go` 作成（TokenSource実装）
5. ✅ ユニットテスト作成（23テスト全て成功）
6. ✅ OAuth2サンプルアプリケーション作成
7. ✅ 詳細ドキュメント作成

### ✅ Phase 3 完了（2025-12-14）

1. ✅ `transport/transport.go` 作成（基本構造）
2. ✅ `transport/options.go` 作成（設定オプション）
3. ✅ `transport/ratelimit.go` 作成（レート制限）
4. ✅ `transport/retry.go` 作成（リトライロジック）
5. ✅ `transport/logging.go` 作成（ロギング）
6. ✅ `transport/useragent.go` 作成（User-Agent）
7. ✅ 包括的なテスト作成（42テスト全て成功）

### ✅ Phase 4 完了（2025-12-14）

1. ✅ コード生成設定（`oapi-codegen.yaml`）- 完了（Issue #7）
2. ✅ OpenAPI仕様ファイル取得（`api/openapi.json`）- 完了（Issue #8）
3. ✅ oapi-codegen セットアップ・実行
4. ✅ `internal/gen/` コード生成と検証
5. ✅ エラー型定義とテスト
6. ✅ Makefile・スクリプト整備

### ✅ Phase 5 完了（2025-12-17）

1. ✅ Client構造体とオプションパターン実装 - 完了（Issue #13, PR #35）
2. ✅ AccountingClient Facade設計 - 完了（Issue #14, PR #37）
3. ✅ Deals API全CRUD操作実装 - 完了（Issue #15, PR #38）
4. ✅ Journals API実装 - 完了（PR #51）
5. ✅ Partners API実装 - 完了（PR #52）
6. ✅ ページング実装 - 完了（PR #53）
7. ✅ ユニットテスト充実 - 完了
8. ✅ 統合テスト作成 - 完了（PR #53-#55）

### ✅ Phase 6 完了（2025-12-17）

1. ✅ GoDocコメント充実 - 完了（PR #56）
2. ✅ README.md完全版 - 完了
3. ✅ Examples作成 - 完了（PR #58）
4. ✅ CONTRIBUTING.md - 完了（PR #49）
5. ✅ SECURITY.md - 完了（PR #49）
6. ✅ pkg.go.dev準備 - 完了（PR #59）

### 🎯 Phase 7 次のタスク

1. ⬜ パフォーマンス最適化（Phase 7.1）
2. ✅ より多くの会計API対応 - 完了（PR #57）
3. ⬜ キャッシュ機能（Phase 7.3）
4. ⬜ メトリクス収集（Phase 7.4）
5. ⬜ コミュニティフィードバック対応（Phase 7.5）

---

## 📝 メモ・注意事項

- 各フェーズは順次進めることを推奨（依存関係あり）
- テストは実装と同時に作成（後回しにしない）
- OpenAPI仕様更新時は自動検知・対応
- セキュリティ問題は最優先対応
- コミュニティからのフィードバックを積極的に取り入れる

---

**最終更新**: 2025-12-17
**次のアクション**: v0.1.0 リリース、Phase 7 拡張機能
