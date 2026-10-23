-- key_info_backup: The backup a rotation keeps of a key, at backup/<name>.key and .pub in the key folder, used beside the new key until an admin deletes it; one backup a key
ALTER TABLE key_info ADD COLUMN backup TEXT NOT NULL DEFAULT '';
ALTER TABLE key_info ADD COLUMN backup_fingerprint TEXT NOT NULL DEFAULT '';
ALTER TABLE key_info ADD COLUMN rotated_at DATETIME;
