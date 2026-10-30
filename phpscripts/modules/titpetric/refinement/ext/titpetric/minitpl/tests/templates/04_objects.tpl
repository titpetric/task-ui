#### 1.4. Objects

Objects used by templates must be assigned like any other template
variable. You can call their methods or access their properties.

~~~~~~~~~~~
{$memcache->get()}
{$memcache->variable}
~~~~~~~~~~~

Both objects are resolved from the template's assigned variables:

~~~~~~~~~~~
<?php
	$_v = &$this->vars;
	echo $_v['memcache']->get();
	echo $_v['memcache']->variable;
?>
~~~~~~~~~~~
