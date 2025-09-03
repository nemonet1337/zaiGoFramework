package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/nemonet1337/zaiGoFramework/pkg/inventory"
)

// PostgreSQLStorage implements the Storage interface using PostgreSQL
// PostgreSQLを使用したStorageインターフェースの実装
type PostgreSQLStorage struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewPostgreSQLStorage creates a new PostgreSQL storage instance
// 新しいPostgreSQLストレージインスタンスを作成
func NewPostgreSQLStorage(dsn string, logger *zap.Logger) (*PostgreSQLStorage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("データベース接続に失敗しました: %w", err)
	}

	// 接続テスト
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("データベースpingに失敗しました: %w", err)
	}

	// 接続プール設定
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	storage := &PostgreSQLStorage{
		db:     db,
		logger: logger,
	}

	return storage, nil
}

// Begin starts a new database transaction
// 新しいデータベーストランザクションを開始
func (s *PostgreSQLStorage) Begin(ctx context.Context) (inventory.Transaction, error) {
	// 注意: ここでのTransactionはDBトランザクション、inventory.Transactionは在庫移動記録
	// 実装の簡略化のため、この機能は後で実装
	return nil, fmt.Errorf("未実装")
}

// CreateStock creates a new stock record
// 新しい在庫記録を作成
func (s *PostgreSQLStorage) CreateStock(ctx context.Context, stock *inventory.Stock) error {
	query := `
		INSERT INTO stocks (item_id, location_id, quantity, reserved, available, version, updated_at, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := s.db.ExecContext(ctx, query,
		stock.ItemID,
		stock.LocationID,
		stock.Quantity,
		stock.Reserved,
		stock.Available,
		stock.Version,
		stock.UpdatedAt,
		stock.UpdatedBy,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("在庫記録は既に存在します")
		}
		return fmt.Errorf("在庫記録作成に失敗しました: %w", err)
	}

	return nil
}

// UpdateStock updates an existing stock record
// 既存の在庫記録を更新
func (s *PostgreSQLStorage) UpdateStock(ctx context.Context, stock *inventory.Stock) error {
	query := `
		UPDATE stocks 
		SET quantity = $3, reserved = $4, available = $5, version = $6, updated_at = $7, updated_by = $8
		WHERE item_id = $1 AND location_id = $2 AND version = $9`

	result, err := s.db.ExecContext(ctx, query,
		stock.ItemID,
		stock.LocationID,
		stock.Quantity,
		stock.Reserved,
		stock.Available,
		stock.Version,
		stock.UpdatedAt,
		stock.UpdatedBy,
		stock.Version-1, // 楽観的ロックのための前バージョン
	)

	if err != nil {
		return fmt.Errorf("在庫記録更新に失敗しました: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("更新行数の取得に失敗しました: %w", err)
	}

	if rowsAffected == 0 {
		return inventory.ErrVersionMismatch
	}

	return nil
}

// GetStock retrieves stock information for an item at a location
// 指定ロケーションの商品在庫情報を取得
func (s *PostgreSQLStorage) GetStock(ctx context.Context, itemID, locationID string) (*inventory.Stock, error) {
	query := `
		SELECT item_id, location_id, quantity, reserved, available, version, updated_at, updated_by
		FROM stocks 
		WHERE item_id = $1 AND location_id = $2`

	stock := &inventory.Stock{}
	err := s.db.QueryRowContext(ctx, query, itemID, locationID).Scan(
		&stock.ItemID,
		&stock.LocationID,
		&stock.Quantity,
		&stock.Reserved,
		&stock.Available,
		&stock.Version,
		&stock.UpdatedAt,
		&stock.UpdatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, inventory.ErrStockNotFound
		}
		return nil, fmt.Errorf("在庫取得に失敗しました: %w", err)
	}

	return stock, nil
}

// ListStockByLocation retrieves all stock at a specific location
// 指定ロケーションのすべての在庫を取得
func (s *PostgreSQLStorage) ListStockByLocation(ctx context.Context, locationID string) ([]inventory.Stock, error) {
	query := `
		SELECT item_id, location_id, quantity, reserved, available, version, updated_at, updated_by
		FROM stocks 
		WHERE location_id = $1
		ORDER BY item_id`

	rows, err := s.db.QueryContext(ctx, query, locationID)
	if err != nil {
		return nil, fmt.Errorf("ロケーション在庫取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var stocks []inventory.Stock
	for rows.Next() {
		var stock inventory.Stock
		err := rows.Scan(
			&stock.ItemID,
			&stock.LocationID,
			&stock.Quantity,
			&stock.Reserved,
			&stock.Available,
			&stock.Version,
			&stock.UpdatedAt,
			&stock.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("在庫スキャンに失敗しました: %w", err)
		}
		stocks = append(stocks, stock)
	}

	return stocks, nil
}

// CreateTransaction creates a new transaction record
// 新しいトランザクション記録を作成
func (s *PostgreSQLStorage) CreateTransaction(ctx context.Context, tx *inventory.Transaction) error {
	metadataJSON, err := json.Marshal(tx.Metadata)
	if err != nil {
		return fmt.Errorf("メタデータのJSON変換に失敗しました: %w", err)
	}

	query := `
		INSERT INTO transactions (id, type, item_id, from_location, to_location, quantity, unit_cost, reference, lot_number, expiry_date, metadata, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err = s.db.ExecContext(ctx, query,
		tx.ID,
		tx.Type,
		tx.ItemID,
		tx.FromLocation,
		tx.ToLocation,
		tx.Quantity,
		tx.UnitCost,
		tx.Reference,
		tx.LotNumber,
		tx.ExpiryDate,
		metadataJSON,
		tx.CreatedAt,
		tx.CreatedBy,
	)

	if err != nil {
		return fmt.Errorf("トランザクション記録作成に失敗しました: %w", err)
	}

	return nil
}

// GetTransactionHistory retrieves transaction history for an item
// 商品のトランザクション履歴を取得
func (s *PostgreSQLStorage) GetTransactionHistory(ctx context.Context, itemID string, limit int) ([]inventory.Transaction, error) {
	query := `
		SELECT id, type, item_id, from_location, to_location, quantity, unit_cost, reference, lot_number, expiry_date, metadata, created_at, created_by
		FROM transactions 
		WHERE item_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := s.db.QueryContext(ctx, query, itemID, limit)
	if err != nil {
		return nil, fmt.Errorf("トランザクション履歴取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var transactions []inventory.Transaction
	for rows.Next() {
		var tx inventory.Transaction
		var metadataJSON []byte

		err := rows.Scan(
			&tx.ID,
			&tx.Type,
			&tx.ItemID,
			&tx.FromLocation,
			&tx.ToLocation,
			&tx.Quantity,
			&tx.UnitCost,
			&tx.Reference,
			&tx.LotNumber,
			&tx.ExpiryDate,
			&metadataJSON,
			&tx.CreatedAt,
			&tx.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("トランザクションスキャンに失敗しました: %w", err)
		}

		// メタデータのデシリアライズ
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &tx.Metadata); err != nil {
				s.logger.Warn("メタデータのパースに失敗しました", zap.Error(err))
			}
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// CreateItem creates a new item
// 新しい商品を作成
func (s *PostgreSQLStorage) CreateItem(ctx context.Context, item *inventory.Item) error {
	query := `
		INSERT INTO items (id, name, sku, description, category, unit_cost, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := s.db.ExecContext(ctx, query,
		item.ID,
		item.Name,
		item.SKU,
		item.Description,
		item.Category,
		item.UnitCost,
		item.CreatedAt,
		item.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return inventory.ErrDuplicateItem
		}
		return fmt.Errorf("商品作成に失敗しました: %w", err)
	}

	return nil
}

// GetItem retrieves an item by ID
// IDで商品を取得
func (s *PostgreSQLStorage) GetItem(ctx context.Context, itemID string) (*inventory.Item, error) {
	query := `
		SELECT id, name, sku, description, category, unit_cost, created_at, updated_at
		FROM items 
		WHERE id = $1`

	item := &inventory.Item{}
	err := s.db.QueryRowContext(ctx, query, itemID).Scan(
		&item.ID,
		&item.Name,
		&item.SKU,
		&item.Description,
		&item.Category,
		&item.UnitCost,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, inventory.ErrItemNotFound
		}
		return nil, fmt.Errorf("商品取得に失敗しました: %w", err)
	}

	return item, nil
}

// UpdateItem updates an existing item
// 既存の商品を更新
func (s *PostgreSQLStorage) UpdateItem(ctx context.Context, item *inventory.Item) error {
	query := `
		UPDATE items 
		SET name = $2, sku = $3, description = $4, category = $5, unit_cost = $6, updated_at = $7
		WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query,
		item.ID,
		item.Name,
		item.SKU,
		item.Description,
		item.Category,
		item.UnitCost,
		item.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("商品更新に失敗しました: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("更新行数の取得に失敗しました: %w", err)
	}

	if rowsAffected == 0 {
		return inventory.ErrItemNotFound
	}

	return nil
}

// CreateLocation creates a new location
// 新しいロケーションを作成
func (s *PostgreSQLStorage) CreateLocation(ctx context.Context, location *inventory.Location) error {
	query := `
		INSERT INTO locations (id, name, type, address, capacity, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := s.db.ExecContext(ctx, query,
		location.ID,
		location.Name,
		location.Type,
		location.Address,
		location.Capacity,
		location.IsActive,
		location.CreatedAt,
		location.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return inventory.ErrDuplicateLocation
		}
		return fmt.Errorf("ロケーション作成に失敗しました: %w", err)
	}

	return nil
}

// GetLocation retrieves a location by ID
// IDでロケーションを取得
func (s *PostgreSQLStorage) GetLocation(ctx context.Context, locationID string) (*inventory.Location, error) {
	query := `
		SELECT id, name, type, address, capacity, is_active, created_at, updated_at
		FROM locations 
		WHERE id = $1`

	location := &inventory.Location{}
	err := s.db.QueryRowContext(ctx, query, locationID).Scan(
		&location.ID,
		&location.Name,
		&location.Type,
		&location.Address,
		&location.Capacity,
		&location.IsActive,
		&location.CreatedAt,
		&location.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, inventory.ErrLocationNotFound
		}
		return nil, fmt.Errorf("ロケーション取得に失敗しました: %w", err)
	}

	return location, nil
}

// CreateLot creates a new lot record
// 新しいロット記録を作成
func (s *PostgreSQLStorage) CreateLot(ctx context.Context, lot *inventory.Lot) error {
	query := `
		INSERT INTO lots (id, number, item_id, quantity, unit_cost, expiry_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := s.db.ExecContext(ctx, query,
		lot.ID,
		lot.Number,
		lot.ItemID,
		lot.Quantity,
		lot.UnitCost,
		lot.ExpiryDate,
		lot.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("ロット作成に失敗しました: %w", err)
	}

	return nil
}

// GetLot retrieves a lot by ID
// IDでロットを取得
func (s *PostgreSQLStorage) GetLot(ctx context.Context, lotID string) (*inventory.Lot, error) {
	query := `
		SELECT id, number, item_id, quantity, unit_cost, expiry_date, created_at
		FROM lots 
		WHERE id = $1`

	lot := &inventory.Lot{}
	err := s.db.QueryRowContext(ctx, query, lotID).Scan(
		&lot.ID,
		&lot.Number,
		&lot.ItemID,
		&lot.Quantity,
		&lot.UnitCost,
		&lot.ExpiryDate,
		&lot.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, inventory.ErrLotNotFound
		}
		return nil, fmt.Errorf("ロット取得に失敗しました: %w", err)
	}

	return lot, nil
}

// GetLotsByItem retrieves all lots for a specific item
// 指定商品のすべてのロットを取得
func (s *PostgreSQLStorage) GetLotsByItem(ctx context.Context, itemID string) ([]inventory.Lot, error) {
	query := `
		SELECT id, number, item_id, quantity, unit_cost, expiry_date, created_at
		FROM lots 
		WHERE item_id = $1
		ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, itemID)
	if err != nil {
		return nil, fmt.Errorf("商品ロット取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var lots []inventory.Lot
	for rows.Next() {
		var lot inventory.Lot
		err := rows.Scan(
			&lot.ID,
			&lot.Number,
			&lot.ItemID,
			&lot.Quantity,
			&lot.UnitCost,
			&lot.ExpiryDate,
			&lot.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ロットスキャンに失敗しました: %w", err)
		}
		lots = append(lots, lot)
	}

	return lots, nil
}

// CreateAlert creates a new stock alert
// 新しい在庫アラートを作成
func (s *PostgreSQLStorage) CreateAlert(ctx context.Context, alert *inventory.StockAlert) error {
	query := `
		INSERT INTO stock_alerts (id, type, item_id, location_id, current_qty, threshold, message, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := s.db.ExecContext(ctx, query,
		alert.ID,
		alert.Type,
		alert.ItemID,
		alert.LocationID,
		alert.CurrentQty,
		alert.Threshold,
		alert.Message,
		alert.IsActive,
		alert.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("アラート作成に失敗しました: %w", err)
	}

	return nil
}

// GetActiveAlerts retrieves active alerts for a location
// ロケーションのアクティブアラートを取得
func (s *PostgreSQLStorage) GetActiveAlerts(ctx context.Context, locationID string) ([]inventory.StockAlert, error) {
	query := `
		SELECT id, type, item_id, location_id, current_qty, threshold, message, is_active, created_at, resolved_at
		FROM stock_alerts 
		WHERE location_id = $1 AND is_active = true
		ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, locationID)
	if err != nil {
		return nil, fmt.Errorf("アラート取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var alerts []inventory.StockAlert
	for rows.Next() {
		var alert inventory.StockAlert
		err := rows.Scan(
			&alert.ID,
			&alert.Type,
			&alert.ItemID,
			&alert.LocationID,
			&alert.CurrentQty,
			&alert.Threshold,
			&alert.Message,
			&alert.IsActive,
			&alert.CreatedAt,
			&alert.ResolvedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("アラートスキャンに失敗しました: %w", err)
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// ResolveAlert resolves an alert by setting it inactive
// アラートを非アクティブにして解決
func (s *PostgreSQLStorage) ResolveAlert(ctx context.Context, alertID string) error {
	now := time.Now()
	query := `
		UPDATE stock_alerts 
		SET is_active = false, resolved_at = $2
		WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query, alertID, now)
	if err != nil {
		return fmt.Errorf("アラート解決に失敗しました: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("更新行数の取得に失敗しました: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("アラートが見つかりません: %s", alertID)
	}

	return nil
}

// Ping checks database connectivity
// データベース接続をチェック
func (s *PostgreSQLStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Close closes the database connection
// データベース接続を閉じる
func (s *PostgreSQLStorage) Close() error {
	return s.db.Close()
}
