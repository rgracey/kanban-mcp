-- SQLite does not support DROP COLUMN on older versions; recreate table without assignee
CREATE TABLE tickets_backup AS SELECT id, board_id, epic_id, title, description, status, priority, created_at, updated_at FROM tickets;
DROP TABLE tickets;
ALTER TABLE tickets_backup RENAME TO tickets;
