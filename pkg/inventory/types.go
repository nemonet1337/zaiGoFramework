// Package inventory provides core inventory management functionality
package inventory

import (
	"time"

	"github.com/google/uuid"
)

// Item represents a product or SKU in the inventory system
// 在庫システムにおける商品またはSKUを表現
type Item struct {
	ID          string    `json:"id" db:"id"`                   // 商品ID
	Name        string    `json:"name" db:"name"`               // 商品名
	SKU         string    `json:"sku" db:"sku"`                 // SKU（在庫管理単位）
	Description string    `json:"description" db:"description"` // 商品説明
	Category    string    `json:"category" db:"category"`       // カテゴリ
	UnitCost    float64   `json:"unit_cost" db:"unit_cost"`     // 単価
	CreatedAt   time.Time `json:"created_at" db:"created_at"`   // 作成日時
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`   // 更新日時
}

// Location represents a storage location or warehouse
// 保管場所または倉庫を表現
type Location struct {
	ID        string    `json:"id" db:"id"`                 // ロケーションID
	Name      string    `json:"name" db:"name"`             // ロケーション名
	Type      string    `json:"type" db:"type"`             // タイプ（倉庫、店舗など）
	Address   string    `json:"address" db:"address"`       // 住所
	Capacity  int64     `json:"capacity" db:"capacity"`     // 最大収容量
	IsActive  bool      `json:"is_active" db:"is_active"`   // アクティブ状態
	CreatedAt time.Time `json:"created_at" db:"created_at"` // 作成日時
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // 更新日時
}

// Stock represents current inventory levels at a location
// 特定ロケーションでの現在の在庫レベルを表現
type Stock struct {
	ItemID     string    `json:"item_id" db:"item_id"`         // 商品ID
	LocationID string    `json:"location_id" db:"location_id"` // ロケーションID
	Quantity   int64     `json:"quantity" db:"quantity"`       // 在庫数量
	Reserved   int64     `json:"reserved" db:"reserved"`       // 予約済み数量
	Available  int64     `json:"available" db:"available"`     // 利用可能数量
	Version    int64     `json:"version" db:"version"`         // 楽観的ロック用バージョン
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`   // 最終更新日時
	UpdatedBy  string    `json:"updated_by" db:"updated_by"`   // 更新者
}

// Transaction represents an inventory movement record
// 在庫移動記録を表現
type Transaction struct {
	ID           string            `json:"id" db:"id"`                       // トランザクションID
	Type         TransactionType   `json:"type" db:"type"`                   // トランザクションタイプ
	ItemID       string            `json:"item_id" db:"item_id"`             // 商品ID
	FromLocation *string           `json:"from_location" db:"from_location"` // 移動元ロケーション（nilの場合は入庫）
	ToLocation   *string           `json:"to_location" db:"to_location"`     // 移動先ロケーション（nilの場合は出庫）
	Quantity     int64             `json:"quantity" db:"quantity"`           // 数量
	UnitCost     *float64          `json:"unit_cost" db:"unit_cost"`         // 単価
	Reference    string            `json:"reference" db:"reference"`         // 参照番号（発注書番号など）
	LotNumber    *string           `json:"lot_number" db:"lot_number"`       // ロット番号
	ExpiryDate   *time.Time        `json:"expiry_date" db:"expiry_date"`     // 有効期限
	Metadata     map[string]string `json:"metadata" db:"metadata"`           // 追加メタデータ
	CreatedAt    time.Time         `json:"created_at" db:"created_at"`       // 作成日時
	CreatedBy    string            `json:"created_by" db:"created_by"`       // 作成者
}

// TransactionType defines the type of inventory movement
// 在庫移動のタイプを定義
type TransactionType string

const (
	TransactionTypeInbound  TransactionType = "inbound"  // 入庫
	TransactionTypeOutbound TransactionType = "outbound" // 出庫
	TransactionTypeTransfer TransactionType = "transfer" // 移動
	TransactionTypeAdjust   TransactionType = "adjust"   // 調整
)

// Lot represents a batch of items with the same characteristics
// 同じ特性を持つ商品のバッチを表現
type Lot struct {
	ID         string     `json:"id" db:"id"`                   // ロットID
	Number     string     `json:"number" db:"number"`           // ロット番号
	ItemID     string     `json:"item_id" db:"item_id"`         // 商品ID
	Quantity   int64      `json:"quantity" db:"quantity"`       // 数量
	UnitCost   float64    `json:"unit_cost" db:"unit_cost"`     // 単価
	ExpiryDate *time.Time `json:"expiry_date" db:"expiry_date"` // 有効期限
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`   // 作成日時
}

// StockAlert represents low stock or other inventory alerts
// 低在庫やその他の在庫アラートを表現
type StockAlert struct {
	ID         string      `json:"id" db:"id"`                   // アラートID
	Type       AlertType   `json:"type" db:"type"`               // アラートタイプ
	ItemID     string      `json:"item_id" db:"item_id"`         // 商品ID
	LocationID string      `json:"location_id" db:"location_id"` // ロケーションID
	CurrentQty int64       `json:"current_qty" db:"current_qty"` // 現在数量
	Threshold  int64       `json:"threshold" db:"threshold"`     // 閾値
	Message    string      `json:"message" db:"message"`         // メッセージ
	IsActive   bool        `json:"is_active" db:"is_active"`     // アクティブ状態
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`   // 作成日時
	ResolvedAt *time.Time  `json:"resolved_at" db:"resolved_at"` // 解決日時
}

// AlertType defines types of inventory alerts
// 在庫アラートのタイプを定義
type AlertType string

const (
	AlertTypeLowStock    AlertType = "low_stock"    // 低在庫
	AlertTypeOverStock   AlertType = "over_stock"   // 過剰在庫
	AlertTypeExpiring    AlertType = "expiring"     // 期限切れ間近
	AlertTypeExpired     AlertType = "expired"      // 期限切れ
	AlertTypeDiscrepancy AlertType = "discrepancy"  // 棚卸差異
)

// BatchOperation represents a batch inventory operation
// バッチ在庫操作を表現
type BatchOperation struct {
	ID          string                   `json:"id"`           // バッチID
	Operations  []InventoryOperation     `json:"operations"`   // 操作リスト
	Status      BatchStatus              `json:"status"`       // ステータス
	SuccessCount int                     `json:"success_count"` // 成功数
	FailureCount int                     `json:"failure_count"` // 失敗数
	Errors      []BatchOperationError    `json:"errors"`       // エラーリスト
	CreatedAt   time.Time                `json:"created_at"`   // 作成日時
	CompletedAt *time.Time               `json:"completed_at"` // 完了日時
}

// InventoryOperation represents a single inventory operation
// 単一の在庫操作を表現
type InventoryOperation struct {
	Type       OperationType `json:"type"`        // 操作タイプ
	ItemID     string        `json:"item_id"`     // 商品ID
	LocationID string        `json:"location_id"` // ロケーションID
	Quantity   int64         `json:"quantity"`    // 数量
	Reference  string        `json:"reference"`   // 参照番号
	ToLocationID *string     `json:"to_location_id,omitempty"` // 移動先（移動操作の場合）
}

// OperationType defines types of inventory operations
// 在庫操作のタイプを定義
type OperationType string

const (
	OperationTypeAdd      OperationType = "add"      // 追加
	OperationTypeRemove   OperationType = "remove"   // 削除
	OperationTypeTransfer OperationType = "transfer" // 移動
	OperationTypeAdjust   OperationType = "adjust"   // 調整
)

// BatchStatus defines the status of a batch operation
// バッチ操作のステータスを定義
type BatchStatus string

const (
	BatchStatusPending   BatchStatus = "pending"   // 処理中
	BatchStatusCompleted BatchStatus = "completed" // 完了
	BatchStatusFailed    BatchStatus = "failed"    // 失敗
)

// BatchOperationError represents an error in batch processing
// バッチ処理でのエラーを表現
type BatchOperationError struct {
	OperationIndex int    `json:"operation_index"` // 操作インデックス
	Error          string `json:"error"`           // エラーメッセージ
}

// NewTransactionID generates a new transaction ID
// 新しいトランザクションIDを生成
func NewTransactionID() string {
	return uuid.New().String()
}

// NewBatchID generates a new batch operation ID
// 新しいバッチ操作IDを生成
func NewBatchID() string {
	return uuid.New().String()
}

// Calculate available quantity (total - reserved)
// 利用可能数量を計算（総数量 - 予約済み数量）
func (s *Stock) CalculateAvailable() {
	s.Available = s.Quantity - s.Reserved
}

// IsExpired checks if a lot has expired
// ロットが期限切れかチェック
func (l *Lot) IsExpired() bool {
	if l.ExpiryDate == nil {
		return false
	}
	return time.Now().After(*l.ExpiryDate)
}

// IsExpiringSoon checks if a lot expires within the given duration
// ロットが指定期間内に期限切れになるかチェック
func (l *Lot) IsExpiringSoon(duration time.Duration) bool {
	if l.ExpiryDate == nil {
		return false
	}
	return time.Now().Add(duration).After(*l.ExpiryDate)
}
