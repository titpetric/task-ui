# Workspace

Workspace.

| Name            | Type     | Key | Comment         |
|-----------------|----------|-----|-----------------|
| id              | char(26) | PRI | ID              |
| folder          | varchar  | MUL | Folder          |
| title           | varchar  |     | Title           |
| description     | varchar  |     | Description     |
| created_at      | datetime |     | Created At      |
| updated_at      | datetime |     | Updated At      |
| home_project_id | char(26) | MUL | Home Project ID |
| shallow         | boolean  |     | Shallow         |
| sort            | bigint   |     | Sort            |
