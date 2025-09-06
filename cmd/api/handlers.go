package main

import (
	"context"
	"encoding/json"
	"fmt"
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

	// ID生成（もしIDが指定されていない場合）
	if item.ID == "" {
		item.ID = inventory.NewTransactionID()
	}

	// ItemManagerを使用して商品を作成
	if itemManager, ok := h.manager.(inventory.ItemManager); ok {
		if err := itemManager.CreateItem(r.Context(), &item); err != nil {
			h.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		h.sendError(w, http.StatusNotImplemented, "商品管理機能がサポートされていません")
		return
	}

	h.sendSuccess(w, map[string]interface{}{
		"message": "商品が作成されました",
		"item":    item,
	})
}

// GetItem handles get item requests
// 商品取得リクエストを処理
func (h *Handlers) GetItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["itemId"]

	// ItemManagerを使用して商品を取得
	if itemManager, ok := h.manager.(inventory.ItemManager); ok {
		item, err := itemManager.GetItem(r.Context(), itemID)
		if err != nil {
			if err == inventory.ErrItemNotFound {
				h.sendError(w, http.StatusNotFound, "商品が見つかりません")
			} else {
				h.sendError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		h.sendSuccess(w, item)
	} else {
		h.sendError(w, http.StatusNotImplemented, "商品管理機能がサポートされていません")
	}
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

	// ItemManagerを使用して商品を更新
	if itemManager, ok := h.manager.(inventory.ItemManager); ok {
		if err := itemManager.UpdateItem(r.Context(), &item); err != nil {
			if err == inventory.ErrItemNotFound {
				h.sendError(w, http.StatusNotFound, "商品が見つかりません")
			} else {
				h.sendError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		h.sendSuccess(w, map[string]interface{}{
			"message": "商品が更新されました",
			"item":    item,
		})
	} else {
		h.sendError(w, http.StatusNotImplemented, "商品管理機能がサポートされていません")
	}
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

	// ID生成（もしIDが指定されていない場合）
	if location.ID == "" {
		location.ID = inventory.NewTransactionID()
	}

	// LocationManagerを使用してロケーションを作成
	if locationManager, ok := h.manager.(inventory.LocationManager); ok {
		if err := locationManager.CreateLocation(r.Context(), &location); err != nil {
			h.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		h.sendError(w, http.StatusNotImplemented, "ロケーション管理機能がサポートされていません")
		return
	}

	h.sendSuccess(w, map[string]interface{}{
		"message":  "ロケーションが作成されました",
		"location": location,
	})
}

// GetLocation handles get location requests
// ロケーション取得リクエストを処理
func (h *Handlers) GetLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	locationID := vars["locationId"]

	// LocationManagerを使用してロケーションを取得
	if locationManager, ok := h.manager.(inventory.LocationManager); ok {
		location, err := locationManager.GetLocation(r.Context(), locationID)
		if err != nil {
			if err == inventory.ErrLocationNotFound {
				h.sendError(w, http.StatusNotFound, "ロケーションが見つかりません")
			} else {
				h.sendError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		h.sendSuccess(w, location)
	} else {
		h.sendError(w, http.StatusNotImplemented, "ロケーション管理機能がサポートされていません")
	}
}

// DeleteItem handles delete item requests
// 商品削除リクエストを処理
func (h *Handlers) DeleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["itemId"]

	// ItemManagerを使用して商品を削除
	if itemManager, ok := h.manager.(inventory.ItemManager); ok {
		if err := itemManager.DeleteItem(r.Context(), itemID); err != nil {
			if err == inventory.ErrItemNotFound {
				h.sendError(w, http.StatusNotFound, "商品が見つかりません")
			} else {
				h.sendError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		h.sendSuccess(w, map[string]string{
			"message": "商品が削除されました",
		})
	} else {
		h.sendError(w, http.StatusNotImplemented, "商品管理機能がサポートされていません")
	}
}

// ListItems handles list items requests
// 商品一覧リクエストを処理
func (h *Handlers) ListItems(w http.ResponseWriter, r *http.Request) {
	// offsetとlimitのパラメータを取得
	offset := 0
	limit := 20 // デフォルト

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	// ItemManagerを使用して商品一覧を取得
	if itemManager, ok := h.manager.(inventory.ItemManager); ok {
		items, err := itemManager.ListItems(r.Context(), offset, limit)
		if err != nil {
			h.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
		h.sendSuccess(w, map[string]interface{}{
			"items":  items,
			"offset": offset,
			"limit":  limit,
			"count":  len(items),
		})
	} else {
		h.sendError(w, http.StatusNotImplemented, "商品管理機能がサポートされていません")
	}
}

// SearchItems handles search items requests
// 商品検索リクエストを処理
func (h *Handlers) SearchItems(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		h.sendError(w, http.StatusBadRequest, "検索クエリが指定されていません")
		return
	}

	// ItemManagerを使用して商品を検索
	if itemManager, ok := h.manager.(inventory.ItemManager); ok {
		items, err := itemManager.SearchItems(r.Context(), query)
		if err != nil {
			h.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
		h.sendSuccess(w, map[string]interface{}{
			"items": items,
			"query": query,
			"count": len(items),
		})
	} else {
		h.sendError(w, http.StatusNotImplemented, "商品管理機能がサポートされていません")
	}
}

// UpdateLocation handles update location requests
// ロケーション更新リクエストを処理
func (h *Handlers) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	locationID := vars["locationId"]

	var location inventory.Location
	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		h.sendError(w, http.StatusBadRequest, "無効なリクエスト形式です")
		return
	}

	location.ID = locationID
	location.UpdatedAt = time.Now()

	// LocationManagerを使用してロケーションを更新
	if locationManager, ok := h.manager.(inventory.LocationManager); ok {
		if err := locationManager.UpdateLocation(r.Context(), &location); err != nil {
			if err == inventory.ErrLocationNotFound {
				h.sendError(w, http.StatusNotFound, "ロケーションが見つかりません")
			} else {
				h.sendError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		h.sendSuccess(w, map[string]interface{}{
			"message":  "ロケーションが更新されました",
			"location": location,
		})
	} else {
		h.sendError(w, http.StatusNotImplemented, "ロケーション管理機能がサポートされていません")
	}
}

// DeleteLocation handles delete location requests
// ロケーション削除リクエストを処理
func (h *Handlers) DeleteLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	locationID := vars["locationId"]

	// LocationManagerを使用してロケーションを削除
	if locationManager, ok := h.manager.(inventory.LocationManager); ok {
		if err := locationManager.DeleteLocation(r.Context(), locationID); err != nil {
			if err == inventory.ErrLocationNotFound {
				h.sendError(w, http.StatusNotFound, "ロケーションが見つかりません")
			} else {
				h.sendError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		h.sendSuccess(w, map[string]string{
			"message": "ロケーションが削除されました",
		})
	} else {
		h.sendError(w, http.StatusNotImplemented, "ロケーション管理機能がサポートされていません")
	}
}

// ListLocations handles list locations requests
// ロケーション一覧リクエストを処理
func (h *Handlers) ListLocations(w http.ResponseWriter, r *http.Request) {
	// offsetとlimitのパラメータを取得
	offset := 0
	limit := 20 // デフォルト

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	// LocationManagerを使用してロケーション一覧を取得
	if locationManager, ok := h.manager.(inventory.LocationManager); ok {
		locations, err := locationManager.ListLocations(r.Context(), offset, limit)
		if err != nil {
			h.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
		h.sendSuccess(w, map[string]interface{}{
			"locations": locations,
			"offset":    offset,
			"limit":     limit,
			"count":     len(locations),
		})
	} else {
		h.sendError(w, http.StatusNotImplemented, "ロケーション管理機能がサポートされていません")
	}
}

// ロット管理ハンドラー

// CreateLot handles create lot requests
// ロット作成リクエストを処理
func (h *Handlers) CreateLot(w http.ResponseWriter, r *http.Request) {
	var lot inventory.Lot
	if err := json.NewDecoder(r.Body).Decode(&lot); err != nil {
		h.sendError(w, http.StatusBadRequest, "無効なリクエスト形式です")
		return
	}

	// タイムスタンプ設定
	lot.CreatedAt = time.Now()

	// ID生成（もしIDが指定されていない場合）
	if lot.ID == "" {
		lot.ID = inventory.NewTransactionID()
	}

	// LotManagerを使用してロットを作成
	if lotManager, ok := h.manager.(inventory.LotManager); ok {
		if err := lotManager.CreateLot(r.Context(), &lot); err != nil {
			h.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		h.sendError(w, http.StatusNotImplemented, "ロット管理機能がサポートされていません")
		return
	}

	h.sendSuccess(w, map[string]interface{}{
		"message": "ロットが作成されました",
		"lot":     lot,
	})
}

// GetLot handles get lot requests
// ロット取得リクエストを処理
func (h *Handlers) GetLot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lotID := vars["lotId"]

	// LotManagerを使用してロットを取得
	if lotManager, ok := h.manager.(inventory.LotManager); ok {
		lot, err := lotManager.GetLot(r.Context(), lotID)
		if err != nil {
			h.sendError(w, http.StatusNotFound, "ロットが見つかりません")
			return
		}
		h.sendSuccess(w, lot)
	} else {
		h.sendError(w, http.StatusNotImplemented, "ロット管理機能がサポートされていません")
	}
}

// GetLotsByItem handles get lots by item requests
// 商品別ロット取得リクエストを処理
func (h *Handlers) GetLotsByItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["itemId"]

	// LotManagerを使用して商品のロット一覧を取得
	if lotManager, ok := h.manager.(inventory.LotManager); ok {
		lots, err := lotManager.GetLotsByItem(r.Context(), itemID)
		if err != nil {
			h.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
		h.sendSuccess(w, map[string]interface{}{
			"lots":    lots,
			"item_id": itemID,
			"count":   len(lots),
		})
	} else {
		h.sendError(w, http.StatusNotImplemented, "ロット管理機能がサポートされていません")
	}
}

// GetExpiringLots handles get expiring lots requests
// 期限切れ間近ロット取得リクエストを処理
func (h *Handlers) GetExpiringLots(w http.ResponseWriter, r *http.Request) {
	// within パラメータを取得（日数）
	withinDays := 7 // デフォルト7日
	if withinStr := r.URL.Query().Get("within_days"); withinStr != "" {
		if parsedDays, err := strconv.Atoi(withinStr); err == nil && parsedDays > 0 {
			withinDays = parsedDays
		}
	}

	within := time.Duration(withinDays) * 24 * time.Hour

	// LotManagerを使用して期限切れ間近ロットを取得
	if lotManager, ok := h.manager.(inventory.LotManager); ok {
		lots, err := lotManager.GetExpiringLots(r.Context(), within)
		if err != nil {
			h.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
		h.sendSuccess(w, map[string]interface{}{
			"lots":        lots,
			"within_days": withinDays,
			"count":       len(lots),
		})
	} else {
		h.sendError(w, http.StatusNotImplemented, "ロット管理機能がサポートされていません")
	}
}

// GetExpiredLots handles get expired lots requests
// 期限切れロット取得リクエストを処理
func (h *Handlers) GetExpiredLots(w http.ResponseWriter, r *http.Request) {
	// LotManagerを使用して期限切れロットを取得
	if lotManager, ok := h.manager.(inventory.LotManager); ok {
		lots, err := lotManager.GetExpiredLots(r.Context())
		if err != nil {
			h.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
		h.sendSuccess(w, map[string]interface{}{
			"lots":  lots,
			"count": len(lots),
		})
	} else {
		h.sendError(w, http.StatusNotImplemented, "ロット管理機能がサポートされていません")
	}
}

// 予約管理ハンドラー

// ReserveStock handles reserve stock requests
// 在庫予約リクエストを処理
func (h *Handlers) ReserveStock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ItemID     string `json:"item_id"`
		LocationID string `json:"location_id"`
		Quantity   int64  `json:"quantity"`
		Reference  string `json:"reference"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "無効なリクエスト形式です")
		return
	}

	ctx := context.WithValue(r.Context(), "user_id", "api_user")
	if err := h.manager.Reserve(ctx, req.ItemID, req.LocationID, req.Quantity, req.Reference); err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccess(w, map[string]string{
		"message": "在庫が予約されました",
	})
}

// ReleaseReservation handles release reservation requests
// 予約解除リクエストを処理
func (h *Handlers) ReleaseReservation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ItemID     string `json:"item_id"`
		LocationID string `json:"location_id"`
		Quantity   int64  `json:"quantity"`
		Reference  string `json:"reference"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "無効なリクエスト形式です")
		return
	}

	ctx := context.WithValue(r.Context(), "user_id", "api_user")
	if err := h.manager.ReleaseReservation(ctx, req.ItemID, req.LocationID, req.Quantity, req.Reference); err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccess(w, map[string]string{
		"message": "予約が解除されました",
	})
}

// 履歴管理の追加ハンドラー

// GetHistoryByLocation handles get history by location requests
// ロケーション別履歴取得リクエストを処理
func (h *Handlers) GetHistoryByLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	locationID := vars["locationId"]

	// limitパラメータの取得
	limit := 50 // デフォルト
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	history, err := h.manager.GetHistoryByLocation(r.Context(), locationID, limit)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccess(w, map[string]interface{}{
		"history":     history,
		"location_id": locationID,
		"limit":       limit,
		"count":       len(history),
	})
}

// GetHistoryByDateRange handles get history by date range requests
// 日付範囲別履歴取得リクエストを処理
func (h *Handlers) GetHistoryByDateRange(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["itemId"]

	// 日付パラメータを取得
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		h.sendError(w, http.StatusBadRequest, "from及びtoパラメータが必要です（形式：2006-01-02）")
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "無効なfrom日付形式です（形式：2006-01-02）")
		return
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "無効なto日付形式です（形式：2006-01-02）")
		return
	}

	// 終了日を23:59:59に設定
	to = to.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	history, err := h.manager.GetHistoryByDateRange(r.Context(), itemID, from, to)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccess(w, map[string]interface{}{
		"history": history,
		"item_id": itemID,
		"from":    fromStr,
		"to":      toStr,
		"count":   len(history),
	})
}

// バッチ管理の追加ハンドラー

// GetBatchStatus handles get batch status requests
// バッチステータス取得リクエストを処理
func (h *Handlers) GetBatchStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	batchID := vars["batchId"]

	batch, err := h.manager.GetBatchStatus(r.Context(), batchID)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccess(w, batch)
}

// 在庫評価エンジンハンドラー

// CalculateValue handles calculate inventory value requests
// 在庫評価計算リクエストを処理
func (h *Handlers) CalculateValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["itemId"]
	locationID := vars["locationId"]

	// 評価方法を取得
	methodStr := r.URL.Query().Get("method")
	if methodStr == "" {
		methodStr = string(inventory.ValuationMethodFIFO) // デフォルト
	}

	method := inventory.ValuationMethod(methodStr)

	// ValuationEngineを使用して在庫評価を計算
	if valuationEngine, ok := h.manager.(inventory.ValuationEngine); ok {
		value, err := valuationEngine.CalculateValue(r.Context(), itemID, locationID, method)
		if err != nil {
			h.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
		h.sendSuccess(w, map[string]interface{}{
			"value":       value,
			"item_id":    itemID,
			"location_id": locationID,
			"method":      method,
		})
	} else {
		h.sendError(w, http.StatusNotImplemented, "在庫評価機能がサポートされていません")
	}
}

// CalculateTotalValue handles calculate total inventory value requests
// 全体在庫評価計算リクエストを処理
func (h *Handlers) CalculateTotalValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	locationID := vars["locationId"]

	// 評価方法を取得
	methodStr := r.URL.Query().Get("method")
	if methodStr == "" {
		methodStr = string(inventory.ValuationMethodFIFO) // デフォルト
	}

	method := inventory.ValuationMethod(methodStr)

	// ValuationEngineを使用して全体在庫評価を計算
	if valuationEngine, ok := h.manager.(inventory.ValuationEngine); ok {
		totalValue, err := valuationEngine.CalculateTotalValue(r.Context(), locationID, method)
		if err != nil {
			h.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
		h.sendSuccess(w, map[string]interface{}{
			"total_value": totalValue,
			"location_id": locationID,
			"method":      method,
		})
	} else {
		h.sendError(w, http.StatusNotImplemented, "在庫評価機能がサポートされていません")
	}
}

// GetAverageCost handles get average cost requests
// 平均原価取得リクエストを処理
func (h *Handlers) GetAverageCost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["itemId"]

	// ValuationEngineを使用して平均原価を取得
	if valuationEngine, ok := h.manager.(inventory.ValuationEngine); ok {
		avgCost, err := valuationEngine.GetAverageCost(r.Context(), itemID)
		if err != nil {
			h.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
		h.sendSuccess(w, map[string]interface{}{
			"average_cost": avgCost,
			"item_id":      itemID,
		})
	} else {
		h.sendError(w, http.StatusNotImplemented, "在庫評価機能がサポートされていません")
	}
}

// 在庫分析エンジンハンドラー

// CalculateABCClassification handles ABC classification requests
// ABC分析リクエストを処理
func (h *Handlers) CalculateABCClassification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	locationID := vars["locationId"]

	// AnalyticsEngineを使用してABC分析を実行
	if analyticsEngine, ok := h.manager.(inventory.AnalyticsEngine); ok {
		classification, err := analyticsEngine.CalculateABCClassification(r.Context(), locationID)
		if err != nil {
			h.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
		h.sendSuccess(w, map[string]interface{}{
			"classification": classification,
			"location_id":    locationID,
			"count":          len(classification),
		})
	} else {
		h.sendError(w, http.StatusNotImplemented, "在庫分析機能がサポートされていません")
	}
}

// GetTurnoverRate handles turnover rate requests
// 回転率取得リクエストを処理
func (h *Handlers) GetTurnoverRate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["itemId"]

	// 期間パラメータを取得（日数）
	periodDays := 30 // デフォルト30日
	if periodStr := r.URL.Query().Get("period_days"); periodStr != "" {
		if parsedDays, err := strconv.Atoi(periodStr); err == nil && parsedDays > 0 {
			periodDays = parsedDays
		}
	}

	period := time.Duration(periodDays) * 24 * time.Hour

	// AnalyticsEngineを使用して回転率を取得
	if analyticsEngine, ok := h.manager.(inventory.AnalyticsEngine); ok {
		turnoverRate, err := analyticsEngine.GetTurnoverRate(r.Context(), itemID, period)
		if err != nil {
			h.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
		h.sendSuccess(w, map[string]interface{}{
			"turnover_rate": turnoverRate,
			"item_id":       itemID,
			"period_days":   periodDays,
		})
	} else {
		h.sendError(w, http.StatusNotImplemented, "在庫分析機能がサポートされていません")
	}
}

// GetSlowMovingItems handles slow moving items requests
// 低回転商品取得リクエストを処理
func (h *Handlers) GetSlowMovingItems(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	locationID := vars["locationId"]

	// 闾値パラメータを取得（日数）
	thresholdDays := 90 // デフォルト90日
	if thresholdStr := r.URL.Query().Get("threshold_days"); thresholdStr != "" {
		if parsedDays, err := strconv.Atoi(thresholdStr); err == nil && parsedDays > 0 {
			thresholdDays = parsedDays
		}
	}

	threshold := time.Duration(thresholdDays) * 24 * time.Hour

	// AnalyticsEngineを使用して低回転商品を取得
	if analyticsEngine, ok := h.manager.(inventory.AnalyticsEngine); ok {
		slowMovingItems, err := analyticsEngine.GetSlowMovingItems(r.Context(), locationID, threshold)
		if err != nil {
			h.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
		h.sendSuccess(w, map[string]interface{}{
			"slow_moving_items": slowMovingItems,
			"location_id":       locationID,
			"threshold_days":    thresholdDays,
			"count":             len(slowMovingItems),
		})
	} else {
		h.sendError(w, http.StatusNotImplemented, "在庫分析機能がサポートされていません")
	}
}

// GenerateStockReport handles stock report generation requests
// 在庫レポート生成リクエストを処理
func (h *Handlers) GenerateStockReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	locationID := vars["locationId"]

	// レポートタイプを取得
	reportTypeStr := r.URL.Query().Get("type")
	if reportTypeStr == "" {
		reportTypeStr = string(inventory.ReportTypeStock) // デフォルト
	}

	reportType := inventory.ReportType(reportTypeStr)

	// AnalyticsEngineを使用してレポートを生成
	if analyticsEngine, ok := h.manager.(inventory.AnalyticsEngine); ok {
		reportData, err := analyticsEngine.GenerateStockReport(r.Context(), locationID, reportType)
		if err != nil {
			h.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// レポートをバイナリデータとして返す
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=stock_report_%s_%s.pdf", locationID, reportType))
		w.WriteHeader(http.StatusOK)
		w.Write(reportData)
	} else {
		h.sendError(w, http.StatusNotImplemented, "在庫分析機能がサポートされていません")
	}
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
