# Secret Read

Secret Read.

| Name       | Type     | Key | Comment    |
|------------|----------|-----|------------|
| id         | char(26) | PRI | ID         |
| secret_id  | char(26) | MUL | Secret ID  |
| name       | varchar  |     | Name       |
| run_id     | char(26) | MUL | Run ID     |
| folder     | varchar  |     | Folder     |
| dir        | varchar  |     | Dir        |
| job        | varchar  |     | Job        |
| user_name  | varchar  |     | User Name  |
| created_at | datetime | MUL | Created At |
