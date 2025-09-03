package inventory

import (
	"errors"
	"fmt"
)

// Common inventory errors
// 共通の在庫エラー定義

var (
	// ErrItemNotFound is returned when an item doesn't exist
	// 商品が存在しない場合のエラー
	ErrItemNotFound = errors.New("商品が見つかりません")

	// ErrLocationNotFound is returned when a location doesn't exist
	// ロケーションが存在しない場合のエラー
	ErrLocationNotFound = errors.New("ロケーションが見つかりません")

	// ErrInsufficientStock is returned when there's not enough stock
	// 在庫不足の場合のエラー
	ErrInsufficientStock = errors.New("在庫が不足しています")

	// ErrNegativeQuantity is returned when a negative quantity is provided
	// 負の数量が指定された場合のエラー
	ErrNegativeQuantity = errors.New("数量は正の値である必要があります")

	// ErrStockNotFound is returned when stock record doesn't exist
	// 在庫記録が存在しない場合のエラー
	ErrStockNotFound = errors.New("在庫記録が見つかりません")

	// ErrVersionMismatch is returned when optimistic locking fails
	// 楽観的ロック失敗時のエラー
	ErrVersionMismatch = errors.New("バージョンが一致しません。他のユーザーによって更新されています")

	// ErrDuplicateItem is returned when trying to create an item that already exists
	// 既に存在する商品を作成しようとした場合のエラー
	ErrDuplicateItem = errors.New("商品は既に存在します")

	// ErrDuplicateLocation is returned when trying to create a location that already exists
	// 既に存在するロケーションを作成しようとした場合のエラー
	ErrDuplicateLocation = errors.New("ロケーションは既に存在します")

	// ErrInvalidReference is returned when reference is invalid
	// 参照番号が無効な場合のエラー
	ErrInvalidReference = errors.New("無効な参照番号です")

	// ErrTransactionFailed is returned when a transaction fails
	// トランザクション失敗時のエラー
	ErrTransactionFailed = errors.New("トランザクションが失敗しました")

	// ErrLotNotFound is returned when a lot doesn't exist
	// ロットが存在しない場合のエラー
	ErrLotNotFound = errors.New("ロットが見つかりません")

	// ErrExpiredLot is returned when trying to use an expired lot
	// 期限切れロットを使用しようとした場合のエラー
	ErrExpiredLot = errors.New("ロットの有効期限が切れています")

	// ErrReservationNotFound is returned when reservation doesn't exist
	// 予約が存在しない場合のエラー
	ErrReservationNotFound = errors.New("予約が見つかりません")

	// ErrInsufficientReservation is returned when trying to release more than reserved
	// 予約量を超えて解除しようとした場合のエラー
	ErrInsufficientReservation = errors.New("予約量が不足しています")
)

// ValidationError represents a validation error with details
// 詳細付きバリデーションエラーを表現
type ValidationError struct {
	Field   string `json:"field"`   // エラーフィールド
	Message string `json:"message"` // エラーメッセージ
	Value   string `json:"value"`   // 無効な値
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("バリデーションエラー [%s]: %s (値: %s)", e.Field, e.Message, e.Value)
}

// BusinessRuleError represents a business rule violation
// ビジネスルール違反を表現
type BusinessRuleError struct {
	Rule    string `json:"rule"`    // ルール名
	Message string `json:"message"` // エラーメッセージ
	Context string `json:"context"` // コンテキスト情報
}

func (e BusinessRuleError) Error() string {
	return fmt.Sprintf("ビジネスルール違反 [%s]: %s (コンテキスト: %s)", e.Rule, e.Message, e.Context)
}

// ConcurrencyError represents a concurrency-related error
// 同時実行関連のエラーを表現
type ConcurrencyError struct {
	Operation string `json:"operation"` // 操作名
	Resource  string `json:"resource"`  // リソース
	Message   string `json:"message"`   // エラーメッセージ
}

func (e ConcurrencyError) Error() string {
	return fmt.Sprintf("同時実行エラー [%s:%s]: %s", e.Operation, e.Resource, e.Message)
}

// StorageError represents a storage layer error
// ストレージ層のエラーを表現
type StorageError struct {
	Operation string `json:"operation"` // 操作名
	Message   string `json:"message"`   // エラーメッセージ
	Cause     error  `json:"cause"`     // 原因エラー
}

func (e StorageError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("ストレージエラー [%s]: %s (原因: %v)", e.Operation, e.Message, e.Cause)
	}
	return fmt.Sprintf("ストレージエラー [%s]: %s", e.Operation, e.Message)
}

func (e StorageError) Unwrap() error {
	return e.Cause
}

// NewValidationError creates a new validation error
// 新しいバリデーションエラーを作成
func NewValidationError(field, message, value string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

// NewBusinessRuleError creates a new business rule error
// 新しいビジネスルールエラーを作成
func NewBusinessRuleError(rule, message, context string) *BusinessRuleError {
	return &BusinessRuleError{
		Rule:    rule,
		Message: message,
		Context: context,
	}
}

// NewConcurrencyError creates a new concurrency error
// 新しい同時実行エラーを作成
func NewConcurrencyError(operation, resource, message string) *ConcurrencyError {
	return &ConcurrencyError{
		Operation: operation,
		Resource:  resource,
		Message:   message,
	}
}

// NewStorageError creates a new storage error
// 新しいストレージエラーを作成
func NewStorageError(operation, message string, cause error) *StorageError {
	return &StorageError{
		Operation: operation,
		Message:   message,
		Cause:     cause,
	}
}
