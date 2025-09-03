# マルチステージビルドでGoアプリケーションを最適化
FROM golang:1.21-alpine AS builder

# 必要なパッケージをインストール
RUN apk add --no-cache git ca-certificates

# 作業ディレクトリを設定
WORKDIR /app

# Go modulesファイルをコピーして依存関係をダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# アプリケーションをビルド
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# 本番用の軽量イメージ
FROM alpine:latest

# セキュリティ更新とCA証明書をインストール
RUN apk --no-cache add ca-certificates tzdata

# 作業ディレクトリを作成
WORKDIR /root/

# ビルドしたバイナリをコピー
COPY --from=builder /app/main .

# ログディレクトリを作成
RUN mkdir -p /app/logs

# ポート8080を公開
EXPOSE 8080

# ヘルスチェック設定
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# アプリケーションを実行
CMD ["./main"]
