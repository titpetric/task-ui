<?php $_v=&$this->vars;?>
#### 1.14. Escaping

Printed variables are escaped by default, so `<?php echo htmlspecialchars($_v['variable'], ENT_QUOTES);?>
` is safe
wherever it lands. A value that is already markup opts out with the
`raw` modifier, spelled `<?php echo $_v['variable'];?>
`, or its `unescape` alias.

~~~~~~~~~~~~~~
<?php echo htmlspecialchars($_v['variable'], ENT_QUOTES);echo htmlspecialchars($_v['variable'], ENT_QUOTES);echo $_v['variable'];echo $_v['variable'];echo htmlspecialchars(strtolower($_v['variable']), ENT_QUOTES);echo strtolower($_v['variable']);echo htmlspecialchars(strtoupper($_v['variable']), ENT_QUOTES);?>

~~~~~~~~~~~~~~
