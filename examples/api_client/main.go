package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// REST API クライアントの使用例
func main() {
	fmt.Println("=== zaiGoFramework REST API クライアント例 ===")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	baseURL := "http://localhost:8080/api/v1"

	// 1. ヘルスチェック
	fmt.Println("\n1. ヘルスチェック")
	resp, err := client.Get("http://localhost:8080/health")
	if err != nil {
		log.Printf("ヘルスチェックエラー: %v", err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			fmt.Println("✓ APIサーバーは正常に動作しています")
		} else {
			fmt.Printf("✗ APIサーバーエラー: %d\n", resp.StatusCode)
		}
	}

	// 2. 在庫追加
	fmt.Println("\n2. 在庫追加")
	addStockReq := map[string]interface{}{
		"item_id":     "ITEM001",
		"location_id": "DEFAULT",
		"quantity":    100,
		"reference":   "API-TEST-001",
	}
	
	err = makeRequest(client, "POST", baseURL+"/inventory/add", addStockReq)
	if err != nil {
		log.Printf("在庫追加エラー: %v", err)
	} else {
		fmt.Println("✓ 在庫追加完了")
	}

	// 3. 在庫確認
	fmt.Println("\n3. 在庫確認")
	resp, err = client.Get(baseURL + "/inventory/ITEM001/DEFAULT")
	if err != nil {
		log.Printf("在庫確認エラー: %v", err)
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		
		var apiResp APIResponse
		if err := json.Unmarshal(body, &apiResp); err == nil && apiResp.Success {
			fmt.Printf("✓ 現在在庫: %v\n", apiResp.Data)
		} else {
			fmt.Printf("✗ 在庫確認失敗: %s\n", string(body))
		}
	}

	// 4. 在庫削除
	fmt.Println("\n4. 在庫削除")
	removeStockReq := map[string]interface{}{
		"item_id":     "ITEM001",
		"location_id": "DEFAULT",
		"quantity":    20,
		"reference":   "API-SHIP-001",
	}
	
	err = makeRequest(client, "POST", baseURL+"/inventory/remove", removeStockReq)
	if err != nil {
		log.Printf("在庫削除エラー: %v", err)
	} else {
		fmt.Println("✓ 在庫削除完了")
	}

	// 5. バッチ操作
	fmt.Println("\n5. バッチ操作")
	batchReq := []map[string]interface{}{
		{
			"type":        "add",
			"item_id":     "ITEM002",
			"location_id": "DEFAULT",
			"quantity":    50,
			"reference":   "BATCH-001",
		},
		{
			"type":        "adjust",
			"item_id":     "ITEM001",
			"location_id": "DEFAULT",
			"quantity":    100,
			"reference":   "BATCH-002",
		},
	}
	
	err = makeRequest(client, "POST", baseURL+"/inventory/batch", batchReq)
	if err != nil {
		log.Printf("バッチ操作エラー: %v", err)
	} else {
		fmt.Println("✓ バッチ操作完了")
	}

	// 6. 履歴確認
	fmt.Println("\n6. トランザクション履歴")
	resp, err = client.Get(baseURL + "/inventory/ITEM001/history?limit=5")
	if err != nil {
		log.Printf("履歴取得エラー: %v", err)
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		
		var apiResp APIResponse
		if err := json.Unmarshal(body, &apiResp); err == nil && apiResp.Success {
			fmt.Printf("✓ 履歴取得完了: %d件\n", len(apiResp.Data.([]interface{})))
		} else {
			fmt.Printf("✗ 履歴取得失敗\n")
		}
	}

	fmt.Println("\n=== API クライアント例完了 ===")
}

// makeRequest makes HTTP request with JSON payload
// JSON ペイロードでHTTPリクエストを作成
func makeRequest(client *http.Client, method, url string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// APIResponse represents standard API response
// 標準的なAPIレスポンスを表現
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
