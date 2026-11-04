# mcp-access

Input scope: `data_mcp`, from `mod_user_edit`.

| Field      | Type   | What it is                                                    |
|------------|--------|---------------------------------------------------------------|
| `name`     | string | The user the block speaks about, which the confirms name      |
| `api`      | string | Where New token posts, and the stem each row's Delete extends |
| `done`     | string | The page Done returns to, with the notice that says it worked |
| `on`       | bool   | Whether the user holds a token at all                         |
| `agent`    | bool   | Whether the user's row allows MCP at all                      |
| `enable`   | bool   | Whether the reader may make one                               |
| `manage`   | bool   | Whether the reader may make one or delete one                 |
| `tokens`   | list   | One row per token, newest first                               |
| `fill`     | string | The placeholder task-ui.js fills with the token               |
| `until`    | string | The placeholder it fills with the expiry                      |
| `named`    | string | The placeholder it fills with the label the token took        |
| `add`      | string | The `claude mcp add` line, with the placeholder inside it     |
| `settings` | string | The same connection as a settings file, with the placeholder  |

Each entry of `tokens`:

| Field       | Type   | What it is                                                        |
|-------------|--------|-------------------------------------------------------------------|
| `name`      | string | The label the token took, unique among the user's MCP ones        |
| `delete`    | string | Where this row's Delete sends its DELETE                          |
| `created`   | string | How long ago the token was made                                   |
| `createdAt` | string | The stamp behind it                                               |
| `expires`   | bool   | Whether the token ends, which a token from the CLI does not       |
| `expiresAt` | string | The stamp it ends at                                              |
| `expired`   | bool   | Whether that stamp has passed                                     |
| `tone`      | string | The badge class: `bad` once it has passed, `signal` within a week |
| `used`      | bool   | Whether the token has ever answered a request                     |
| `usedAgo`   | string | How long ago that was                                             |
| `usedAt`    | string | The stamp behind it                                               |

The MCP tokens of one user, as the page that edits them shows them: a table of what each token is, and the one control the reader is allowed on each, which is Delete. A token is never rotated; one is made and one is deleted.

A token is shown once. The POST answers with it, task-ui.js reveals `#mcp-token` and fills every `data-fill` whose value matches a key of the answer — which is why `fill`, `until`, `named`, `add` and `settings` are fields and not markup: a template writes `{token}` as a variable, so the literal has to arrive as data.

Done is a navigation and not this page's own anchor: the reader is already at `#mcp`, so a link there scrolls nowhere and leaves the revealed token standing over a table that was never read again.

The three states are tokens, none with the row allowing one, and none with the row not. The last says where the switch is rather than offering it, because the switch is Roles above and belongs to an admin.
