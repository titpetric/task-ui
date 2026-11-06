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

Two meta tags paint the browser around the page. `theme-color`, one per scheme,
tints the address bar and the status bar of a phone with `--ink-0`, the colour
the header already holds, so the chrome reads as part of the page; a stylesheet
cannot reach it, so the hex is written here and in `css/tokens.css`.
`color-scheme` says the same as the `:root` declaration and arrives before the
stylesheet does, so the form controls and the scrollbars start dark instead of
flashing white.
