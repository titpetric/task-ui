# Refinement Comment

Refinement Comment.

| Name           | Type     | Key | Comment        |
|----------------|----------|-----|----------------|
| id             | char(26) | PRI | ID             |
| thread_id      | char(26) | MUL | Thread ID      |
| user           | varchar  |     | User           |
| body           | varchar  |     | Body           |
| note_id        | char(26) | MUL | Note ID        |
| kind           | varchar  |     | Kind           |
| collapsed_into | char(26) |     | Collapsed Into |
| created_at     | datetime | MUL | Created At     |
