# Note

Note.

| Name                  | Type     | Key | Comment               |
|-----------------------|----------|-----|-----------------------|
| id                    | char(26) | PRI | ID                    |
| channel               | varchar  | MUL | Channel               |
| parent_id             | char(26) | MUL | Parent ID             |
| user_id               | char(26) | MUL | User ID               |
| user_name             | varchar  |     | User Name             |
| body                  | varchar  |     | Body                  |
| mentions              | varchar  |     | Mentions              |
| state                 | varchar  | MUL | State                 |
| run_id                | char(26) | MUL | Run ID                |
| created_at            | datetime |     | Created At            |
| updated_at            | datetime |     | Updated At            |
| about_run_id          | char(26) | MUL | About Run ID          |
| about_workflow_run_id | char(26) | MUL | About Workflow Run ID |
| area                  | varchar  |     | Area                  |
| edited_at             | datetime |     | Edited At             |
| workspace_id          | char(26) | MUL | Workspace ID          |
| dir                   | varchar  | MUL | Dir                   |
