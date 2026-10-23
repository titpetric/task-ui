{{/* actions is the actions bar of a folder page and of a directory browsed
     inside a folder: the git line of the checkout the place is in, then a
     chip per atkins skill that applies there, and a chip of the place's own
     atkins jobs. Every action is a Run of the place's folder, run where the
     job runs; a running one attaches, one the user may not run links to
     its page. At the top right, each behind its permission: Enter Shell
     (shell) and Enter Claude (shell and a Claude token). */}}
{{- define "actions" }}
{{- with .Actions }}{{ if not .Empty }}
{{- $a := . }}
<div class="actionsbar" data-live>
	{{- with .Git }}
	<div class="gitline" title="The checkout {{ if .Checkout }}{{ .Checkout }}{{ else }}at the root{{ end }}{{ with .Upstream }}, tracking {{ . }}{{ end }}">
		<span class="branch"><span class="glyph" aria-hidden="true">⎇</span> {{ if .Branch }}{{ .Branch }}{{ else }}<span class="faint">detached</span> {{ .Detached }}{{ end }}</span>
		{{- if .Ahead }}<span class="ahead" title="{{ .Ahead }} commits to push">↑{{ .Ahead }}</span>{{ end }}
		{{- if .Behind }}<span class="behind" title="{{ .Behind }} commits to pull">↓{{ .Behind }}</span>{{ end }}
		{{- if .Dirty }}<span class="dirty" title="{{ .Dirty }} files changed or untracked">●{{ .Dirty }}</span>{{ end }}
		{{- if and .Upstream (not .Ahead) (not .Behind) (not .Dirty) }}<span class="clean">up to date</span>{{ end }}
		<span class="where">· {{ if .Checkout }}{{ .Checkout }}{{ else }}the root{{ end }}</span>
	</div>
	{{- end }}
	{{- if or .Shell .Claude }}
	<div class="bar-end">
		{{- with .Shell }}{{ template "enterTerminal" (list $ $a.Place . "Shell" "bash -l") }}{{ end }}
		{{- with .Claude }}{{ template "enterTerminal" (list $ $a.Place . "Claude" "Claude Code") }}{{ end }}
	</div>
	{{- end }}
	{{- if or .Skills .Pipeline }}
	<div class="chips">
		{{- range $s := .Skills }}
		<span class="chip" title="{{ with $s.Title }}{{ . }}{{ else }}{{ $s.ID }}{{ end }}{{ with $s.When }}{{ with .String }}: when {{ . }}{{ end }}{{ end }}{{ if $s.Dir }}; runs in {{ $s.Dir }}/{{ end }}">
			<span class="chip-name">{{ $s.ID }}</span>
			{{- range $i, $j := $s.Jobs }}{{ if $i }}<span class="sep" aria-hidden="true">·</span>{{ end }}{{ template "chipAction" (list $ $a.Place $j ($a.Confirm $s $j)) }}{{ end }}
		</span>
		{{- end }}
		{{- with .Pipeline }}
		<span class="chip" title="The atkins pipeline of {{ $a.Place.Title }}">
			<span class="chip-name">atkins</span>
			{{- range $i, $j := . }}{{ if $i }}<span class="sep" aria-hidden="true">·</span>{{ end }}{{ template "chipAction" (list $ $a.Place $j "") }}{{ end }}
		</span>
		{{- end }}
	</div>
	{{- end }}
</div>
{{- end }}{{ end }}
{{- end }}

{{/* chipAction is one job of a chip: Run, Attach while it runs, or a link
     to its page for a user who may not run it. A job that takes args
     opens a field for them under the chips first. */}}
{{- define "chipAction" }}
{{- $ := index . 0 }}{{ $p := index . 1 }}{{ $r := index . 2 }}{{ $confirm := index . 3 }}
{{- $run := $.RunningRun $r.Key }}
{{- $what := print $p.ID (or (and $r.Dir (print "/" $r.Dir)) "") " / " $r.Name }}
{{- if $run -}}
<a class="chip-act running" href="/runs/{{ $run.ID }}" data-attach="{{ $run.ID }}" title="Attach to {{ $what }}, running"><i class="dot running"></i>{{ $r.Name }}</a>
{{- else if $.CanRun $p $r -}}
<button type="button" class="chip-act" data-run data-folder="{{ $p.ID }}"{{ with $r.Dir }} data-dir="{{ . }}"{{ end }} data-job="{{ $r.ID }}" data-title="{{ $what }}"{{ with $confirm }} data-confirm="{{ . }}" data-confirm-ok="Run {{ $r.Name }}"{{ end }}{{ with $r.ArgsCommand }} data-args="{{ . }}"{{ end }} title="{{ with $r.Description }}{{ . }}{{ else }}Run {{ $what }}{{ end }}{{ with $r.ArgsCommand }}: arguments after {{ . }}{{ end }}">{{ $r.Name }}{{ if $r.TakesArgs }}...{{ end }}</button>
{{- else -}}
<a class="chip-act" href="{{ $.JobHref $p $r }}" title="{{ $what }}: open its page and history">{{ $r.Name }}</a>
{{- end }}
{{- end }}

{{/* enterTerminal is Enter Shell or Enter Claude: a terminal job of the
     place, which opens the run window, or Attach while it runs. */}}
{{- define "enterTerminal" }}
{{- $ := index . 0 }}{{ $p := index . 1 }}{{ $j := index . 2 }}{{ $label := index . 3 }}{{ $what := index . 4 }}
{{- $where := print $p.ID (or (and $j.Dir (print "/" $j.Dir)) "") }}
{{- with $.RunningRun $j.Key }}
		<a class="btn btn-sm shell-enter running" href="/runs/{{ .ID }}" data-attach="{{ .ID }}" title="Attach to {{ $what }} in {{ $where }}/, running"><i class="dot running"></i>{{ $label }}</a>
{{- else }}
		<button type="button" class="btn btn-sm shell-enter" data-run data-folder="{{ $p.ID }}"{{ with $j.Dir }} data-dir="{{ . }}"{{ end }} data-job="{{ $j.ID }}" data-title="{{ $where }} / {{ $j.Name }}" title="{{ $what }} in {{ $where }}/, with the same setup as its jobs"><svg class="ic"><use href="#i-terminal"/></svg>Enter {{ $label }}</button>
{{- end }}
{{- end }}
