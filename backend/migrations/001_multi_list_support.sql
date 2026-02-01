-- Phase 1: Multi-list support migration
-- Run this in Supabase SQL Editor

-- 1. Create the lists table
CREATE TABLE IF NOT EXISTS lists (
    id VARCHAR(32) PRIMARY KEY DEFAULT substr(gen_random_uuid()::text, 1, 32),
    name VARCHAR(15) NOT NULL,
    emoji TEXT,
    hex_color VARCHAR(6) DEFAULT '42b883',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. Create a default list for existing items
INSERT INTO lists (id, name, emoji, hex_color)
VALUES ('default', 'Shopping', '🛒', '42b883');

-- 3. Add new columns to items table
ALTER TABLE items
ADD COLUMN IF NOT EXISTS list_id VARCHAR(32) REFERENCES lists(id) ON DELETE CASCADE,
ADD COLUMN IF NOT EXISTS sort_order FLOAT DEFAULT 0,
ADD COLUMN IF NOT EXISTS is_separator BOOLEAN DEFAULT FALSE;

-- 4. Migrate existing items to the default list
UPDATE items SET list_id = 'default' WHERE list_id IS NULL;

-- 5. Make list_id NOT NULL after migration
ALTER TABLE items ALTER COLUMN list_id SET NOT NULL;

-- 6. Create index for faster queries by list
CREATE INDEX IF NOT EXISTS idx_items_list_id ON items(list_id);
CREATE INDEX IF NOT EXISTS idx_items_sort_order ON items(list_id, sort_order);
