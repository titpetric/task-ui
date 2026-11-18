# Meeting

Meeting.

| Name              | Type     | Key | Comment           |
|-------------------|----------|-----|-------------------|
| id                | char(26) | PRI | ID                |
| title             | varchar  |     | Title             |
| user_id           | char(26) | MUL | User ID           |
| user_name         | varchar  |     | User Name         |
| state             | varchar  | MUL | State             |
| started_at        | datetime | MUL | Started At        |
| ended_at          | datetime |     | Ended At          |
| limit_seconds     | bigint   |     | Limit Seconds     |
| empty_since       | datetime |     | Empty Since       |
| summary           | varchar  |     | Summary           |
| summary_state     | varchar  |     | Summary State     |
| summary_message   | varchar  |     | Summary Message   |
| summary_user_name | varchar  |     | Summary User Name |
| created_at        | datetime |     | Created At        |
| updated_at        | datetime |     | Updated At        |
