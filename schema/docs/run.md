# Run

Run.

| Name        | Type     | Key | Comment     |
|-------------|----------|-----|-------------|
| id          | char(26) | PRI | ID          |
| folder      | varchar  | MUL | Folder      |
| job         | varchar  | MUL | Job         |
| user_id     | char(26) | MUL | User ID     |
| user_name   | varchar  |     | User Name   |
| state       | varchar  | MUL | State       |
| exit_code   | bigint   |     | Exit Code   |
| forced      | bigint   |     | Forced      |
| scheduled   | bigint   |     | Scheduled   |
| term_cols   | bigint   |     | Term Cols   |
| term_rows   | bigint   |     | Term Rows   |
| cast_path   | varchar  |     | Cast Path   |
| reason      | varchar  |     | Reason      |
| started_at  | datetime | MUL | Started At  |
| finished_at | datetime |     | Finished At |
| created_at  | datetime |     | Created At  |
| updated_at  | datetime |     | Updated At  |
| dir         | varchar  |     | Dir         |
| args        | varchar  |     | Args        |
| target      | varchar  |     | Target      |
| batch_id    | char(26) | MUL | Batch ID    |
| message     | varchar  |     | Message     |
