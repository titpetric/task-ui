# show-source

Input scope: `data_source`, from `Task\Place::source()`.

| Field    | Type   | What it is                                              |
|----------|--------|---------------------------------------------------------|
| `folder` | string | The folder the file is read under                       |
| `file`   | string | The file, as the source window asks for it              |
| `label`  | string | What the button says: how many lines it holds           |

The Show button of a file a place's pipeline reads: it opens the file,
coloured, in the source window. The Folder panel draws one beside each source
it was detected in, and one beside each file the pipeline includes.
