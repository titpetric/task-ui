# Note Event

Note Event.

| Name        | Type     | Key | Comment     |
|-------------|----------|-----|-------------|
| id          | char(26) | PRI | ID          |
| note_id     | char(26) | MUL | Note ID     |
| run_id      | char(26) | MUL | Run ID      |
| question_id | char(26) | MUL | Question ID |
| seq         | bigint   | MUL | Seq         |
| kind        | varchar  |     | Kind        |
| tool_id     | varchar  | MUL | Tool ID     |
| body        | varchar  |     | Body        |
| data        | varchar  |     | Data        |
| created_at  | datetime |     | Created At  |
