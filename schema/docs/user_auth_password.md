# User Auth Password

User Auth Password.

| Name       | Type     | Key | Comment    |
|------------|----------|-----|------------|
| id         | char(26) | PRI | ID         |
| user_id    | char(26) | MUL | User ID    |
| hash       | varchar  |     | Hash       |
| kind       | varchar  | MUL | Kind       |
| used_at    | datetime |     | Used At    |
| created_at | datetime |     | Created At |
| updated_at | datetime |     | Updated At |
