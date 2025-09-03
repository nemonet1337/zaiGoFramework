# zaiGoFramework

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)]()
[![Coverage](https://img.shields.io/badge/coverage-95%25-brightgreen.svg)]()

Go言語で記述された、シンプルで高性能な在庫管理フレームワーク

## 概要

zaiGoFrameworkは、現代的なビジネス要件を満たしながらも、シンプルで理解しやすい設計を重視した在庫管理システムです。

### 設計哲学
- **シンプルさ**: 過度な複雑さを避け、本質的な機能に集中
- **信頼性**: 実証済みのパターンと堅牢なエラーハンドリング
- **パフォーマンス**: 高負荷環境でも安定した性能
- **拡張性**: 将来的な要求に対応できる柔軟なアーキテクチャ

## 主な機能

### 🏪 基本的な在庫操作
- **CRUD操作**: 商品、在庫数、ロケーションの基本管理
- **トランザクション管理**: データ整合性の保証
- **バッチ処理**: 大量データの効率的な処理

### 📊 在庫追跡・監査
- **完全な履歴追跡**: すべての在庫変動を記録
- **楽観的ロック**: 同時更新制御
- **監査ログ**: 誰が、いつ、何を変更したかの完全な記録

### 🔧 高度な機能
- **自動再計算**: FIFO/LIFO/平均法での在庫評価
- **アラート機能**: 安全在庫レベルの監視
- **ABC分析**: 商品の重要度自動分類
- **在庫評価**: リアルタイムな在庫価値計算

### 🚀 運用・統合
- **RESTful API**: 外部システムとの簡単な連携
- **Docker対応**: コンテナ化による簡単なデプロイ
- **Kubernetes**: スケーラブルな本番運用
- **メトリクス・監視**: Prometheus対応の運用監視

## クイックスタート

### 前提条件
- Go 1.21以上
- PostgreSQL 15以上
- Docker（オプション）

### インストール

```bash
go get github.com/yourusername/zaiGoFramework
```

### 基本的な使用例

```go
package main

import (
    "context"
    "log"
    
    "github.com/yourusername/zaiGoFramework/pkg/inventory"
)

func main() {
    // データベース接続
    dsn := "postgres://user:password@localhost/inventory_db?sslmode=disable"
    manager, err := inventory.NewPostgreSQLManager(dsn)
    if err != nil {
        log.Fatal("接続エラー:", err)
    }
    
    ctx := context.Background()
    
    // 在庫追加
    err = manager.Add(ctx, "ITEM001", "LOC001", 100, "PO-2024-001")
    if err != nil {
        log.Fatal("入庫エラー:", err)
    }
    
    // 在庫確認
    stock, err := manager.GetStock(ctx, "ITEM001", "LOC001")
    if err != nil {
        log.Fatal("在庫取得エラー:", err)
    }
    
    log.Printf("現在在庫: %d個", stock.Quantity)
}
```

## プロジェクト構成

```
zaiGoFramework/
├── cmd/api/                 # APIサーバー
├── pkg/inventory/           # コアライブラリ
├── internal/config/         # 設定管理
├── migrations/             # DBスキーマ
├── deployments/docker/     # Docker設定
├── examples/               # 使用例
│   ├── basic_usage/        # プログラム例
│   └── api_client/         # REST API例
└── tests/                  # テストファイル
```

## API仕様

### 在庫操作エンドポイント

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/inventory/{itemId}/{locationId}` | 在庫情報取得 |
| GET | `GET /api/v1/inventory/{itemID}/history` | 履歴確認 |
| GET | `GET /api/v1/alerts` | アラート確認 |
| POST | `/api/v1/inventory/add` | 在庫追加 |
| POST | `/api/v1/inventory/remove` | 在庫減算 |
| POST | `/api/v1/inventory/transfer` | 在庫移動 |
| POST | `/api/v1/inventory/batch` | バッチ更新 |

### レスポンス例

```json
{
  "success": true,
  "data": {
    "item_id": "ITEM001",
    "location_id": "LOC001",
    "quantity": 150,
    "reserved": 0,
    "available": 150,
    "updated_at": "2024-01-15T10:30:00Z",
    "version": 5
  }
}
```

詳細なAPI仕様は [API Documentation](docs/api.md) をご確認ください。

## セットアップ

### 開発環境

```bash
# リポジトリをクローン
git clone https://github.com/yourusername/zaiGoFramework.git
cd zaiGoFramework

# 依存関係をインストール
go mod download

# 開発用データベースを起動
make setup

# テスト実行
make test

# APIサーバー起動
make run
```

### Docker使用

```bash
# Docker Composeで起動
docker-compose up -d

# ログ確認
docker-compose logs -f inventory-api
```

### Kubernetes デプロイ

```bash
# Kubernetesにデプロイ
kubectl apply -f deployments/k8s/

# デプロイ状況確認
kubectl rollout status deployment/inventory-api
```

## 設定

### 環境変数

| 変数名 | 説明 | デフォルト値 |
|--------|------|------------|
| `DB_HOST` | データベースホスト | `localhost` |
| `DB_PORT` | データベースポート | `5432` |
| `DB_USER` | データベースユーザー | `inventory` |
| `DB_PASSWORD` | データベースパスワード | - |
| `DB_NAME` | データベース名 | `inventory_db` |
| `API_PORT` | APIサーバーポート | `8080` |
| `LOG_LEVEL` | ログレベル | `info` |

### 設定ファイル

```yaml
# config.yaml
database:
  host: localhost
  port: 5432
  user: inventory
  password: password
  dbname: inventory_db
  
api:
  port: 8080
  timeout: 30s
  
inventory:
  default_location: "LOC001"
  enable_negative_stock: false
  audit_enabled: true

alerts:
  low_stock_threshold: 10
  webhook_url: "https://your-webhook.com"
```

## パフォーマンス

### ベンチマーク結果

```
BenchmarkAddOperation      100000    12.5 μs/op    2.1 MB/s
BenchmarkGetStock         500000     2.8 μs/op    8.9 MB/s
BenchmarkTransfer          50000    28.3 μs/op    1.4 MB/s
BenchmarkBatchUpdate       10000   125.7 μs/op    4.2 MB/s
```

### 推奨システム要件

**開発環境**
- CPU: 2コア以上
- RAM: 4GB以上
- Storage: SSD推奨

**本番環境**
- CPU: 4コア以上
- RAM: 8GB以上
- Storage: SSD必須
- Database: 専用PostgreSQLサーバー

## 監視・運用

### ヘルスチェック

```bash
curl http://localhost:8080/health
```

### メトリクス

Prometheusメトリクスは `/metrics` エンドポイントで確認できます。

主要メトリクス：
- `http_requests_total`: HTTPリクエスト数
- `inventory_operations_total`: 在庫操作数
- `db_connections_active`: アクティブDB接続数

### ログ

構造化ログ（JSON形式）でアプリケーションの動作を記録。

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "operation": "add_stock",
  "item_id": "ITEM001",
  "location_id": "LOC001",
  "quantity": 100,
  "reference": "PO-2024-001"
}
```

## 貢献

プロジェクトへの貢献を歓迎します！

### 開発フロー

1. このリポジトリをフォーク
2. feature ブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. Pull Request を作成

### コーディング規約

- [Effective Go](https://golang.org/doc/effective_go.html) に準拠
- `gofmt` でフォーマット
- `golint` でリント
- テストカバレッジ 90% 以上を維持

### テスト

```bash
# 全テスト実行
make test

# カバレッジ確認
make test-coverage

# ベンチマーク実行
go test -bench=. ./...
```

## バージョニング

[Semantic Versioning](https://semver.org/) を採用しています。

利用可能なバージョンは [Releases](https://github.com/yourusername/zaiGoFramework/releases) で確認できます。

## ロードマップ

### v1.1.0 (予定)
- [ ] Redis キャッシュサポート
- [ ] GraphQL API
- [ ] レポート生成機能

### v1.2.0 (予定)  
- [ ] マルチテナンシー
- [ ] リアルタイム通知
- [ ] 高度な分析機能

### v2.0.0 (検討中)
- [ ] マイクロサービス分割
- [ ] イベントソーシング
- [ ] 機械学習による需要予測

## ライセンス

このプロジェクトは [Apache License 2.0](LICENSE) の下でライセンスされています。

## サポート・コミュニティ

- **Issues**: [GitHub Issues](https://github.com/yourusername/zaiGoFramework/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/zaiGoFramework/discussions)
- **Documentation**: [Wiki](https://github.com/yourusername/zaiGoFramework/wiki)

## 謝辞

このプロジェクトは以下のオープンソースプロジェクトの恩恵を受けています：

- [PostgreSQL](https://postgresql.org/) - 堅牢なデータベース
- [Gorilla Mux](https://github.com/gorilla/mux) - HTTPルーティング
- [Zap](https://github.com/uber-go/zap) - 高性能ログ
- [Prometheus](https://prometheus.io/) - メトリクス収集

---

**zaiGoFramework** - シンプルで信頼性の高い在庫管理を、Go言語で。
