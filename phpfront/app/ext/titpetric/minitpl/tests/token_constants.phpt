name: token constants
description: PHP tokenizer token IDs can be compared with their named constants.
---
<?php
$tokens = token_get_all('<?php $object->method();');
$matched = array();

foreach ($tokens as $token) {
	if (is_array($token) && $token[0] == T_VARIABLE) {
		$matched[] = "T_VARIABLE";
	}
	if (is_array($token) && $token[0] == T_OBJECT_OPERATOR) {
		$matched[] = "T_OBJECT_OPERATOR";
	}
}

echo implode(",", $matched);
---
T_VARIABLE,T_OBJECT_OPERATOR
