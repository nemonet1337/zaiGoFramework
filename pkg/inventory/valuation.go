package inventory

import (
	"context"
	"fmt"
	"sort"
	"time"

	"go.uber.org/zap"
)

// ValuationEngineImpl implements the ValuationEngine interface
// ValuationEngineインターフェースの実装
type ValuationEngineImpl struct {
	storage Storage
	logger  *zap.Logger
}

// NewValuationEngine creates a new valuation engine
// 新しい在庫評価エンジンを作成
func NewValuationEngine(storage Storage, logger *zap.Logger) *ValuationEngineImpl {
	return &ValuationEngineImpl{
		storage: storage,
		logger:  logger,
	}
}

// CalculateValue calculates inventory value using specified method
// 指定された方法で在庫価値を計算
func (v *ValuationEngineImpl) CalculateValue(ctx context.Context, itemID, locationID string, method ValuationMethod) (float64, error) {
	// 現在の在庫を取得
	stock, err := v.storage.GetStock(ctx, itemID, locationID)
	if err != nil {
		return 0, NewStorageError("get_stock", "在庫取得に失敗しました", err)
	}

	if stock.Quantity <= 0 {
		return 0, nil
	}

	// 評価方法に応じて計算
	switch method {
	case ValuationMethodFIFO:
		return v.calculateFIFO(ctx, itemID, locationID, stock.Quantity)
	case ValuationMethodLIFO:
		return v.calculateLIFO(ctx, itemID, locationID, stock.Quantity)
	case ValuationMethodAverage:
		return v.calculateAverage(ctx, itemID, locationID, stock.Quantity)
	case ValuationMethodStandard:
		return v.calculateStandard(ctx, itemID, stock.Quantity)
	default:
		return 0, fmt.Errorf("未対応の評価方法です: %s", method)
	}
}

// CalculateTotalValue calculates total inventory value for a location
// ロケーションの総在庫価値を計算
func (v *ValuationEngineImpl) CalculateTotalValue(ctx context.Context, locationID string, method ValuationMethod) (float64, error) {
	// ロケーションの全在庫を取得
	stocks, err := v.storage.ListStockByLocation(ctx, locationID)
	if err != nil {
		return 0, NewStorageError("list_stock_by_location", "ロケーション在庫取得に失敗しました", err)
	}

	totalValue := 0.0
	for _, stock := range stocks {
		if stock.Quantity > 0 {
			value, err := v.CalculateValue(ctx, stock.ItemID, locationID, method)
			if err != nil {
				v.logger.Warn("商品価値計算でエラーが発生しました",
					zap.String("item_id", stock.ItemID),
					zap.String("location_id", locationID),
					zap.Error(err),
				)
				continue
			}
			totalValue += value
		}
	}

	return totalValue, nil
}

// GetAverageCost calculates average cost for an item
// 商品の平均原価を計算
func (v *ValuationEngineImpl) GetAverageCost(ctx context.Context, itemID string) (float64, error) {
	// 入庫トランザクションから平均原価を計算
	transactions, err := v.storage.GetTransactionHistory(ctx, itemID, 1000)
	if err != nil {
		return 0, NewStorageError("get_transaction_history", "トランザクション履歴取得に失敗しました", err)
	}

	totalCost := 0.0
	totalQuantity := int64(0)

	for _, tx := range transactions {
		if tx.Type == TransactionTypeInbound && tx.UnitCost != nil && *tx.UnitCost > 0 {
			totalCost += *tx.UnitCost * float64(tx.Quantity)
			totalQuantity += tx.Quantity
		}
	}

	if totalQuantity == 0 {
		return 0, fmt.Errorf("平均原価計算用のデータが不足しています")
	}

	return totalCost / float64(totalQuantity), nil
}

// calculateFIFO calculates inventory value using FIFO method
// FIFO法で在庫価値を計算
func (v *ValuationEngineImpl) calculateFIFO(ctx context.Context, itemID, locationID string, quantity int64) (float64, error) {
	// 入庫トランザクションを古い順に取得
	transactions, err := v.getInboundTransactions(ctx, itemID, locationID)
	if err != nil {
		return 0, err
	}

	// 古い順にソート
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].CreatedAt.Before(transactions[j].CreatedAt)
	})

	return v.calculateValueFromTransactions(transactions, quantity), nil
}

// calculateLIFO calculates inventory value using LIFO method
// LIFO法で在庫価値を計算
func (v *ValuationEngineImpl) calculateLIFO(ctx context.Context, itemID, locationID string, quantity int64) (float64, error) {
	// 入庫トランザクションを新しい順に取得
	transactions, err := v.getInboundTransactions(ctx, itemID, locationID)
	if err != nil {
		return 0, err
	}

	// 新しい順にソート
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].CreatedAt.After(transactions[j].CreatedAt)
	})

	return v.calculateValueFromTransactions(transactions, quantity), nil
}

// calculateAverage calculates inventory value using weighted average method
// 加重平均法で在庫価値を計算
func (v *ValuationEngineImpl) calculateAverage(ctx context.Context, itemID, locationID string, quantity int64) (float64, error) {
	averageCost, err := v.GetAverageCost(ctx, itemID)
	if err != nil {
		return 0, err
	}

	return averageCost * float64(quantity), nil
}

// calculateStandard calculates inventory value using standard cost method
// 標準原価法で在庫価値を計算
func (v *ValuationEngineImpl) calculateStandard(ctx context.Context, itemID string, quantity int64) (float64, error) {
	// 商品の標準原価を取得
	item, err := v.storage.GetItem(ctx, itemID)
	if err != nil {
		return 0, NewStorageError("get_item", "商品取得に失敗しました", err)
	}

	if item.UnitCost <= 0 {
		return 0, fmt.Errorf("商品に標準原価が設定されていません")
	}

	return item.UnitCost * float64(quantity), nil
}

// getInboundTransactions gets inbound transactions for an item at a location
// 指定商品・ロケーションの入庫トランザクションを取得
func (v *ValuationEngineImpl) getInboundTransactions(ctx context.Context, itemID, locationID string) ([]Transaction, error) {
	// 全トランザクション履歴を取得（実際にはより効率的な方法で実装）
	allTransactions, err := v.storage.GetTransactionHistory(ctx, itemID, 10000)
	if err != nil {
		return nil, NewStorageError("get_transaction_history", "トランザクション履歴取得に失敗しました", err)
	}

	var inboundTransactions []Transaction
	for _, tx := range allTransactions {
		// 指定ロケーションへの入庫または移動を対象
		if (tx.Type == TransactionTypeInbound && tx.ToLocation != nil && *tx.ToLocation == locationID) ||
			(tx.Type == TransactionTypeTransfer && tx.ToLocation != nil && *tx.ToLocation == locationID) {
			if tx.UnitCost != nil && *tx.UnitCost > 0 {
				inboundTransactions = append(inboundTransactions, tx)
			}
		}
	}

	return inboundTransactions, nil
}

// calculateValueFromTransactions calculates value from sorted transactions
// ソートされたトランザクションから価値を計算
func (v *ValuationEngineImpl) calculateValueFromTransactions(transactions []Transaction, quantity int64) float64 {
	totalValue := 0.0
	remainingQty := quantity

	for _, tx := range transactions {
		if remainingQty <= 0 {
			break
		}

		if tx.UnitCost == nil {
			continue
		}

		// このトランザクションから使用する数量
		useQty := tx.Quantity
		if useQty > remainingQty {
			useQty = remainingQty
		}

		totalValue += *tx.UnitCost * float64(useQty)
		remainingQty -= useQty
	}

	return totalValue
}

// AnalyticsEngineImpl implements the AnalyticsEngine interface
// AnalyticsEngineインターフェースの実装
type AnalyticsEngineImpl struct {
	storage Storage
	logger  *zap.Logger
}

// NewAnalyticsEngine creates a new analytics engine
// 新しい分析エンジンを作成
func NewAnalyticsEngine(storage Storage, logger *zap.Logger) *AnalyticsEngineImpl {
	return &AnalyticsEngineImpl{
		storage: storage,
		logger:  logger,
	}
}

// CalculateABCClassification performs ABC analysis on inventory
// 在庫のABC分析を実行
func (a *AnalyticsEngineImpl) CalculateABCClassification(ctx context.Context, locationID string) (map[string]string, error) {
	// ロケーションの全在庫を取得
	stocks, err := a.storage.ListStockByLocation(ctx, locationID)
	if err != nil {
		return nil, NewStorageError("list_stock_by_location", "ロケーション在庫取得に失敗しました", err)
	}

	// 各商品の年間売上高を計算（簡略化版）
	itemValues := make(map[string]float64)
	for _, stock := range stocks {
		// 実際には過去12ヶ月の出庫データから計算すべき
		// ここでは簡略化して在庫数量 × 単価で代用
		item, err := a.storage.GetItem(ctx, stock.ItemID)
		if err != nil {
			continue
		}
		
		// 年間出庫予想値として在庫数量の10倍を使用（仮定）
		estimatedAnnualSales := float64(stock.Quantity * 10) * item.UnitCost
		itemValues[stock.ItemID] = estimatedAnnualSales
	}

	// 値でソートして分類
	return a.classifyABC(itemValues), nil
}

// classifyABC classifies items into A, B, C categories
// 商品をA、B、Cカテゴリに分類
func (a *AnalyticsEngineImpl) classifyABC(itemValues map[string]float64) map[string]string {
	// 値の順序でアイテムをソート
	type ItemValue struct {
		ItemID string
		Value  float64
	}

	var items []ItemValue
	totalValue := 0.0
	for itemID, value := range itemValues {
		items = append(items, ItemValue{ItemID: itemID, Value: value})
		totalValue += value
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Value > items[j].Value
	})

	// ABC分類（80-15-5の法則）
	classification := make(map[string]string)
	cumulativeValue := 0.0
	
	for _, item := range items {
		cumulativeValue += item.Value
		percentage := cumulativeValue / totalValue
		
		if percentage <= 0.8 {
			classification[item.ItemID] = "A"
		} else if percentage <= 0.95 {
			classification[item.ItemID] = "B"
		} else {
			classification[item.ItemID] = "C"
		}
	}

	return classification
}

// GetTurnoverRate calculates inventory turnover rate for an item
// 商品の在庫回転率を計算
func (a *AnalyticsEngineImpl) GetTurnoverRate(ctx context.Context, itemID string, period time.Duration) (float64, error) {
	// 指定期間の出庫量を計算
	transactions, err := a.storage.GetTransactionHistory(ctx, itemID, 10000)
	if err != nil {
		return 0, NewStorageError("get_transaction_history", "トランザクション履歴取得に失敗しました", err)
	}

	cutoffDate := time.Now().Add(-period)
	outboundQuantity := int64(0)

	for _, tx := range transactions {
		if tx.CreatedAt.After(cutoffDate) && tx.Type == TransactionTypeOutbound {
			outboundQuantity += tx.Quantity
		}
	}

	// 平均在庫量を計算（簡略化：現在の総在庫量を使用）
	// TODO: より正確な平均在庫計算を実装
	avgInventory := int64(100) // 仮の値

	if avgInventory == 0 {
		return 0, nil
	}

	// 回転率 = 期間中の出庫量 / 平均在庫量
	turnoverRate := float64(outboundQuantity) / float64(avgInventory)
	
	// 年間回転率に換算
	daysInPeriod := period.Hours() / 24
	yearlyTurnoverRate := turnoverRate * (365 / daysInPeriod)

	return yearlyTurnoverRate, nil
}

// GetSlowMovingItems identifies slow-moving items
// 動きの遅い商品を特定
func (a *AnalyticsEngineImpl) GetSlowMovingItems(ctx context.Context, locationID string, threshold time.Duration) ([]string, error) {
	stocks, err := a.storage.ListStockByLocation(ctx, locationID)
	if err != nil {
		return nil, NewStorageError("list_stock_by_location", "ロケーション在庫取得に失敗しました", err)
	}

	var slowMovingItems []string
	cutoffDate := time.Now().Add(-threshold)

	for _, stock := range stocks {
		// 各商品の最新出庫日を確認
		transactions, err := a.storage.GetTransactionHistory(ctx, stock.ItemID, 100)
		if err != nil {
			continue
		}

		hasRecentActivity := false
		for _, tx := range transactions {
			if tx.Type == TransactionTypeOutbound && tx.CreatedAt.After(cutoffDate) {
				hasRecentActivity = true
				break
			}
		}

		if !hasRecentActivity && stock.Quantity > 0 {
			slowMovingItems = append(slowMovingItems, stock.ItemID)
		}
	}

	return slowMovingItems, nil
}

// GenerateStockReport generates inventory reports
// 在庫レポートを生成
func (a *AnalyticsEngineImpl) GenerateStockReport(ctx context.Context, locationID string, reportType ReportType) ([]byte, error) {
	switch reportType {
	case ReportTypeStock:
		return a.generateStockReport(ctx, locationID)
	case ReportTypeABC:
		return a.generateABCReport(ctx, locationID)
	default:
		return nil, fmt.Errorf("未対応のレポートタイプです: %s", reportType)
	}
}

// generateStockReport generates basic stock report
// 基本在庫レポートを生成
func (a *AnalyticsEngineImpl) generateStockReport(ctx context.Context, locationID string) ([]byte, error) {
	stocks, err := a.storage.ListStockByLocation(ctx, locationID)
	if err != nil {
		return nil, err
	}

	// 簡略化：CSVフォーマットで出力
	report := "商品ID,在庫数量,予約済み,利用可能,最終更新\n"
	for _, stock := range stocks {
		line := fmt.Sprintf("%s,%d,%d,%d,%s\n",
			stock.ItemID, stock.Quantity, stock.Reserved, stock.Available,
			stock.UpdatedAt.Format("2006-01-02 15:04:05"))
		report += line
	}

	return []byte(report), nil
}

// generateABCReport generates ABC analysis report
// ABC分析レポートを生成
func (a *AnalyticsEngineImpl) generateABCReport(ctx context.Context, locationID string) ([]byte, error) {
	classification, err := a.CalculateABCClassification(ctx, locationID)
	if err != nil {
		return nil, err
	}

	// 簡略化：CSVフォーマットで出力
	report := "商品ID,分類\n"
	for itemID, class := range classification {
		line := fmt.Sprintf("%s,%s\n", itemID, class)
		report += line
	}

	return []byte(report), nil
}
