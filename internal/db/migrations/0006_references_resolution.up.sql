-- references: JSON array of {kind, target, label} objects
-- e.g. [{"kind":"file","target":"src/api/handler.go:142","label":"handler"},
--        {"kind":"pr","target":"https://github.com/org/repo/pull/12"}]
ALTER TABLE tickets ADD COLUMN "references" TEXT NOT NULL DEFAULT '[]';

-- resolution: JSON object set when an agent closes a ticket
-- e.g. {"commit_sha":"abc123","pr_url":"https://...","notes":"Fixed by ...","resolved_at":"2026-01-01T00:00:00Z"}
ALTER TABLE tickets ADD COLUMN resolution TEXT;
