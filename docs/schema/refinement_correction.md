# Refinement Correction

Refinement Correction.

| Name        | Type     | Key | Comment     |
|-------------|----------|-----|-------------|
| id          | char(26) | PRI | ID          |
| thread_id   | char(26) | MUL | Thread ID   |
| comment_id  | char(26) | MUL | Comment ID  |
| corrects_id | char(26) | MUL | Corrects ID |
| created_at  | datetime | MUL | Created At  |
