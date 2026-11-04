# mcp-access

Input scope: `data_mcp`, from `mod_user_edit`.

| Field       | Type   | What it is                                                        |
|-------------|--------|-------------------------------------------------------------------|
| `name`      | string | The user the block speaks about, which the confirms name          |
| `api`       | string | Where Enable and Rotate post and Disable deletes                  |
| `done`      | string | The page Done returns to, which is this one at the block          |
| `on`        | bool   | Whether a token exists                                            |
| `agent`     | bool   | Whether the user's row allows MCP at all                          |
| `enable`    | bool   | Whether the reader may turn it on                                 |
| `manage`    | bool   | Whether the reader may rotate or disable the token there is       |
| `created`   | string | How long ago the token was made                                   |
| `createdAt` | string | The stamp behind it                                               |
| `expires`   | bool   | Whether the token ends, which a token from the CLI does not       |
| `expiresAt` | string | The stamp it ends at                                              |
| `expired`   | bool   | Whether that stamp has passed                                     |
| `tone`      | string | The badge class: `bad` once it has passed, `signal` within a week |
| `used`      | bool   | Whether the token has ever answered a request                     |
| `usedAgo`   | string | How long ago that was                                             |
| `usedAt`    | string | The stamp behind it                                               |
| `fill`      | string | The placeholder task-ui.js fills with the token                   |
| `until`     | string | The placeholder it fills with the expiry                          |
| `add`       | string | The `claude mcp add` line, with the placeholder inside it         |

One user's MCP token, as the page that edits them shows it: what the token is, and the one control the reader is allowed.

A token is shown once. The POST answers with it, task-ui.js reveals `#mcp-token` and fills every `data-fill` whose value matches a key of the answer — which is why `fill`, `until` and `add` are fields and not markup: a template writes `{token}` as a variable, so the literal has to arrive as data.

The three states are a token, no token with the row allowing one, and no token with the row not. The last says where the switch is rather than offering it, because the switch is Roles above and belongs to an admin.
