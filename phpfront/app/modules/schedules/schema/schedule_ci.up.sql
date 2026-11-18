-- source names what registered a schedule: ci:daily for the row the index reconciled, empty for a person's.
ALTER TABLE schedule ADD COLUMN source TEXT NOT NULL DEFAULT '';
