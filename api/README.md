# api

freee API の OpenAPI 仕様を格納するディレクトリ。

## ファイル

- `openapi.yaml`: freee 会計API の OpenAPI 3.0 仕様

## 更新方法

```bash
# 自動更新スクリプトを実行
./tools/update-openapi.sh

# または手動でダウンロード
curl -o api/openapi.yaml https://developer.freee.co.jp/...
```

## バージョン管理

OpenAPI仕様はバージョン管理の対象です。
freee側の仕様変更を追跡するため、変更履歴を保持します。

## 生成コードへの反映

```bash
# OpenAPI仕様から internal/gen/ のコードを再生成
go generate ./tools
```
