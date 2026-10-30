<?php $_v=&$this->vars;?>
#### 1.4. Objects

Objects used by templates must be assigned like any other template
variable. You can call their methods or access their properties.

~~~~~~~~~~~
<?php echo htmlspecialchars($_v['memcache']->get(), ENT_QUOTES);echo htmlspecialchars($_v['memcache']->variable, ENT_QUOTES);?>

~~~~~~~~~~~

Both objects are resolved from the template's assigned variables:

~~~~~~~~~~~
<?php
	$_v = &$this->vars;
	echo $_v['memcache']->get();
	echo $_v['memcache']->variable;
?>
~~~~~~~~~~~
