# who-field

Input scope: `data_whofield`, from `Task\WhoField::data()`.

| Field              | Type   | What it is                                           |
|--------------------|--------|------------------------------------------------------|
| `compact`          | bool   | One wrapped line and no summary, for a table cell    |
| `empty`            | string | What nothing checked means; empty means no run gets it |
| `name`             | string | The roles' boxes' name                               |
| `userName`         | string | The users' boxes' name                               |
| `type`             | string | The boxes' `data-type`: `set` or `joined`            |
| `roles`            | list   | A box each: `name`, `checked`, `count`               |
| `users`            | list   | A box each: `name`, `role`, `checked`                |
| `extra`            | list   | Checked names no user holds, kept so a save does not drop them |
