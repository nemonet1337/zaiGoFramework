package inventory

import (
	"context"
	"time"
)

// InventoryManager defines the core interface for inventory management
// 在庫管理のコアインターフェースを定義
type InventoryManager interface {
	// 基本的な在庫操作 - Basic inventory operations
	Add(ctx context.Context, itemID, locationID string, quantity int64, reference string) error
	Remove(ctx context.Context, itemID, locationID string, quantity int64, reference string) error
	Transfer(ctx context.Context, itemID, fromLocationID, toLocationID string, quantity int64, reference string) error
	Adjust(ctx context.Context, itemID, locationID string, newQuantity int64, reference string) error

	// 在庫照会 - Stock inquiry
	GetStock(ctx context.Context, itemID, locationID string) (*Stock, error)
	GetTotalStock(ctx context.Context, itemID string) (int64, error)
	GetStockByLocation(ctx context.Context, locationID string) ([]Stock, error)

	// 履歴管理 - History management
	GetHistory(ctx context.Context, itemID string, limit int) ([]Transaction, error)
	GetHistoryByLocation(ctx context.Context, locationID string, limit int) ([]Transaction, error)
	GetHistoryByDateRange(ctx context.Context, itemID string, from, to time.Time) ([]Transaction, error)

	// バッチ処理 - Batch operations
	ExecuteBatch(ctx context.Context, operations []InventoryOperation) (*BatchOperation, error)
	GetBatchStatus(ctx context.Context, batchID string) (*BatchOperation, error)

	// 予約管理 - Reservation management
	Reserve(ctx context.Context, itemID, locationID string, quantity int64, reference string) error
	ReleaseReservation(ctx context.Context, itemID, locationID string, quantity int64, reference string) error

	// アラート管理 - Alert management
	GetAlerts(ctx context.Context, locationID string) ([]StockAlert, error)
	ResolveAlert(ctx context.Context, alertID string) error
}

// ItemManager defines interface for item management
// 商品管理のインターフェースを定義
type ItemManager interface {
	CreateItem(ctx context.Context, item *Item) error
	GetItem(ctx context.Context, itemID string) (*Item, error)
	UpdateItem(ctx context.Context, item *Item) error
	DeleteItem(ctx context.Context, itemID string) error
	ListItems(ctx context.Context, offset, limit int) ([]Item, error)
	SearchItems(ctx context.Context, query string) ([]Item, error)
}

// LocationManager defines interface for location management
// ロケーション管理のインターフェースを定義
type LocationManager interface {
	CreateLocation(ctx context.Context, location *Location) error
	GetLocation(ctx context.Context, locationID string) (*Location, error)
	UpdateLocation(ctx context.Context, location *Location) error
	DeleteLocation(ctx context.Context, locationID string) error
	ListLocations(ctx context.Context, offset, limit int) ([]Location, error)
}

// LotManager defines interface for lot/batch management
// ロット/バッチ管理のインターフェースを定義
type LotManager interface {
	CreateLot(ctx context.Context, lot *Lot) error
	GetLot(ctx context.Context, lotID string) (*Lot, error)
	GetLotsByItem(ctx context.Context, itemID string) ([]Lot, error)
	GetExpiringLots(ctx context.Context, within time.Duration) ([]Lot, error)
	GetExpiredLots(ctx context.Context) ([]Lot, error)
}

// ValuationEngine defines interface for inventory valuation
// 在庫評価エンジンのインターフェースを定義
type ValuationEngine interface {
	CalculateValue(ctx context.Context, itemID, locationID string, method ValuationMethod) (float64, error)
	CalculateTotalValue(ctx context.Context, locationID string, method ValuationMethod) (float64, error)
	GetAverageCost(ctx context.Context, itemID string) (float64, error)
}

// ValuationMethod defines inventory valuation methods
// 在庫評価方法を定義
type ValuationMethod string

const (
	ValuationMethodFIFO     ValuationMethod = "FIFO"     // 先入先出
	ValuationMethodLIFO     ValuationMethod = "LIFO"     // 後入先出
	ValuationMethodAverage  ValuationMethod = "AVERAGE"  // 平均法
	ValuationMethodStandard ValuationMethod = "STANDARD" // 標準原価
)

// AnalyticsEngine defines interface for inventory analytics
// 在庫分析エンジンのインターフェースを定義
type AnalyticsEngine interface {
	CalculateABCClassification(ctx context.Context, locationID string) (map[string]string, error)
	GetTurnoverRate(ctx context.Context, itemID string, period time.Duration) (float64, error)
	GetSlowMovingItems(ctx context.Context, locationID string, threshold time.Duration) ([]string, error)
	GenerateStockReport(ctx context.Context, locationID string, reportType ReportType) ([]byte, error)
}

// ReportType defines types of inventory reports
// 在庫レポートのタイプを定義
type ReportType string

const (
	ReportTypeStock      ReportType = "stock"      // 在庫レポート
	ReportTypeMovement   ReportType = "movement"   // 移動レポート
	ReportTypeValuation  ReportType = "valuation"  // 評価レポート
	ReportTypeABC        ReportType = "abc"        // ABC分析レポート
	ReportTypeTurnover   ReportType = "turnover"   // 回転率レポート
)

// Storage defines the interface for data persistence layer
// データ永続化層のインターフェースを定義
type Storage interface {
	// Transaction management
	Begin(ctx context.Context) (Transaction, error)
	
	// Stock operations
	CreateStock(ctx context.Context, stock *Stock) error
	UpdateStock(ctx context.Context, stock *Stock) error
	GetStock(ctx context.Context, itemID, locationID string) (*Stock, error)
	ListStockByLocation(ctx context.Context, locationID string) ([]Stock, error)
	
	// Transaction history
	CreateTransaction(ctx context.Context, tx *Transaction) error
	GetTransactionHistory(ctx context.Context, itemID string, limit int) ([]Transaction, error)
	
	// Item management
	CreateItem(ctx context.Context, item *Item) error
	GetItem(ctx context.Context, itemID string) (*Item, error)
	UpdateItem(ctx context.Context, item *Item) error
	
	// Location management
	CreateLocation(ctx context.Context, location *Location) error
	GetLocation(ctx context.Context, locationID string) (*Location, error)
	
	// Lot management
	CreateLot(ctx context.Context, lot *Lot) error
	GetLot(ctx context.Context, lotID string) (*Lot, error)
	GetLotsByItem(ctx context.Context, itemID string) ([]Lot, error)
	
	// Alert management
	CreateAlert(ctx context.Context, alert *StockAlert) error
	GetActiveAlerts(ctx context.Context, locationID string) ([]StockAlert, error)
	ResolveAlert(ctx context.Context, alertID string) error
	
	// Health check
	Ping(ctx context.Context) error
	Close() error
}

// EventPublisher defines interface for publishing inventory events
// 在庫イベント発行のインターフェースを定義
type EventPublisher interface {
	PublishStockChanged(ctx context.Context, event StockChangedEvent) error
	PublishLowStockAlert(ctx context.Context, event LowStockAlertEvent) error
	PublishItemTransferred(ctx context.Context, event ItemTransferredEvent) error
}

// Events for inventory operations
// 在庫操作のイベント定義

// StockChangedEvent represents a stock level change
// 在庫レベル変更イベントを表現
type StockChangedEvent struct {
	ItemID       string    `json:"item_id"`
	LocationID   string    `json:"location_id"`
	OldQuantity  int64     `json:"old_quantity"`
	NewQuantity  int64     `json:"new_quantity"`
	ChangeType   string    `json:"change_type"`
	Reference    string    `json:"reference"`
	TransactionID string   `json:"transaction_id"`
	Timestamp    time.Time `json:"timestamp"`
	UserID       string    `json:"user_id"`
}

// LowStockAlertEvent represents a low stock alert
// 低在庫アラートイベントを表現
type LowStockAlertEvent struct {
	ItemID      string    `json:"item_id"`
	LocationID  string    `json:"location_id"`
	CurrentQty  int64     `json:"current_qty"`
	Threshold   int64     `json:"threshold"`
	Timestamp   time.Time `json:"timestamp"`
}

// ItemTransferredEvent represents an item transfer
// 商品移動イベントを表現
type ItemTransferredEvent struct {
	ItemID         string    `json:"item_id"`
	FromLocationID string    `json:"from_location_id"`
	ToLocationID   string    `json:"to_location_id"`
	Quantity       int64     `json:"quantity"`
	Reference      string    `json:"reference"`
	TransactionID  string    `json:"transaction_id"`
	Timestamp      time.Time `json:"timestamp"`
	UserID         string    `json:"user_id"`
}
