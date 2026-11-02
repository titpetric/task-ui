# footer

Input scope: `data`.

| Field          | Type   | What it is                                     |
|----------------|--------|------------------------------------------------|
| `assets`       | string | Where the static files answer, outside the base |
| `assetVersion` | string | The content hash the asset URLs carry as `?v=`  |

The three scripts are the whole of the page's behaviour: the terminal, its fit
addon and the site's own. A page writes an attribute and this fills it; the
contract is in docs/framework.md.
