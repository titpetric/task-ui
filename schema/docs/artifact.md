# Artifact

Artifact.

| Name         | Type     | Key | Comment      |
|--------------|----------|-----|--------------|
| id           | char(26) | PRI | ID           |
| run_id       | char(26) | MUL | Run ID       |
| folder       | varchar  | MUL | Folder       |
| job          | varchar  | MUL | Job          |
| name         | varchar  |     | Name         |
| path         | varchar  |     | Path         |
| byte_size    | bigint   |     | Byte Size    |
| content_type | varchar  |     | Content Type |
| created_at   | datetime |     | Created At   |
