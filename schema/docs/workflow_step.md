# Workflow Step

Workflow Step.

| Name         | Type     | Key | Comment      |
|--------------|----------|-----|--------------|
| id           | char(26) | PRI | ID           |
| workflow_id  | char(26) | MUL | Workflow ID  |
| position     | bigint   | MUL | Position     |
| folder       | varchar  |     | Folder       |
| job          | varchar  |     | Job          |
| workflow_ref | char(26) |     | Workflow Ref |
