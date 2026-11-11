<?php $_v=&$this->vars;?>
#### 1.6. Includes

#### 1.5. Constants

Value starting with `_` is assumed to be a constant. The template
system will output the value of the constant if defined,
or just the name of the constant. Item `<?php echo htmlspecialchars(_MY_CONSTANT, ENT_QUOTES);?>
` will be
used as a constant, because of those rules, however `<?php echo htmlspecialchars($_v['_my_variable'], ENT_QUOTES);?>
`
wouldnt be, since it starts with the variable identifier `$`.

~~~~~~~~~~~
<?php echo htmlspecialchars(_MY_CONSTANT, ENT_QUOTES);echo htmlspecialchars(_this_is_also_a_constant, ENT_QUOTES);echo htmlspecialchars($_v['_my_variable'], ENT_QUOTES);?>

~~~~~~~~~~~
