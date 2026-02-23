-- Rename the comments table to notes.
-- SQLite does not support ALTER TABLE RENAME TABLE directly in older versions,
-- but it does support it since 3.25.0 (2018). We target modern SQLite.
ALTER TABLE comments RENAME TO notes;
