# Settings

Generated from the module by `atkins docs`. DO NOT EDIT.

![The tables this module reads](schema.svg)

mod_settings is the index of the settings: the steps in the order an instance is set up in, each with how much it holds.

The page reads and does not write. It is the one step that is not a step itself, so it draws no trail at the foot: a reader is at the start of the walk and every row is a way into it.

## Pages

| Method | Path                                          | Answered by                                                        |
|--------|-----------------------------------------------|--------------------------------------------------------------------|
| `GET`  | `/settings/instance`                          | [`mod_instance.php`](mod_instance.php)                             |
| `GET`  | `/settings/integrations`                      | [`mod_integrations.php`](mod_integrations.php)                     |
| `GET`  | `/settings/keys/{name}`                       | [`mod_key_edit.php`](mod_key_edit.php)                             |
| `GET`  | `/settings/keys/generate`                     | [`mod_key_generate.php`](mod_key_generate.php)                     |
| `GET`  | `/settings/keys/upload`                       | [`mod_key_upload.php`](mod_key_upload.php)                         |
| `GET`  | `/settings/keys`                              | [`mod_keys.php`](mod_keys.php)                                     |
| `GET`  | `/settings/permissions`                       | [`mod_permissions.php`](mod_permissions.php)                       |
| `GET`  | `/settings/projects/{rest...}`                | [`mod_project_edit.php`](mod_project_edit.php)                     |
| `GET`  | `/settings/projects/new`                      | [`mod_project_new.php`](mod_project_new.php)                       |
| `GET`  | `/settings/projects`                          | [`mod_projects.php`](mod_projects.php)                             |
| `GET`  | `/settings/retention`                         | [`mod_retention.php`](mod_retention.php)                           |
| `GET`  | `/settings/secrets/edit`                      | [`mod_secret_edit.php`](mod_secret_edit.php)                       |
| `GET`  | `/settings/secrets/new`                       | [`mod_secret_new.php`](mod_secret_new.php)                         |
| `GET`  | `/settings/secrets/{id}/reads`                | [`mod_secret_reads.php`](mod_secret_reads.php)                     |
| `GET`  | `/settings/secrets`                           | [`mod_secrets.php`](mod_secrets.php)                               |
| `GET`  | `/settings`                                   | [`mod_settings.php`](mod_settings.php)                             |
| `GET`  | `/settings/sidebar`                           | [`mod_sidebar.php`](mod_sidebar.php)                               |
| `GET`  | `/settings/telemetry`                         | [`mod_telemetry.php`](mod_telemetry.php)                           |
| `GET`  | `/settings/users/{id}`                        | [`mod_user_by_id.php`](mod_user_by_id.php)                         |
| `GET`  | `/settings/users/new`                         | [`mod_user_new.php`](mod_user_new.php)                             |
| `GET`  | `/settings/users`                             | [`mod_users.php`](mod_users.php)                                   |
| `GET`  | `/settings/workspaces/{folder}`               | [`mod_workspace_edit.php`](mod_workspace_edit.php)                 |
| `GET`  | `/settings/workspaces/{folder}/keys/generate` | [`mod_workspace_key_generate.php`](mod_workspace_key_generate.php) |
| `GET`  | `/settings/workspaces/{folder}/keys/upload`   | [`mod_workspace_key_upload.php`](mod_workspace_key_upload.php)     |
| `GET`  | `/settings/workspaces/{folder}/keys`          | [`mod_workspace_keys.php`](mod_workspace_keys.php)                 |
| `GET`  | `/settings/workspaces/new`                    | [`mod_workspace_new.php`](mod_workspace_new.php)                   |
| `GET`  | `/settings/workspaces`                        | [`mod_workspaces.php`](mod_workspaces.php)                         |

## The data model

### `env_file`

Env File

| Column       | Type        | Key     |
|--------------|-------------|---------|
| `id`         | `text`      | primary |
| `folder`     | `text`      | indexed |
| `path`       | `text`      | indexed |
| `content`    | `text`      |         |
| `created_at` | `timestamp` |         |
| `updated_at` | `timestamp` |         |

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

### `project`

Project

| Column           | Type        | Key     |
|------------------|-------------|---------|
| `id`             | `text`      | primary |
| `folder`         | `text`      | indexed |
| `title`          | `text`      |         |
| `description`    | `text`      |         |
| `source_url`     | `text`      |         |
| `branch`         | `text`      |         |
| `allow`          | `text`      |         |
| `commands`       | `text`      |         |
| `sources`        | `text`      |         |
| `job_allow`      | `text`      |         |
| `created_at`     | `timestamp` |         |
| `updated_at`     | `timestamp` |         |
| `shallow`        | `boolean`   |         |
| `sidebar_hidden` | `boolean`   |         |

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

### `secret`

Secret

| Column       | Type        | Key     |
|--------------|-------------|---------|
| `id`         | `text`      | primary |
| `scope`      | `text`      | indexed |
| `name`       | `text`      | indexed |
| `content`    | `text`      |         |
| `created_at` | `timestamp` |         |
| `updated_at` | `timestamp` |         |
| `is_secret`  | `integer`   |         |

### `secret_read`

Secret Read

| Column       | Type        | Key     |
|--------------|-------------|---------|
| `id`         | `text`      | primary |
| `secret_id`  | `text`      | indexed |
| `name`       | `text`      |         |
| `run_id`     | `text`      | indexed |
| `folder`     | `text`      |         |
| `dir`        | `text`      |         |
| `job`        | `text`      |         |
| `user_name`  | `text`      |         |
| `created_at` | `timestamp` | indexed |

### `setting`

Setting

| Column       | Type        | Key     |
|--------------|-------------|---------|
| `name`       | `text`      | primary |
| `value`      | `text`      |         |
| `updated_at` | `timestamp` |         |

### `sidebar`

Sidebar

| Column       | Type        | Key     |
|--------------|-------------|---------|
| `id`         | `text`      | primary |
| `kind`       | `text`      |         |
| `href`       | `text`      | indexed |
| `title`      | `text`      |         |
| `icon`       | `text`      |         |
| `position`   | `integer`   | indexed |
| `hidden`     | `boolean`   |         |
| `rule`       | `text`      |         |
| `created_at` | `timestamp` |         |
| `updated_at` | `timestamp` |         |

### `tool`

Tool

| Column        | Type        | Key     |
|---------------|-------------|---------|
| `name`        | `text`      | primary |
| `category`    | `text`      |         |
| `version`     | `text`      |         |
| `path`        | `text`      |         |
| `description` | `text`      |         |
| `updated_at`  | `timestamp` |         |

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

### `Keys`

Keys reads the key tables: the instance SSH keys, and the two tables that say who gets one, by the role a user acts with and by their name. The Keys page is this module's and the statements are here with it. The tables are the Go keystore's to write and migrate, under keystore/schema/, which is why this module's own schema/ creates none of them. No private material is here: the keystore holds it and it does not enter PHP.

| Method              | What it does                                                                       |
|---------------------|------------------------------------------------------------------------------------|
| `ssh()`             | ssh is the name of every instance SSH key, by name; a workspace's own is no run's. |
| `roles()`           | roles is the roles each key is granted to, by key name.                            |
| `holders()`         | holders is the users each key is granted to, by key name.                          |
| `links($statement)` | links reads one of the two grant tables as the values of each key name.            |

### `Policies`

Policies reads and writes the permission_policy table: an effect per path, per subject. A subject is a role or a user, by name; the admin role holds no entries. The table is this module's, which draws the permissions page. What an entry means to a reader is Task\Policy.

| Method                                 | What it does                                                                                          |
|----------------------------------------|-------------------------------------------------------------------------------------------------------|
| `policy($kind, $subject)`              | policy is a subject's entries.                                                                        |
| `put($kind, $subject, $path, $effect)` | put sets a subject's entry for a path, or removes it with an empty effect so the path inherits again. |
| `forget($kind, $subject)`              | forget removes every entry of a subject, as when its user goes.                                       |

### `Projects`

Projects reads the project table: which projects the sidebar hides, the folder of every project row, and one row by folder. The table is this module's, which draws the Projects page. What a project means to a page - its workspace, its crumbs, what a move of it leaves behind - is Task\Project.

| Method         | What it does                                                                                                                                                                                                                                                |
|----------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `hidden()`     | hidden is the folder of every project set not to show in the sidebar.                                                                                                                                                                                       |
| `folders()`    | folders is the folder of every project row, which the workspaces count beside the directories the host walked. A row whose directory is not there yet is a project of its workspace all the same: the clone failed and the page lists it with git's output. |
| `row($folder)` | row is the project a folder names, null where no row does. A directory somebody made by hand has no row, and the page of it draws the settings of a project the instance keeps nothing about.                                                               |

### `Refs`

Refs reads the rows of this module's tables that name a folder: a variable and an env file, which a move of the project leaves where they were. Each owning module answers for its own tables, and Task\Project asks all of them. An ID may hold the underscore LIKE matches any character with, so the match is on the ID itself or on the ID and a slash as a prefix. storage/folder_ref_storage.go.

| Method    | What it does                                                          |
|-----------|-----------------------------------------------------------------------|
| `of($id)` | of is the variables and the env files a folder and its children name. |

### `Secrets`

Secrets reads the secret and secret_read tables: the environment variables and secrets a workspace's runs are given, and which run got which. The tables are this module's, which draws the Secrets page. No value is here: a secret's value is encrypted with the instance's key and the key does not enter PHP, so nothing answers more than what a page counts or lists. What a row means to a page is Task\Secret.

| Method                                | What it does                                                                                                                                                                                                                                                                                                                                                                                                  |
|---------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `count()`                             | count is how many variables the settings list, the integrations' apart.                                                                                                                                                                                                                                                                                                                                       |
| `scopes()`                            | scopes is the scope of every variable the settings list, one entry per variable, which the keys page counts by workspace.                                                                                                                                                                                                                                                                                     |
| `all()`                               | all is every variable the settings list, in the order the page draws them: those of every workspace first, then each workspace's, then a folder's, and by name inside each. storage/secret_storage.go:81.                                                                                                                                                                                                     |
| `counts()`                            | counts is how many runs got each variable, by its ID.                                                                                                                                                                                                                                                                                                                                                         |
| `last()`                              | last is the newest read of each variable, by its ID. The newest is the greatest rowid and not the greatest timestamp, for the reason reads() gives: the reads of one run share a millisecond.                                                                                                                                                                                                                 |
| `byId($id)`                           | byId is one variable the settings list, or null. An integration's own setting is no variable of the settings, so an ID that names one answers as an ID that names nothing.                                                                                                                                                                                                                                    |
| `edited($id, $scope, $name)`          | edited is the variable a form was opened on: the row an ID names, or the row a scope and a name name where no ID was given, null where neither names one. frontend/secret.go:127. The integration scope is not left out here, which byId() leaves out: the form opens whatever row it was pointed at, and a service's own setting has a form like any other variable.                                         |
| `reads($id, $before, $after, $limit)` | reads is a page of the runs that got a variable, newest first, or oldest first for a page read after a cursor, which is reversed here. The order is the row's and not the timestamp's: the reads of one run are written in one transaction and share a millisecond, so a timestamp orders none of them and rowid orders all of them. The cursor is a read's ID, and a cursor whose row is gone keeps nothing. |

### `Settings`

Settings reads the setting table: the instance's own values, by name. The table is this module's. What a value means - the retention rules, the sidebar's limits - is read off it by the class that owns that question.

| Method     | What it does                                                                                                                                                  |
|------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `values()` | values is the table by name. The whole table is read in one statement, as the cleanup reads it: the rows are a handful and the names of two of them per kind. |

### `Sidebars`

Sidebars reads and writes the sidebar table: the top block of the sidebar, in its order, and the rows the Sidebar settings page edits. The table is this module's. What a row draws and who sees it is Task\Sidebar.

| Method                   | What it does                                                                    |
|--------------------------|---------------------------------------------------------------------------------|
| `all()`                  | all is every row, hidden ones included, in the order the table holds.           |
| `byHref($href)`          | byHref is one row by its link, or null.                                         |
| `order($hrefs, $hidden)` | order writes the position and the hidden flag of each link, in one transaction. |

### `Tools`

Tools reads the tool table: what the last check recorded of the runtimes and the tools the image carries. The table is this module's, which draws the Instance page. What a row means and which category it belongs to is Task\Toolset.

| Method  | What it does                                                             |
|---------|--------------------------------------------------------------------------|
| `all()` | all is every tool the last check recorded, by category and then by name. |

### `Workspaces`

Workspaces reads the workspace table: the folder of every workspace, the rows themselves, each title, and the sidebar limit of the ones that set one. The table is this module's, which draws the Workspaces page. What a workspace means to a page - where its home is, which projects are inside it - is Task\Workspace.

| Method         | What it does                                                                                                                                    |
|----------------|-------------------------------------------------------------------------------------------------------------------------------------------------|
| `folders()`    | folders is the folder of every named workspace, in the sidebar's order.                                                                         |
| `names()`      | names is the folder of every workspace row, the default one first and the named ones in the sidebar's order after it.                           |
| `all()`        | all is every workspace row, the default one first and the named ones in the sidebar's order after it, which is storage/workspace_storage.go:69. |
| `row($folder)` | row is one workspace row by folder, or null.                                                                                                    |
| `titles()`     | titles is the title of every workspace, by folder.                                                                                              |
| `limits()`     | limits is how many projects the sidebar lists for a workspace that limits them, by folder; a workspace that lists every one is not here.        |

---

The tables come from `tables()` on `mod_instance`, `mod_integrations`, `mod_key_edit`, `mod_key_generate`, `mod_key_upload`, `mod_keys`, `mod_permissions`, `mod_project_edit`, `mod_project_new`, `mod_projects`, `mod_retention`, `mod_secret_edit`, `mod_secret_new`, `mod_secret_reads`, `mod_secrets`, `mod_settings`, `mod_sidebar`, `mod_telemetry`, `mod_user_by_id`, `mod_user_new`, `mod_users`, `mod_workspace_edit`, `mod_workspace_key_generate`, `mod_workspace_key_upload`, `mod_workspace_keys`, `mod_workspace_new`, `mod_workspaces`. The operations are the classes of `storage/`, which hold every statement this module runs.
