# zaiGoFramework 使い方ガイド

本書は zaiGoFramework のセットアップから起動、API 利用方法までを Windows 11 + Docker を前提に簡潔に説明します。

- リポジトリ: `c:\git\zaiGoFramework`
- 主な構成: `cmd/api/` (APIサーバー), `pkg/inventory/` (コア), `migrations/` (DB初期化SQL), `examples/` (使用例)

---

## 前提条件

- Docker Desktop for Windows
- PowerShell
- (任意) Go 1.21+（ローカルビルド/サンプル実行用）

---

## クイックスタート（Docker）

`docker-compose.yml` を用いて API と PostgreSQL を起動します。

1) 依存イメージとコンテナを起動

```powershell
# プロジェクトルート（c:\git\zaiGoFramework）で実行
docker-compose up -d --build
```

2) 起動確認

```powershell
# コンテナ一覧
docker ps

# API ログを追尾
docker-compose logs -f inventory-api
```

3) ヘルスチェック（HTTP）

```powershell
# PowerShellのcurl (Invoke-WebRequest エイリアス) でOK
curl http://localhost:8080/health
```

4) 停止/削除

```powershell
# 停止
docker-compose down

# ボリュームも削除（DBを初期化したい場合）
docker-compose down -v
```

メモ:
- 初回起動時、`migrations/` が `postgres` にマウントされ、`001_initial_schema.sql` が自動実行されます（`docker-compose.yml` 参照）。
- API コンテナのログは `./logs` を `/app/logs` にマウントしています。

---

## 環境変数（主要）

`internal/config/config.go` での読み取りと `docker-compose.yml` のデフォルト値に基づきます。

- データベース
  - `DB_HOST` (default: `localhost` / Compose では `postgres`)
  - `DB_PORT` (default: `5432`)
  - `DB_USER` (default: `inventory`)
  - `DB_PASSWORD` (default: `password`)
  - `DB_NAME` (default: `inventory_db`)
  - `DB_SSLMODE` (default: `disable`)

- API
  - `API_PORT` (default: `8080`)
  - `API_READ_TIMEOUT` (default: `30s`)
  - `API_WRITE_TIMEOUT` (default: `30s`)
  - `API_IDLE_TIMEOUT` (default: `60s`)
  - `API_ENABLE_CORS` (default: `true`)
  - `API_ENABLE_METRICS` (default: `true`)

- 在庫設定
  - `INVENTORY_ALLOW_NEGATIVE_STOCK` (default: `false`)
  - `INVENTORY_DEFAULT_LOCATION` (default: `DEFAULT`)
  - `INVENTORY_AUDIT_ENABLED` (default: `true`)
  - `INVENTORY_LOW_STOCK_THRESHOLD` (default: `10`)
  - `INVENTORY_ALERT_TIMEOUT_HOURS` (default: `24`)

- ログ
  - `LOG_LEVEL` (default: `info`)
  - `LOG_FORMAT` (default: `json`)
  - `LOG_OUTPUT` (default: `stdout`)

---

## API エンドポイント

`cmd/api/main.go` のルーティングに基づく一覧です。未実装のものは注記します。

- ヘルス/メトリクス
  - GET `/health` ヘルスチェック
  - GET `/metrics` メトリクス（プレースホルダー）

- 在庫操作（POST）
  - `/api/v1/inventory/add` 在庫追加
  - `/api/v1/inventory/remove` 在庫削除
  - `/api/v1/inventory/transfer` 在庫移動
  - `/api/v1/inventory/adjust` 在庫調整
  - `/api/v1/inventory/batch` バッチ操作

- 在庫照会（GET）
  - `/api/v1/inventory/{itemId}/{locationId}` 在庫取得
  - `/api/v1/inventory/{itemId}/total` 総在庫取得
  - `/api/v1/inventory/location/{locationId}` ロケーション別在庫

- 履歴（GET）
  - `/api/v1/inventory/{itemId}/history?limit={n}` 履歴取得（`limit` 省略時 50）

- アラート
  - GET `/api/v1/alerts/{locationId}` アラート一覧
  - POST `/api/v1/alerts/{alertId}/resolve` アラート解決

- 商品・ロケーション（現在は未実装のスタブ）
  - POST `/api/v1/items` 商品作成（未実装）
  - GET `/api/v1/items/{itemId}` 商品取得（未実装）
  - PUT `/api/v1/items/{itemId}` 商品更新（未実装）
  - POST `/api/v1/locations` ロケーション作成（未実装）
  - GET `/api/v1/locations/{locationId}` ロケーション取得（未実装）

---

## リクエスト例（PowerShell）

1) 在庫追加

```powershell
$body = @{ item_id = "ITEM001"; location_id = "DEFAULT"; quantity = 100; reference = "API-TEST-001" } | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri "http://localhost:8080/api/v1/inventory/add" -ContentType "application/json" -Body $body
```

2) 在庫取得

```powershell
Invoke-RestMethod -Method Get -Uri "http://localhost:8080/api/v1/inventory/ITEM001/DEFAULT"
```

3) バッチ操作

```powershell
$payload = @(
  @{ type = "add"; item_id = "ITEM002"; location_id = "DEFAULT"; quantity = 50; reference = "BATCH-001" },
  @{ type = "adjust"; item_id = "ITEM001"; location_id = "DEFAULT"; quantity = 100; reference = "BATCH-002" }
) | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri "http://localhost:8080/api/v1/inventory/batch" -ContentType "application/json" -Body $payload
```

4) 履歴取得（最新5件）

```powershell
Invoke-RestMethod -Method Get -Uri "http://localhost:8080/api/v1/inventory/ITEM001/history?limit=5"
```

---

## サンプルの実行

- ローカルの PostgreSQL は `docker-compose up -d` で起動済みを前提とします。

1) 基本使用例（CLI）

```powershell
# 例: examples/basic_usage
go run .\examples\basic_usage\main.go
```

2) REST API クライアント例

```powershell
# 例: examples/api_client
go run .\examples\api_client\main.go
```

---

## ローカル開発（任意）

Docker を用いずに API を起動する場合:

```powershell
# 依存取得
go mod download

# 環境変数（必要に応じて）
$env:DB_HOST = "localhost"
$env:DB_PORT = "5432"
$env:DB_USER = "inventory"
$env:DB_PASSWORD = "password"
$env:DB_NAME = "inventory_db"
$env:API_PORT = "8080"

# API 起動
go run .\cmd\api\main.go
```

---

## トラブルシューティング

- ポート競合: `8080` や `5432` が使用中の場合、`docker-compose.yml` のポートや `API_PORT`/`DB_PORT` を変更。
- DB 初期化失敗: `docker-compose down -v` でボリューム削除後、再起動。`migrations/` の SQL を確認。
- 接続失敗: API コンテナから DB に到達できるか確認（`DB_HOST=postgres`）。API ログ（`docker-compose logs -f inventory-api`）と Postgres ログを確認。
- CORS: 開発用にワイルドカード許可（`cmd/api/main.go` のミドルウェア）。本番環境では適切に制限してください。

---

## 参考

- ルーティング: `cmd/api/main.go`
- ハンドラー: `cmd/api/handlers.go`
- 設定読み込み: `internal/config/config.go`
- DB初期化: `migrations/001_initial_schema.sql`
- 使用例: `examples/basic_usage/`, `examples/api_client/`
