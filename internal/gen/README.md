# internal/gen

OpenAPIスキーマから自動生成されたコード。

## 注意

- このパッケージのコードは `oapi-codegen` により自動生成されます
- **直接編集しないでください**
- 公開API（internal/）として外部から利用できません
- 再生成: `go generate ./tools`

## 生成元

- OpenAPI仕様: `api/openapi.json`
- 生成設定: `oapi-codegen.yaml`
- 生成スクリプト: `tools/generate.go`
- 生成ツール: `oapi-codegen v2.5.1`

## 生成時の問題と対応

### 問題: 参照深度エラー

oapi-codegen v2.5.1でfreee APIのOpenAPI仕様からコード生成を試みた際、以下のエラーが発生しました：

```
error generating code: error creating operation definitions:
error generating body definitions: error generating request body definition:
error turning reference (#/paths/~1api~11~1account_items~1code~1upsert/put/requestBody/content/application~1json/schema)
into a Go type: unexpected reference depth: 8 for ref:
#/paths/~1api~11~1account_items~1code~1upsert/put/requestBody/content/application~1json/schema local: true
```

### 原因

**oapi-codegenの参照深度制限**

oapi-codegen v2.5.1には、スキーマ参照の深度チェック機能があり、深くネストしたスキーマ構造を処理できません。freee APIの一部のエンドポイント（特に`upsert_by_code`パターンのエンドポイント）は、インラインスキーマが深くネストしており、この制限に抵触します。

**該当するエンドポイントの特徴**：
- リクエストボディに複雑なネストした構造を持つ
- `$ref`を使わずにインラインでスキーマ定義している
- パスのエンコーディング（`~1`）により参照パスが長くなる

**既知の問題**：
- GitHub Issue: [Code Generation Issue #1892](https://github.com/oapi-codegen/oapi-codegen/issues/1892)
- PR（ドラフト）: [fix: Unexpected reference depth #1950](https://github.com/oapi-codegen/oapi-codegen/pull/1950)
  - 参照深度チェックを削除する修正がドラフトPRとして提出されているが、v2.5.1時点ではまだマージされていない

### 対応策

**エンドポイントの除外**

問題のある5つのエンドポイントを除外してコード生成を行いました。

**除外されたエンドポイント**：

| パス | Operation ID | 説明 |
|------|-------------|------|
| `/api/1/account_items/code/upsert` | `api/v1/account_items#upsert_by_code` | 勘定科目のコード指定登録・更新 |
| `/api/1/items/code/upsert` | `api/v1/items#upsert_by_code` | 品目のコード指定登録・更新 |
| `/api/1/sections/code/upsert` | `api/v1/sections#upsert_by_code` | 部門のコード指定登録・更新 |
| `/api/1/partners/upsert_by_code` | `api/v1/partners#upsert_by_code` | 取引先のコード指定登録・更新 |
| `/api/1/segments/{segment_id}/tags/code/upsert` | `upsert_segment_tag` | セグメントタグのコード指定登録・更新 |

**除外設定の実装**：

`tools/generate.go`で以下のように除外パラメータを指定：

```go
//go:generate sh -c "oapi-codegen -package gen -generate types,client -exclude-operation-ids 'api/v1/account_items#upsert_by_code,api/v1/items#upsert_by_code,api/v1/sections#upsert_by_code,upsert_segment_tag,api/v1/partners#upsert_by_code' ../api/openapi.json > ../internal/gen/client.gen.go"
```

### 影響範囲

**カバレッジ**：
- 全エンドポイント数: 89
- 生成成功: 84エンドポイント
- 除外: 5エンドポイント
- **カバレッジ: 94%**

**機能的影響**：
- 除外された5エンドポイントは、すべて「コード指定での登録・更新」機能
- 代替手段として、ID指定での登録・更新エンドポイントは正常に生成されている
- 必要に応じて、Phase 5（Facade層）で手動実装可能

### 今後の対応

**短期的な対応**：
1. **現状のまま運用**: 94%のエンドポイントカバレッジで十分実用的
2. **手動実装**: 必要に応じてFacade層で除外エンドポイントを手動実装

**長期的な対応**：
1. **oapi-codegen更新待ち**: PR #1950がマージされた後、oapi-codegenをアップデート
2. **OpenAPI仕様の修正**: freee公式リポジトリにフィードバック（スキーマのフラット化提案）
3. **別ツールの検討**: 他のOpenAPIコード生成ツール（go-swagger, openapi-generator等）の評価

### 参考資料

- [oapi-codegen公式リポジトリ](https://github.com/oapi-codegen/oapi-codegen)
- [freee API Schema公式リポジトリ](https://github.com/freee/freee-api-schema)
- [Issue #1892: Code Generation Issue](https://github.com/oapi-codegen/oapi-codegen/issues/1892)
- [PR #1950: fix: Unexpected reference depth](https://github.com/oapi-codegen/oapi-codegen/pull/1950)

### 生成統計

- **生成コード行数**: ~46,000行
- **生成日時**: 2025-12-14
- **oapi-codegenバージョン**: v2.5.1
- **OpenAPIバージョン**: 3.0.1
- **freee API バージョン**: v1.0
