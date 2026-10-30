<?php $_v=&$this->vars;?>
#### 1.2. Arrays

There is a shorthand syntax for using arrays inside a template.
The array index separator is a dot (`.`). Consecutive dots can be
used for traversing into array depth like `<?php echo htmlspecialchars($_v['array']['items']['0'], ENT_QUOTES);?>
`, or
even variables starting with `$`, like `<?php echo htmlspecialchars($_v['sections'][$_v['news_section']]['title'], ENT_QUOTES);?>
`.
PHP syntax for arrays is also supported.

~~~~~~~~~~~~
<?php echo htmlspecialchars($_v['array']['items'], ENT_QUOTES);?>
 is the same as <?php echo htmlspecialchars($_v['array']['items'], ENT_QUOTES);echo htmlspecialchars($_v['array']['items']['0'], ENT_QUOTES);?>
 is the same as <?php echo htmlspecialchars($_v['array']['items'][0], ENT_QUOTES);echo htmlspecialchars($_v['array'][$_v['items']]['0'], ENT_QUOTES);?>
 is the same as <?php echo htmlspecialchars($_v['array'][$_v['items']][0], ENT_QUOTES);?>

~~~~~~~~~~~~
