# Meeting Line

Meeting Line.

| Name       | Type     | Key | Comment    |
|------------|----------|-----|------------|
| id         | char(26) | PRI | ID         |
| meeting_id | char(26) | MUL | Meeting ID |
| user_id    | char(26) | MUL | User ID    |
| user_name  | varchar  |     | User Name  |
| kind       | varchar  |     | Kind       |
| body       | varchar  |     | Body       |
| start_ms   | bigint   |     | Start Ms   |
| end_ms     | bigint   |     | End Ms     |
| created_at | datetime | MUL | Created At |
