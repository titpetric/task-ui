# User Auth Token

User Auth Token.

| Name       | Type     | Key | Comment    |
|------------|----------|-----|------------|
| id         | char(26) | PRI | ID         |
| user_id    | char(26) | MUL | User ID    |
| name       | varchar  | MUL | Name       |
| hash       | varchar  | MUL | Hash       |
| used_at    | datetime |     | Used At    |
| created_at | datetime |     | Created At |
| updated_at | datetime |     | Updated At |
| expires_at | datetime |     | Expires At |
