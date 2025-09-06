package inventory

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Manager implements the InventoryManager interface
// InventoryManagerインターフェースの実装
type Manager struct {
	storage   Storage         // ストレージ層
	publisher EventPublisher  // イベント発行者
	logger    *zap.Logger     // ログ
	config    *Config         // 設定
}

// すべてのインターフェースを実装することを明示
var (
	_ InventoryManager = (*Manager)(nil)
	_ ItemManager     = (*Manager)(nil)
	_ LocationManager = (*Manager)(nil)
	_ LotManager      = (*Manager)(nil)
)

// Config holds configuration for the inventory manager
// 在庫マネージャーの設定を保持
type Config struct {
	AllowNegativeStock bool          `yaml:"allow_negative_stock"` // 負の在庫を許可
	DefaultLocation    string        `yaml:"default_location"`     // デフォルトロケーション
	AuditEnabled       bool          `yaml:"audit_enabled"`        // 監査ログ有効
	LowStockThreshold  int64         `yaml:"low_stock_threshold"`  // 低在庫閾値
	AlertTimeout       time.Duration `yaml:"alert_timeout"`        // アラートタイムアウト
}

// NewManager creates a new inventory manager
// 新しい在庫マネージャーを作成
func NewManager(storage Storage, publisher EventPublisher, logger *zap.Logger, config *Config) *Manager {
	if config == nil {
		config = &Config{
			AllowNegativeStock: false,
			DefaultLocation:    "DEFAULT",
			AuditEnabled:       true,
			LowStockThreshold:  10,
			AlertTimeout:       time.Hour * 24,
		}
	}

	return &Manager{
		storage:   storage,
		publisher: publisher,
		logger:    logger,
		config:    config,
	}
}

// Add adds inventory to a specific location
// 指定ロケーションに在庫を追加
func (m *Manager) Add(ctx context.Context, itemID, locationID string, quantity int64, reference string) error {
	if quantity <= 0 {
		return NewValidationError("quantity", "数量は正の値である必要があります", fmt.Sprintf("%d", quantity))
	}

	// 商品とロケーションの存在確認
	if err := m.validateItemAndLocation(ctx, itemID, locationID); err != nil {
		return err
	}

	// 現在の在庫を取得または初期化
	stock, err := m.storage.GetStock(ctx, itemID, locationID)
	if err != nil && err != ErrStockNotFound {
		return NewStorageError("get_stock", "在庫取得に失敗しました", err)
	}

	oldQuantity := int64(0)
	if stock == nil {
		// 新しい在庫記録を作成
		stock = &Stock{
			ItemID:     itemID,
			LocationID: locationID,
			Quantity:   quantity,
			Reserved:   0,
			Version:    1,
			UpdatedAt:  time.Now(),
			UpdatedBy:  m.getUserFromContext(ctx),
		}
		stock.CalculateAvailable()

		if err := m.storage.CreateStock(ctx, stock); err != nil {
			return NewStorageError("create_stock", "在庫作成に失敗しました", err)
		}
	} else {
		// 既存の在庫を更新
		oldQuantity = stock.Quantity
		stock.Quantity += quantity
		stock.Version++
		stock.UpdatedAt = time.Now()
		stock.UpdatedBy = m.getUserFromContext(ctx)
		stock.CalculateAvailable()

		if err := m.storage.UpdateStock(ctx, stock); err != nil {
			return NewStorageError("update_stock", "在庫更新に失敗しました", err)
		}
	}

	// イベント発行
	if m.publisher != nil {
		event := StockChangedEvent{
			ItemID:        itemID,
			LocationID:    locationID,
			OldQuantity:   oldQuantity,
			NewQuantity:   stock.Quantity,
			ChangeType:    "add",
			Reference:     reference,
			TransactionID: NewTransactionID(),
			Timestamp:     time.Now(),
			UserID:        m.getUserFromContext(ctx),
		}
		if err := m.publisher.PublishStockChanged(ctx, event); err != nil {
			m.logger.Error("イベント発行に失敗しました", zap.Error(err))
		}
	}

	// トランザクション記録
	tx := &Transaction{
		ID:         NewTransactionID(),
		Type:       TransactionTypeInbound,
		ItemID:     itemID,
		ToLocation: &locationID,
		Quantity:   quantity,
		Reference:  reference,
		CreatedAt:  time.Now(),
		CreatedBy:  m.getUserFromContext(ctx),
	}

	if err := m.storage.CreateTransaction(ctx, tx); err != nil {
		m.logger.Error("トランザクション記録に失敗しました", zap.Error(err))
	}

	m.logger.Info("在庫追加完了",
		zap.String("item_id", itemID),
		zap.String("location_id", locationID),
		zap.Int64("quantity", quantity),
		zap.String("reference", reference),
	)

	return nil
}

// Remove removes inventory from a specific location
// 指定ロケーションから在庫を削除
func (m *Manager) Remove(ctx context.Context, itemID, locationID string, quantity int64, reference string) error {
	if quantity <= 0 {
		return NewValidationError("quantity", "数量は正の値である必要があります", fmt.Sprintf("%d", quantity))
	}

	// 商品とロケーションの存在確認
	if err := m.validateItemAndLocation(ctx, itemID, locationID); err != nil {
		return err
	}

	// 現在の在庫を取得
	stock, err := m.storage.GetStock(ctx, itemID, locationID)
	if err != nil {
		if err == ErrStockNotFound {
			return ErrInsufficientStock
		}
		return NewStorageError("get_stock", "在庫取得に失敗しました", err)
	}

	// 在庫不足チェック
	if stock.Available < quantity {
		return ErrInsufficientStock
	}

	// 在庫更新
	oldQuantity := stock.Quantity
	stock.Quantity -= quantity
	stock.Version++
	stock.UpdatedAt = time.Now()
	stock.UpdatedBy = m.getUserFromContext(ctx)
	stock.CalculateAvailable()

	// 負の在庫チェック
	if !m.config.AllowNegativeStock && stock.Quantity < 0 {
		return NewBusinessRuleError("negative_stock", "負の在庫は許可されていません", fmt.Sprintf("商品ID: %s, ロケーション: %s", itemID, locationID))
	}

	if err := m.storage.UpdateStock(ctx, stock); err != nil {
		return NewStorageError("update_stock", "在庫更新に失敗しました", err)
	}

	// イベント発行
	if m.publisher != nil {
		event := StockChangedEvent{
			ItemID:        itemID,
			LocationID:    locationID,
			OldQuantity:   oldQuantity,
			NewQuantity:   stock.Quantity,
			ChangeType:    "remove",
			Reference:     reference,
			TransactionID: NewTransactionID(),
			Timestamp:     time.Now(),
			UserID:        m.getUserFromContext(ctx),
		}
		if err := m.publisher.PublishStockChanged(ctx, event); err != nil {
			m.logger.Error("イベント発行に失敗しました", zap.Error(err))
		}
	}

	// 低在庫アラートチェック
	if stock.Quantity <= m.config.LowStockThreshold {
		m.triggerLowStockAlert(ctx, itemID, locationID, stock.Quantity)
	}

	// トランザクション記録
	tx := &Transaction{
		ID:           NewTransactionID(),
		Type:         TransactionTypeOutbound,
		ItemID:       itemID,
		FromLocation: &locationID,
		Quantity:     quantity,
		Reference:    reference,
		CreatedAt:    time.Now(),
		CreatedBy:    m.getUserFromContext(ctx),
	}

	if err := m.storage.CreateTransaction(ctx, tx); err != nil {
		m.logger.Error("トランザクション記録に失敗しました", zap.Error(err))
	}

	m.logger.Info("在庫削除完了",
		zap.String("item_id", itemID),
		zap.String("location_id", locationID),
		zap.Int64("quantity", quantity),
		zap.String("reference", reference),
	)

	return nil
}

// Transfer moves inventory between locations
// ロケーション間で在庫を移動
func (m *Manager) Transfer(ctx context.Context, itemID, fromLocationID, toLocationID string, quantity int64, reference string) error {
	if quantity <= 0 {
		return NewValidationError("quantity", "数量は正の値である必要があります", fmt.Sprintf("%d", quantity))
	}

	if fromLocationID == toLocationID {
		return NewValidationError("location", "移動元と移動先が同じです", fmt.Sprintf("%s -> %s", fromLocationID, toLocationID))
	}

	// 商品とロケーションの存在確認
	if err := m.validateItemAndLocation(ctx, itemID, fromLocationID); err != nil {
		return err
	}
	if err := m.validateItemAndLocation(ctx, itemID, toLocationID); err != nil {
		return err
	}

	// 移動元から在庫を削除
	if err := m.Remove(ctx, itemID, fromLocationID, quantity, reference); err != nil {
		return err
	}

	// 移動先に在庫を追加
	if err := m.Add(ctx, itemID, toLocationID, quantity, reference); err != nil {
		// ロールバック処理（移動元に戻す）
		if rollbackErr := m.Add(ctx, itemID, fromLocationID, quantity, reference+"_ROLLBACK"); rollbackErr != nil {
			m.logger.Error("ロールバック失敗", zap.Error(rollbackErr))
		}
		return err
	}

	// 移動イベント発行
	if m.publisher != nil {
		event := ItemTransferredEvent{
			ItemID:         itemID,
			FromLocationID: fromLocationID,
			ToLocationID:   toLocationID,
			Quantity:       quantity,
			Reference:      reference,
			TransactionID:  NewTransactionID(),
			Timestamp:      time.Now(),
			UserID:         m.getUserFromContext(ctx),
		}
		if err := m.publisher.PublishItemTransferred(ctx, event); err != nil {
			m.logger.Error("移動イベント発行に失敗しました", zap.Error(err))
		}
	}

	// 移動トランザクション記録
	tx := &Transaction{
		ID:           NewTransactionID(),
		Type:         TransactionTypeTransfer,
		ItemID:       itemID,
		FromLocation: &fromLocationID,
		ToLocation:   &toLocationID,
		Quantity:     quantity,
		Reference:    reference,
		CreatedAt:    time.Now(),
		CreatedBy:    m.getUserFromContext(ctx),
	}

	if err := m.storage.CreateTransaction(ctx, tx); err != nil {
		m.logger.Error("移動トランザクション記録に失敗しました", zap.Error(err))
	}

	m.logger.Info("在庫移動完了",
		zap.String("item_id", itemID),
		zap.String("from_location", fromLocationID),
		zap.String("to_location", toLocationID),
		zap.Int64("quantity", quantity),
		zap.String("reference", reference),
	)

	return nil
}

// Adjust adjusts inventory to a specific quantity
// 在庫を指定数量に調整
func (m *Manager) Adjust(ctx context.Context, itemID, locationID string, newQuantity int64, reference string) error {
	if newQuantity < 0 && !m.config.AllowNegativeStock {
		return NewValidationError("quantity", "負の在庫は許可されていません", fmt.Sprintf("%d", newQuantity))
	}

	// 商品とロケーションの存在確認
	if err := m.validateItemAndLocation(ctx, itemID, locationID); err != nil {
		return err
	}

	// 現在の在庫を取得
	stock, err := m.storage.GetStock(ctx, itemID, locationID)
	if err != nil && err != ErrStockNotFound {
		return NewStorageError("get_stock", "在庫取得に失敗しました", err)
	}

	oldQuantity := int64(0)
	if stock == nil {
		// 新しい在庫記録を作成
		stock = &Stock{
			ItemID:     itemID,
			LocationID: locationID,
			Quantity:   newQuantity,
			Reserved:   0,
			Version:    1,
			UpdatedAt:  time.Now(),
			UpdatedBy:  m.getUserFromContext(ctx),
		}
		stock.CalculateAvailable()

		if err := m.storage.CreateStock(ctx, stock); err != nil {
			return NewStorageError("create_stock", "在庫作成に失敗しました", err)
		}
	} else {
		// 既存の在庫を調整
		oldQuantity = stock.Quantity
		stock.Quantity = newQuantity
		stock.Version++
		stock.UpdatedAt = time.Now()
		stock.UpdatedBy = m.getUserFromContext(ctx)
		stock.CalculateAvailable()

		if err := m.storage.UpdateStock(ctx, stock); err != nil {
			return NewStorageError("update_stock", "在庫更新に失敗しました", err)
		}
	}

	// 調整イベント発行
	if m.publisher != nil {
		event := StockChangedEvent{
			ItemID:        itemID,
			LocationID:    locationID,
			OldQuantity:   oldQuantity,
			NewQuantity:   stock.Quantity,
			ChangeType:    "adjust",
			Reference:     reference,
			TransactionID: NewTransactionID(),
			Timestamp:     time.Now(),
			UserID:        m.getUserFromContext(ctx),
		}
		if err := m.publisher.PublishStockChanged(ctx, event); err != nil {
			m.logger.Error("調整イベント発行に失敗しました", zap.Error(err))
		}
	}

	// 調整トランザクション記録
	tx := &Transaction{
		ID:         NewTransactionID(),
		Type:       TransactionTypeAdjust,
		ItemID:     itemID,
		ToLocation: &locationID,
		Quantity:   newQuantity - oldQuantity, // 差分を記録
		Reference:  reference,
		CreatedAt:  time.Now(),
		CreatedBy:  m.getUserFromContext(ctx),
	}

	if err := m.storage.CreateTransaction(ctx, tx); err != nil {
		m.logger.Error("調整トランザクション記録に失敗しました", zap.Error(err))
	}

	m.logger.Info("在庫調整完了",
		zap.String("item_id", itemID),
		zap.String("location_id", locationID),
		zap.Int64("old_quantity", oldQuantity),
		zap.Int64("new_quantity", newQuantity),
		zap.String("reference", reference),
	)

	return nil
}

// GetStock gets current stock for an item at a location
// 指定ロケーションの商品在庫を取得
func (m *Manager) GetStock(ctx context.Context, itemID, locationID string) (*Stock, error) {
	return m.storage.GetStock(ctx, itemID, locationID)
}

// GetTotalStock gets total stock across all locations for an item
// 商品の全ロケーション合計在庫を取得
func (m *Manager) GetTotalStock(ctx context.Context, itemID string) (int64, error) {
	// 商品の存在確認
	if _, err := m.storage.GetItem(ctx, itemID); err != nil {
		if err == ErrItemNotFound {
			return 0, ErrItemNotFound
		}
		return 0, NewStorageError("get_item", "商品取得に失敗しました", err)
	}

	totalStock, err := m.storage.GetTotalStockByItem(ctx, itemID)
	if err != nil {
		m.logger.Error("合計在庫数取得に失敗しました", zap.String("item_id", itemID), zap.Error(err))
		return 0, fmt.Errorf("合計在庫数取得に失敗しました: %w", err)
	}

	m.logger.Info("総在庫数取得完了",
		zap.String("item_id", itemID),
		zap.Int64("total_stock", totalStock),
	)

	return totalStock, nil
}

// GetStockByLocation gets all stock at a specific location
// 指定ロケーションのすべての在庫を取得
func (m *Manager) GetStockByLocation(ctx context.Context, locationID string) ([]Stock, error) {
	return m.storage.ListStockByLocation(ctx, locationID)
}

// GetHistory gets transaction history for an item
// 商品のトランザクション履歴を取得
func (m *Manager) GetHistory(ctx context.Context, itemID string, limit int) ([]Transaction, error) {
	return m.storage.GetTransactionHistory(ctx, itemID, limit)
}

// GetHistoryByLocation gets transaction history for a location
// ロケーションのトランザクション履歴を取得
func (m *Manager) GetHistoryByLocation(ctx context.Context, locationID string, limit int) ([]Transaction, error) {
	if locationID == "" {
		return nil, NewValidationError("location_id", "ロケーションIDが指定されていません", "")
	}

	if limit <= 0 {
		limit = 100 // デフォルト値
	}

	// ロケーションの存在確認
	if _, err := m.storage.GetLocation(ctx, locationID); err != nil {
		if err == ErrLocationNotFound {
			return nil, ErrLocationNotFound
		}
		return nil, NewStorageError("get_location", "ロケーション取得に失敗しました", err)
	}

	transactions, err := m.storage.GetTransactionHistoryByLocation(ctx, locationID, limit)
	if err != nil {
		m.logger.Error("ロケーション履歴取得に失敗しました", zap.String("location_id", locationID), zap.Error(err))
		return nil, fmt.Errorf("ロケーション履歴取得に失敗しました: %w", err)
	}

	m.logger.Info("ロケーション履歴取得完了",
		zap.String("location_id", locationID),
		zap.Int("limit", limit),
		zap.Int("count", len(transactions)),
	)

	return transactions, nil
}

// GetHistoryByDateRange gets transaction history within a date range
// 日付範囲でトランザクション履歴を取得
func (m *Manager) GetHistoryByDateRange(ctx context.Context, itemID string, from, to time.Time) ([]Transaction, error) {
	if itemID == "" {
		return nil, NewValidationError("item_id", "商品IDが指定されていません", "")
	}

	if from.After(to) {
		return nil, NewValidationError("date_range", "開始日が終了日より後になっています", fmt.Sprintf("%s > %s", from.Format("2006-01-02"), to.Format("2006-01-02")))
	}

	// 商品の存在確認
	if _, err := m.storage.GetItem(ctx, itemID); err != nil {
		if err == ErrItemNotFound {
			return nil, ErrItemNotFound
		}
		return nil, NewStorageError("get_item", "商品取得に失敗しました", err)
	}

	transactions, err := m.storage.GetTransactionHistoryByDateRange(ctx, itemID, from, to)
	if err != nil {
		m.logger.Error("日付範囲履歴取得に失敗しました", zap.String("item_id", itemID), zap.Error(err))
		return nil, fmt.Errorf("日付範囲履歴取得に失敗しました: %w", err)
	}

	m.logger.Info("日付範囲履歴取得完了",
		zap.String("item_id", itemID),
		zap.String("from", from.Format("2006-01-02")),
		zap.String("to", to.Format("2006-01-02")),
		zap.Int("count", len(transactions)),
	)

	return transactions, nil
}

// ExecuteBatch executes a batch of inventory operations
// バッチ在庫操作を実行
func (m *Manager) ExecuteBatch(ctx context.Context, operations []InventoryOperation) (*BatchOperation, error) {
	batch := &BatchOperation{
		ID:          NewBatchID(),
		Operations:  operations,
		Status:      BatchStatusPending,
		CreatedAt:   time.Now(),
		Errors:      make([]BatchOperationError, 0),
	}

	for i, op := range operations {
		var err error
		switch op.Type {
		case OperationTypeAdd:
			err = m.Add(ctx, op.ItemID, op.LocationID, op.Quantity, op.Reference)
		case OperationTypeRemove:
			err = m.Remove(ctx, op.ItemID, op.LocationID, op.Quantity, op.Reference)
		case OperationTypeTransfer:
			if op.ToLocationID == nil {
				err = fmt.Errorf("移動先ロケーションが指定されていません")
			} else {
				err = m.Transfer(ctx, op.ItemID, op.LocationID, *op.ToLocationID, op.Quantity, op.Reference)
			}
		case OperationTypeAdjust:
			err = m.Adjust(ctx, op.ItemID, op.LocationID, op.Quantity, op.Reference)
		default:
			err = fmt.Errorf("未知の操作タイプ: %s", op.Type)
		}

		if err != nil {
			batch.Errors = append(batch.Errors, BatchOperationError{
				OperationIndex: i,
				Error:          err.Error(),
			})
			batch.FailureCount++
		} else {
			batch.SuccessCount++
		}
	}

	now := time.Now()
	batch.CompletedAt = &now
	
	if batch.FailureCount > 0 {
		batch.Status = BatchStatusFailed
	} else {
		batch.Status = BatchStatusCompleted
	}

	return batch, nil
}

// GetBatchStatus gets the status of a batch operation
// バッチ操作のステータスを取得
func (m *Manager) GetBatchStatus(ctx context.Context, batchID string) (*BatchOperation, error) {
	if batchID == "" {
		return nil, NewValidationError("batch_id", "バッチIDが指定されていません", "")
	}

	// TODO: 実際の実装では、バッチ操作の状態をストレージに永続化し、
	// ここで取得する必要がある。現在は簡易実装として固定値を返す。
	batch := &BatchOperation{
		ID:           batchID,
		Operations:   make([]InventoryOperation, 0),
		Status:       BatchStatusCompleted,
		SuccessCount: 0,
		FailureCount: 0,
		Errors:       make([]BatchOperationError, 0),
		CreatedAt:    time.Now().Add(-time.Hour), // 1時間前に作成されたと仮定
		CompletedAt:  &[]time.Time{time.Now()}[0],
	}

	m.logger.Info("バッチステータス取得完了",
		zap.String("batch_id", batchID),
		zap.String("status", string(batch.Status)),
	)

	return batch, nil
}

// Reserve reserves inventory
// 在庫を予約
func (m *Manager) Reserve(ctx context.Context, itemID, locationID string, quantity int64, reference string) error {
	if quantity <= 0 {
		return NewValidationError("quantity", "数量は正の値である必要があります", fmt.Sprintf("%d", quantity))
	}

	// 現在の在庫を取得
	stock, err := m.storage.GetStock(ctx, itemID, locationID)
	if err != nil {
		return NewStorageError("get_stock", "在庫取得に失敗しました", err)
	}

	// 予約可能量チェック
	if stock.Available < quantity {
		return ErrInsufficientStock
	}

	// 予約量更新
	stock.Reserved += quantity
	stock.Version++
	stock.UpdatedAt = time.Now()
	stock.UpdatedBy = m.getUserFromContext(ctx)
	stock.CalculateAvailable()

	if err := m.storage.UpdateStock(ctx, stock); err != nil {
		return NewStorageError("update_stock", "在庫更新に失敗しました", err)
	}

	m.logger.Info("在庫予約完了",
		zap.String("item_id", itemID),
		zap.String("location_id", locationID),
		zap.Int64("quantity", quantity),
		zap.String("reference", reference),
	)

	return nil
}

// ReleaseReservation releases reserved inventory
// 予約された在庫を解除
func (m *Manager) ReleaseReservation(ctx context.Context, itemID, locationID string, quantity int64, reference string) error {
	if quantity <= 0 {
		return NewValidationError("quantity", "数量は正の値である必要があります", fmt.Sprintf("%d", quantity))
	}

	// 現在の在庫を取得
	stock, err := m.storage.GetStock(ctx, itemID, locationID)
	if err != nil {
		return NewStorageError("get_stock", "在庫取得に失敗しました", err)
	}

	// 予約量チェック
	if stock.Reserved < quantity {
		return ErrInsufficientReservation
	}

	// 予約量更新
	stock.Reserved -= quantity
	stock.Version++
	stock.UpdatedAt = time.Now()
	stock.UpdatedBy = m.getUserFromContext(ctx)
	stock.CalculateAvailable()

	if err := m.storage.UpdateStock(ctx, stock); err != nil {
		return NewStorageError("update_stock", "在庫更新に失敗しました", err)
	}

	m.logger.Info("在庫予約解除完了",
		zap.String("item_id", itemID),
		zap.String("location_id", locationID),
		zap.Int64("quantity", quantity),
		zap.String("reference", reference),
	)

	return nil
}

// GetAlerts gets active alerts for a location
// ロケーションのアクティブアラートを取得
func (m *Manager) GetAlerts(ctx context.Context, locationID string) ([]StockAlert, error) {
	return m.storage.GetActiveAlerts(ctx, locationID)
}

// ResolveAlert resolves an alert
// アラートを解決
func (m *Manager) ResolveAlert(ctx context.Context, alertID string) error {
	return m.storage.ResolveAlert(ctx, alertID)
}

// ヘルパーメソッド

// validateItemAndLocation validates that item and location exist
// 商品とロケーションの存在を確認
func (m *Manager) validateItemAndLocation(ctx context.Context, itemID, locationID string) error {
	// 商品の存在確認
	if _, err := m.storage.GetItem(ctx, itemID); err != nil {
		if err == ErrItemNotFound {
			return ErrItemNotFound
		}
		return NewStorageError("get_item", "商品取得に失敗しました", err)
	}

	// ロケーションの存在確認
	if _, err := m.storage.GetLocation(ctx, locationID); err != nil {
		if err == ErrLocationNotFound {
			return ErrLocationNotFound
		}
		return NewStorageError("get_location", "ロケーション取得に失敗しました", err)
	}

	return nil
}

// getUserFromContext extracts user ID from context
// コンテキストからユーザーIDを取得
func (m *Manager) getUserFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return "system"
}

// triggerLowStockAlert creates a low stock alert
// 低在庫アラートを作成
func (m *Manager) triggerLowStockAlert(ctx context.Context, itemID, locationID string, currentQty int64) {
	alert := &StockAlert{
		ID:         NewTransactionID(),
		Type:       AlertTypeLowStock,
		ItemID:     itemID,
		LocationID: locationID,
		CurrentQty: currentQty,
		Threshold:  m.config.LowStockThreshold,
		Message:    fmt.Sprintf("商品 %s のロケーション %s での在庫が低下しています (現在: %d, 閾値: %d)", itemID, locationID, currentQty, m.config.LowStockThreshold),
		IsActive:   true,
		CreatedAt:  time.Now(),
	}

	if err := m.storage.CreateAlert(ctx, alert); err != nil {
		m.logger.Error("アラート作成に失敗しました", zap.Error(err))
		return
	}

	// 低在庫アラートイベント発行
	if m.publisher != nil {
		event := LowStockAlertEvent{
			ItemID:     itemID,
			LocationID: locationID,
			CurrentQty: currentQty,
			Threshold:  m.config.LowStockThreshold,
			Timestamp:  time.Now(),
		}
		if err := m.publisher.PublishLowStockAlert(ctx, event); err != nil {
			m.logger.Error("低在庫アラートイベント発行に失敗しました", zap.Error(err))
		}
	}
}
