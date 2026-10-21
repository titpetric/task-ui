# User Session

User Session.

| Name       | Type     | Key | Comment    |
|------------|----------|-----|------------|
| id         | char(26) | PRI | ID         |
| user_id    | char(26) | MUL | User ID    |
| hash       | varchar  | MUL | Hash       |
| expires_at | datetime |     | Expires At |
| used_at    | datetime |     | Used At    |
| created_at | datetime |     | Created At |
| updated_at | datetime |     | Updated At |
