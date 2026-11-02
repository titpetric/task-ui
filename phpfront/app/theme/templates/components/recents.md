# recents

Input scope: `data`, the right rail: the latest runs and workflow runs by
anyone, grouped by when they started.

| Field                     | Type   | What it is                                      |
|---------------------------|--------|-------------------------------------------------|
| `href`                    | string | Where the label and the last slot go            |
| `groups`                  | list   | The groups, newest first                        |
| `groups[].label`          | string | What the group heading says                     |
| `groups[].entries`        | list   | The runs of the group                           |
| `entries[].href`          | string | The run or the workflow, without the base       |
| `entries[].name`          | string | The job's name, or the workflow's               |
| `entries[].place`         | string | Where it ran                                    |
| `entries[].state`         | string | The run state, which colours the dot            |
| `entries[].ended`         | bool   | Whether the run is over                         |
| `entries[].act`           | string | `play` or `eye`, the icon for what the reader may do |
| `entries[].flow`          | bool   | Whether the entry is a workflow run             |
| `entries[].user`          | string | Who started it                                  |
| `entries[].role`          | string | Their role, which colours the avatar            |
| `entries[].initials`      | string | The one or two letters the avatar draws         |

The last slot is the script's: it puts "All runs" in the last slot that fits
when the rail cannot show every entry.

The groups arrive with the `runs` module, which owns the run table. Until then
`groups` is empty and the component draws the empty line.
