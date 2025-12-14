# tools

コード生成・開発ツール用のディレクトリ。

## ファイル

- `generate.go`: OpenAPIからのコード生成スクリプト
- `update-openapi.sh`: OpenAPI仕様の自動更新スクリプト

## コード生成

```bash
# OpenAPIからGoコードを生成
go generate ./tools

# または
make generate
```

## OpenAPI仕様更新

```bash
# 最新のfreee API仕様を取得
./tools/update-openapi.sh

# 差分確認
git diff api/openapi.yaml
```

## 依存関係

```bash
# oapi-codegen のインストール
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
```
