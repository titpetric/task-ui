# host-jobs

Input scope: `data_hostjobs`.

| Field        | Type   | What it is                                              |
|--------------|--------|---------------------------------------------------------|
| `[]`         | list   | One option per host job, `@host:*` first, else empty    |
| `[].id`      | string | `@host:<host>`, which is the value a schedule is written with |
| `[].name`    | string | What the option says, which is the ID                   |
| `[].folder`  | string | The folder the option narrows under, always `internal`  |
| `[].title`   | string | What the option's tooltip says the job does             |
| `[].selected`| bool   | Whether the select opens on this one                    |

The `Hosts` group of a job picker, from `Task\HostJobs::of()`. The `<select>`
around it is the page's, and the folder jobs before it are too: this draws the
group and nothing else, and draws nothing at all with no host.
