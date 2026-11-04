# run-as-field

Input scope: `data_runas`.

| Field     | Type   | What it is                                               |
|-----------|--------|----------------------------------------------------------|
| `id`      | string | The id of the select, which its label points at          |
| `options` | object | The users to pick from, read by `user-options`           |

The Runs as field of a form that schedules something: the schedules page's Add
schedule and the same form on a job's page. Both write `sch-user`, because
neither page holds two of them.

An empty `options.users` draws nothing at all, field and label included: only
an admin picks who a run is as, so `Task\Chrome::runAsUsers()` answers an
empty list to everyone else and the form keeps the user it is submitted by.

The field is the label, the select and the line under it, and not the `<form>`
around them: the page writes that, with the endpoint and the notice of its own.
