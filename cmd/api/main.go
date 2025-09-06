package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/nemonet1337/zaiGoFramework/internal/config"
	"github.com/nemonet1337/zaiGoFramework/pkg/inventory"
	"github.com/nemonet1337/zaiGoFramework/pkg/inventory/storage"
)

func main() {
	// ログ設定
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("ログ初期化に失敗しました:", err)
	}
	defer logger.Sync()

	// 設定読み込み
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("設定読み込みに失敗しました", zap.Error(err))
	}

	// データベース接続
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
	)

	storage, err := storage.NewPostgreSQLStorage(dsn, logger)
	if err != nil {
		logger.Fatal("データベース接続に失敗しました", zap.Error(err))
	}
	defer storage.Close()

	// 在庫マネージャー初期化
	inventoryConfig := &inventory.Config{
		AllowNegativeStock: cfg.Inventory.AllowNegativeStock,
		DefaultLocation:    cfg.Inventory.DefaultLocation,
		AuditEnabled:       cfg.Inventory.AuditEnabled,
		LowStockThreshold:  cfg.Inventory.LowStockThreshold,
		AlertTimeout:       time.Duration(cfg.Inventory.AlertTimeoutHours) * time.Hour,
	}

	manager := inventory.NewManager(storage, nil, logger, inventoryConfig)

	// HTTPハンドラー設定
	handlers := NewHandlers(manager, logger)
	router := setupRouter(handlers)

	// HTTPサーバー設定
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.API.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// グレースフルシャットダウン設定
	go func() {
		logger.Info("在庫管理APIサーバーを開始します", zap.Int("port", cfg.API.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("サーバー開始に失敗しました", zap.Error(err))
		}
	}()

	// シャットダウンシグナル待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("サーバーをシャットダウンしています...")

	// グレースフルシャットダウン
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("サーバーシャットダウンに失敗しました", zap.Error(err))
	}

	logger.Info("サーバーが正常に停止しました")
}

// setupRouter sets up HTTP routes
// HTTPルートを設定
func setupRouter(handlers *Handlers) *mux.Router {
	router := mux.NewRouter()

	// ヘルスチェック
	router.HandleFunc("/health", handlers.HealthCheck).Methods("GET")
	router.HandleFunc("/metrics", handlers.Metrics).Methods("GET")

	// API v1ルート
	api := router.PathPrefix("/api/v1").Subrouter()

	// 在庫操作
	api.HandleFunc("/inventory/add", handlers.AddStock).Methods("POST")
	api.HandleFunc("/inventory/remove", handlers.RemoveStock).Methods("POST")
	api.HandleFunc("/inventory/transfer", handlers.TransferStock).Methods("POST")
	api.HandleFunc("/inventory/adjust", handlers.AdjustStock).Methods("POST")
	api.HandleFunc("/inventory/batch", handlers.BatchOperation).Methods("POST")

	// 在庫照会
	api.HandleFunc("/inventory/{itemId}/{locationId}", handlers.GetStock).Methods("GET")
	api.HandleFunc("/inventory/{itemId}/total", handlers.GetTotalStock).Methods("GET")
	api.HandleFunc("/inventory/location/{locationId}", handlers.GetStockByLocation).Methods("GET")

	// 履歴
	api.HandleFunc("/inventory/{itemId}/history", handlers.GetHistory).Methods("GET")

	// アラート
	api.HandleFunc("/alerts/{locationId}", handlers.GetAlerts).Methods("GET")
	api.HandleFunc("/alerts/{alertId}/resolve", handlers.ResolveAlert).Methods("POST")

	// 商品管理
	api.HandleFunc("/items", handlers.CreateItem).Methods("POST")
	api.HandleFunc("/items", handlers.ListItems).Methods("GET")
	api.HandleFunc("/items/search", handlers.SearchItems).Methods("GET")
	api.HandleFunc("/items/{itemId}", handlers.GetItem).Methods("GET")
	api.HandleFunc("/items/{itemId}", handlers.UpdateItem).Methods("PUT")
	api.HandleFunc("/items/{itemId}", handlers.DeleteItem).Methods("DELETE")

	// ロケーション管理
	api.HandleFunc("/locations", handlers.CreateLocation).Methods("POST")
	api.HandleFunc("/locations", handlers.ListLocations).Methods("GET")
	api.HandleFunc("/locations/{locationId}", handlers.GetLocation).Methods("GET")
	api.HandleFunc("/locations/{locationId}", handlers.UpdateLocation).Methods("PUT")
	api.HandleFunc("/locations/{locationId}", handlers.DeleteLocation).Methods("DELETE")

	// ロット管理
	api.HandleFunc("/lots", handlers.CreateLot).Methods("POST")
	api.HandleFunc("/lots/{lotId}", handlers.GetLot).Methods("GET")
	api.HandleFunc("/lots/item/{itemId}", handlers.GetLotsByItem).Methods("GET")
	api.HandleFunc("/lots/expiring", handlers.GetExpiringLots).Methods("GET")
	api.HandleFunc("/lots/expired", handlers.GetExpiredLots).Methods("GET")

	// 予約管理
	api.HandleFunc("/inventory/reserve", handlers.ReserveStock).Methods("POST")
	api.HandleFunc("/inventory/release-reservation", handlers.ReleaseReservation).Methods("POST")

	// 履歴管理（追加）
	api.HandleFunc("/inventory/history/location/{locationId}", handlers.GetHistoryByLocation).Methods("GET")
	api.HandleFunc("/inventory/{itemId}/history/date-range", handlers.GetHistoryByDateRange).Methods("GET")

	// バッチ管理（追加）
	api.HandleFunc("/inventory/batch/{batchId}/status", handlers.GetBatchStatus).Methods("GET")

	// 在庫評価エンジン
	api.HandleFunc("/valuation/{itemId}/{locationId}", handlers.CalculateValue).Methods("GET")
	api.HandleFunc("/valuation/total/{locationId}", handlers.CalculateTotalValue).Methods("GET")
	api.HandleFunc("/valuation/average-cost/{itemId}", handlers.GetAverageCost).Methods("GET")

	// 在庫分析エンジン
	api.HandleFunc("/analytics/abc/{locationId}", handlers.CalculateABCClassification).Methods("GET")
	api.HandleFunc("/analytics/turnover/{itemId}", handlers.GetTurnoverRate).Methods("GET")
	api.HandleFunc("/analytics/slow-moving/{locationId}", handlers.GetSlowMovingItems).Methods("GET")
	api.HandleFunc("/analytics/report/{locationId}", handlers.GenerateStockReport).Methods("GET")

	// CORS設定（開発用）
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// ログ機能
	router.Use(loggingMiddleware(handlers.logger))

	return router
}

// loggingMiddleware logs HTTP requests
// HTTPリクエストをログ出力するミドルウェア
func loggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// リクエスト処理
			next.ServeHTTP(w, r)

			// ログ出力
			logger.Info("HTTPリクエスト",
				zap.String("method", r.Method),
				zap.String("url", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}
