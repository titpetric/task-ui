# Search Document

Search Document.

| Name        | Type     | Key | Comment     |
|-------------|----------|-----|-------------|
| id          | char(26) | PRI | ID          |
| kind        | varchar  | MUL | Kind        |
| path        | varchar  | MUL | Path        |
| file        | varchar  |     | File        |
| folder_id   | varchar  | MUL | Folder ID   |
| workspace   | varchar  |     | Workspace   |
| project     | varchar  |     | Project     |
| target      | varchar  | MUL | Target      |
| name        | varchar  | MUL | Name        |
| dir         | varchar  |     | Dir         |
| title       | varchar  |     | Title       |
| hash        | varchar  |     | Hash        |
| size        | bigint   |     | Size        |
| modified_at | datetime |     | Modified At |
| indexed_at  | datetime |     | Indexed At  |
