# Permission Policy

Permission Policy.

| Name         | Type     | Key | Comment      |
|--------------|----------|-----|--------------|
| id           | char(26) | PRI | ID           |
| subject_kind | varchar  | MUL | Subject Kind |
| subject      | varchar  | MUL | Subject      |
| path         | varchar  | MUL | Path         |
| effect       | varchar  |     | Effect       |
| created_at   | datetime |     | Created At   |
