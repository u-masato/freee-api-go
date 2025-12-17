# freee Public API Go Client Library 計画書

## 1. 目的（Purpose）

本プロジェクトの目的は、freee（フリー株式会社）が提供する **Public API（主に会計API）** を安全かつ使いやすく利用するための **Go製クライアントライブラリ（SDK）** を設計・実装し、GitHub上で公開することである。

特に以下を重視する。

- バックエンド用途（サーバーサイド）での実用性
- OAuth 2.0 認可コードフローを前提とした堅牢な認証処理
- OpenAPIスキーマを活用しつつ、利用者にとって扱いやすいAPI設計
- freee API仕様変更への耐性（生成コードと手書きコードの分離）
- 初めてSDK/Clientを作成する際の「模範例」となる設計

---

## 2. 対象範囲・非対象範囲（Scope）

### 対象範囲
- freee 会計API（Accounting API）
- OAuth 2.0 Authorization Code Grant
  - 認可URL生成
  - 認可コード → アクセストークン取得
  - リフレッシュトークンによる更新

### 非対象範囲
- フロントエンド用SDK
- Implicit Flow / PKCE専用クライアント
- 請求書API（会計APIとは別系統のため初期スコープ外）
- UIやCLIツールの提供

---

## 3. 提供機能（Functional Requirements）

### 3.1 認証・認可（OAuth2）
- 認可URLの生成
- アクセストークン取得
- リフレッシュトークン取得・更新
- トークンの自動更新（TokenSource）[optional]

### 3.2 APIアクセス
- freee APIへのHTTPリクエスト送信
- Authorizationヘッダ自動付与
- Context対応（キャンセル・タイムアウト）
- User-Agent付与

### 3.3 会計Facade
- サービス単位のFacade提供
- CRUD操作の簡易化
- ページング処理の隠蔽（Iterator / Pager）

### 3.4 信頼性・運用性
- レート制限（HTTP 429）対応
- リトライ制御（条件付き）
- freee APIエラーを統一的に扱えるエラー型
- ログ出力時の機密情報マスキング

---

## 4. アーキテクチャ概要（Architecture）

### 4.1 全体構成

利用者コード
│
▼
Facade（accounting/*）
│
▼
Generated API Client（internal/gen）
│
▼
Transport（http.Client / RoundTripper）
│
▼
freee Public API

### 4.2 レイヤ責務

| レイヤ | 責務 |
|------|------|
| Facade | ユースケース単位のAPI、ページング統合 |
| Generated | OpenAPIから生成された生API |
| Auth | OAuth2フロー・トークン管理 |
| Transport | HTTP共通処理（Retry / RateLimit / UA） |

---

## 5. パッケージ構成（Proposed Structure）

freee-api-go/
├─ client/            # 公開API（Client, Options, Error）
├─ auth/              # OAuth2（認可URL, トークン取得/更新）
├─ accounting/        # 会計API Facade
│   ├─ deals.go
│   ├─ journals.go
│   ├─ partners.go
│   └─ …
├─ transport/         # HTTPミドルウェア（Retry, RateLimit, Logging）
├─ internal/
│   ├─ gen/           # OpenAPI生成コード（非公開）
│   └─ testutil/      # テストユーティリティ
├─ examples/          # サンプルコード
│   ├─ oauth/
│   ├─ basic/
│   └─ advanced/
└─ tools/             # OpenAPI生成・管理スクリプト

### 5.1 OpenAPI仕様の取得

freee APIのOpenAPI仕様は以下から取得:

- freee公式ドキュメント: https://developer.freee.co.jp/
- OpenAPI Specification (v3): 公式サイトから最新版をダウンロード
- tools/スクリプトで自動取得・バージョン管理

### 5.2 技術スタック（Technology Stack）

| カテゴリ | 選定技術 | 用途 |
|---------|---------|------|
| OpenAPIジェネレーター | [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) | Go構造体・クライアント生成 |
| OAuth2ライブラリ | [golang.org/x/oauth2](https://pkg.go.dev/golang.org/x/oauth2) | OAuth2フロー実装 |
| HTTPクライアント | net/http (標準ライブラリ) | API通信 |
| レート制限 | [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate) | トークンバケット実装 |
| リトライ | [hashicorp/go-retryablehttp](https://github.com/hashicorp/go-retryablehttp) | 自動リトライ |
| テスト | httptest (標準), testify | モックサーバー・アサーション |
| Linter | golangci-lint | 静的解析 |
| CI/CD | GitHub Actions | ビルド・テスト・リリース |

---

## 6. 設計方針（Design Principles）

### 6.1 公開APIの安定性
- 生成コードは `internal/` に閉じる
- 公開APIはFacade層のみ
- セマンティックバージョニング（SemVer）を採用

### 6.2 利用者中心設計
- freee APIの生仕様を直接露出しない
- ページングやエラー処理を隠蔽
- 「forで回せる」「Contextを渡すだけ」を重視

### 6.3 OAuth責務分離
- 認証処理はSDKで補助するが、Webフロー制御は利用者に委ねる

### 6.4 拡張性・保守性
- Transport層で横断的関心事を集約
- サービス追加が容易な構成
- OpenAPI更新に追従しやすい生成運用

---

## 7. 非機能要件（Non-Functional Requirements）

### 7.1 コード品質

- Go Modules対応（go 1.24+）
- GoDocによるAPIドキュメント生成
- golangci-lint による静的解析
- コードカバレッジ 80%以上を目標

### 7.2 テスト戦略

| テスト種別 | 手法 | 対象 |
|----------|------|------|
| ユニットテスト | httptest.Server | 各パッケージ |
| 統合テスト | モックサーバー | エンドツーエンド |
| 契約テスト | OpenAPI Validation | APIレスポンス検証 |
| セキュリティテスト | go-sec | 脆弱性スキャン |

**モック方針:**
- internal/testutil でテスト用ヘルパー提供
- httptest.Server で freee API をモック
- golden file パターンでレスポンス管理

### 7.3 CI/CD パイプライン

**GitHub Actions ワークフロー:**
- **Lint**: golangci-lint実行
- **Test**: ユニットテスト・カバレッジレポート
- **Build**: マルチOS（Linux, macOS, Windows）ビルド検証
- **OpenAPI Check**: スキーマ差分検知・自動PR作成
- **Release**: セマンティックバージョニング・自動リリース

### 7.4 セキュリティ

- トークン・シークレットのログ出力抑制
- 環境変数からの認証情報読み込み
- TLS 1.2+ 必須
- 依存ライブラリの脆弱性スキャン（dependabot）

---

## 8. 想定ユースケース（Example）

- freee会計データを定期取得し、分析・検証するバックエンド処理
- 会計仕訳のAI自動チェック・評価ツール/ダッシュボード
- 内部業務システムとfreeeのデータ連携

---

## 9. 今後の展開（Future Work）

- キャッシュレイヤの公式サポート
- 請求書API・HR API対応
- Webhook受信サポート
- メトリクス・トレーシング統合（OpenTelemetry）

---

## 10. 実装フェーズ（Implementation Phases）

### Phase 1: プロジェクト基盤（Foundation）

**目標**: 開発環境・ビルド基盤の構築

- [ ] リポジトリ初期化（go.mod, .gitignore, LICENSE）
- [ ] ディレクトリ構造作成
- [ ] GitHub Actions CI/CD設定
- [ ] golangci-lint設定
- [ ] OpenAPI仕様ファイル取得
- [ ] oapi-codegen セットアップ
- [ ] README.md基本構造

**成果物**: ビルド可能な空プロジェクト

### Phase 2: OAuth2認証（Authentication）

**目標**: freee OAuth2フロー実装

- [ ] auth/ パッケージ構造設計
- [ ] 認可URL生成機能
- [ ] アクセストークン取得
- [ ] リフレッシュトークン処理
- [ ] TokenSource実装
- [ ] ユニットテスト（モック）
- [ ] examples/oauth/ サンプル作成

**成果物**: 認証可能なライブラリ

### Phase 3: HTTP Transport層（Transport）

**目標**: 共通HTTP処理の実装

- [ ] transport/ パッケージ設計
- [ ] カスタムRoundTripper実装
- [ ] レート制限（rate.Limiter統合）
- [ ] リトライロジック
- [ ] ロギング（構造化ログ）
- [ ] User-Agent付与
- [ ] ユニットテスト

**成果物**: 堅牢なHTTP通信基盤

### Phase 4: Generated API Client（Code Generation）

**目標**: OpenAPIからクライアント生成

- [ ] oapi-codegenテンプレート設定
- [ ] internal/gen/ コード生成
- [ ] 生成コードの検証
- [ ] エラー型定義（freee APIエラー）
- [ ] 基本的なAPI呼び出しテスト
- [ ] 生成スクリプト整備（tools/）

**成果物**: 会計API呼び出し可能なクライアント

### Phase 5: Accounting Facade（User-Facing API）

**目標**: 使いやすいFacade API提供

- [ ] client/ パッケージ設計（Client構造体）
- [ ] accounting/ Facade設計
- [ ] 取引（Deals）API実装
- [ ] 仕訳（Journals）API実装
- [ ] 取引先（Partners）API実装
- [ ] ページング実装（Iterator/Pager）
- [ ] ユニットテスト
- [ ] 統合テスト（E2E with mock）

**成果物**: v0.1.0リリース候補

### Phase 6: ドキュメント・サンプル（Documentation）

**目標**: ユーザー向けドキュメント整備

- [ ] GoDoc コメント充実
- [ ] README.md完全版
- [ ] examples/ 複数パターン
- [ ] CONTRIBUTING.md
- [ ] セキュリティポリシー（SECURITY.md）
- [ ] APIリファレンス生成

**成果物**: v0.1.0正式リリース

### Phase 7: 拡張・改善（Enhancement）

**目標**: フィードバック反映・機能拡充

- [ ] パフォーマンス最適化
- [ ] より多くの会計API対応
- [ ] キャッシュ機能（オプション）
- [ ] メトリクス収集
- [ ] コミュニティフィードバック対応

**成果物**: v0.2.0+

---

## 11. ライセンス・公開方針（License & Publishing）

### ライセンス

- **MIT License** を採用
  - 商用利用可能
  - 改変・再配布自由
  - 責任免責

### 公開戦略

- GitHub上でオープンソースとして公開
- pkg.go.devでドキュメント自動公開
- Semantic Versioning準拠
- リリースノート自動生成
- issueテンプレート・PR テンプレート整備

### コントリビューション

- CONTRIBUTING.md でガイドライン明示
- Code of Conduct 設定
- issue/PR歓迎のスタンス

---

## 12. まとめ

本プロジェクトは、
**「OpenAPI × Go × OAuth2」を前提とした、実運用に耐えるfreee APIクライアントのリファレンス実装**
を目指す。

### 重視する価値

1. **実用性**: 実務でそのまま使える機能・設計
2. **保守性**: OpenAPI変更への追従性、明確な責務分離
3. **可読性**: 初学者でも理解できるコード・ドキュメント
4. **安全性**: OAuth2セキュリティ、エラーハンドリング、テスト
5. **拡張性**: 新機能追加が容易なアーキテクチャ

### 成功基準

- ✅ freee会計APIをGoから簡単に呼び出せる
- ✅ OAuth2フローが安全・確実に動作する
- ✅ レート制限・リトライが自動で処理される
- ✅ ページングが透過的に扱える
- ✅ エラーが適切に処理・報告される
- ✅ テストカバレッジ80%以上
- ✅ 充実したドキュメント・サンプル

本計画に基づき、段階的に実装を進める。
