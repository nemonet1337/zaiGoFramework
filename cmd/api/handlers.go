package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/nemonet1337/zaiGoFramework/pkg/inventory"
)

// Handlers holds HTTP handlers for the inventory API
// 在庫API用のHTTPハンドラーを保持
type Handlers struct {
	manager inventory.InventoryManager
	logger  *zap.Logger
}

// NewHandlers creates new HTTP handlers
// 新しいHTTPハンドラーを作成
func NewHandlers(manager inventory.InventoryManager, logger *zap.Logger) *Handlers {
	return &Handlers{
		manager: manager,
		logger:  logger,
	}
}

// APIResponse represents standard API response format
// 標準的なAPIレスポンス形式を表現
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// AddStockRequest represents request to add stock
// 在庫追加リクエストを表現
type AddStockRequest struct {
	ItemID     string `json:"item_id"`
	LocationID string `json:"location_id"`
	Quantity   int64  `json:"quantity"`
	Reference  string `json:"reference"`
}

// RemoveStockRequest represents request to remove stock
// 在庫削除リクエストを表現
type RemoveStockRequest struct {
	ItemID     string `json:"item_id"`
	LocationID string `json:"location_id"`
	Quantity   int64  `json:"quantity"`
	Reference  string `json:"reference"`
}

// TransferStockRequest represents request to transfer stock
// 在庫移動リクエストを表現
type TransferStockRequest struct {
	ItemID         string `json:"item_id"`
	FromLocationID string `json:"from_location_id"`
	ToLocationID   string `json:"to_location_id"`
	Quantity       int64  `json:"quantity"`
	Reference      string `json:"reference"`
}

// AdjustStockRequest represents request to adjust stock
// 在庫調整リクエストを表現
type AdjustStockRequest struct {
	ItemID      string `json:"item_id"`
	LocationID  string `json:"location_id"`
	NewQuantity int64  `json:"new_quantity"`
	Reference   string `json:"reference"`
}

// HealthCheck handles health check requests
// ヘルスチェックリクエストを処理
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now(),
			"service":   "zaiGoFramework",
		},
	}
	
	json.NewEncoder(w).Encode(response)
}

// Metrics handles metrics requests (placeholder)
// メトリクスリクエストを処理（プレースホルダー）
func (h *Handlers) Metrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("# メトリクス機能は後で実装予定\n"))
}

// AddStock handles add stock requests
// 在庫追加リクエストを処理
func (h *Handlers) AddStock(w http.ResponseWriter, r *http.Request) {
	var req AddStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "無効なリクエスト形式です")
		return
	}

	ctx := context.WithValue(r.Context(), "user_id", "api_user")
	if err := h.manager.Add(ctx, req.ItemID, req.LocationID, req.Quantity, req.Reference); err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccess(w, map[string]string{
		"message": "在庫追加が完了しました",
	})
}

// RemoveStock handles remove stock requests
// 在庫削除リクエストを処理
func (h *Handlers) RemoveStock(w http.ResponseWriter, r *http.Request) {
	var req RemoveStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "無効なリクエスト形式です")
		return
	}

	ctx := context.WithValue(r.Context(), "user_id", "api_user")
	if err := h.manager.Remove(ctx, req.ItemID, req.LocationID, req.Quantity, req.Reference); err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccess(w, map[string]string{
		"message": "在庫削除が完了しました",
	})
}

// TransferStock handles transfer stock requests
// 在庫移動リクエストを処理
func (h *Handlers) TransferStock(w http.ResponseWriter, r *http.Request) {
	var req TransferStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "無効なリクエスト形式です")
		return
	}

	ctx := context.WithValue(r.Context(), "user_id", "api_user")
	if err := h.manager.Transfer(ctx, req.ItemID, req.FromLocationID, req.ToLocationID, req.Quantity, req.Reference); err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccess(w, map[string]string{
		"message": "在庫移動が完了しました",
	})
}

// AdjustStock handles adjust stock requests
// 在庫調整リクエストを処理
func (h *Handlers) AdjustStock(w http.ResponseWriter, r *http.Request) {
	var req AdjustStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "無効なリクエスト形式です")
		return
	}

	ctx := context.WithValue(r.Context(), "user_id", "api_user")
	if err := h.manager.Adjust(ctx, req.ItemID, req.LocationID, req.NewQuantity, req.Reference); err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccess(w, map[string]string{
		"message": "在庫調整が完了しました",
	})
}

// BatchOperation handles batch operations
// バッチ操作を処理
func (h *Handlers) BatchOperation(w http.ResponseWriter, r *http.Request) {
	var operations []inventory.InventoryOperation
	if err := json.NewDecoder(r.Body).Decode(&operations); err != nil {
		h.sendError(w, http.StatusBadRequest, "無効なリクエスト形式です")
		return
	}

	ctx := context.WithValue(r.Context(), "user_id", "api_user")
	batch, err := h.manager.ExecuteBatch(ctx, operations)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccess(w, batch)
}

// GetStock handles get stock requests
// 在庫取得リクエストを処理
func (h *Handlers) GetStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["itemId"]
	locationID := vars["locationId"]

	stock, err := h.manager.GetStock(r.Context(), itemID, locationID)
	if err != nil {
		if err == inventory.ErrStockNotFound {
			h.sendError(w, http.StatusNotFound, "在庫が見つかりません")
		} else {
			h.sendError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.sendSuccess(w, stock)
}

// GetTotalStock handles get total stock requests
// 総在庫取得リクエストを処理
func (h *Handlers) GetTotalStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["itemId"]

	total, err := h.manager.GetTotalStock(r.Context(), itemID)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccess(w, map[string]int64{
		"total_quantity": total,
	})
}

// GetStockByLocation handles get stock by location requests
// ロケーション別在庫取得リクエストを処理
func (h *Handlers) GetStockByLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	locationID := vars["locationId"]

	stocks, err := h.manager.GetStockByLocation(r.Context(), locationID)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccess(w, stocks)
}

// GetHistory handles get history requests
// 履歴取得リクエストを処理
func (h *Handlers) GetHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["itemId"]

	// limitパラメータの取得
	limit := 50 // デフォルト
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	history, err := h.manager.GetHistory(r.Context(), itemID, limit)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccess(w, history)
}

// GetAlerts handles get alerts requests
// アラート取得リクエストを処理
func (h *Handlers) GetAlerts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	locationID := vars["locationId"]

	alerts, err := h.manager.GetAlerts(r.Context(), locationID)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccess(w, alerts)
}

// ResolveAlert handles resolve alert requests
// アラート解決リクエストを処理
func (h *Handlers) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alertID := vars["alertId"]

	if err := h.manager.ResolveAlert(r.Context(), alertID); err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccess(w, map[string]string{
		"message": "アラートが解決されました",
	})
}

// CreateItem handles create item requests
// 商品作成リクエストを処理
func (h *Handlers) CreateItem(w http.ResponseWriter, r *http.Request) {
	var item inventory.Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		h.sendError(w, http.StatusBadRequest, "無効なリクエスト形式です")
		return
	}

	// タイムスタンプ設定
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	// ストレージ直接アクセス（簡略化のため）
	// 実際の実装では適切なマネージャーインターフェースを使用
	h.sendError(w, http.StatusNotImplemented, "商品作成機能は未実装です")
}

// GetItem handles get item requests
// 商品取得リクエストを処理
func (h *Handlers) GetItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["itemId"]

	// ストレージ直接アクセス（簡略化のため）
	// 実際の実装では適切なマネージャーインターフェースを使用
	_ = itemID
	h.sendError(w, http.StatusNotImplemented, "商品取得機能は未実装です")
}

// UpdateItem handles update item requests
// 商品更新リクエストを処理
func (h *Handlers) UpdateItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["itemId"]

	var item inventory.Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		h.sendError(w, http.StatusBadRequest, "無効なリクエスト形式です")
		return
	}

	item.ID = itemID
	item.UpdatedAt = time.Now()

	// ストレージ直接アクセス（簡略化のため）
	h.sendError(w, http.StatusNotImplemented, "商品更新機能は未実装です")
}

// CreateLocation handles create location requests
// ロケーション作成リクエストを処理
func (h *Handlers) CreateLocation(w http.ResponseWriter, r *http.Request) {
	var location inventory.Location
	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		h.sendError(w, http.StatusBadRequest, "無効なリクエスト形式です")
		return
	}

	// タイムスタンプ設定
	now := time.Now()
	location.CreatedAt = now
	location.UpdatedAt = now

	h.sendError(w, http.StatusNotImplemented, "ロケーション作成機能は未実装です")
}

// GetLocation handles get location requests
// ロケーション取得リクエストを処理
func (h *Handlers) GetLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	locationID := vars["locationId"]

	_ = locationID
	h.sendError(w, http.StatusNotImplemented, "ロケーション取得機能は未実装です")
}

// ヘルパーメソッド

// sendSuccess sends a successful API response
// 成功APIレスポンスを送信
func (h *Handlers) sendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := APIResponse{
		Success: true,
		Data:    data,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("レスポンス送信に失敗しました", zap.Error(err))
	}
}

// sendError sends an error API response
// エラーAPIレスポンスを送信
func (h *Handlers) sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := APIResponse{
		Success: false,
		Error:   message,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("エラーレスポンス送信に失敗しました", zap.Error(err))
	}
}
