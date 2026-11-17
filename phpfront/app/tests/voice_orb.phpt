name: voice orb
description: orb.json gives every state two actions, and a saved note leaves the stage and back on screen alone.
---
<?php

require "ext/autoload.php";

$orb = json_decode(file_get_contents("theme/assets/voice/orb.json"), true);
$tpl = file_get_contents("theme/templates/components/voice.tpl");
$icons = file_get_contents("theme/templates/components/icons.tpl");
$js = file_get_contents("theme/assets/voice/voice.js");

// shown is the actions a state puts on screen, as id:action and in order.
function shown($view) {
	$out = array();

	foreach ($view["actions"] as $action) {
		$out[] = $action["id"] . ":" . $action["action"];
	}

	return implode(" ", $out);
}

// back names the icon the second action wears, which is trash only where there
// is a recording to throw away.
function back($view) {
	$action = $view["actions"][1];

	return isset($action["icon"]) ? $action["icon"] : "";
}

echo count($orb["states"]), "\n";
echo implode(",", array_map("count", array_column($orb["states"], "actions"))), "\n";

foreach ($orb["states"] as $name => $view) {
	echo trim($name . " " . shown($view)), "\n";
}

echo "question-card ", implode(" ", array_column($orb["states"]["question"]["card"], "id")), "\n";

// A voice note: the states the window passes through, in order, each with the
// two actions it leaves on screen.
foreach (array("idle", "recording", "saving", "saved") as $name) {
	echo "note ", $name, " ", shown($orb["states"][$name]), "\n";
}

foreach (array("idle", "recording", "paused", "saved") as $name) {
	echo $name, " back:", back($orb["states"][$name]), "\n";
}

// The line the user reads after leaving a note: the transcript is above it in
// voice-said and this is what the note became.
echo "saved ", $orb["states"]["saved"]["status"], "\n";

// Every element the map names is in the template, and the buttons it replaced
// are gone: the stage is the only record and send control.
$elements = true;

foreach ($orb["elements"] as $name => $element) {
	if (strpos($tpl, 'id="' . $element["id"] . '"') === false) {
		$elements = false;
		echo "missing ", $name, "\n";
	}
}

echo "elements ", var_export($elements, true), "\n";
echo "removed ", var_export(
	strpos($tpl, 'id="voice-send"') === false && strpos($tpl, 'id="voice-replay"') === false,
	true
), "\n";

// One button under the stage, the back button.
preg_match('/<div class="voice-acts">(.*?)<\/div>/s', $tpl, $acts);
echo "acts ", substr_count($acts[1], "<button"), "\n";

$wanted = array();

foreach ($orb["states"] as $view) {
	foreach ($view["actions"] as $action) {
		if (isset($action["icon"])) {
			$wanted[$action["icon"]] = true;
		}
	}
}

$missing = array();

foreach (array_keys($wanted) as $name) {
	if (strpos($icons, 'id="i-' . $name . '"') === false) {
		$missing[] = $name;
	}
}

echo "icons ", $missing ? implode(",", $missing) : "none", "\n";

$allowed = array("record", "begin", "stop", "send", "back", "none");
$bad = array();

foreach ($orb["states"] as $name => $view) {
	foreach ($view["actions"] as $action) {
		if (!in_array($action["action"], $allowed, true)) {
			$bad[] = $name . ":" . $action["action"];
		}
	}
}

echo "actions ", $bad ? implode(",", $bad) : "none", "\n";

// The script reads the map and writes the saved status the map only starts.
echo "map ", var_export(strpos($js, "orb.json") !== false && strpos($js, "applyView") !== false, true), "\n";
echo "saved-status ", var_export(
	strpos($js, "Transcript saved") !== false && strpos($js, "thread created") !== false,
	true
), "\n";
---
9
2,2,2,2,2,2,2,2,0
idle stage:record back:back
listening stage:begin back:back
recording stage:stop back:back
paused stage:send back:back
saving stage:none back:back
saved stage:record back:back
answered stage:record back:back
failed stage:record back:back
question
question-card quizCancel quizSubmit
note idle stage:record back:back
note recording stage:stop back:back
note saving stage:none back:back
note saved stage:record back:back
idle back:back
recording back:trash
paused back:trash
saved back:back
saved Transcript saved
elements true
removed true
acts 1
icons none
actions none
map true
saved-status true
