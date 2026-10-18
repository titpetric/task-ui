# Workflow Run Step

Workflow Run Step.

| Name            | Type     | Key | Comment         |
|-----------------|----------|-----|-----------------|
| id              | char(26) | PRI | ID              |
| workflow_run_id | char(26) | MUL | Workflow Run ID |
| position        | bigint   | MUL | Position        |
| folder          | varchar  |     | Folder          |
| job             | varchar  |     | Job             |
| run_id          | char(26) | MUL | Run ID          |
| state           | varchar  |     | State           |
| created_at      | datetime |     | Created At      |
| updated_at      | datetime |     | Updated At      |
