# Sign in

Generated from the module by `atkins docs`. DO NOT EDIT.

![The tables this module reads](schema.svg)

## Pages

| Method | Path                 | Answered by                              |
|--------|----------------------|------------------------------------------|
| `GET`  | `/login`             | [`mod_login.php`](mod_login.php)         |
| `GET`  | `/keys`              | [`mod_my_keys.php`](mod_my_keys.php)     |
| `GET`  | `/users/{name}`      | [`mod_profile.php`](mod_profile.php)     |
| `GET`  | `/users/{name}/edit` | [`mod_user_edit.php`](mod_user_edit.php) |

## The data model

### `note`

Note

| Column                  | Type        | Key     |
|-------------------------|-------------|---------|
| `id`                    | `text`      | primary |
| `channel`               | `TEXT`      | indexed |
| `parent_id`             | `text`      | indexed |
| `user_id`               | `text`      | indexed |
| `user_name`             | `text`      |         |
| `body`                  | `text`      |         |
| `mentions`              | `text`      |         |
| `state`                 | `TEXT`      | indexed |
| `run_id`                | `text`      | indexed |
| `created_at`            | `timestamp` |         |
| `updated_at`            | `timestamp` |         |
| `about_run_id`          | `text`      | indexed |
| `about_workflow_run_id` | `text`      | indexed |
| `area`                  | `text`      |         |
| `edited_at`             | `timestamp` |         |
| `deleted_at`            | `timestamp` | indexed |
| `audio_path`            | `text`      |         |
| `spoken`                | `text`      |         |
| `workspace_id`          | `text`      | indexed |
| `dir`                   | `text`      | indexed |

### `permission_policy`

Permission Policy

| Column         | Type        | Key     |
|----------------|-------------|---------|
| `id`           | `text`      | primary |
| `subject_kind` | `TEXT`      | indexed |
| `subject`      | `text`      | indexed |
| `path`         | `text`      | indexed |
| `effect`       | `TEXT`      |         |
| `created_at`   | `timestamp` |         |

### `run`

Run

| Column        | Type        | Key     |
|---------------|-------------|---------|
| `id`          | `text`      | primary |
| `folder`      | `text`      | indexed |
| `job`         | `text`      | indexed |
| `user_id`     | `text`      | indexed |
| `user_name`   | `text`      |         |
| `state`       | `TEXT`      | indexed |
| `exit_code`   | `integer`   |         |
| `forced`      | `integer`   |         |
| `scheduled`   | `integer`   |         |
| `term_cols`   | `integer`   |         |
| `term_rows`   | `integer`   |         |
| `cast_path`   | `text`      |         |
| `reason`      | `text`      |         |
| `started_at`  | `timestamp` | indexed |
| `finished_at` | `timestamp` |         |
| `created_at`  | `timestamp` |         |
| `updated_at`  | `timestamp` |         |
| `dir`         | `text`      |         |
| `args`        | `text`      |         |
| `target`      | `text`      |         |
| `batch_id`    | `text`      | indexed |
| `message`     | `text`      |         |
| `title`       | `text`      |         |
| `parent_id`   | `text`      | indexed |
| `parent_kind` | `text`      |         |

### `user`

User

| Column        | Type        | Key     |
|---------------|-------------|---------|
| `id`          | `text`      | primary |
| `name`        | `text`      | indexed |
| `role`        | `TEXT`      |         |
| `folders`     | `text`      |         |
| `created_at`  | `timestamp` |         |
| `updated_at`  | `timestamp` |         |
| `admin`       | `boolean`   |         |
| `mcp`         | `boolean`   |         |
| `agent_reply` | `boolean`   |         |

### `user_auth_token`

User Auth Token

| Column       | Type        | Key     |
|--------------|-------------|---------|
| `id`         | `text`      | primary |
| `user_id`    | `text`      | indexed |
| `name`       | `text`      | indexed |
| `hash`       | `text`      | indexed |
| `used_at`    | `timestamp` |         |
| `created_at` | `timestamp` |         |
| `updated_at` | `timestamp` |         |
| `expires_at` | `timestamp` |         |
| `kind`       | `text`      | indexed |

### `workflow_run`

Workflow Run

| Column        | Type        | Key     |
|---------------|-------------|---------|
| `id`          | `text`      | primary |
| `workflow_id` | `text`      | indexed |
| `name`        | `text`      |         |
| `user_id`     | `text`      | indexed |
| `user_name`   | `text`      |         |
| `state`       | `TEXT`      | indexed |
| `message`     | `text`      |         |
| `started_at`  | `timestamp` | indexed |
| `finished_at` | `timestamp` |         |
| `created_at`  | `timestamp` |         |
| `updated_at`  | `timestamp` |         |
| `input`       | `text`      |         |
| `answer`      | `text`      |         |

### `workflow_run_step`

Workflow Run Step

| Column             | Type        | Key     |
|--------------------|-------------|---------|
| `id`               | `text`      | primary |
| `workflow_run_id`  | `text`      | indexed |
| `position`         | `integer`   | indexed |
| `folder`           | `text`      |         |
| `job`              | `text`      |         |
| `run_id`           | `text`      | indexed |
| `state`            | `TEXT`      |         |
| `created_at`       | `timestamp` |         |
| `updated_at`       | `timestamp` |         |
| `prompt`           | `text`      |         |
| `answer`           | `text`      |         |
| `workflow_ref`     | `text`      |         |
| `workflow_run_ref` | `text`      |         |
| `kind`             | `TEXT`      |         |
| `markers`          | `text`      |         |

### `workspace`

Workspace

| Column            | Type        | Key     |
|-------------------|-------------|---------|
| `id`              | `text`      | primary |
| `folder`          | `text`      | indexed |
| `title`           | `text`      |         |
| `description`     | `text`      |         |
| `created_at`      | `timestamp` |         |
| `updated_at`      | `timestamp` |         |
| `home_project_id` | `text`      | indexed |
| `home_url`        | `text`      |         |
| `home_branch`     | `text`      |         |
| `hours`           | `text`      |         |
| `shallow`         | `boolean`   |         |
| `sidebar_limit`   | `integer`   |         |
| `sort`            | `integer`   |         |

## Operations

### `Tokens`

Tokens reads the user_auth_token table: the tokens that sign a client in without a session. A row's kind says what issued it: mcp for the ones a user's MCP access issues, which expire and which their page lists and deletes, and empty for the ones `task-ui token` writes, which never expire and which no page shows. A name is unique per user and kind, so a user holds as many MCP tokens as they care to name. Nothing here answers with the hash and nothing answers with the token. A token is shown once, by whoever made it, and a page that lists one says when it was made, when it runs out and when it was last used.

| Method                 | What it does                                                                                                                                                    |
|------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `ofKind($user, $kind)` | ofKind is a user's tokens of one kind, newest first: the order their page lists them in, because the one they just issued is the one they came for.             |
| `expired($expires)`    | expired reports whether an expiry has passed. A token from `task-ui token` has none and never expires, which reads as not expired.                              |
| `tone($expires)`       | tone is the colour of the expiry badge: bad once it has passed, signal within a week of it, and the plain badge before that and for a token that never expires. |

### `Users`

Users reads the user table: one user by ID or by name, every user by name, how many there are, how many are admins, and the agents a reply may be written as. The table is this module's, which draws the sign in page, a user's own page and the form that edits one. What a user may do - the role, the folders, the permissions - is Task\User.

| Method          | What it does                                   |
|-----------------|------------------------------------------------|
| `get($id)`      | get is one user by ID, or null.                |
| `byName($name)` | byName is one user by name, or null.           |
| `all()`         | all is every user, by name.                    |
| `count()`       | count is how many users there are.             |
| `admins()`      | admins is how many users carry the admin flag. |
| `agents()`      | agents is every user an agent may reply as.    |

### `Connect`

Connect is how a client is told to connect with a token it has just been given: the command for Claude Code, the same connection as a settings file, and the placeholders task-ui.js replaces in either of them. A placeholder is a constant here and not markup of a template, because MiniTPL compiles any {word} into an echo: a template cannot write the literal, so it arrives as a field of the scope. The same strings are read by the user edit page, which reveals them, and by the one-click buttons of the profile and the users list, which copy them.

| Method            | What it does                                                      |
|-------------------|-------------------------------------------------------------------|
| `add($site)`      | add is ADD against a site, which is the instance's own base link. |
| `settings($site)` | settings is SETTINGS against a site.                              |

---

The tables come from `tables()` on `mod_login`, `mod_my_keys`, `mod_profile`, `mod_user_edit`. The operations are the classes of `storage/`, which hold every statement this module runs.
