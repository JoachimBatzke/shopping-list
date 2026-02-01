-- Item History table for recommendations
-- Run this SQL in your Supabase SQL editor to enable recommendations

CREATE TABLE IF NOT EXISTS item_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id VARCHAR(32) NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    item_name VARCHAR(100) NOT NULL,
    added_count INTEGER DEFAULT 1,
    last_added_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    avg_days_between FLOAT DEFAULT 7,
    dismissed BOOLEAN DEFAULT false,
    UNIQUE(list_id, item_name)
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_item_history_list_id ON item_history(list_id);
CREATE INDEX IF NOT EXISTS idx_item_history_lookup ON item_history(list_id, dismissed, added_count);
