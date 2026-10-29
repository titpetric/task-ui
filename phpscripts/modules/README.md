# modules

The PHP source tree embedded in the binary, one directory per module, named
`<vendor>/<plugin>`. `phpscripts.Extract` writes it into the instance's
modules directory on start and the Manager walks that directory.

The contract a module follows is [docs/modules.md](../../docs/modules.md).
