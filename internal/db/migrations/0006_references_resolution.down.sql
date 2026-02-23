-- SQLite does not support DROP COLUMN on older versions; recreate table to remove columns.
-- This migration is intentionally left as a no-op for safety.
-- To downgrade, restore from a backup taken before migration 0006.
SELECT 1;
