package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"

	"github.com/nemonet1337/zaiGoFramework/pkg/inventory"
	"github.com/nemonet1337/zaiGoFramework/pkg/inventory/storage"
)

// 基本的な在庫操作の使用例
func main() {
	fmt.Println("=== zaiGoFramework 基本使用例 ===")

	// ログ設定
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("ログ初期化に失敗しました:", err)
	}
	defer logger.Sync()

	// データベース接続（ローカル開発用）
	dsn := "host=localhost port=5432 user=inventory password=password dbname=inventory_db sslmode=disable"
	storage, err := storage.NewPostgreSQLStorage(dsn, logger)
	if err != nil {
		log.Fatal("ストレージ初期化に失敗しました:", err)
	}
	defer storage.Close()

	// 在庫マネージャー初期化
	config := &inventory.Config{
		AllowNegativeStock: false,
		DefaultLocation:    "WAREHOUSE-01",
		AuditEnabled:       true,
		LowStockThreshold:  50,
		AlertTimeout:       24 * time.Hour,
	}
	manager := inventory.NewManager(storage, nil, logger, config)

	ctx := context.Background()

	// 1. 商品とロケーションの作成
	fmt.Println("\n1. 商品とロケーションの初期設定...")
	err = createSampleData(ctx, storage)
	if err != nil {
		log.Printf("サンプルデータ作成エラー: %v", err)
	}

	// 2. 基本的な在庫操作
	fmt.Println("\n2. 基本的な在庫操作のデモ")
	
	// 在庫追加
	fmt.Println("✓ 在庫追加: 商品A を倉庫01に100個追加")
	err = manager.Add(ctx, "ITEM-A", "WAREHOUSE-01", 100, "PO-2024-001")
	if err != nil {
		log.Printf("在庫追加エラー: %v", err)
	} else {
		fmt.Println("  → 在庫追加完了")
	}

	// 在庫確認
	stock, err := manager.GetStock(ctx, "ITEM-A", "WAREHOUSE-01")
	if err != nil {
		log.Printf("在庫確認エラー: %v", err)
	} else {
		fmt.Printf("  → 現在在庫: %d個 (利用可能: %d個)\n", stock.Quantity, stock.Available)
	}

	// 在庫予約
	fmt.Println("\n✓ 在庫予約: 商品A を30個予約")
	err = manager.Reserve(ctx, "ITEM-A", "WAREHOUSE-01", 30, "ORDER-2024-001")
	if err != nil {
		log.Printf("在庫予約エラー: %v", err)
	} else {
		stock, _ = manager.GetStock(ctx, "ITEM-A", "WAREHOUSE-01")
		fmt.Printf("  → 予約後在庫: %d個 (予約済み: %d個, 利用可能: %d個)\n", 
			stock.Quantity, stock.Reserved, stock.Available)
	}

	// 在庫削除
	fmt.Println("\n✓ 在庫削除: 商品A を20個出庫")
	err = manager.Remove(ctx, "ITEM-A", "WAREHOUSE-01", 20, "SHIP-2024-001")
	if err != nil {
		log.Printf("在庫削除エラー: %v", err)
	} else {
		stock, _ = manager.GetStock(ctx, "ITEM-A", "WAREHOUSE-01")
		fmt.Printf("  → 出庫後在庫: %d個 (利用可能: %d個)\n", stock.Quantity, stock.Available)
	}

	// 在庫移動
	fmt.Println("\n✓ 在庫移動: 商品A を倉庫01から倉庫02に15個移動")
	err = manager.Transfer(ctx, "ITEM-A", "WAREHOUSE-01", "WAREHOUSE-02", 15, "TRANSFER-001")
	if err != nil {
		log.Printf("在庫移動エラー: %v", err)
	} else {
		fmt.Println("  → 在庫移動完了")
		
		// 移動後の在庫確認
		stock1, _ := manager.GetStock(ctx, "ITEM-A", "WAREHOUSE-01")
		stock2, _ := manager.GetStock(ctx, "ITEM-A", "WAREHOUSE-02")
		fmt.Printf("  → 倉庫01: %d個, 倉庫02: %d個\n", stock1.Quantity, stock2.Quantity)
	}

	// 3. 履歴確認
	fmt.Println("\n3. トランザクション履歴の確認")
	history, err := manager.GetHistory(ctx, "ITEM-A", 10)
	if err != nil {
		log.Printf("履歴取得エラー: %v", err)
	} else {
		fmt.Printf("✓ 商品Aの履歴 (%d件):\n", len(history))
		for i, tx := range history {
			fmt.Printf("  %d. %s - %s - 数量:%d - 参照:%s - 日時:%s\n",
				i+1, tx.Type, tx.ID[:8], tx.Quantity, tx.Reference, tx.CreatedAt.Format("15:04:05"))
		}
	}

	// 4. バッチ操作
	fmt.Println("\n4. バッチ操作のデモ")
	operations := []inventory.InventoryOperation{
		{
			Type:       inventory.OperationTypeAdd,
			ItemID:     "ITEM-B",
			LocationID: "WAREHOUSE-01",
			Quantity:   50,
			Reference:  "BATCH-ADD-001",
		},
		{
			Type:       inventory.OperationTypeAdd,
			ItemID:     "ITEM-B",
			LocationID: "WAREHOUSE-02",
			Quantity:   30,
			Reference:  "BATCH-ADD-002",
		},
		{
			Type:       inventory.OperationTypeAdjust,
			ItemID:     "ITEM-A",
			LocationID: "WAREHOUSE-01",
			Quantity:   100,
			Reference:  "BATCH-ADJUST-001",
		},
	}

	batch, err := manager.ExecuteBatch(ctx, operations)
	if err != nil {
		log.Printf("バッチ操作エラー: %v", err)
	} else {
		fmt.Printf("✓ バッチ操作完了 - 成功: %d件, 失敗: %d件\n", 
			batch.SuccessCount, batch.FailureCount)
		if batch.FailureCount > 0 {
			for _, batchErr := range batch.Errors {
				fmt.Printf("  エラー[%d]: %s\n", batchErr.OperationIndex, batchErr.Error)
			}
		}
	}

	// 5. アラート確認
	fmt.Println("\n5. アラートの確認")
	alerts, err := manager.GetAlerts(ctx, "WAREHOUSE-01")
	if err != nil {
		log.Printf("アラート取得エラー: %v", err)
	} else {
		if len(alerts) > 0 {
			fmt.Printf("✓ アクティブなアラート (%d件):\n", len(alerts))
			for _, alert := range alerts {
				fmt.Printf("  - %s: %s (商品: %s)\n", alert.Type, alert.Message, alert.ItemID)
			}
		} else {
			fmt.Println("✓ アクティブなアラートはありません")
		}
	}

	fmt.Println("\n=== デモ完了 ===")
}

// createSampleData creates sample items and locations
// サンプルの商品とロケーションを作成
func createSampleData(ctx context.Context, storage inventory.Storage) error {
	// サンプル商品
	items := []inventory.Item{
		{
			ID:          "ITEM-A",
			Name:        "商品A - ノートPC",
			SKU:         "SKU-LAPTOP-001",
			Description: "高性能ビジネスノートPC",
			Category:    "electronics",
			UnitCost:    80000.00,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "ITEM-B",
			Name:        "商品B - マウス",
			SKU:         "SKU-MOUSE-001",
			Description: "ワイヤレス光学マウス",
			Category:    "accessories",
			UnitCost:    2000.00,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, item := range items {
		err := storage.CreateItem(ctx, &item)
		if err != nil {
			// 既存の場合はエラーを無視
			continue
		}
		fmt.Printf("  商品作成: %s - %s\n", item.ID, item.Name)
	}

	// サンプルロケーション
	locations := []inventory.Location{
		{
			ID:        "WAREHOUSE-01",
			Name:      "メイン倉庫",
			Type:      "warehouse",
			Address:   "東京都江東区豊洲1-1-1",
			Capacity:  10000,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "WAREHOUSE-02",
			Name:      "サブ倉庫",
			Type:      "warehouse",
			Address:   "東京都大田区羽田1-1-1",
			Capacity:  5000,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, location := range locations {
		err := storage.CreateLocation(ctx, &location)
		if err != nil {
			// 既存の場合はエラーを無視
			continue
		}
		fmt.Printf("  ロケーション作成: %s - %s\n", location.ID, location.Name)
	}

	return nil
}
