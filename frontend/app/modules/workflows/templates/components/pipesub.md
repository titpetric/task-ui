# pipesub

Input scope: `data_sub`, one entry of a `pipenode`'s `sub`, from `Task\Pipeline::nodes()`.

The fields are `pipenode`'s own; see `pipenode.md` for the schema.

A step of the workflow run a step made. The markup is `pipenode`'s, less the inner list: a step of this depth draws the empty `<ol class="pipeline sub">` of the workflow run it made and no steps inside it, because a pipeline draws one level down and no further.

The component exists because `{include}` is a compile-time replace that cannot nest into itself, so the depth is unrolled rather than recursed.
