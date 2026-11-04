# Workflows

Generated from the module by `atkins docs`. DO NOT EDIT.

![The tables this module reads](schema.svg)

mod_workflows lists the workflows the reader may see, each with the schedules that start it and how its last run ended, and under them the newest runs of any of them.

A workflow the reader may not see every folder of is not listed at all, so the page and the sidebar hold the same set: Task\Workflow::visible decides it once.

## Pages

| Method | Path                         | Answered by                                      |
|--------|------------------------------|--------------------------------------------------|
| `GET`  | `/workflows/{id}`            | [`mod_workflow.php`](mod_workflow.php)           |
| `GET`  | `/workflows/{id}/edit`       | [`mod_workflow_edit.php`](mod_workflow_edit.php) |
| `GET`  | `/workflows/new`             | [`mod_workflow_new.php`](mod_workflow_new.php)   |
| `GET`  | `/workflows/{id}/runs/{run}` | [`mod_workflow_run.php`](mod_workflow_run.php)   |
| `GET`  | `/workflows`                 | [`mod_workflows.php`](mod_workflows.php)         |

## The data model

### `key_info`

Key Info

| Column               | Type        | Key     |
|----------------------|-------------|---------|
| `name`               | `text`      | primary |
| `purpose`            | `TEXT`      | indexed |
| `algo`               | `text`      |         |
| `bits`               | `integer`   |         |
| `fingerprint`        | `text`      | indexed |
| `scope`              | `text`      | indexed |
| `comment`            | `text`      |         |
| `created_at`         | `timestamp` |         |
| `updated_at`         | `timestamp` |         |
| `backup`             | `text`      |         |
| `backup_fingerprint` | `text`      |         |
| `rotated_at`         | `timestamp` |         |

### `key_role`

Key Role

| Column       | Type        | Key     |
|--------------|-------------|---------|
| `id`         | `text`      | primary |
| `key_name`   | `text`      | indexed |
| `role`       | `TEXT`      | indexed |
| `created_at` | `timestamp` |         |

### `key_user`

Key User

| Column       | Type        | Key     |
|--------------|-------------|---------|
| `id`         | `text`      | primary |
| `key_name`   | `text`      | indexed |
| `user_name`  | `text`      | indexed |
| `created_at` | `timestamp` |         |

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
| `workspace_id`          | `text`      | indexed |
| `dir`                   | `text`      | indexed |

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

### `schedule`

Schedule

| Column        | Type        | Key     |
|---------------|-------------|---------|
| `id`          | `text`      | primary |
| `folder`      | `text`      | indexed |
| `job`         | `text`      | indexed |
| `user_id`     | `text`      | indexed |
| `user_name`   | `text`      |         |
| `expression`  | `text`      |         |
| `enabled`     | `integer`   | indexed |
| `last_run_at` | `timestamp` |         |
| `next_run_at` | `timestamp` | indexed |
| `created_at`  | `timestamp` |         |
| `updated_at`  | `timestamp` |         |
| `target`      | `text`      |         |
| `args`        | `text`      |         |
| `triage`      | `integer`   |         |
| `workflow_id` | `text`      | indexed |

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

### `workflow`

Workflow

| Column           | Type        | Key     |
|------------------|-------------|---------|
| `id`             | `text`      | primary |
| `name`           | `text`      | indexed |
| `user_id`        | `text`      | indexed |
| `user_name`      | `text`      |         |
| `created_at`     | `timestamp` |         |
| `updated_at`     | `timestamp` |         |
| `description`    | `text`      |         |
| `prompt`         | `text`      |         |
| `prompt_folder`  | `text`      |         |
| `run_as_user_id` | `text`      | indexed |
| `runs_over`      | `text`      |         |
| `sidebar_hidden` | `boolean`   |         |

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

### `workflow_step`

Workflow Step

| Column         | Type      | Key     |
|----------------|-----------|---------|
| `id`           | `text`    | primary |
| `workflow_id`  | `text`    | indexed |
| `position`     | `integer` | indexed |
| `folder`       | `text`    |         |
| `job`          | `text`    |         |
| `disabled`     | `boolean` |         |
| `prompt`       | `text`    |         |
| `dir`          | `text`    |         |
| `workflow_ref` | `text`    |         |
| `target`       | `text`    |         |
| `args`         | `text`    |         |
| `kind`         | `TEXT`    |         |
| `markers`      | `text`    |         |
| `tries`        | `integer` |         |

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

This module owns no data model of its own. It reads the instance's tables through `code/Task`.

---

The tables come from `tables()` on `mod_workflow`, `mod_workflow_edit`, `mod_workflow_new`, `mod_workflow_run`, `mod_workflows`. The operations are the classes of `storage/`, which hold every statement this module runs.
