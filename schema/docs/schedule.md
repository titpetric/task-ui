# Schedule

Schedule.

| Name        | Type     | Key | Comment     |
|-------------|----------|-----|-------------|
| id          | char(26) | PRI | ID          |
| folder      | varchar  | MUL | Folder      |
| job         | varchar  | MUL | Job         |
| user_id     | char(26) | MUL | User ID     |
| user_name   | varchar  |     | User Name   |
| expression  | varchar  |     | Expression  |
| enabled     | bigint   | MUL | Enabled     |
| last_run_at | datetime |     | Last Run At |
| next_run_at | datetime | MUL | Next Run At |
| created_at  | datetime |     | Created At  |
| updated_at  | datetime |     | Updated At  |
| target      | varchar  |     | Target      |
| args        | varchar  |     | Args        |
