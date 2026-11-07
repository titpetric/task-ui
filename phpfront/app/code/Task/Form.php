<?php

namespace Task;

// Form is the ```form fence of a note's body: where it is, the form its yaml
// declares, and the markup the theme draws for it.
//
// It is the port of the Go form package, and the two are read as one pair: the
// dialect, the field a bad declaration is named in, which fields a submission
// shows and the markup are the same on both sides, so a reply reads the same
// whichever front end served it. A message that differs is a bug in one of
// them.
//
// The yaml is decoded by the runtime and the rest is held here, because no
// binding crosses a parsed form: see docs/phpfront.md.
class Form {
	/** COMPONENTS are the inputs the dialect names; a field names one of them. */
	const COMPONENTS = array(
		"text",
		"textarea",
		"number",
		"bool",
		"hidden",
		"choice",
		"multi",
		"folder",
		"job",
		"role",
		"user",
		"key",
		"workspace",
	);

	/** PICKERS are the components that take their options from a source of their own name. */
	const PICKERS = array("folder", "job", "role", "user", "key", "workspace");

	/** TRUTHS are the defaults a bool field reads as a yes. */
	const TRUTHS = array("1", "t", "y", "on", "yes", "true");

	/** LABEL is the button's label a form that names none takes. */
	const LABEL = "Run";

	/** TICKS is the least backticks a fence opens with, and WORD the first word after them. */
	const TICKS = 3;
	const WORD = "form";

	/** KEYS are the keys each level of the dialect takes, and nothing else. */
	const FORM_KEYS = array("title", "table", "fields", "submit");
	const FIELD_KEYS = array(
		"key",
		"label",
		"component",
		"help",
		"placeholder",
		"default",
		"source",
		"when",
		"required",
		"options",
	);
	const OPTION_KEYS = array("value", "label", "next");
	const SUBMIT_KEYS = array("label", "job", "folder", "dir", "prompt");

	/** KEY_PATTERN is the word a field's key is, and PROMPT_PATTERN a placeholder of a prompt. */
	const KEY_PATTERN = '/^[a-z0-9_-]+$/';
	const PROMPT_PATTERN = '/\{\{([^{}]*)\}\}/';

	/**
	 * blocks is every form fence of a body in order: the offset it opens at,
	 * the offset just past its closing backticks, the form it declares and
	 * what the dialect refused it for.
	 */
	static function blocks($body) {
		$out = array();
		$length = strlen($body);
		$pos = 0;

		while ($pos < $length) {
			$line = self::lineAt($body, $pos);
			$fence = self::fenceOf($line["text"]);

			if ($fence === null) {
				$pos = $line["next"];
				continue;
			}

			$block = self::scan($body, $fence, $pos, $line["next"]);

			if ($fence["form"]) {
				$read = self::parse($block["body"]);
				$block["form"] = $read["form"];
				$block["err"] = $read["err"];
				$out[] = $block;
			}

			$pos = $block["after"];
		}

		return $out;
	}

	/** parse is the form a yaml document declares, or null with the message naming the field it refuses. */
	static function parse($yaml) {
		$document = null;

		try {
			$document = yaml_decode($yaml);
		} catch (\Throwable $e) {
			return array("form" => null, "err" => "form: " . $e->getMessage());
		}

		if ($document === null) {
			return array("form" => null, "err" => "form: it declares nothing");
		}

		if (!is_array($document)) {
			return array("form" => null, "err" => "form: it declares nothing");
		}

		$form = self::shape($document);
		$err = self::validate($document, $form);

		if ($err !== "") {
			return array("form" => null, "err" => $err);
		}

		return array("form" => $form, "err" => "");
	}

	/** shape is the document read into the dialect's own keys, each one filled. */
	static function shape($document) {
		$submit = Row::items($document, "submit");
		$fields = array();

		foreach (Row::items($document, "fields") as $declared) {
			$fields[] = self::shapeField($declared);
		}

		return array(
			"title" => self::scalar($document, "title"),
			"table" => self::scalar($document, "table"),
			"fields" => $fields,
			"submit" => array(
				"label" => self::scalar($submit, "label") === "" ? self::LABEL : self::scalar($submit, "label"),
				"job" => self::scalar($submit, "job"),
				"folder" => self::scalar($submit, "folder"),
				"dir" => self::scalar($submit, "dir"),
				"prompt" => self::scalar($submit, "prompt"),
			),
		);
	}

	/** shapeField is one declared field read into the dialect's own keys. */
	static function shapeField($declared) {
		$options = array();

		foreach (Row::items($declared, "options") as $option) {
			$options[] = array(
				"value" => self::scalar($option, "value"),
				"label" => self::scalar($option, "label"),
				"next" => self::scalar($option, "next"),
			);
		}

		return array(
			"key" => self::scalar($declared, "key"),
			"label" => self::scalar($declared, "label"),
			"component" => self::scalar($declared, "component"),
			"help" => self::scalar($declared, "help"),
			"placeholder" => self::scalar($declared, "placeholder"),
			"default" => self::scalar($declared, "default"),
			"source" => self::scalar($declared, "source"),
			"when" => self::scalar($declared, "when"),
			"required" => Row::flag($declared, "required"),
			"options" => $options,
		);
	}

	/** validate is the first thing the dialect refuses, the references and the prompt last. */
	static function validate($document, $form) {
		$unknown = self::unknown($document, $form);

		if ($unknown !== "") {
			return $unknown;
		}

		if (count($form["fields"]) === 0) {
			return "form: it declares no field";
		}

		$keys = array();

		foreach ($form["fields"] as $index => $field) {
			$err = self::validateField($field, $index);

			if ($err !== "") {
				return $err;
			}

			if (isset($keys[$field["key"]])) {
				return "form: field " . self::quote($field["key"]) . ": declared twice";
			}

			$keys[$field["key"]] = $field;
		}

		foreach ($form["fields"] as $field) {
			$err = self::validateRefs($field, $keys);

			if ($err !== "") {
				return $err;
			}
		}

		return self::validatePrompt($form, $keys);
	}

	/** validateField is what a field itself gets wrong, its index naming one with no usable key. */
	static function validateField($field, $index) {
		if ($field["key"] === "") {
			return "form: field " . ($index + 1) . ": it has no key";
		}

		if (preg_match(self::KEY_PATTERN, $field["key"]) !== 1) {
			return "form: field " . ($index + 1) . ": key " . self::quote($field["key"]) . " is no [a-z0-9_-] word";
		}

		if ($field["component"] === "") {
			return "form: field " . self::quote($field["key"]) . ": it names no component";
		}

		if (!in_array($field["component"], self::COMPONENTS, true)) {
			return "form: field " . self::quote($field["key"]) . ": no component " . self::quote($field["component"]) . "; one of " . implode(", ", self::COMPONENTS);
		}

		return self::validateOptions($field);
	}

	/** validateOptions is the options a component may not hold, a source that supplies them, and a value twice. */
	static function validateOptions($field) {
		$picks = $field["component"] === "choice" || $field["component"] === "multi";
		$count = count($field["options"]);

		if ($count > 0 && !$picks) {
			return "form: field " . self::quote($field["key"]) . ": options are for choice and multi, not " . $field["component"];
		}

		if ($count > 0 && $field["source"] !== "") {
			return "form: field " . self::quote($field["key"]) . ": source " . self::quote($field["source"]) . " supplies the options, so it declares none";
		}

		if ($picks && $count === 0 && $field["source"] === "") {
			return "form: field " . self::quote($field["key"]) . ": a " . $field["component"] . " needs options or a source";
		}

		$seen = array();

		foreach ($field["options"] as $index => $option) {
			if ($option["value"] === "") {
				return "form: field " . self::quote($field["key"]) . ": option " . ($index + 1) . ": it has no value";
			}

			if (isset($seen[$option["value"]])) {
				return "form: field " . self::quote($field["key"]) . ": option " . self::quote($option["value"]) . " twice";
			}

			$seen[$option["value"]] = true;
		}

		return "";
	}

	/** validateRefs is a next or a when of the field that names no field of the form. */
	static function validateRefs($field, $keys) {
		foreach ($field["options"] as $option) {
			if ($option["next"] !== "" && !isset($keys[$option["next"]])) {
				return "form: field " . self::quote($field["key"]) . ": option " . self::quote($option["value"]) . ": next names no field " . self::quote($option["next"]);
			}
		}

		if ($field["when"] === "") {
			return "";
		}

		$cut = self::cut($field["when"]);

		if (!$cut["ok"] || $cut["key"] === "" || $cut["value"] === "") {
			return "form: field " . self::quote($field["key"]) . ": when " . self::quote($field["when"]) . " is no <key>=<value>";
		}

		if (!isset($keys[$cut["key"]])) {
			return "form: field " . self::quote($field["key"]) . ": when names no field " . self::quote($cut["key"]);
		}

		$named = $keys[$cut["key"]];

		if (count($named["options"]) > 0 && !self::hasOption($named, $cut["value"])) {
			return "form: field " . self::quote($field["key"]) . ": when " . self::quote($cut["value"]) . " is no option of field " . self::quote($cut["key"]);
		}

		return "";
	}

	/** validatePrompt is the first placeholder of the prompt that names no field. */
	static function validatePrompt($form, $keys) {
		$count = preg_match_all(self::PROMPT_PATTERN, $form["submit"]["prompt"], $found);

		if ($count < 1) {
			return "";
		}

		foreach ($found[1] as $index => $name) {
			$name = trim($name);

			if (!isset($keys[$name])) {
				return "form: submit: prompt " . $found[0][$index] . " names no field " . self::quote($name);
			}
		}

		return "";
	}

	/**
	 * unknown is the first key of the document the dialect does not name.
	 *
	 * The Go parser refuses one through the yaml decoder's own strict mode, so
	 * its message is the decoder's and this one is not. A fence that misspells
	 * a key reads as refused on both sides and the sentence differs; see
	 * docs/phpfront.md.
	 */
	static function unknown($document, $form) {
		$err = self::unknownIn($document, self::FORM_KEYS, "form");

		if ($err !== "") {
			return $err;
		}

		$err = self::unknownIn(Row::items($document, "submit"), self::SUBMIT_KEYS, "submit");

		if ($err !== "") {
			return $err;
		}

		foreach (Row::items($document, "fields") as $index => $declared) {
			$err = self::unknownIn($declared, self::FIELD_KEYS, "field " . ($index + 1));

			if ($err !== "") {
				return $err;
			}

			foreach (Row::items($declared, "options") as $at => $option) {
				$err = self::unknownIn($option, self::OPTION_KEYS, "field " . ($index + 1) . ", option " . ($at + 1));

				if ($err !== "") {
					return $err;
				}
			}
		}

		return "";
	}

	/** unknownIn is the first key of a mapping that is not in the list it takes. */
	static function unknownIn($mapping, $allowed, $where) {
		if (!is_array($mapping)) {
			return "";
		}

		foreach (array_keys($mapping) as $key) {
			if (!in_array($key, $allowed, true)) {
				return "form: " . $where . ": no key " . self::quote((string)$key);
			}
		}

		return "";
	}

	/** visible is the keys the values show; a field whose own when-field is hidden stays hidden. */
	static function visible($form, $values) {
		$shown = array();

		foreach ($form["fields"] as $field) {
			if ($field["when"] !== "") {
				$cut = self::cut($field["when"]);
				$held = isset($values[$cut["key"]]) ? trim($values[$cut["key"]]) : "";

				if (!isset($shown[$cut["key"]]) || $held !== $cut["value"]) {
					continue;
				}
			}

			$shown[$field["key"]] = true;
		}

		return $shown;
	}

	/** render is the prompt with every placeholder replaced by the submitted value, "" for a field with none. */
	static function render($form, $values) {
		$out = $form["submit"]["prompt"];

		foreach (self::placeholders($out) as $match => $name) {
			$out = str_replace($match, isset($values[$name]) ? $values[$name] : "", $out);
		}

		return $out;
	}

	/** placeholders is each placeholder of a prompt as written, by the field it names. */
	static function placeholders($prompt) {
		$out = array();
		$count = preg_match_all(self::PROMPT_PATTERN, $prompt, $found);

		if ($count < 1) {
			return $out;
		}

		foreach ($found[0] as $index => $match) {
			$out[$match] = trim($found[1][$index]);
		}

		return $out;
	}

	/** html is a fence as the body holds it: the form it declares, or the error it failed to parse with. */
	static function html($block, $chrome) {
		$paths = array(View::THEME);
		$tpl = new \MiniTPL\Template($paths);

		$tpl->load("components/form-block.tpl");
		$tpl->assign("data_form", self::block($block, $chrome));

		return $tpl->get();
	}

	/** block is a fence as components/form-block.tpl reads it, the defaults deciding which field is hidden. */
	static function block($block, $chrome) {
		$form = isset($block["form"]) ? $block["form"] : null;

		if ($form === null) {
			return array(
				"err" => Row::text($block, "err"),
				"title" => "",
				"label" => "",
				"folder" => "",
				"dir" => "",
				"job" => "",
				"prompt" => "",
				"fields" => array(),
			);
		}

		$values = array();

		foreach ($form["fields"] as $field) {
			$values[$field["key"]] = $field["default"];
		}

		$shown = self::visible($form, $values);
		$fields = array();

		foreach ($form["fields"] as $field) {
			$fields[] = self::field($field, !isset($shown[$field["key"]]), $chrome);
		}

		return array(
			"err" => "",
			"title" => $form["title"],
			"label" => $form["submit"]["label"],
			"folder" => $form["submit"]["folder"],
			"dir" => $form["submit"]["dir"],
			"job" => $form["submit"]["job"],
			"prompt" => $form["submit"]["prompt"],
			"fields" => $fields,
		);
	}

	/** field is one declared field as the template draws it, its options from the source it names. */
	static function field($field, $hidden, $chrome) {
		$options = array();

		foreach ($field["options"] as $option) {
			$options[] = array(
				"value" => $option["value"],
				"label" => $option["label"] === "" ? $option["value"] : $option["label"],
				"next" => $option["next"],
				"selected" => $option["value"] === $field["default"],
			);
		}

		if (count($options) === 0) {
			$options = self::sourced(self::sourceOf($field), $field["default"], $chrome);
		}

		return array(
			"key" => $field["key"],
			"label" => $field["label"] === "" ? $field["key"] : $field["label"],
			"component" => $field["component"],
			"help" => $field["help"],
			"placeholder" => $field["placeholder"],
			"value" => $field["default"],
			"when" => $field["when"],
			"required" => $field["required"],
			"hidden" => $hidden,
			"checked" => in_array(strtolower(trim($field["default"])), self::TRUTHS, true),
			"options" => $options,
		);
	}

	/** sourceOf names where a field's options come from: the one it declares, else its own picker component. */
	static function sourceOf($field) {
		if ($field["source"] !== "") {
			return $field["source"];
		}

		return in_array($field["component"], self::PICKERS, true) ? $field["component"] : "";
	}

	/** sourced is the options of a source, the one the default names selected. */
	static function sourced($source, $default, $chrome) {
		$out = array();

		foreach (self::source($source, $chrome) as $value) {
			$out[] = array(
				"value" => $value,
				"label" => $value,
				"next" => "",
				"selected" => $value === $default,
			);
		}

		return $out;
	}

	/**
	 * source is the values a source names, the one lookup point of a form's
	 * source: the folders the reader sees, the roles, the users and the
	 * workspaces in the sidebar's order.
	 *
	 * A source nothing here reaches has no options, which draws an empty
	 * picker. job and key are the two: neither is on the page a note is read
	 * from in either front end.
	 */
	static function source($source, $chrome) {
		if ($source === "role") {
			return User::ROLES;
		}

		if ($chrome === null) {
			return array();
		}

		switch ($source) {

			case "folder":
				$out = array();

				foreach ($chrome->folders() as $folder) {
					$out[] = Row::text($folder, "id");
				}

				return $out;
			case "user":
				$names = array_keys($chrome->roles());

				sort($names);

				return $names;
			case "workspace":
				return $chrome->workspaces();
		}

		return array();
	}

	/** hasOption reports whether a value is one of a field's static options. */
	static function hasOption($field, $value) {
		foreach ($field["options"] as $option) {
			if ($option["value"] === $value) {
				return true;
			}
		}

		return false;
	}

	/** cut splits a when on its first =, saying whether it held one. */
	static function cut($when) {
		$at = strpos($when, "=");

		if ($at === false) {
			return array(
				"key" => $when,
				"value" => "",
				"ok" => false,
			);
		}

		return array(
			"key" => substr($when, 0, $at),
			"value" => substr($when, $at + 1),
			"ok" => true,
		);
	}

	/** scalar is a key of a mapping as a string, a yaml truth as the word that spells it. */
	static function scalar($mapping, $key) {
		if (!is_array($mapping) || !isset($mapping[$key])) {
			return "";
		}

		$value = $mapping[$key];

		if (is_bool($value)) {
			return $value ? "true" : "false";
		}

		if (is_string($value) || is_int($value) || is_float($value)) {
			return (string)$value;
		}

		return "";
	}

	/** quote is a value as the Go verb %q writes it, which every message of the dialect names a field with. */
	static function quote($value) {
		return '"' . str_replace(array("\\", '"', "\n", "\r", "\t"), array("\\\\", '\\"', '\\n', '\\r', '\\t'), $value) . '"';
	}

	/** lineAt is the line at an offset without its break, and the offset the next one opens at. */
	static function lineAt($body, $pos) {
		$at = strpos($body, "\n", $pos);

		if ($at === false) {
			return array("text" => rtrim(substr($body, $pos), "\r"), "next" => strlen($body));
		}

		return array("text" => rtrim(substr($body, $pos, $at - $pos), "\r"), "next" => $at + 1);
	}

	/** fenceOf reads a line as a fence opener: its indent, its backticks and whether it says form; null for none. */
	static function fenceOf($line) {
		$parts = self::fenceParts($line);

		if ($parts["ticks"] < self::TICKS || strpos($parts["rest"], "`") !== false) {
			return null;
		}

		return array(
			"indent" => $parts["indent"],
			"ticks" => $parts["ticks"],
			"form" => self::firstWord($parts["rest"]) === self::WORD,
		);
	}

	/** scan is the block a fence opens at start and the offset past its closing line, the body de-indented. */
	static function scan($body, $fence, $start, $pos) {
		$lines = array();
		$length = strlen($body);

		while ($pos < $length) {
			$line = self::lineAt($body, $pos);

			if (self::closes($fence, $line["text"])) {
				return array(
					"start" => $start,
					"end" => $pos + strlen($line["text"]),
					"body" => self::body($lines),
					"after" => $line["next"],
					"form" => null,
					"err" => "",
				);
			}

			$lines[] = self::deindent($line["text"], $fence["indent"]);
			$pos = $line["next"];
		}

		return array(
			"start" => $start,
			"end" => $length,
			"body" => self::body($lines),
			"after" => $length,
			"form" => null,
			"err" => "",
		);
	}

	/** closes reports whether a line closes a fence: as many backticks, and nothing after them. */
	static function closes($fence, $line) {
		$parts = self::fenceParts($line);

		return $parts["ticks"] >= $fence["ticks"] && $parts["rest"] === "";
	}

	/** fenceParts splits a line into its indent, how many backticks follow it and the rest, trimmed. */
	static function fenceParts($line) {
		$ticked = ltrim($line, " \t");
		$rest = ltrim($ticked, "`");

		return array(
			"indent" => substr($line, 0, strlen($line) - strlen($ticked)),
			"ticks" => strlen($ticked) - strlen($rest),
			"rest" => trim($rest),
		);
	}

	/** deindent takes a fence's own indent off one of its lines. */
	static function deindent($line, $indent) {
		if ($indent === "" || strpos($line, $indent) !== 0) {
			return $line;
		}

		return substr($line, strlen($indent));
	}

	/** body is a fence's lines as one document, "" for a fence that holds none. */
	static function body($lines) {
		if (count($lines) === 0) {
			return "";
		}

		return implode("\n", $lines) . "\n";
	}

	/** firstWord is the first word of a string, "" for none. */
	static function firstWord($text) {
		$found = preg_split('/[ \t]/', $text, 2);

		return $found[0];
	}
}
