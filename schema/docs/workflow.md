# Workflow

Workflow.

| Name           | Type     | Key | Comment        |
|----------------|----------|-----|----------------|
| id             | char(26) | PRI | ID             |
| name           | varchar  | MUL | Name           |
| user_id        | char(26) | MUL | User ID        |
| user_name      | varchar  |     | User Name      |
| created_at     | datetime |     | Created At     |
| updated_at     | datetime |     | Updated At     |
| prompt         | varchar  |     | Prompt         |
| prompt_folder  | varchar  |     | Prompt Folder  |
| sidebar_hidden | boolean  |     | Sidebar Hidden |
