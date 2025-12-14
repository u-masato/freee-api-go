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

freee/
├─ client/            # 公開API（Client, Options, Error）
├─ auth/              # OAuth2（認可URL, トークン取得/更新）
├─ accounting/        # 会計API Facade
│   ├─ deals.go
│   ├─ journals.go
│   ├─ partners.go
│   └─ …
├─ transport/         # HTTPミドルウェア
├─ internal/
│   └─ gen/           # OpenAPI生成コード（非公開）
└─ tools/             # OpenAPI生成スクリプト

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

- Go Modules対応
- GoDocによるAPIドキュメント生成
- ユニットテスト（httptest.Server活用）
- CIでのOpenAPI差分検知
- セキュリティ考慮（Secret非出力）

---

## 8. 想定ユースケース（Example）

- freee会計データを定期取得し、分析・検証するバックエンド処理
- 会計仕訳のAI自動チェック・評価ツール/ダッシュボード
- 内部業務システムとfreeeのデータ連携

---

## 9. 今後の展開（Future Work）

- キャッシュレイヤの公式サポート

---

## 10. まとめ

本プロジェクトは、  
**「OpenAPI × Go × OAuth2」を前提とした、実運用に耐えるfreee APIクライアントのリファレンス実装**  
を目指す。

SDKとしての完成度・保守性・可読性を重視し、  
初学者が読んでも理解でき、実務でそのまま使える設計を採用する。
