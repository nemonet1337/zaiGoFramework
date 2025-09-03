# zaiGoFramework Makefile

.PHONY: build test setup run clean docker-build docker-up docker-down

# 変数定義
APP_NAME=zai-inventory-api
DOCKER_IMAGE=zai-go-framework:latest
POSTGRES_CONTAINER=zai-postgres

# デフォルトターゲット
all: test build

# Goアプリケーションをビルド
build:
	@echo "Goアプリケーションをビルドしています..."
	go build -o bin/$(APP_NAME) ./cmd/api

# テスト実行
test:
	@echo "テストを実行しています..."
	go test -v ./...

# テストカバレッジ確認
test-coverage:
	@echo "テストカバレッジを確認しています..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "カバレッジレポート: coverage.html"

# 依存関係をダウンロード
deps:
	@echo "依存関係をダウンロードしています..."
	go mod download
	go mod tidy

# 開発環境セットアップ
setup: deps
	@echo "開発環境をセットアップしています..."
	docker-compose up -d postgres
	@echo "データベースの起動を待機しています..."
	timeout 30 && docker-compose exec postgres pg_isready -U inventory || true

# アプリケーション実行
run: build
	@echo "アプリケーションを実行しています..."
	./bin/$(APP_NAME)

# ローカル開発用（ホットリロード）
dev:
	@echo "開発モードで実行しています..."
	go run ./cmd/api

# クリーンアップ
clean:
	@echo "クリーンアップしています..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Dockerイメージをビルド
docker-build:
	@echo "Dockerイメージをビルドしています..."
	docker build -t $(DOCKER_IMAGE) -f Dockerfile .

# Docker Composeで全体を起動
docker-up:
	@echo "Docker環境を起動しています..."
	docker-compose up -d
	@echo "サービスが起動しました。ログ確認: make docker-logs"

# Docker環境を停止
docker-down:
	@echo "Docker環境を停止しています..."
	docker-compose down

# Docker環境を再起動
docker-restart: docker-down docker-up

# Dockerログを確認
docker-logs:
	docker-compose logs -f

# データベースマイグレーション実行
migrate:
	@echo "データベースマイグレーションを実行しています..."
	docker-compose exec postgres psql -U inventory -d inventory_db -f /docker-entrypoint-initdb.d/001_initial_schema.sql

# データベースに接続
db-connect:
	@echo "データベースに接続しています..."
	docker-compose exec postgres psql -U inventory -d inventory_db

# データベースをリセット
db-reset:
	@echo "データベースをリセットしています..."
	docker-compose exec postgres psql -U inventory -d inventory_db -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	make migrate

# APIのヘルスチェック
health-check:
	@echo "APIのヘルスチェックを実行しています..."
	curl -f http://localhost:8080/health || echo "API is not responding"

# コードフォーマット
fmt:
	@echo "コードをフォーマットしています..."
	go fmt ./...

# リント実行
lint:
	@echo "リントを実行しています..."
	golangci-lint run

# セキュリティチェック
security:
	@echo "セキュリティチェックを実行しています..."
	gosec ./...

# ベンチマーク実行
benchmark:
	@echo "ベンチマークを実行しています..."
	go test -bench=. -benchmem ./...

# すべてのチェックを実行
check: fmt lint test security
	@echo "すべてのチェックが完了しました"

# 本番ビルド
build-prod:
	@echo "本番用ビルドを実行しています..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o bin/$(APP_NAME) ./cmd/api

# ヘルプ
help:
	@echo "利用可能なコマンド:"
	@echo "  build          - アプリケーションをビルド"
	@echo "  test           - テストを実行"
	@echo "  test-coverage  - テストカバレッジを確認"
	@echo "  setup          - 開発環境をセットアップ"
	@echo "  run            - アプリケーションを実行"
	@echo "  dev            - 開発モードで実行"
	@echo "  clean          - クリーンアップ"
	@echo "  docker-build   - Dockerイメージをビルド"
	@echo "  docker-up      - Docker環境を起動"
	@echo "  docker-down    - Docker環境を停止"
	@echo "  docker-logs    - Dockerログを確認"
	@echo "  migrate        - データベースマイグレーション実行"
	@echo "  db-connect     - データベースに接続"
	@echo "  health-check   - APIのヘルスチェック"
	@echo "  fmt            - コードをフォーマット"
	@echo "  lint           - リントを実行"
	@echo "  benchmark      - ベンチマークを実行"
	@echo "  check          - すべてのチェックを実行"
