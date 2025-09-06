package inventory

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// ValidateItemID 商品IDの形式をバリデーション
func ValidateItemID(itemID string) error {
	if itemID == "" {
		return NewValidationError("item_id", "商品IDが空です", itemID)
	}
	if len(itemID) > 255 {
		return NewValidationError("item_id", "商品IDが長すぎます", itemID)
	}
	// 英数字、ハイフン、アンダースコアのみ許可
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validPattern.MatchString(itemID) {
		return NewValidationError("item_id", "商品IDに無効な文字が含まれています", itemID)
	}
	return nil
}

// ValidateQuantity 数量をバリデーション
func ValidateQuantity(quantity int64, allowNegative bool) error {
	if !allowNegative && quantity < 0 {
		return NewValidationError("quantity", "負の数量は許可されていません", fmt.Sprintf("%d", quantity))
	}
	if quantity < -999999999 || quantity > 999999999 {
		return NewValidationError("quantity", "数量が有効範囲を超えています", fmt.Sprintf("%d", quantity))
	}
	return nil
}

// ValidateItemName 商品名をバリデーション
func ValidateItemName(name string) error {
	if strings.TrimSpace(name) == "" {
		return NewValidationError("name", "商品名が空です", name)
	}
	if len(name) > 500 {
		return NewValidationError("name", "商品名が長すぎます", name)
	}
	return nil
}

// ValidateLocationID ロケーションIDの形式をバリデーション
func ValidateLocationID(locationID string) error {
	if locationID == "" {
		return NewValidationError("location_id", "ロケーションIDが空です", locationID)
	}
	if len(locationID) > 255 {
		return NewValidationError("location_id", "ロケーションIDが長すぎます", locationID)
	}
	// 英数字、ハイフン、アンダースコアのみ許可
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validPattern.MatchString(locationID) {
		return NewValidationError("location_id", "ロケーションIDに無効な文字が含まれています", locationID)
	}
	return nil
}

// ValidateLocationName ロケーション名をバリデーション
func ValidateLocationName(name string) error {
	if strings.TrimSpace(name) == "" {
		return NewValidationError("name", "ロケーション名が空です", name)
	}
	if len(name) > 500 {
		return NewValidationError("name", "ロケーション名が長すぎます", name)
	}
	return nil
}

// ValidateSKU SKUの形式をバリデーション
func ValidateSKU(sku string) error {
	if sku == "" {
		return nil // SKUは任意
	}
	if len(sku) > 255 {
		return NewValidationError("sku", "SKUが長すぎます", sku)
	}
	// 英数字、ハイフン、アンダースコア、ドットのみ許可
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)
	if !validPattern.MatchString(sku) {
		return NewValidationError("sku", "SKUに無効な文字が含まれています", sku)
	}
	return nil
}

// ValidateCategory カテゴリの形式をバリデーション
func ValidateCategory(category string) error {
	if category == "" {
		return nil // カテゴリは任意
	}
	if len(category) > 255 {
		return NewValidationError("category", "カテゴリが長すぎます", category)
	}
	return nil
}

// ValidateDescription 説明の形式をバリデーション
func ValidateDescription(description string) error {
	if description == "" {
		return nil // 説明は任意
	}
	if len(description) > 2000 {
		return NewValidationError("description", "説明が長すぎます", description)
	}
	return nil
}

// ValidateReference 参照番号の形式をバリデーション
func ValidateReference(reference string) error {
	if reference == "" {
		return nil // 参照番号は任意
	}
	if len(reference) > 500 {
		return NewValidationError("reference", "参照番号が長すぎます", reference)
	}
	return nil
}

// ValidateLotNumber ロット番号の形式をバリデーション
func ValidateLotNumber(lotNumber string) error {
	if lotNumber == "" {
		return NewValidationError("lot_number", "ロット番号が空です", lotNumber)
	}
	if len(lotNumber) > 255 {
		return NewValidationError("lot_number", "ロット番号が長すぎます", lotNumber)
	}
	// 英数字、ハイフン、アンダースコア、ドットのみ許可
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)
	if !validPattern.MatchString(lotNumber) {
		return NewValidationError("lot_number", "ロット番号に無効な文字が含まれています", lotNumber)
	}
	return nil
}

// ValidateUnitCost 単価をバリデーション
func ValidateUnitCost(unitCost float64) error {
	if unitCost < 0 {
		return NewValidationError("unit_cost", "単価は0以上である必要があります", fmt.Sprintf("%.2f", unitCost))
	}
	if unitCost > 999999.9999 {
		return NewValidationError("unit_cost", "単価が有効範囲を超えています", fmt.Sprintf("%.2f", unitCost))
	}
	return nil
}

// ValidateThreshold 閾値をバリデーション
func ValidateThreshold(threshold int64) error {
	if threshold < 0 {
		return NewValidationError("threshold", "閾値は0以上である必要があります", fmt.Sprintf("%d", threshold))
	}
	if threshold > 999999999 {
		return NewValidationError("threshold", "閾値が有効範囲を超えています", fmt.Sprintf("%d", threshold))
	}
	return nil
}

// ValidateCapacity 容量をバリデーション
func ValidateCapacity(capacity int64) error {
	if capacity < 0 {
		return NewValidationError("capacity", "容量は0以上である必要があります", fmt.Sprintf("%d", capacity))
	}
	if capacity > 999999999999 {
		return NewValidationError("capacity", "容量が有効範囲を超えています", fmt.Sprintf("%d", capacity))
	}
	return nil
}

// ValidateVersion バージョンをバリデーション
func ValidateVersion(version int64) error {
	if version < 1 {
		return NewValidationError("version", "バージョンは1以上である必要があります", fmt.Sprintf("%d", version))
	}
	return nil
}

// ValidateUserID ユーザーIDをバリデーション
func ValidateUserID(userID string) error {
	if userID == "" {
		return NewValidationError("user_id", "ユーザーIDが空です", userID)
	}
	if len(userID) > 255 {
		return NewValidationError("user_id", "ユーザーIDが長すぎます", userID)
	}
	return nil
}

// ValidateTransactionType トランザクション種別をバリデーション
func ValidateTransactionType(transactionType string) error {
	validTypes := map[string]bool{
		TransactionTypeInbound:  true,
		TransactionTypeOutbound: true,
		TransactionTypeTransfer: true,
		TransactionTypeAdjust:   true,
	}
	
	if !validTypes[transactionType] {
		return NewValidationError("transaction_type", "無効なトランザクション種別です", transactionType)
	}
	return nil
}

// ValidateAlertType アラート種別をバリデーション
func ValidateAlertType(alertType string) error {
	validTypes := map[string]bool{
		AlertTypeLowStock:    true,
		AlertTypeExpiry:      true,
		AlertTypeOverstock:   true,
		AlertTypeSystemError: true,
	}
	
	if !validTypes[alertType] {
		return NewValidationError("alert_type", "無効なアラート種別です", alertType)
	}
	return nil
}

// ValidateOperationType オペレーション種別をバリデーション
func ValidateOperationType(operationType string) error {
	validTypes := map[string]bool{
		OperationTypeAdd:      true,
		OperationTypeRemove:   true,
		OperationTypeTransfer: true,
		OperationTypeAdjust:   true,
	}
	
	if !validTypes[operationType] {
		return NewValidationError("operation_type", "無効なオペレーション種別です", operationType)
	}
	return nil
}

// ValidateItem 商品全体をバリデーション
func ValidateItem(item *Item) error {
	if item == nil {
		return NewValidationError("item", "商品が指定されていません", "nil")
	}

	if err := ValidateItemID(item.ID); err != nil {
		return err
	}
	if err := ValidateItemName(item.Name); err != nil {
		return err
	}
	if err := ValidateSKU(item.SKU); err != nil {
		return err
	}
	if err := ValidateCategory(item.Category); err != nil {
		return err
	}
	if err := ValidateDescription(item.Description); err != nil {
		return err
	}
	if err := ValidateUnitCost(item.UnitCost); err != nil {
		return err
	}

	return nil
}

// ValidateLocation ロケーション全体をバリデーション
func ValidateLocation(location *Location) error {
	if location == nil {
		return NewValidationError("location", "ロケーションが指定されていません", "nil")
	}

	if err := ValidateLocationID(location.ID); err != nil {
		return err
	}
	if err := ValidateLocationName(location.Name); err != nil {
		return err
	}
	if err := ValidateCapacity(location.Capacity); err != nil {
		return err
	}

	return nil
}

// ValidateStock 在庫全体をバリデーション
func ValidateStock(stock *Stock, allowNegative bool) error {
	if stock == nil {
		return NewValidationError("stock", "在庫が指定されていません", "nil")
	}

	if err := ValidateItemID(stock.ItemID); err != nil {
		return err
	}
	if err := ValidateLocationID(stock.LocationID); err != nil {
		return err
	}
	if err := ValidateQuantity(stock.Quantity, allowNegative); err != nil {
		return err
	}
	if err := ValidateQuantity(stock.Reserved, false); err != nil {
		return err
	}
	if err := ValidateVersion(stock.Version); err != nil {
		return err
	}
	if err := ValidateUserID(stock.UpdatedBy); err != nil {
		return err
	}

	return nil
}

// ValidateLot ロット全体をバリデーション
func ValidateLot(lot *Lot) error {
	if lot == nil {
		return NewValidationError("lot", "ロットが指定されていません", "nil")
	}

	if err := ValidateItemID(lot.ItemID); err != nil {
		return err
	}
	if err := ValidateLotNumber(lot.Number); err != nil {
		return err
	}
	if err := ValidateQuantity(lot.Quantity, false); err != nil {
		return err
	}
	if err := ValidateUnitCost(lot.UnitCost); err != nil {
		return err
	}

	return nil
}

// ValidateTransaction トランザクション全体をバリデーション
func ValidateTransaction(tx *Transaction) error {
	if tx == nil {
		return NewValidationError("transaction", "トランザクションが指定されていません", "nil")
	}

	if err := ValidateTransactionType(tx.Type); err != nil {
		return err
	}
	if err := ValidateItemID(tx.ItemID); err != nil {
		return err
	}
	if err := ValidateQuantity(tx.Quantity, true); err != nil {
		return err
	}
	if err := ValidateReference(tx.Reference); err != nil {
		return err
	}
	if err := ValidateUserID(tx.CreatedBy); err != nil {
		return err
	}

	// ロケーションの存在確認（任意フィールド）
	if tx.FromLocation != nil {
		if err := ValidateLocationID(*tx.FromLocation); err != nil {
			return err
		}
	}
	if tx.ToLocation != nil {
		if err := ValidateLocationID(*tx.ToLocation); err != nil {
			return err
		}
	}

	// ロット番号の確認（任意フィールド）
	if tx.LotNumber != "" {
		if err := ValidateLotNumber(tx.LotNumber); err != nil {
			return err
		}
	}

	// 単価の確認（任意フィールド）
	if tx.UnitCost != nil {
		if err := ValidateUnitCost(*tx.UnitCost); err != nil {
			return err
		}
	}

	return nil
}

// ValidateStockAlert アラート全体をバリデーション
func ValidateStockAlert(alert *StockAlert) error {
	if alert == nil {
		return NewValidationError("alert", "アラートが指定されていません", "nil")
	}

	if err := ValidateAlertType(alert.Type); err != nil {
		return err
	}
	if err := ValidateItemID(alert.ItemID); err != nil {
		return err
	}
	if err := ValidateLocationID(alert.LocationID); err != nil {
		return err
	}
	if err := ValidateQuantity(alert.CurrentQty, true); err != nil {
		return err
	}
	if err := ValidateThreshold(alert.Threshold); err != nil {
		return err
	}

	if strings.TrimSpace(alert.Message) == "" {
		return NewValidationError("message", "アラートメッセージが空です", alert.Message)
	}

	return nil
}

// IsASCII 文字列がASCII文字のみかをチェック
func IsASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// ContainsOnlyAlphanumeric 文字列が英数字のみかをチェック
func ContainsOnlyAlphanumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// IsValidEmail メールアドレスの形式をチェック
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
