# Meeting Participant

Meeting Participant.

| Name         | Type     | Key | Comment      |
|--------------|----------|-----|--------------|
| meeting_id   | char(26) | PRI | Meeting ID   |
| user_id      | char(26) | PRI | User ID      |
| user_name    | varchar  |     | User Name    |
| joined_at    | datetime |     | Joined At    |
| last_seen_at | datetime |     | Last Seen At |
| left_at      | datetime | MUL | Left At      |
| joins        | bigint   |     | Joins        |
