#### 1.14. Escaping

Printed variables are escaped by default, so `{variable}` is safe
wherever it lands. A value that is already markup opts out with the
`raw` modifier, spelled `{variable|raw}`, or its `unescape` alias.

~~~~~~~~~~~~~~
{variable}
{variable|escape}
{variable|raw}
{variable|unescape}
{variable|tolower}
{variable|tolower|raw}
{variable|toupper|escape}
~~~~~~~~~~~~~~
