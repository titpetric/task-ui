<?php

namespace MiniTPL;

/*

Tit Petric <black@titpetric.com>, MIT license, see LICENSE.

 */

/** Template compiler class */
class Compiler {
	/** Markup contexts a template variable can be printed in */
	const CONTEXT_TEXT = "text";
	const CONTEXT_TAG = "tag";
	const CONTEXT_ATTRIBUTE_DQ = "attribute_dq";
	const CONTEXT_ATTRIBUTE_SQ = "attribute_sq";
	const CONTEXT_COMMENT = "comment";
	const CONTEXT_RAW = "raw";
	const CONTEXT_PHP = "php";

	protected $hooks = array(Hook::POSITION_PRE => array(), Hook::POSITION_POST => array());

	public $_tag_php_open = "<" . "?php";
	public $_tag_php_close = "?" . ">\n";
	public $_global_variables = array();
	public $_literals = array();

	/** Auto-escape printed variables unless the context or a modifier says otherwise */
	public $_escape = true;

	function set_hooks($hooks) {
		$this->hooks = $hooks;
	}

	/** Enable or disable auto-escaping for this compilation */
	function set_escape($escape) {
		$this->_escape = $escape ? true : false;
	}

	protected function load_contents($filename) {
		$contents = file_get_contents($filename);
		if ($contents !== false) {
			if (substr($contents, 0, 3) == "\xEF\xBB\xBF") {
				return substr($contents, 3);
			}
		}

		return $contents;
	}

	/** Compile template file into php code */
	function compile($filename, $output_filename, $find_path, $nocache) {
		$contents = $this->load_contents($filename);

		foreach ($this->hooks[Hook::POSITION_PRE] as $hook) {
			$contents = $hook->execute($filename, $contents);
		}

		$r = 0;
		if ($contents !== false && $contents !== "") {
			while (preg_match_all("/\{include\ (.*?)\}/s", $contents, $matches)) {
				$matches = array_unique($matches[1]);
				foreach ($matches as $file) {
					$cn = "<!-- " . $file . " -->";
					$fn = call_user_func($find_path, $file);
					if ($fn !== false) {
						$cn = $this->load_contents($fn . $file);
					}

					$contents = str_replace("{include " . $file . "}", $cn, $contents);
				}
			}

			while (preg_match_all("/\{load\ (.*?)\}/s", $contents, $matches)) {
				$matches = array_unique($matches[1]);
				foreach ($matches as $file) {
					$file_var = (substr($file, 0, 1) == '$') ? $this->_split_exp($file) : '"' . $file . '"';
					$cn = $this->_code('$this->push();$this->load(' . $file_var . ');$this->assign($_v);$this->render();$this->pop();');
					$contents = str_replace("{load " . $file . "}", $cn, $contents);
				}
			}

			$nocache = $nocache ? $this->_code("@unlink(__FILE__);") : "";
			$contents = str_replace("{*nocache*}", $nocache, $contents);
			// A template that isn't markup - json, plain text mail - opts out
			// of auto-escaping for its whole body.
			if (strpos($contents, "{*noescape*}") !== false) {
				$this->_escape = false;
				$contents = str_replace("{*noescape*}", "", $contents);
			}

			$contents = $this->_strip_comments($contents);
			$contents = $this->_parse_constants($contents);
			$contents = $this->_parse_functions($contents, $filename);
			$contents = $this->_parse_expressions($contents);
			$contents = $this->_parse_variables($contents);
			if (!empty($this->_global_variables)) {
				$globals = array_unique($this->_global_variables);
				$contents = $this->_code('global ' . implode(", ", $globals) . ';') . $contents;
			}

			$contents = $this->_template_cleanup($contents);

			foreach ($this->hooks[Hook::POSITION_POST] as $hook) {
				$contents = $hook->execute($filename, $contents);
			}

			$this->_r_mkdir(dirname($output_filename));

			// Written beside the target and renamed into place, never opened
			// over it. fopen($output_filename, "w") truncates the file to zero
			// before the first byte is written, so anything that include()s a
			// compiled template while a compile is running reads an empty or
			// half-written file - a page that renders its doctype and its head
			// and then stops. rename(2) is atomic within a filesystem, so a
			// reader sees either the previous complete file or the new one.
			//
			// Two concurrent renders of the same template are ordinary: a test
			// suite running fixtures in parallel, or several server processes
			// sharing one cache directory. The compile is idempotent, so the
			// loser of the rename has written the same bytes as the winner.
			//
			// The temporary name carries the pid where one is available, so
			// separate processes do not collide on the temporary itself.
			$suffix = function_exists("posix_getpid") ? posix_getpid() : "tmp";
			$tmp = $output_filename . "." . $suffix . ".tmp";

			$f = fopen($tmp, "w");
			if ($f) {
				fwrite($f, $contents);
				fclose($f);

				if (rename($tmp, $output_filename)) {
					$r = 1;
				} else {
					unlink($tmp);
				}
			}
		}

		return $r;
	}

	function _r_mkdir($dir) {
		if (file_exists($dir)) {
			return;
		}

		if (!file_exists(dirname($dir))) {
			$this->_r_mkdir(dirname($dir));
		}

		mkdir($dir);
	}

	/** Insert system configuration, clean up code */
	function _template_cleanup($contents) {
		// set up variables
		$contents = $this->_code('$_v=&$this->vars;') . $contents;
		// strip unnecessary php tags
		$contents = str_replace($this->_tag_php_close . $this->_tag_php_open, "", $contents);
		// strip new line whitespace between php code
		$contents = str_replace($this->_tag_php_close . "\n" . $this->_tag_php_open . ' ', "", $contents);
		$contents = str_replace("echo ;", "", $contents);
		foreach ($this->_literals as $key => $value) {
			$contents = str_replace("[[" . $key . "]]", $value, $contents);
		}

		return $contents;
	}

	/** Strip template style comments */
	function _strip_comments($contents) {
		$contents = preg_replace("/\{\*.+\*\}/sU", "", $contents);
		return $contents;
	}

	/** Replace constant definitions */
	function _parse_constants($contents) {
		return $this->_replace_tags($contents, true);
	}

	/** Whether a tag names a constant rather than a variable */
	function _is_constant($v) {
		return preg_match("/^\_[a-zA-Z0-9\_]+$/", $v) == 1;
	}

	/** Search and replace for function blocks and inline definitions */
	function _parse_functions($contents, $filename) {
		$inlines = array();
		$blocks = array();

		if (preg_match_all("/\{(block|inline)\ ([a-zA-Z0-9\_\-]+)\}(.*?)\{\/\\1\}/s", $contents, $matches)) {
			foreach ($matches[0] as $k => $ma) {
				$m = array("content" => trim($matches[3][$k]), "src" => $ma);
				if ($matches[1][$k] == "block") {
					$blocks[$matches[2][$k]] = $m;
				} else {
					$inlines[$matches[2][$k]] = $m;
				}
			}
		}

		if (preg_match_all("/\<script\ ([^\>]+)\>(.*?)\<\/script\>/s", $contents, $matches)) {
			foreach ($matches[0] as $k => $parameters) {
				if (strpos($parameters, "text/template") !== false || strpos($parameters, "text/x-jquery") !== false) {
					$key = count($this->_literals) . "_literal";
					$this->_literals[$key] = $matches[2][$k];
					$contents = str_replace($matches[2][$k], "[[" . $key . "]]", $contents);
				}
			}
		}

		foreach ($blocks as $name => $code) {
			$lambda = sprintf("%u", crc32($code['content'])) . "_" . sprintf("%u", crc32($filename));
			$block_code = "if (!function_exists('" . $name . "_" . $lambda . "')) { function " . $name . "_" . $lambda . "(\$_v) {" . $this->_tag_php_close . $code['content'] . $this->_tag_php_open . " } }";
			$contents = str_replace($code['src'], $this->_code($block_code), $contents);
		}

		foreach ($inlines as $name => $code) {
			$contents = str_replace($code['src'], '', $contents);
		}

		$matches = array();
		while (preg_match_all("/\{inline\:([a-zA-Z0-9\_\-]+)\}/s", $contents, $matches)) {
			foreach ($matches[0] as $k => $ma) {
				$contents = str_replace($ma, $inlines[$matches[1][$k]]['content'], $contents);
			}
		}

		foreach ($blocks as $name => $code) {
			$contents = str_replace("{block:" . $name . "}", $this->_code($name . "_" . $lambda . "(&\$_v);"), $contents);
		}

		return $contents;
	}

	/** Parse expression syntax: if, elseif, foreach, else, for, eval, eval_literal */
	function _parse_expressions($contents) {
		// foreach parsing
		if (preg_match_all("/\{foreach (.+)\}/sU", $contents, $matches)) {
			foreach ($matches[1] as $k => $exp) {
				$exp = trim(trim($exp, "()"));
				list($e_left, $e_right) = explode(" as ", $exp);
				$e_right = explode("=>", $e_right);

				$left_exp = $this->_split_exp($e_left);
				$code = "";
				if (substr($left_exp, 0, 5) !== "array") {
					$code = "if(!empty(" . $left_exp . "))";
				}

				$code .= "foreach(" . $left_exp . " as " . $this->_split_exp($e_right[0]);
				if (count($e_right) == 2) {
					$code .= '=>' . $this->_split_exp($e_right[1]);
				}

				$code .= '){';
				$contents = str_replace($matches[0][$k], trim($this->_code($code)), $contents);
			}
		}

		// if & for & elseif parsing
		if (preg_match_all("/\{(if|elseif|for|while) (.+)\}/sU", $contents, $matches)) {
			foreach ($matches[1] as $k => $v) {
				if ($v == "for") {
					$matches[2][$k] = trim($matches[2][$k], "()");
				}

				$code = $v . "(" . $this->_split_exp($matches[2][$k]) . "){";
				if ($v == "elseif") {
					$code = "}" . $code;
				}

				$contents = str_replace($matches[0][$k], $this->_code($code), $contents);
			}
		}

		// eval & eval_literal parsing
		if (preg_match_all("/\{(eval|eval_literal) (.+)\}/sU", $contents, $matches)) {
			foreach ($matches[1] as $k => $type) {
				$code = rtrim(trim($matches[2][$k]), ';');
				if ($type == "eval") {
					$code = $this->_split_exp($code);
				}

				$code .= ";";
				$contents = str_replace($matches[0][$k], $this->_code($code), $contents);
			}
		}

		$contents = str_replace("{else}", $this->_code("}else{"), $contents);
		$contents = str_replace(array("{/foreach}", "{/while}", "{/for}", "{/if}"), trim($this->_code("}")), $contents);
		return $contents;
	}

	/** Parse variables */
	function _parse_variables($contents) {
		return $this->_replace_tags($contents, false);
	}

	/**
	 * Rewrite {tags} into php, escaping each by the markup context it sits in.
	 *
	 * $constants_only restricts the pass to {_CONSTANT} tags. Constants are
	 * replaced earlier than variables, before <script type="text/template">
	 * bodies are stashed, so they are still substituted inside a literal block
	 * where a variable deliberately is not.
	 */
	function _replace_tags($contents, $constants_only) {
		$matches = array();
		// [a-zA-Z\_\$\"\'\[\]\ ]
		if (!preg_match_all("/\{([^\{]+?)\}/s", $contents, $matches, PREG_OFFSET_CAPTURE)) {
			return $contents;
		}

		$map = $this->_html_map($contents);
		// Each tag is replaced where it stands, so two occurrences of the same
		// name in different markup contexts escape differently. The output is
		// assembled in one pass, which keeps the cost linear in the template
		// size.
		$output = "";
		$copied = 0;
		foreach ($matches[0] as $k => $match) {
			list($tag, $offset) = $match;
			$v = $matches[1][$k][0];
			if (strstr($v, "\n") !== false || $v[0] == " ") {
				continue;
			}

			$context = $this->_context_at($map, $offset);
			if ($context === self::CONTEXT_PHP) {
				// generated code, not a template tag
				continue;
			}

			$modifiers = explode("|", $v);
			if ($constants_only && !$this->_is_constant($modifiers[0])) {
				continue;
			}

			$output .= substr($contents, $copied, $offset - $copied);
			$output .= $this->_code($this->_variable_code($v, $context));
			$copied = $offset + strlen($tag);
		}

		return $output . substr($contents, $copied);
	}

	/** Build the php echo statement for a single {variable|modifier} tag */
	function _variable_code($v, $context) {
		$modifiers = explode("|", $v);
		$v = $modifiers[0];
		if ($this->_is_constant($v)) {
			// {_CONSTANT} echoes the php constant by name. Its value is whatever
			// define() was given, so it escapes like any other value.
			$code = $v;
		} else {
			if ($v[0] != '$' && !in_array($v[0], array("'", '"'))) {
				// shorthand variables {v}
				$v = '$' . $v;
			}

			$code = $this->_split_exp($v);
		}

		$escape = $this->_escape_context($context);
		$count = count($modifiers);
		// index 0 is the expression, the rest are modifiers in the order written
		for ($i = 1; $i < $count; $i++) {
			$modifier = $modifiers[$i];
			switch ($modifier) {

				case "raw":
				case "unescape":
					$escape = false;
					break;
				case "escape":
					$escape = true;
					break;
				case "toupper":
					$code = "strtoupper(" . $code . ")";
					break;
				case "tolower":
					$code = "strtolower(" . $code . ")";
					break;
				default:
					$code = $modifier . "(" . $code . ")";
			}
		}

		if ($escape) {
			// escaping is the outermost step, whatever order the modifiers came in
			$code = "htmlspecialchars(" . $code . ", ENT_QUOTES)";
		}

		return "echo " . $code . ";";
	}

	/** Whether a variable printed in this markup context is escaped by default */
	function _escape_context($context) {
		if (!$this->_escape) {
			return false;
		}

		// Script and style bodies are not markup. Escaping there would corrupt
		// the javascript or css rather than protect it.
		return $context !== self::CONTEXT_RAW;
	}

	/**
	 * Scan the template and record where the markup context changes.
	 *
	 * The result is an ascending list of array($offset, $context) pairs, each
	 * marking the first byte covered by that context.
	 */
	function _html_map($contents) {
		$map = array(array(0, self::CONTEXT_TEXT));
		$state = self::CONTEXT_TEXT;
		$resume = self::CONTEXT_TEXT;
		$rawtag = "";
		$matches = array();

		// Every delimiter that can change the context, in document order. A
		// {template tag} is matched too, so that quotes inside an expression
		// are skipped whole rather than read as attribute delimiters.
		if (!preg_match_all("/\{[^\{]+?\}|\<\?php|\?\>|\<!--|--\>|\<(\/?)([a-zA-Z][a-zA-Z0-9:_\-]*)|\>|\"|'/s", $contents, $matches, PREG_OFFSET_CAPTURE)) {
			return $map;
		}

		foreach ($matches[0] as $k => $match) {
			list($token, $offset) = $match;
			if (substr($token, 0, 1) == "{") {
				continue;
			}

			switch ($state) {

				case self::CONTEXT_PHP:
					if ($token == "?" . ">") {
						$state = $resume;
						$map[] = array($offset + 2, $state);
					}

					break;

				case self::CONTEXT_COMMENT:
					if ($token == "--" . ">") {
						$state = self::CONTEXT_TEXT;
						$map[] = array($offset + 3, $state);
					}

					break;

				case self::CONTEXT_RAW:
					if ($token == $this->_tag_php_open) {
						$resume = $state;
						$state = self::CONTEXT_PHP;
						$map[] = array($offset, $state);
						break;
					}

					// only the matching close tag ends a script or style body
					if ($matches[1][$k][0] == "/" && strtolower($matches[2][$k][0]) == $rawtag) {
						$rawtag = "";
						$state = self::CONTEXT_TAG;
						$map[] = array($offset, $state);
					}

					break;

				case self::CONTEXT_ATTRIBUTE_DQ:
				case self::CONTEXT_ATTRIBUTE_SQ:
					if ($token == $this->_tag_php_open) {
						$resume = $state;
						$state = self::CONTEXT_PHP;
						$map[] = array($offset, $state);
						break;
					}

					$quote = ($state === self::CONTEXT_ATTRIBUTE_DQ) ? '"' : "'";
					if ($token == $quote) {
						$state = self::CONTEXT_TAG;
						$map[] = array($offset, $state);
					}

					break;

				case self::CONTEXT_TAG:
					if ($token == $this->_tag_php_open) {
						$resume = $state;
						$state = self::CONTEXT_PHP;
						$map[] = array($offset, $state);
						break;
					}

					if ($token == '"' || $token == "'") {
						$state = ($token == '"') ? self::CONTEXT_ATTRIBUTE_DQ : self::CONTEXT_ATTRIBUTE_SQ;
						$map[] = array($offset + 1, $state);
						break;
					}

					if ($token == ">") {
						$state = ($rawtag != "") ? self::CONTEXT_RAW : self::CONTEXT_TEXT;
						$map[] = array($offset + 1, $state);
					}

					break;
				default:
					if ($token == $this->_tag_php_open) {
						$resume = self::CONTEXT_TEXT;
						$state = self::CONTEXT_PHP;
						$map[] = array($offset, $state);
						break;
					}

					if ($token == "<!--") {
						$state = self::CONTEXT_COMMENT;
						$map[] = array($offset, $state);
						break;
					}

					if ($matches[2][$k][0] !== "") {
						$tag = strtolower($matches[2][$k][0]);
						// an opening script or style tag turns its body into
						// raw text, which the closing > below switches to
						$rawtag = ($matches[1][$k][0] != "/" && ($tag == "script" || $tag == "style")) ? $tag : "";
						$state = self::CONTEXT_TAG;
						$map[] = array($offset, $state);
					}

					break;
			}
		}

		return $map;
	}

	/**
	 * Look up the markup context covering an offset in a map from _html_map().
	 *
	 * The map is ascending, so this binary searches for the last breakpoint at
	 * or before the offset. A template holds as many breakpoints as it does
	 * markup, and every printed variable needs a lookup.
	 */
	function _context_at($map, $offset) {
		$context = self::CONTEXT_TEXT;
		$low = 0;
		$high = count($map) - 1;
		while ($low <= $high) {
			$mid = (int)(($low + $high) / 2);
			if ($map[$mid][0] <= $offset) {
				$context = $map[$mid][1];
				$low = $mid + 1;
			} else {
				$high = $mid - 1;
			}
		}

		return $context;
	}

	/** Split up variables from a php expression and replace them with actual variable locations */
	function _split_exp($exp) {
		$code = str_replace(".", "__1", "<" . "?php if (" . $exp . ") { ?" . ">");
		$tokens = token_get_all($code);
		$objects = array();
		$variable = null;
		$variables = array();
		$variable_continues = false;
		foreach ($tokens as $k => $v) {
			if (is_array($v)) {
				if ($v[0] == T_OBJECT_OPERATOR) {
					$variable_continues = false;
					$objects[] = $variable;
				}

				if ($v[0] == T_VARIABLE) {
					if (!$variable_continues && isset($variable) && !in_array($variable, $objects)) {
						$variables[] = $variable;
					}

					$variable = $variable_continues ? $variable . $v[1] : $v[1];
					if (strstr($variable, "__1") !== false) {
						$variable = str_replace("__1", ".", $variable);
						$variable_continues = false;
						if (substr($variable, -1) == ".") {
							$variable_continues = true;
						}
					} elseif ($variable_continues) {
						$variable_continues = false;
					}
				}

				$v[0] = token_name($v[0]);
				$tokens[$k] = $v;
			}
		}

		if (isset($variable) && !in_array($variable, $variables) && !in_array($variable, $objects)) {
			$variables[] = $variable;
		}

		// globalize objects. The isset() guard keeps this working on runtimes
		// without a $GLOBALS superglobal (phpscript): there the object is simply
		// treated as a template variable instead.
		foreach ($objects as $object) {
			$name = substr($object, 1);
			if ($object != '$this' && isset($GLOBALS[$name]) && is_object($GLOBALS[$name])) {
				$this->_global_variables[] = $object;
			} else {
				$variables[] = $object;
			}
		}

		// closure to sort vars by length and alphabetically
		usort($variables, function($a, $b) {
			if (strlen($a) == strlen($b)) {
				if ($a == $b) {
					return 0;
				}

				return ($a < $b) ? 1 : -1;
			}

			return (strlen($a) < strlen($b)) ? 1 : -1;
		});

		foreach ($variables as $var) {
			if ($var != '$this') {
				$exp = str_replace($var, $this->_get_var($var), $exp);
			}
		}

		return $exp;
	}

	/** Helper function for replacing tags into actual variable locations */
	function _get_var($var) {
		$left_modifier = substr($var, 1); // remove $
		$retval = $var;
		if ($var[0] != '"' && $var[0] != "'") {
			$retval = '$_v';
			if (strstr($left_modifier, '.') !== false) {
				// we have ourselves a table index
				$table_indices = explode('.', $left_modifier);
				foreach ($table_indices as $v) {
					$retval .= (($v[0] == '$') ? "[" . $this->_get_var($v) . "]" : "['" . $v . "']");
				}
			} else {
				$retval .= (($left_modifier[0] == '$') ? "[" . $this->_get_var($left_modifier) . "]" : "['" . $left_modifier . "']");
			}
		}

		return $retval;
	}

	/** Helper function for php code shorthand syntax, optimizing compiler size */
	function _code($s) {
		return $this->_tag_php_open . " " . $s . $this->_tag_php_close;
	}
}
