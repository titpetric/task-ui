# Refinement Artifact

Refinement Artifact.

| Name       | Type     | Key | Comment    |
|------------|----------|-----|------------|
| id         | char(26) | PRI | ID         |
| thread_id  | char(26) | MUL | Thread ID  |
| comment_id | char(26) | MUL | Comment ID |
| path       | varchar  | MUL | Path       |
| kind       | varchar  |     | Kind       |
| created_at | datetime |     | Created At |
