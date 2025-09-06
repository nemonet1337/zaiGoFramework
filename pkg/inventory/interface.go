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
//
// このインターフェースは在庫管理システムのデータ永続化を抽象化し、
// PostgreSQL、MySQL、その他のデータベースシステムに対応できる設計となっています。
// 全てのメソッドはコンテキストを受け取り、適切なタイムアウトとキャンセレーション処理を行います。
type Storage interface {
	// Transaction management - トランザクション管理
	// データベーストランザクションを開始し、ACID特性を保証します
	Begin(ctx context.Context) (Transaction, error)
	
	// Stock operations - 在庫操作
	// 新しい在庫記録を作成します。既存の記録がある場合はエラーを返します
	CreateStock(ctx context.Context, stock *Stock) error
	// 既存の在庫記録を更新します。楽観的ロックによる同時実行制御を行います
	UpdateStock(ctx context.Context, stock *Stock) error
	// 指定された商品とロケーションの在庫情報を取得します
	GetStock(ctx context.Context, itemID, locationID string) (*Stock, error)
	// 指定されたロケーションの全ての在庫情報を取得します
	ListStockByLocation(ctx context.Context, locationID string) ([]Stock, error)
	// 指定された商品の全ロケーションでの合計在庫数を取得します
	GetTotalStockByItem(ctx context.Context, itemID string) (int64, error)
	
	// Transaction history - トランザクション履歴
	// 新しいトランザクション記録を作成します（監査証跡として使用）
	CreateTransaction(ctx context.Context, tx *Transaction) error
	// 指定された商品のトランザクション履歴を取得します（最新順）
	GetTransactionHistory(ctx context.Context, itemID string, limit int) ([]Transaction, error)
	// 指定されたロケーションのトランザクション履歴を取得します（最新順）
	GetTransactionHistoryByLocation(ctx context.Context, locationID string, limit int) ([]Transaction, error)
	// 指定された商品の指定日付範囲のトランザクション履歴を取得します
	GetTransactionHistoryByDateRange(ctx context.Context, itemID string, from, to time.Time) ([]Transaction, error)
	
	// Item management - 商品管理
	// 新しい商品を作成します。重複するIDの場合はエラーを返します
	CreateItem(ctx context.Context, item *Item) error
	// 指定されたIDの商品情報を取得します
	GetItem(ctx context.Context, itemID string) (*Item, error)
	// 既存の商品情報を更新します
	UpdateItem(ctx context.Context, item *Item) error
	
	// Location management - ロケーション管理
	// 新しいロケーションを作成します
	CreateLocation(ctx context.Context, location *Location) error
	// 指定されたIDのロケーション情報を取得します
	GetLocation(ctx context.Context, locationID string) (*Location, error)
	
	// Lot management - ロット管理
	// 新しいロット（バッチ）を作成します
	CreateLot(ctx context.Context, lot *Lot) error
	// 指定されたIDのロット情報を取得します
	GetLot(ctx context.Context, lotID string) (*Lot, error)
	// 指定された商品の全てのロット情報を取得します
	GetLotsByItem(ctx context.Context, itemID string) ([]Lot, error)
	
	// Alert management - アラート管理
	// 新しいアラートを作成します（低在庫、期限切れなど）
	CreateAlert(ctx context.Context, alert *StockAlert) error
	// 指定されたロケーションのアクティブなアラートを取得します
	GetActiveAlerts(ctx context.Context, locationID string) ([]StockAlert, error)
	// 指定されたアラートを解決済みとしてマークします
	ResolveAlert(ctx context.Context, alertID string) error
	
	// Health check - ヘルスチェック
	// データベース接続の健全性を確認します
	Ping(ctx context.Context) error
	// データベース接続を安全に閉じます
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
