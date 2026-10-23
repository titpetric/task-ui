# Workflow Step

Workflow Step.

| Name         | Type     | Key | Comment      |
|--------------|----------|-----|--------------|
| id           | char(26) | PRI | ID           |
| workflow_id  | char(26) | MUL | Workflow ID  |
| position     | bigint   | MUL | Position     |
| folder       | varchar  |     | Folder       |
| job          | varchar  |     | Job          |
| disabled     | boolean  |     | Disabled     |
| prompt       | varchar  |     | Prompt       |
| dir          | varchar  |     | Dir          |
| workflow_ref | char(26) |     | Workflow Ref |
| target       | varchar  |     | Target       |
| args         | varchar  |     | Args         |
| kind         | varchar  |     | Kind         |
