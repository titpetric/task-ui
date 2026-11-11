# recursive-field

Input scope: `data_recursive`, the Recursive toggle of a workspace or project form.

| Field     | Type   | What it is                                      |
|-----------|--------|-------------------------------------------------|
| `id`      | string | The `id` of the checkbox, which its label names |
| `checked` | bool   | Whether discovery walks every folder inside     |

One field for four forms: Add workspace, a workspace's settings, Add project and a project's settings each draw it, which is why it is a component rather than markup on each page. The `id` differs per form because two of them draw other fields with ids of their own, and a label points at one id.

Off, discovery lists the folder and its top-level folders alone. The toggle posts as `recursive`, and the forms that edit something read the stored flag the other way round: a row stores `shallow`, and this is checked where it is not set.
