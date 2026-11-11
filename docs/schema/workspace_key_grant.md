# Workspace Key Grant

Workspace Key Grant.

| Name         | Type     | Key | Comment      |
|--------------|----------|-----|--------------|
| id           | char(26) | PRI | ID           |
| workspace_id | char(26) | MUL | Workspace ID |
| key_name     | varchar  | MUL | Key Name     |
| role         | varchar  | MUL | Role         |
| user_name    | varchar  | MUL | User Name    |
| created_at   | datetime |     | Created At   |
