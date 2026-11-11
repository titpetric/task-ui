# Secret

Secret.

| Name       | Type     | Key | Comment    |
|------------|----------|-----|------------|
| id         | char(26) | PRI | ID         |
| scope      | varchar  | MUL | Scope      |
| name       | varchar  | MUL | Name       |
| content    | varchar  |     | Content    |
| created_at | datetime |     | Created At |
| updated_at | datetime |     | Updated At |
| is_secret  | bigint   |     | Is Secret  |
