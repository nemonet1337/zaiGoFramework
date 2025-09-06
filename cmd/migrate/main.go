package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"

	_ "github.com/lib/pq"
	"github.com/nemonet1337/zaiGoFramework/internal/config"
)

func main() {
	log.Println("zaiGoFramework マイグレーション実行ツール")
	
	// 設定読み込み
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("設定読み込みに失敗しました:", err)
	}

	// データベース接続
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.DBName)

	log.Printf("データベースに接続中: %s:%d/%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("データベース接続に失敗しました:", err)
	}
	defer db.Close()

	// 接続テスト
	if err := db.Ping(); err != nil {
		log.Fatal("データベースpingに失敗しました:", err)
	}

	log.Println("データベース接続が確立されました")

	// マイグレーションディレクトリの確認
	migrationDir := "migrations"
	if len(os.Args) > 1 {
		migrationDir = os.Args[1]
	}

	if _, err := os.Stat(migrationDir); os.IsNotExist(err) {
		log.Fatalf("マイグレーションディレクトリが見つかりません: %s", migrationDir)
	}

	// マイグレーション履歴テーブルの作成
	if err := createMigrationTable(db); err != nil {
		log.Fatal("マイグレーション履歴テーブル作成に失敗しました:", err)
	}

	// マイグレーション実行
	if err := runMigrations(db, migrationDir); err != nil {
		log.Fatal("マイグレーション実行に失敗しました:", err)
	}

	log.Println("すべてのマイグレーションが完了しました")
}

// createMigrationTable マイグレーション履歴テーブルを作成
func createMigrationTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) NOT NULL UNIQUE,
			executed_at TIMESTAMP NOT NULL DEFAULT NOW(),
			checksum VARCHAR(64) NOT NULL
		)`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("マイグレーション履歴テーブル作成エラー: %w", err)
	}

	log.Println("マイグレーション履歴テーブルを確認/作成しました")
	return nil
}

// runMigrations マイグレーションを実行
func runMigrations(db *sql.DB, migrationDir string) error {
	// .sqlファイルを取得
	files, err := filepath.Glob(filepath.Join(migrationDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("マイグレーションファイル検索エラー: %w", err)
	}

	if len(files) == 0 {
		log.Printf("マイグレーションファイルが見つかりません: %s", migrationDir)
		return nil
	}

	// ファイル名でソート
	sort.Strings(files)

	// 実行済みマイグレーションを取得
	executedMigrations, err := getExecutedMigrations(db)
	if err != nil {
		return fmt.Errorf("実行済みマイグレーション取得エラー: %w", err)
	}

	// 各マイグレーションファイルを処理
	for _, file := range files {
		filename := filepath.Base(file)

		// 既に実行済みかチェック
		if _, executed := executedMigrations[filename]; executed {
			log.Printf("スキップ (実行済み): %s", filename)
			continue
		}

		log.Printf("実行中: %s", filename)

		// ファイル内容を読み込み
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("ファイル読み込みエラー %s: %w", filename, err)
		}

		// チェックサムを計算
		checksum := calculateChecksum(content)

		// トランザクション開始
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("トランザクション開始エラー %s: %w", filename, err)
		}

		// マイグレーション実行
		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("マイグレーション実行エラー %s: %w", filename, err)
		}

		// マイグレーション履歴に記録
		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (filename, checksum) VALUES ($1, $2)",
			filename, checksum,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("マイグレーション履歴記録エラー %s: %w", filename, err)
		}

		// コミット
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("トランザクションコミットエラー %s: %w", filename, err)
		}

		log.Printf("完了: %s", filename)
	}

	return nil
}

// getExecutedMigrations 実行済みマイグレーションを取得
func getExecutedMigrations(db *sql.DB) (map[string]bool, error) {
	executed := make(map[string]bool)

	rows, err := db.Query("SELECT filename FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		executed[filename] = true
	}

	return executed, rows.Err()
}

// calculateChecksum ファイル内容のチェックサムを計算
func calculateChecksum(content []byte) string {
	// 簡易的なチェックサム（実際の実装ではSHA256などを使用）
	sum := 0
	for _, b := range content {
		sum += int(b)
	}
	return fmt.Sprintf("%x", sum)
}
