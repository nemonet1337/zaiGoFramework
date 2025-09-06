-- 在庫管理システムの初期スキーマ
-- zaiGoFramework initial database schema

-- 商品テーブル
CREATE TABLE items (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(500) NOT NULL,
    sku VARCHAR(255) UNIQUE,
    description TEXT,
    category VARCHAR(255),
    unit_cost DECIMAL(12,4) DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ロケーションテーブル
CREATE TABLE locations (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(500) NOT NULL,
    type VARCHAR(100),
    address TEXT,
    capacity BIGINT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 在庫テーブル
CREATE TABLE stocks (
    item_id VARCHAR(255) NOT NULL,
    location_id VARCHAR(255) NOT NULL,
    quantity BIGINT NOT NULL DEFAULT 0,
    reserved BIGINT NOT NULL DEFAULT 0,
    available BIGINT NOT NULL DEFAULT 0,
    version BIGINT NOT NULL DEFAULT 1,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by VARCHAR(255) NOT NULL DEFAULT 'system',
    PRIMARY KEY (item_id, location_id),
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE,
    FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE CASCADE
);

-- トランザクションテーブル
CREATE TABLE transactions (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    item_id VARCHAR(255) NOT NULL,
    from_location VARCHAR(255),
    to_location VARCHAR(255),
    quantity BIGINT NOT NULL,
    unit_cost DECIMAL(12,4),
    reference VARCHAR(500),
    lot_number VARCHAR(255),
    expiry_date TIMESTAMP,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL DEFAULT 'system',
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE,
    FOREIGN KEY (from_location) REFERENCES locations(id) ON DELETE SET NULL,
    FOREIGN KEY (to_location) REFERENCES locations(id) ON DELETE SET NULL
);

-- ロットテーブル
CREATE TABLE lots (
    id VARCHAR(255) PRIMARY KEY,
    number VARCHAR(255) NOT NULL,
    item_id VARCHAR(255) NOT NULL,
    quantity BIGINT NOT NULL,
    unit_cost DECIMAL(12,4) NOT NULL,
    expiry_date TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE
);

-- アラートテーブル
CREATE TABLE stock_alerts (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    item_id VARCHAR(255) NOT NULL,
    location_id VARCHAR(255) NOT NULL,
    current_qty BIGINT NOT NULL,
    threshold BIGINT NOT NULL,
    message TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP,
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE,
    FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE CASCADE
);

-- パフォーマンス向上のためのインデックス
CREATE INDEX idx_items_sku ON items(sku);
CREATE INDEX idx_items_category ON items(category);
CREATE INDEX idx_stocks_item_id ON stocks(item_id);
CREATE INDEX idx_stocks_location_id ON stocks(location_id);
CREATE INDEX idx_stocks_quantity ON stocks(quantity);
CREATE INDEX idx_transactions_item_id ON transactions(item_id);
CREATE INDEX idx_transactions_created_at ON transactions(created_at DESC);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_lots_item_id ON lots(item_id);
CREATE INDEX idx_lots_expiry_date ON lots(expiry_date);
CREATE INDEX idx_stock_alerts_location_id ON stock_alerts(location_id);
CREATE INDEX idx_stock_alerts_is_active ON stock_alerts(is_active);

-- 初期データ挿入

-- デフォルトロケーション
INSERT INTO locations (id, name, type, address, capacity, is_active) 
VALUES ('DEFAULT', 'デフォルト倉庫', 'warehouse', '東京都', 10000, true)
ON CONFLICT (id) DO NOTHING;

-- サンプル商品
INSERT INTO items (id, name, sku, description, category, unit_cost) 
VALUES 
    ('ITEM001', 'サンプル商品A', 'SKU-001', 'テスト用サンプル商品A', 'sample', 100.00),
    ('ITEM002', 'サンプル商品B', 'SKU-002', 'テスト用サンプル商品B', 'sample', 200.00)
ON CONFLICT (id) DO NOTHING;

-- トリガー：在庫テーブルのavailable列を自動計算
CREATE OR REPLACE FUNCTION calculate_available_stock()
RETURNS TRIGGER AS $$
BEGIN
    NEW.available = NEW.quantity - NEW.reserved;
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_calculate_available_stock
    BEFORE INSERT OR UPDATE ON stocks
    FOR EACH ROW
    EXECUTE FUNCTION calculate_available_stock();

-- トリガー：商品・ロケーションの更新日時自動更新
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_items_updated_at
    BEFORE UPDATE ON items
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_update_locations_updated_at
    BEFORE UPDATE ON locations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();
