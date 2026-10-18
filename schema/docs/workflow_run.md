# Workflow Run

Workflow Run.

| Name        | Type     | Key | Comment     |
|-------------|----------|-----|-------------|
| id          | char(26) | PRI | ID          |
| workflow_id | char(26) | MUL | Workflow ID |
| name        | varchar  |     | Name        |
| user_id     | char(26) | MUL | User ID     |
| user_name   | varchar  |     | User Name   |
| state       | varchar  | MUL | State       |
| message     | varchar  |     | Message     |
| started_at  | datetime | MUL | Started At  |
| finished_at | datetime |     | Finished At |
| created_at  | datetime |     | Created At  |
| updated_at  | datetime |     | Updated At  |
