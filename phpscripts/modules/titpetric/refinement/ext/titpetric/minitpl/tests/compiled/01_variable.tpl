<?php $_v=&$this->vars;?>
#### 1.1. Variables

Using variables from templates is easy. Variables are enclosed
in curly braces like so: `<?php echo htmlspecialchars($_v['variable'], ENT_QUOTES);?>
`. Since variables are parsed
on the last stage, prefixing variables with the dollar sign is
optional. So, `<?php echo htmlspecialchars($_v['variable'], ENT_QUOTES);?>
` is the same as `<?php echo htmlspecialchars($_v['variable'], ENT_QUOTES);?>
`.

~~~~~~~~~~~~~
<?php echo htmlspecialchars($_v['variable'], ENT_QUOTES);?>
 is the same as <?php echo htmlspecialchars($_v['variable'], ENT_QUOTES);?>

~~~~~~~~~~~~~
