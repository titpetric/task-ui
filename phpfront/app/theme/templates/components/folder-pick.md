# folder-pick

Input scope: `data_pick`. `choices` comes from `Task\Folders::choices()`.

| Field               | Type   | What it is                                           |
|---------------------|--------|------------------------------------------------------|
| `all`               | bool   | Whether the user sees every folder, which All folders picks |
| `label`             | bool   | Whether the field writes its own Folders heading      |
| `choices`           | list   | A box each                                            |
| `choices[].id`      | string | The folder, which is the value the form puts          |
| `choices[].checked` | bool   | Whether the user's row lists it                       |

The folders one user sees, as the user forms ask for them: All folders, or
Custom with a box per folder. An empty list on the row means every folder, so
All folders is the empty list and not a list of all of them — a folder
discovered after the form was saved comes with it.

The boxes stay drawn under All folders and go disabled, so picking Custom shows
what was ticked before without a round trip.

`label` is false where the block's own heading already says Folders, which is
the page that edits a user; the form that adds one sets it, because there the
field stands among others with nothing above it.
