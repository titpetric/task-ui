name: form
description: Task\Form reads the ```form fence of a note's body - where each one is, the field a bad declaration is in, which fields the defaults show, the prompt it fills and the markup the theme draws for each component.
---
<?php

require "ext/autoload.php";

// A fence with a choice that reveals a field, which is the form the Go
// page's own test renders: the markup below is what both front ends emit.
$deploy = "```form\n" .
	"title: Deploy\n" .
	"fields:\n" .
	"  - key: env\n" .
	"    label: Environment\n" .
	"    component: choice\n" .
	"    help: Where it goes\n" .
	"    required: true\n" .
	"    options:\n" .
	"      - value: prod\n" .
	"        label: Production\n" .
	"        next: ticket\n" .
	"      - value: stage\n" .
	"        label: Staging\n" .
	"  - key: ticket\n" .
	"    label: Change ticket\n" .
	"    component: text\n" .
	"    when: env=prod\n" .
	"    placeholder: OPS-1\n" .
	"submit:\n" .
	"  label: Deploy it\n" .
	"  folder: infra\n" .
	"  job: deploy\n" .
	"  prompt: Deploy {{env}}, ticket {{ticket}}.\n" .
	"```";

// A body with one fence: the offsets name the prose on either side of it,
// so the renderer draws markdown, the form, then markdown.
$body = "Pick one:\n\n" . $deploy . "\n\nIt runs as you.";
$blocks = Task\Form::blocks($body);

echo count($blocks), "\n";
echo "[", substr($body, 0, $blocks[0]["start"]), "]\n";
echo "[", substr($body, $blocks[0]["end"]), "]\n";

// A body with two fences, and one with none.
echo count(Task\Form::blocks($deploy . "\n\nor else\n\n" . $deploy)), "\n";
echo count(Task\Form::blocks("plain text and a ```yaml\nkey: value\n``` fence that is no form")), "\n";

// A fence the dialect refuses carries the message and no form, which the
// renderer draws in the form's place. Every message names the field.
$refused = array(
	"",
	"fields:\n  - key: env\n",
	"fields:\n  - {component: text}\n",
	"fields:\n  - {key: Env, component: text}\n",
	"fields:\n  - {key: a, component: nosuch}\n",
	"fields:\n  - {key: a, component: text, options: [{value: x}]}\n",
	"fields:\n  - {key: a, component: choice}\n",
	"fields:\n  - {key: a, component: choice, source: role, options: [{value: x}]}\n",
	"fields:\n  - {key: a, component: choice, options: [{label: x}]}\n",
	"fields:\n  - {key: a, component: choice, options: [{value: x}, {value: x}]}\n",
	"fields:\n  - {key: a, component: text}\n  - {key: a, component: text}\n",
	"fields:\n  - {key: a, component: choice, options: [{value: x, next: nope}]}\n",
	"fields:\n  - {key: a, component: text, when: bare}\n",
	"fields:\n  - {key: a, component: text, when: nope=1}\n",
	"fields:\n  - {key: a, component: choice, options: [{value: x}]}\n  - {key: b, component: text, when: a=y}\n",
	"fields:\n  - {key: a, component: text}\nsubmit: {prompt: \"{{nope}}\"}\n",
);

foreach ($refused as $yaml) {
	$read = Task\Form::parse($yaml);

	echo $read["err"], "\n";
}

// The visible rule: a field shows where the field its when names is shown
// and holds that value, and a field under a hidden one stays hidden.
$chain = Task\Form::parse("fields:\n  - {key: a, component: choice, options: [{value: \"1\"}, {value: \"2\"}]}\n  - {key: b, component: choice, when: a=1, options: [{value: \"9\"}]}\n  - {key: c, component: text, when: b=9}\nsubmit: {prompt: \"{{a}}{{b}}{{c}}\"}\n");
$form = $chain["form"];

echo implode(",", array_keys(Task\Form::visible($form, array()))), "\n";
echo implode(",", array_keys(Task\Form::visible($form, array("a" => "1")))), "\n";
echo implode(",", array_keys(Task\Form::visible($form, array("a" => "1", "b" => "9")))), "\n";
echo implode(",", array_keys(Task\Form::visible($form, array("a" => "2", "b" => "9")))), "\n";

// The prompt takes the submitted values, "" for a field with none.
echo "[", Task\Form::render($form, array("a" => "1", "b" => "9", "c" => "go")), "]\n";
echo "[", Task\Form::render($form, array()), "]\n";

// The markup of the fence: one form, the button that starts it, and a
// field carrying its key, its component and its when.
echo Task\Form::html($blocks[0], null), "\n";

// A field the defaults show is not hidden, which is the visible rule drawn.
$shown = Task\Form::blocks(str_replace("    component: choice\n", "    component: choice\n    default: prod\n", $deploy));
$drawn = Task\Form::block($shown[0], null);

echo $drawn["fields"][0]["options"][0]["selected"] ? "selected" : "-", " ", $drawn["fields"][1]["hidden"] ? "hidden" : "shown", "\n";

// A component each, the pickers among them: the control the theme has for
// it, with no source to fill one from.
$every = "```form\nfields:\n" .
	"  - {key: name, label: Name, component: text, placeholder: web, required: true, help: The host}\n" .
	"  - {key: why, label: Why, component: textarea, default: because}\n" .
	"  - {key: count, label: Count, component: number, default: \"3\"}\n" .
	"  - {key: force, label: Force, component: bool, default: \"yes\"}\n" .
	"  - {key: quiet, label: Quiet, component: bool}\n" .
	"  - {key: batch, label: Batch, component: hidden, default: b1}\n" .
	"  - {key: env, label: Env, component: choice, options: [{value: prod, label: Production, next: why}, {value: stage}]}\n" .
	"  - {key: tags, label: Tags, component: multi, options: [{value: a}, {value: b, label: Bee}]}\n" .
	"  - {key: place, label: Place, component: folder}\n" .
	"  - {key: who, label: Who, component: role, required: true}\n" .
	"  - {key: owner, component: user}\n" .
	"  - {key: ws, component: workspace}\n" .
	"  - {key: jobs, component: job}\n" .
	"  - {key: sshkey, component: key}\n" .
	"submit: {prompt: \"go {{name}}\"}\n```";
$all = Task\Form::blocks($every);

echo Task\Form::html($all[0], null), "\n";

// A source names where a field's options come from: the one it declares,
// else the picker component it is. The roles are the theme's own and need
// no reader; the rest are the page's, and job and key are wired nowhere.
echo Task\Form::sourceOf(array("source" => "role", "component" => "choice")), "\n";
echo Task\Form::sourceOf(array("source" => "", "component" => "workspace")), "\n";
echo "[", Task\Form::sourceOf(array("source" => "", "component" => "text")), "]\n";
echo implode(",", Task\Form::source("role", null)), "\n";
echo count(Task\Form::source("job", null)), count(Task\Form::source("key", null)), count(Task\Form::source("", null)), "\n";
---
1
[Pick one:

]
[

It runs as you.]
2
0
form: it declares nothing
form: field "env": it names no component
form: field 1: it has no key
form: field 1: key "Env" is no [a-z0-9_-] word
form: field "a": no component "nosuch"; one of text, textarea, number, bool, hidden, choice, multi, folder, job, role, user, key, workspace
form: field "a": options are for choice and multi, not text
form: field "a": a choice needs options or a source
form: field "a": source "role" supplies the options, so it declares none
form: field "a": option 1: it has no value
form: field "a": option "x" twice
form: field "a": declared twice
form: field "a": option "x": next names no field "nope"
form: field "a": when "bare" is no <key>=<value>
form: field "a": when names no field "nope"
form: field "b": when "y" is no option of field "a"
form: submit: prompt {{nope}} names no field "nope"
a
a,b
a,b,c
a
[19go]
[]
<form class="note-form run-form" data-run-form data-folder="infra" data-job="deploy" data-prompt="Deploy {{env}}, ticket {{ticket}}.">
<p class="form-title">Deploy</p>
<div class="field" data-key="env" data-component="choice" data-required>
<span class="label">Environment</span>
<select name="env" aria-label="Environment" required><option value="prod" data-next="ticket">Production</option><option value="stage">Staging</option></select>
<small class="faint">Where it goes</small>
</div>
<div class="field" data-key="ticket" data-component="text" data-when="env=prod" hidden>
<span class="label">Change ticket</span>
<input type="text" name="ticket" value="" aria-label="Change ticket" placeholder="OPS-1">
</div>
<div class="form-foot"><span class="faint form-state"></span><button class="btn btn-primary" type="submit">Deploy it</button></div>
</form>
selected shown
<form class="note-form run-form" data-run-form data-prompt="go {{name}}">
<div class="field" data-key="name" data-component="text" data-required>
<span class="label">Name</span>
<input type="text" name="name" value="" aria-label="Name" placeholder="web" required>
<small class="faint">The host</small>
</div>
<div class="field" data-key="why" data-component="textarea">
<span class="label">Why</span>
<textarea name="why" rows="4" spellcheck="false" aria-label="Why">because</textarea>
</div>
<div class="field" data-key="count" data-component="number">
<span class="label">Count</span>
<input type="number" name="count" value="3" aria-label="Count">
</div>
<div class="field" data-key="force" data-component="bool">
<label class="check"><input type="checkbox" name="force" data-type="toggle" checked> Force</label>
</div>
<div class="field" data-key="quiet" data-component="bool">
<label class="check"><input type="checkbox" name="quiet" data-type="toggle"> Quiet</label>
</div>
<input type="hidden" name="batch" value="b1" data-key="batch" data-component="hidden">
<div class="field" data-key="env" data-component="choice">
<span class="label">Env</span>
<select name="env" aria-label="Env"><option value=""></option><option value="prod" data-next="why">Production</option><option value="stage">stage</option></select>
</div>
<div class="field" data-key="tags" data-component="multi">
<span class="label">Tags</span>
<div class="checks"><label class="check"><input type="checkbox" name="tags" value="a" data-type="set"> a</label><label class="check"><input type="checkbox" name="tags" value="b" data-type="set"> Bee</label></div>
</div>
<div class="field" data-key="place" data-component="folder">
<span class="label">Place</span>
<select name="place" aria-label="Place"><option value=""></option></select>
</div>
<div class="field" data-key="who" data-component="role" data-required>
<span class="label">Who</span>
<select name="who" aria-label="Who" required><option value="admin">admin</option><option value="runner">runner</option><option value="viewer">viewer</option></select>
</div>
<div class="field" data-key="owner" data-component="user">
<span class="label">owner</span>
<select name="owner" aria-label="owner"><option value=""></option></select>
</div>
<div class="field" data-key="ws" data-component="workspace">
<span class="label">ws</span>
<select name="ws" aria-label="ws"><option value=""></option></select>
</div>
<div class="field" data-key="jobs" data-component="job">
<span class="label">jobs</span>
<select name="jobs" aria-label="jobs"><option value=""></option></select>
</div>
<div class="field" data-key="sshkey" data-component="key">
<span class="label">sshkey</span>
<select name="sshkey" aria-label="sshkey"><option value=""></option></select>
</div>
<div class="form-foot"><span class="faint form-state"></span><button class="btn btn-primary" type="submit">Run</button></div>
</form>
role
workspace
[]
admin,runner,viewer
000
