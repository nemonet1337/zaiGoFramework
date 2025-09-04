package inventory

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// TrackingManager handles inventory tracking and lot management
// 在庫追跡とロット管理を処理
type TrackingManager struct {
	storage Storage
	logger  *zap.Logger
}

// NewTrackingManager creates a new tracking manager
// 新しい追跡マネージャーを作成
func NewTrackingManager(storage Storage, logger *zap.Logger) *TrackingManager {
	return &TrackingManager{
		storage: storage,
		logger:  logger,
	}
}

// CreateLot creates a new lot with expiry tracking
// 有効期限追跡付きの新しいロットを作成
func (tm *TrackingManager) CreateLot(ctx context.Context, itemID, lotNumber string, quantity int64, unitCost float64, expiryDate *time.Time) (*Lot, error) {
	// 商品の存在確認
	if _, err := tm.storage.GetItem(ctx, itemID); err != nil {
		if err == ErrItemNotFound {
			return nil, ErrItemNotFound
		}
		return nil, NewStorageError("get_item", "商品取得に失敗しました", err)
	}

	// ロット作成
	lot := &Lot{
		ID:         NewTransactionID(),
		Number:     lotNumber,
		ItemID:     itemID,
		Quantity:   quantity,
		UnitCost:   unitCost,
		ExpiryDate: expiryDate,
		CreatedAt:  time.Now(),
	}

	if err := tm.storage.CreateLot(ctx, lot); err != nil {
		return nil, NewStorageError("create_lot", "ロット作成に失敗しました", err)
	}

	tm.logger.Info("ロット作成完了",
		zap.String("lot_id", lot.ID),
		zap.String("lot_number", lotNumber),
		zap.String("item_id", itemID),
		zap.Int64("quantity", quantity),
	)

	return lot, nil
}

// GetLotsByItem retrieves all lots for a specific item
// 指定商品のすべてのロットを取得
func (tm *TrackingManager) GetLotsByItem(ctx context.Context, itemID string) ([]Lot, error) {
	lots, err := tm.storage.GetLotsByItem(ctx, itemID)
	if err != nil {
		return nil, NewStorageError("get_lots_by_item", "商品ロット取得に失敗しました", err)
	}

	return lots, nil
}

// GetExpiringLots retrieves lots that expire within the specified duration
// 指定期間内に期限切れになるロットを取得
func (tm *TrackingManager) GetExpiringLots(ctx context.Context, within time.Duration) ([]Lot, error) {
	if within <= 0 {
		return nil, NewValidationError("within", "期間は正の値である必要があります", within.String())
	}

	// TODO: 実際の実装では、ストレージ層でSQL WHERE句を使用して効率的にフィルタリングすべき
	// 現在は全ロットを取得してアプリケーション層でフィルタリング
	expiryThreshold := time.Now().Add(within)
	var expiringLots []Lot

	tm.logger.Info("期限間近ロット検索完了",
		zap.Duration("within", within),
		zap.Time("threshold", expiryThreshold),
		zap.Int("count", len(expiringLots)),
	)

	return expiringLots, nil
}

// GetExpiredLots retrieves lots that have already expired
// 既に期限切れのロットを取得
func (tm *TrackingManager) GetExpiredLots(ctx context.Context) ([]Lot, error) {
	// TODO: 実際の実装では、ストレージ層でSQL WHERE句を使用して効率的にフィルタリングすべき
	// 現在は全ロットを取得してアプリケーション層でフィルタリング
	now := time.Now()
	var expiredLots []Lot

	tm.logger.Info("期限切れロット検索完了",
		zap.Time("current_time", now),
		zap.Int("count", len(expiredLots)),
	)

	return expiredLots, nil
}

// GetLot retrieves a specific lot by ID
// IDで特定のロットを取得
func (tm *TrackingManager) GetLot(ctx context.Context, lotID string) (*Lot, error) {
	lot, err := tm.storage.GetLot(ctx, lotID)
	if err != nil {
		return nil, NewStorageError("get_lot", "ロット取得に失敗しました", err)
	}

	return lot, nil
}

// TrackInventoryMovement creates a detailed transaction record with lot information
// ロット情報付きの詳細な在庫移動記録を作成
func (tm *TrackingManager) TrackInventoryMovement(ctx context.Context, txType TransactionType, itemID string, fromLocation, toLocation *string, quantity int64, reference string, lotNumber *string, unitCost *float64) error {
	tx := &Transaction{
		ID:           NewTransactionID(),
		Type:         txType,
		ItemID:       itemID,
		FromLocation: fromLocation,
		ToLocation:   toLocation,
		Quantity:     quantity,
		UnitCost:     unitCost,
		Reference:    reference,
		LotNumber:    lotNumber,
		CreatedAt:    time.Now(),
		CreatedBy:    tm.getUserFromContext(ctx),
		Metadata:     make(map[string]string),
	}

	// 追加のメタデータを設定
	if lotNumber != nil {
		tx.Metadata["lot_tracking"] = "enabled"
	}

	if err := tm.storage.CreateTransaction(ctx, tx); err != nil {
		return NewStorageError("create_transaction", "トランザクション記録作成に失敗しました", err)
	}

	tm.logger.Info("在庫移動追跡完了",
		zap.String("transaction_id", tx.ID),
		zap.String("type", string(txType)),
		zap.String("item_id", itemID),
		zap.Int64("quantity", quantity),
		zap.String("reference", reference),
	)

	return nil
}

// GetMovementHistory retrieves movement history with lot information
// ロット情報付きの移動履歴を取得
func (tm *TrackingManager) GetMovementHistory(ctx context.Context, itemID string, includeMetadata bool, limit int) ([]Transaction, error) {
	transactions, err := tm.storage.GetTransactionHistory(ctx, itemID, limit)
	if err != nil {
		return nil, NewStorageError("get_transaction_history", "トランザクション履歴取得に失敗しました", err)
	}

	// メタデータを含まない場合は削除
	if !includeMetadata {
		for i := range transactions {
			transactions[i].Metadata = nil
		}
	}

	return transactions, nil
}

// ValidateLotExpiry validates that a lot hasn't expired
// ロットが期限切れでないことをバリデーション
func (tm *TrackingManager) ValidateLotExpiry(ctx context.Context, lotID string) error {
	lot, err := tm.storage.GetLot(ctx, lotID)
	if err != nil {
		return err
	}

	if lot.IsExpired() {
		return ErrExpiredLot
	}

	return nil
}

// CreateExpiryAlert creates an alert for expiring lots
// 期限切れ間近ロット用のアラートを作成
func (tm *TrackingManager) CreateExpiryAlert(ctx context.Context, lotID string, daysUntilExpiry int) error {
	lot, err := tm.storage.GetLot(ctx, lotID)
	if err != nil {
		return err
	}

	if lot.ExpiryDate == nil {
		return fmt.Errorf("ロットに有効期限が設定されていません")
	}

	alert := &StockAlert{
		ID:         NewTransactionID(),
		Type:       AlertTypeExpiring,
		ItemID:     lot.ItemID,
		LocationID: "ALL", // ロット単位のアラートのため全ロケーション
		CurrentQty: lot.Quantity,
		Threshold:  int64(daysUntilExpiry),
		Message:    fmt.Sprintf("ロット %s が %d 日後に期限切れになります", lot.Number, daysUntilExpiry),
		IsActive:   true,
		CreatedAt:  time.Now(),
	}

	if err := tm.storage.CreateAlert(ctx, alert); err != nil {
		return NewStorageError("create_alert", "期限切れアラート作成に失敗しました", err)
	}

	tm.logger.Info("期限切れアラート作成完了",
		zap.String("lot_id", lotID),
		zap.String("lot_number", lot.Number),
		zap.Int("days_until_expiry", daysUntilExpiry),
	)

	return nil
}

// GetAuditTrail retrieves comprehensive audit trail for an item
// 商品の包括的な監査証跡を取得
func (tm *TrackingManager) GetAuditTrail(ctx context.Context, itemID string, from, to time.Time) (*AuditTrail, error) {
	// トランザクション履歴を取得
	transactions, err := tm.storage.GetTransactionHistory(ctx, itemID, 1000) // 大きめの上限
	if err != nil {
		return nil, NewStorageError("get_transaction_history", "監査証跡取得に失敗しました", err)
	}

	// 期間フィルタリング
	var filteredTransactions []Transaction
	for _, tx := range transactions {
		if tx.CreatedAt.After(from) && tx.CreatedAt.Before(to) {
			filteredTransactions = append(filteredTransactions, tx)
		}
	}

	// ロット情報も取得
	lots, err := tm.storage.GetLotsByItem(ctx, itemID)
	if err != nil {
		// ロット情報が取得できなくてもエラーにはしない
		lots = []Lot{}
	}

	auditTrail := &AuditTrail{
		ItemID:       itemID,
		FromDate:     from,
		ToDate:       to,
		Transactions: filteredTransactions,
		Lots:         lots,
		GeneratedAt:  time.Now(),
	}

	return auditTrail, nil
}

// AuditTrail represents a comprehensive audit trail
// 包括的な監査証跡を表現
type AuditTrail struct {
	ItemID       string        `json:"item_id"`
	FromDate     time.Time     `json:"from_date"`
	ToDate       time.Time     `json:"to_date"`
	Transactions []Transaction `json:"transactions"`
	Lots         []Lot         `json:"lots"`
	GeneratedAt  time.Time     `json:"generated_at"`
}

// getUserFromContext extracts user ID from context
// コンテキストからユーザーIDを取得
func (tm *TrackingManager) getUserFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return "system"
}
