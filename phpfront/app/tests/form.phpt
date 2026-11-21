name: form
description: Task\Form reads the ```ask fence of a note's body - where each one is, the field a bad declaration is in, which fields the defaults show, the prompt it fills and the markup the theme draws for each component and each section.
---
<?php

require "ext/autoload.php";

// A fence with a choice that reveals a field, which is the form the Go
// page's own test renders: the markup below is what both front ends emit.
$deploy = "```ask\n" .
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
echo count(Task\Form::blocks("plain text and a ```yaml\nkey: value\n``` fence that is no ask")), "\n";

// A fence the dialect refuses carries the message and no form, which the
// renderer draws in the form's place. Every message names the field.
$refused = [
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
	"questions:\n  - {id: a, title: A, kind: qcm_free_text, options: [{id: x, label: X}]}\n",
	"questions:\n  - {id: a, title: A, kind: ranking}\n",
	"fields:\n  - {key: a, component: text, section: nope}\n",
	"sections: [{id: s}, {id: s}]\nfields:\n  - {key: a, component: text}\n",
	"fields: [{key: a, component: text}]\nquestions: [{id: b, title: B, kind: free_text}]\n",
];

foreach ($refused as $yaml) {
	$read = Task\Form::parse($yaml);

	echo $read["err"], "\n";
}

// The visible rule: a field shows where the field its when names is shown
// and holds that value, and a field under a hidden one stays hidden.
$chain = Task\Form::parse("fields:\n  - {key: a, component: choice, options: [{value: \"1\"}, {value: \"2\"}]}\n  - {key: b, component: choice, when: a=1, options: [{value: \"9\"}]}\n  - {key: c, component: text, when: b=9}\nsubmit: {prompt: \"{{a}}{{b}}{{c}}\"}\n");
$form = $chain["form"];

echo implode(",", array_keys(Task\Form::visible($form, []))), "\n";
echo implode(",", array_keys(Task\Form::visible($form, ["a" => "1"]))), "\n";
echo implode(",", array_keys(Task\Form::visible($form, ["a" => "1", "b" => "9"]))), "\n";
echo implode(",", array_keys(Task\Form::visible($form, ["a" => "2", "b" => "9"]))), "\n";

// The prompt takes the submitted values, "" for a field with none.
echo "[", Task\Form::render($form, ["a" => "1", "b" => "9", "c" => "go"]), "]\n";
echo "[", Task\Form::render($form, []), "]\n";

// The markup of the fence: one form, the button that starts it, and a
// field carrying its key, its component and its when.
echo Task\Form::html($blocks[0], null), "\n";

// A field the defaults show is not hidden, which is the visible rule drawn.
$shown = Task\Form::blocks(str_replace("    component: choice\n", "    component: choice\n    default: prod\n", $deploy));
$drawn = Task\Form::block($shown[0], null);

echo $drawn["fields"][0]["options"][0]["selected"] ? "selected" : "-", " ", $drawn["fields"][1]["hidden"] ? "hidden" : "shown", "\n";

// A component each, the pickers among them: the control the theme has for
// it, with no source to fill one from.
$every = "```ask\nfields:\n" .
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

// A questionnaire: sections, a question a pick reveals a reason for and a
// checklist, with the fields the ungrouped ones after the sections.
$ask = "```ask\n" .
	"id: release\n" .
	"project: malta\n" .
	"title: Release\n" .
	"subtitle: Two decisions\n" .
	"context: The tests are green.\n" .
	"prerequisites:\n" .
	"  - the image is pushed\n" .
	"sections:\n" .
	"  - {id: rollout, title: Rollout, description: How it ships}\n" .
	"questions:\n" .
	"  - id: target\n" .
	"    title: Where\n" .
	"    kind: qcm_free_text\n" .
	"    section: rollout\n" .
	"    required: true\n" .
	"    options:\n" .
	"      - {id: staging, label: Staging, rationale: Rehearsal, recommended: true, default_checked: true}\n" .
	"      - {id: prod, label: Production}\n" .
	"    free_text: true\n" .
	"    free_text_label: Why\n" .
	"  - {id: scope, title: What ships, kind: boolean_checklist, section: rollout, options: [{id: api, label: The API}, {id: ui, label: The UI}]}\n" .
	"  - {id: notes, title: Anything else, kind: free_text}\n" .
	"submit: {prompt: \"Ship {{target}}\"}\n```";
$q = Task\Form::blocks($ask);
$qform = Task\Form::block($q[0], null, "01hz", 0);

echo $qform["title"], "|", $qform["subtitle"], "|", $qform["context"], "|", implode(",", $qform["prerequisites"]), "|", $qform["note"], "|", $qform["index"], "\n";
echo count($qform["sections"]), " ", $qform["sections"][0]["title"], " ", count($qform["sections"][0]["fields"]), "\n";
echo $qform["fields"][0]["component"], " ", $qform["fields"][0]["kind"], " ", $qform["fields"][1]["component"], " ", $qform["fields"][2]["component"], "\n";
echo Task\Form::html($q[0], null, "01hz", 0), "\n";

// A source names where a field's options come from: the one it declares,
// else the picker component it is. The roles are the theme's own and need
// no reader; the rest are the page's, and job and key are wired nowhere.
echo Task\Form::sourceOf(["source" => "role", "component" => "choice"]), "\n";
echo Task\Form::sourceOf(["source" => "", "component" => "workspace"]), "\n";
echo "[", Task\Form::sourceOf(["source" => "", "component" => "text"]), "]\n";
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
ask: it declares nothing
ask: field "env": it names no component
ask: field 1: it has no key
ask: field 1: key "Env" is no [a-z0-9_-] word
ask: field "a": no component "nosuch"; one of text, textarea, number, bool, hidden, choice, multi, checklist, folder, job, role, user, key, workspace
ask: field "a": options are for choice, multi and checklist, not text
ask: field "a": a choice needs options or a source
ask: field "a": source "role" supplies the options, so it declares none
ask: field "a": option 1: it has no value
ask: field "a": option "x" twice
ask: field "a": declared twice
ask: field "a": option "x": next names no field "nope"
ask: field "a": when "bare" is no <key>=<value>
ask: field "a": when names no field "nope"
ask: field "b": when "y" is no option of field "a"
ask: submit: prompt {{nope}} names no field "nope"
ask: field "a": a qcm_free_text needs two options
ask: field "a": no kind "ranking"; one of free_text, qcm_free_text, boolean_checklist
ask: field "a": section "nope" is not declared
ask: section "s": declared twice
ask: it declares fields and questions
a
a,b
a,b,c
a
[19go]
[]
<form class="note-form run-form" data-run-form data-note="" data-block="0" data-folder="infra" data-job="deploy" data-prompt="Deploy {{env}}, ticket {{ticket}}.">
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
<div class="form-note"><textarea name="general_note" rows="2" placeholder="Anything else? (optional)" aria-label="A note with the answers"></textarea></div>
<div class="form-foot"><span class="faint form-runs">Runs infra/deploy</span><span class="faint form-state"></span><button class="btn btn-primary" type="submit">Deploy it</button></div>
</form>
selected shown
<form class="note-form run-form" data-run-form data-note="" data-block="0" data-prompt="go {{name}}">
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
<div class="checks"><label class="check"><input type="checkbox" name="tags" value="a" data-type="set"> <span>a</span></label><label class="check"><input type="checkbox" name="tags" value="b" data-type="set"> <span>Bee</span></label></div>
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
<div class="form-note"><textarea name="general_note" rows="2" placeholder="Anything else? (optional)" aria-label="A note with the answers"></textarea></div>
<div class="form-foot"><span class="faint form-state"></span><button class="btn btn-primary" type="submit">Run</button></div>
</form>
Release|Two decisions|The tests are green.|the image is pushed|01hz|0
2 Rollout 2
choice qcm_free_text checklist textarea
<form class="note-form run-form" data-run-form data-note="01hz" data-block="0" data-prompt="Ship {{target}}">
<p class="form-title">Release</p>
<p class="form-subtitle faint">Two decisions</p>
<div class="form-context prose">The tests are green.</div>
<ul class="form-prerequisites faint"><li>the image is pushed</li></ul>
<h3 class="form-section">Rollout</h3>
<p class="faint form-section-note">How it ships</p>
<div class="field" data-key="target" data-component="choice" data-kind="qcm_free_text" data-required>
<span class="label">Where</span>
<div class="checks"><label class="check"><input type="radio" name="target" value="staging" checked> <span>Staging</span> <span class="badge signal">recommended</span><small class="faint">Rehearsal</small></label><label class="check"><input type="radio" name="target" value="prod"> <span>Production</span></label></div>
<input class="free-text" type="text" name="target.text" data-free-text maxlength="32768" placeholder="" aria-label="Why">
</div>
<div class="field" data-key="scope" data-component="checklist" data-kind="boolean_checklist">
<span class="label">What ships</span>
<div class="checks"><label class="check"><input type="checkbox" name="scope" value="api" data-type="set"> <span>The API</span></label><label class="check"><input type="checkbox" name="scope" value="ui" data-type="set"> <span>The UI</span></label></div>
</div>
<div class="field" data-key="notes" data-component="textarea" data-kind="free_text">
<span class="label">Anything else</span>
<textarea name="notes" rows="4" spellcheck="false" aria-label="Anything else"></textarea>
</div>
<div class="form-note"><textarea name="general_note" rows="2" placeholder="Anything else? (optional)" aria-label="A note with the answers"></textarea></div>
<div class="form-foot"><span class="faint form-state"></span><button class="btn btn-primary" type="submit">Run</button></div>
</form>
role
workspace
[]
admin,runner,viewer
000
