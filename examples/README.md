# zaiGoFramework 使用例

このディレクトリには、zaiGoFrameworkの基本的な使用例が含まれています。

## 📁 ディレクトリ構成

```
examples/
├── basic_usage/     # 基本的なプログラム使用例
├── api_client/      # REST API クライアント例
└── README.md        # このファイル
```

## 🚀 実行方法

### 1. 開発環境の準備

```bash
# プロジェクトルートで実行
make setup
```

### 2. 基本使用例の実行

```bash
# 基本的なプログラム使用例
cd examples/basic_usage
go run main.go
```

### 3. API サーバーの起動

```bash
# 別のターミナルでAPIサーバーを起動
make docker-up

# または直接起動
make run
```

### 4. API クライアント例の実行

```bash
# APIサーバーが起動していることを確認してから実行
cd examples/api_client  
go run main.go
```

## 📋 使用例の内容

### basic_usage/main.go

プログラムから直接在庫管理ライブラリを使用する例：

- ✅ 在庫の追加・削除・移動
- ✅ 在庫の予約・解除
- ✅ バッチ操作
- ✅ トランザクション履歴の確認
- ✅ アラート機能

### api_client/main.go

REST APIを通じて在庫管理システムを操作する例：

- ✅ ヘルスチェック
- ✅ 在庫操作（追加・削除・移動）
- ✅ バッチ処理
- ✅ 履歴確認

## 🔧 環境変数

以下の環境変数を設定できます：

```bash
# データベース設定
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=inventory
export DB_PASSWORD=password
export DB_NAME=inventory_db

# API設定
export API_PORT=8080
export LOG_LEVEL=info

# 在庫設定
export INVENTORY_LOW_STOCK_THRESHOLD=10
export INVENTORY_ALLOW_NEGATIVE_STOCK=false
```

## 📊 サンプルデータ

例では以下のサンプルデータを使用します：

### 商品
- **ITEM-A**: ノートPC (¥80,000)
- **ITEM-B**: マウス (¥2,000)

### ロケーション
- **WAREHOUSE-01**: メイン倉庫 (東京都江東区)
- **WAREHOUSE-02**: サブ倉庫 (東京都大田区)

## 🛠️ トラブルシューティング

### データベース接続エラー

```bash
# PostgreSQLが起動していることを確認
make docker-up

# データベースの状態確認
make db-connect
```

### API接続エラー

```bash
# APIサーバーのヘルスチェック
make health-check

# ログの確認
make docker-logs
```

### 在庫データのリセット

```bash
# データベースをリセット
make db-reset
```

## 📚 詳細な使用方法

より詳細な使用方法については、プロジェクトルートの `README.md` および `docs/` ディレクトリを参照してください。

## 🤝 貢献

バグ報告や機能改善の提案は、GitHub Issues でお知らせください。
