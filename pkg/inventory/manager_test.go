package inventory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockStorage はテスト用のStorageモック
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Begin(ctx context.Context) (Transaction, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return Transaction{}, args.Error(1)
	}
	return args.Get(0).(Transaction), args.Error(1)
}

func (m *MockStorage) CreateStock(ctx context.Context, stock *Stock) error {
	args := m.Called(ctx, stock)
	return args.Error(0)
}

func (m *MockStorage) UpdateStock(ctx context.Context, stock *Stock) error {
	args := m.Called(ctx, stock)
	return args.Error(0)
}

func (m *MockStorage) GetStock(ctx context.Context, itemID, locationID string) (*Stock, error) {
	args := m.Called(ctx, itemID, locationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Stock), args.Error(1)
}

func (m *MockStorage) ListStockByLocation(ctx context.Context, locationID string) ([]Stock, error) {
	args := m.Called(ctx, locationID)
	return args.Get(0).([]Stock), args.Error(1)
}

func (m *MockStorage) CreateTransaction(ctx context.Context, tx *Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *MockStorage) GetTransactionHistory(ctx context.Context, itemID string, limit int) ([]Transaction, error) {
	args := m.Called(ctx, itemID, limit)
	return args.Get(0).([]Transaction), args.Error(1)
}

func (m *MockStorage) CreateItem(ctx context.Context, item *Item) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockStorage) GetItem(ctx context.Context, itemID string) (*Item, error) {
	args := m.Called(ctx, itemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Item), args.Error(1)
}

func (m *MockStorage) UpdateItem(ctx context.Context, item *Item) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockStorage) CreateLocation(ctx context.Context, location *Location) error {
	args := m.Called(ctx, location)
	return args.Error(0)
}

func (m *MockStorage) GetLocation(ctx context.Context, locationID string) (*Location, error) {
	args := m.Called(ctx, locationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Location), args.Error(1)
}

func (m *MockStorage) CreateLot(ctx context.Context, lot *Lot) error {
	args := m.Called(ctx, lot)
	return args.Error(0)
}

func (m *MockStorage) GetLot(ctx context.Context, lotID string) (*Lot, error) {
	args := m.Called(ctx, lotID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Lot), args.Error(1)
}

func (m *MockStorage) GetLotsByItem(ctx context.Context, itemID string) ([]Lot, error) {
	args := m.Called(ctx, itemID)
	return args.Get(0).([]Lot), args.Error(1)
}

func (m *MockStorage) CreateAlert(ctx context.Context, alert *StockAlert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockStorage) GetActiveAlerts(ctx context.Context, locationID string) ([]StockAlert, error) {
	args := m.Called(ctx, locationID)
	return args.Get(0).([]StockAlert), args.Error(1)
}

func (m *MockStorage) ResolveAlert(ctx context.Context, alertID string) error {
	args := m.Called(ctx, alertID)
	return args.Error(0)
}

func (m *MockStorage) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

// TestManager_Add は在庫追加機能のテスト
func TestManager_Add(t *testing.T) {
	mockStorage := new(MockStorage)
	logger := zap.NewNop()
	config := &Config{
		AllowNegativeStock: false,
		DefaultLocation:    "DEFAULT",
		AuditEnabled:       true,
		LowStockThreshold:  10,
	}

	manager := NewManager(mockStorage, nil, logger, config)
	ctx := context.Background()

	// テスト用のサンプルデータ
	item := &Item{
		ID:       "TEST-ITEM",
		Name:     "テスト商品",
		UnitCost: 1000.0,
	}
	location := &Location{
		ID:   "TEST-LOC",
		Name: "テストロケーション",
	}

	// モックの期待値設定
	mockStorage.On("GetItem", ctx, "TEST-ITEM").Return(item, nil)
	mockStorage.On("GetLocation", ctx, "TEST-LOC").Return(location, nil)
	mockStorage.On("GetStock", ctx, "TEST-ITEM", "TEST-LOC").Return(nil, ErrStockNotFound)
	mockStorage.On("CreateStock", ctx, mock.AnythingOfType("*inventory.Stock")).Return(nil)
	mockStorage.On("CreateTransaction", ctx, mock.AnythingOfType("*inventory.Transaction")).Return(nil)

	// テスト実行
	err := manager.Add(ctx, "TEST-ITEM", "TEST-LOC", 100, "TEST-REF")

	// アサーション
	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}

// TestManager_Remove は在庫削除機能のテスト
func TestManager_Remove(t *testing.T) {
	mockStorage := new(MockStorage)
	logger := zap.NewNop()
	config := &Config{
		AllowNegativeStock: false,
		DefaultLocation:    "DEFAULT",
		AuditEnabled:       true,
		LowStockThreshold:  10,
	}

	manager := NewManager(mockStorage, nil, logger, config)
	ctx := context.Background()

	// テスト用のサンプルデータ
	item := &Item{
		ID:       "TEST-ITEM",
		Name:     "テスト商品",
		UnitCost: 1000.0,
	}
	location := &Location{
		ID:   "TEST-LOC",
		Name: "テストロケーション",
	}
	stock := &Stock{
		ItemID:     "TEST-ITEM",
		LocationID: "TEST-LOC",
		Quantity:   100,
		Reserved:   0,
		Available:  100,
		Version:    1,
	}

	// モックの期待値設定
	mockStorage.On("GetItem", ctx, "TEST-ITEM").Return(item, nil)
	mockStorage.On("GetLocation", ctx, "TEST-LOC").Return(location, nil)
	mockStorage.On("GetStock", ctx, "TEST-ITEM", "TEST-LOC").Return(stock, nil)
	mockStorage.On("UpdateStock", ctx, mock.AnythingOfType("*inventory.Stock")).Return(nil)
	mockStorage.On("CreateTransaction", ctx, mock.AnythingOfType("*inventory.Transaction")).Return(nil)

	// テスト実行
	err := manager.Remove(ctx, "TEST-ITEM", "TEST-LOC", 50, "TEST-REF")

	// アサーション
	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}

// TestManager_InsufficientStock は在庫不足エラーのテスト
func TestManager_InsufficientStock(t *testing.T) {
	mockStorage := new(MockStorage)
	logger := zap.NewNop()
	config := &Config{
		AllowNegativeStock: false,
		DefaultLocation:    "DEFAULT",
		AuditEnabled:       true,
		LowStockThreshold:  10,
	}

	manager := NewManager(mockStorage, nil, logger, config)
	ctx := context.Background()

	// テスト用のサンプルデータ
	item := &Item{
		ID:       "TEST-ITEM",
		Name:     "テスト商品",
		UnitCost: 1000.0,
	}
	location := &Location{
		ID:   "TEST-LOC",
		Name: "テストロケーション",
	}
	stock := &Stock{
		ItemID:     "TEST-ITEM",
		LocationID: "TEST-LOC",
		Quantity:   10,
		Reserved:   0,
		Available:  10,
		Version:    1,
	}

	// モックの期待値設定
	mockStorage.On("GetItem", ctx, "TEST-ITEM").Return(item, nil)
	mockStorage.On("GetLocation", ctx, "TEST-LOC").Return(location, nil)
	mockStorage.On("GetStock", ctx, "TEST-ITEM", "TEST-LOC").Return(stock, nil)

	// テスト実行 - 在庫数を超える削除を試行
	err := manager.Remove(ctx, "TEST-ITEM", "TEST-LOC", 50, "TEST-REF")

	// アサーション - 在庫不足エラーになることを確認
	assert.Equal(t, ErrInsufficientStock, err)
	mockStorage.AssertExpectations(t)
}

// TestManager_Reserve は在庫予約機能のテスト
func TestManager_Reserve(t *testing.T) {
	mockStorage := new(MockStorage)
	logger := zap.NewNop()
	config := &Config{
		AllowNegativeStock: false,
		DefaultLocation:    "DEFAULT",
		AuditEnabled:       true,
		LowStockThreshold:  10,
	}

	manager := NewManager(mockStorage, nil, logger, config)
	ctx := context.Background()

	// テスト用のサンプルデータ
	stock := &Stock{
		ItemID:     "TEST-ITEM",
		LocationID: "TEST-LOC",
		Quantity:   100,
		Reserved:   0,
		Available:  100,
		Version:    1,
	}

	// モックの期待値設定
	mockStorage.On("GetStock", ctx, "TEST-ITEM", "TEST-LOC").Return(stock, nil)
	mockStorage.On("UpdateStock", ctx, mock.AnythingOfType("*inventory.Stock")).Return(nil)

	// テスト実行
	err := manager.Reserve(ctx, "TEST-ITEM", "TEST-LOC", 30, "TEST-RESERVE")

	// アサーション
	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}

// TestManager_BatchOperation はバッチ操作のテスト
func TestManager_BatchOperation(t *testing.T) {
	mockStorage := new(MockStorage)
	logger := zap.NewNop()
	config := &Config{
		AllowNegativeStock: false,
		DefaultLocation:    "DEFAULT",
		AuditEnabled:       true,
		LowStockThreshold:  10,
	}

	manager := NewManager(mockStorage, nil, logger, config)
	ctx := context.Background()

	// テスト用のサンプルデータ
	item := &Item{
		ID:       "TEST-ITEM",
		Name:     "テスト商品",
		UnitCost: 1000.0,
	}
	location := &Location{
		ID:   "TEST-LOC",
		Name: "テストロケーション",
	}

	// バッチ操作
	operations := []InventoryOperation{
		{
			Type:       OperationTypeAdd,
			ItemID:     "TEST-ITEM",
			LocationID: "TEST-LOC",
			Quantity:   100,
			Reference:  "BATCH-001",
		},
	}

	// モックの期待値設定
	mockStorage.On("GetItem", ctx, "TEST-ITEM").Return(item, nil)
	mockStorage.On("GetLocation", ctx, "TEST-LOC").Return(location, nil)
	mockStorage.On("GetStock", ctx, "TEST-ITEM", "TEST-LOC").Return(nil, ErrStockNotFound)
	mockStorage.On("CreateStock", ctx, mock.AnythingOfType("*inventory.Stock")).Return(nil)
	mockStorage.On("CreateTransaction", ctx, mock.AnythingOfType("*inventory.Transaction")).Return(nil)

	// テスト実行
	batch, err := manager.ExecuteBatch(ctx, operations)

	// アサーション
	assert.NoError(t, err)
	assert.NotNil(t, batch)
	assert.Equal(t, 1, batch.SuccessCount)
	assert.Equal(t, 0, batch.FailureCount)
	mockStorage.AssertExpectations(t)
}

// TestManager_GetTotalStock は総在庫取得のテスト
func TestManager_GetTotalStock(t *testing.T) {
	mockStorage := new(MockStorage)
	logger := zap.NewNop()
	config := &Config{}

	manager := NewManager(mockStorage, nil, logger, config)
	ctx := context.Background()

	// テスト用のサンプルデータ
	item := &Item{
		ID:       "TEST-ITEM",
		Name:     "テスト商品",
		UnitCost: 1000.0,
	}

	// モックの期待値設定
	mockStorage.On("GetItem", ctx, "TEST-ITEM").Return(item, nil)

	// テスト実行
	totalStock, err := manager.GetTotalStock(ctx, "TEST-ITEM")

	// アサーション
	assert.NoError(t, err)
	assert.Equal(t, int64(0), totalStock) // 現在の実装では0を返す
	mockStorage.AssertExpectations(t)
}

// TestManager_GetHistoryByDateRange は日付範囲履歴取得のテスト
func TestManager_GetHistoryByDateRange(t *testing.T) {
	mockStorage := new(MockStorage)
	logger := zap.NewNop()
	config := &Config{}

	manager := NewManager(mockStorage, nil, logger, config)
	ctx := context.Background()

	// テスト用のサンプルデータ
	item := &Item{
		ID:       "TEST-ITEM",
		Name:     "テスト商品",
		UnitCost: 1000.0,
	}

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	transactions := []Transaction{
		{
			ID:        "TX-001",
			Type:      TransactionTypeInbound,
			ItemID:    "TEST-ITEM",
			Quantity:  100,
			CreatedAt: time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC),
		},
	}

	// モックの期待値設定
	mockStorage.On("GetItem", ctx, "TEST-ITEM").Return(item, nil)
	mockStorage.On("GetTransactionHistory", ctx, "TEST-ITEM", 10000).Return(transactions, nil)

	// テスト実行
	result, err := manager.GetHistoryByDateRange(ctx, "TEST-ITEM", from, to)

	// アサーション
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "TX-001", result[0].ID)
	mockStorage.AssertExpectations(t)
}

// TestValidationErrors はバリデーションエラーのテスト
func TestValidationErrors(t *testing.T) {
	mockStorage := new(MockStorage)
	logger := zap.NewNop()
	config := &Config{}

	manager := NewManager(mockStorage, nil, logger, config)
	ctx := context.Background()

	// 空の商品IDでのテスト
	_, err := manager.GetTotalStock(ctx, "")
	assert.Error(t, err)
	assert.IsType(t, &ValidationError{}, err)

	// 負の数量でのテスト
	err = manager.Add(ctx, "TEST-ITEM", "TEST-LOC", -10, "TEST-REF")
	assert.Error(t, err)
	assert.IsType(t, &ValidationError{}, err)

	// 無効な日付範囲でのテスト
	from := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err = manager.GetHistoryByDateRange(ctx, "TEST-ITEM", from, to)
	assert.Error(t, err)
	assert.IsType(t, &ValidationError{}, err)
}

// ベンチマークテスト
func BenchmarkManager_Add(b *testing.B) {
	mockStorage := new(MockStorage)
	logger := zap.NewNop()
	config := &Config{
		AllowNegativeStock: false,
		DefaultLocation:    "DEFAULT",
		AuditEnabled:       true,
		LowStockThreshold:  10,
	}

	manager := NewManager(mockStorage, nil, logger, config)
	ctx := context.Background()

	// テスト用のサンプルデータ
	item := &Item{
		ID:       "TEST-ITEM",
		Name:     "テスト商品",
		UnitCost: 1000.0,
	}
	location := &Location{
		ID:   "TEST-LOC",
		Name: "テストロケーション",
	}

	// モックの期待値設定
	mockStorage.On("GetItem", ctx, "TEST-ITEM").Return(item, nil)
	mockStorage.On("GetLocation", ctx, "TEST-LOC").Return(location, nil)
	mockStorage.On("GetStock", ctx, "TEST-ITEM", "TEST-LOC").Return(nil, ErrStockNotFound)
	mockStorage.On("CreateStock", ctx, mock.AnythingOfType("*inventory.Stock")).Return(nil)
	mockStorage.On("CreateTransaction", ctx, mock.AnythingOfType("*inventory.Transaction")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.Add(ctx, "TEST-ITEM", "TEST-LOC", int64(i+1), "BENCH-TEST")
	}
}

// BenchmarkManager_GetStock はGetStock操作のベンチマークテスト
func BenchmarkManager_GetStock(b *testing.B) {
	mockStorage := new(MockStorage)
	logger := zap.NewNop()
	config := &Config{}

	manager := NewManager(mockStorage, nil, logger, config)
	ctx := context.Background()

	stock := &Stock{
		ItemID:     "TEST-ITEM",
		LocationID: "TEST-LOC",
		Quantity:   100,
		Available:  100,
	}

	// モックの期待値設定
	mockStorage.On("GetStock", ctx, "TEST-ITEM", "TEST-LOC").Return(stock, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GetStock(ctx, "TEST-ITEM", "TEST-LOC")
	}
}
