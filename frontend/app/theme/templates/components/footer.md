# footer

Input scope: `data`.

| Field          | Type   | What it is                                     |
|----------------|--------|------------------------------------------------|
| `assets`       | string | Where the static files answer, outside the base |
| `assetVersion` | string | The content hash the asset URLs carry as `?v=`  |
| `scripts`      | list   | The serving module's own scripts, from `Task\View::assets()`, after the site's |

The three scripts are the whole of the page's behaviour: the terminal, its fit
addon and the site's own. A page writes an attribute and this fills it; the
contract is in docs/framework.md.

A module's own script comes after them and adds what a plain form has no
spelling for: the site's script reaches no global, so a module that posts an
ordinary form rather than through the API writes its own confirm and its own
Enter.
