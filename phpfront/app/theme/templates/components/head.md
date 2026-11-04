# head

Input scope: `data`.

| Field          | Type   | What it is                                      |
|----------------|--------|-------------------------------------------------|
| `title`        | string | The page's own title, before the instance's name |
| `assets`       | string | Where the static files answer, outside the base |
| `assetVersion` | string | The content hash the asset URLs carry as `?v=`  |
| `styles`       | list   | The serving module's own stylesheets, from `Task\View::assets()`, after the theme's |

A module's stylesheet comes after the theme's, which is the cascade: it styles
the markup only that module draws and overrides nothing else.
