{{- /* The control buttons of a folder's jobs, the dashboard's control
     centre's and a folder page's Files list's. */}}

{{/* controljob is a job's button in the control centre: Run, an eye on
     the run in progress, or a #! to its page for a user who may not run
     it. In the other column the button carries the job's name. The
     button and the #! take the colour of the job's last run: ok, bad for
     failed, stopped or killed, never without one. The slot around it says
     what the script needs to draw it again as runs start and end. */}}
{{- define "controljob" }}
{{- $ := index . 0 }}{{ $f := index . 1 }}{{ $r := index . 2 }}{{ $named := index . 3 }}
{{- $run := $.RunningRun $r.Key }}
{{- /* A name two sources share is told apart by its kind, as in the sidebar. */}}
{{- $label := $r.Name }}{{ if ne $r.ID $r.Name }}{{ $label = $r.ID }}{{ end }}
{{- $last := "never" }}{{ with $.LastRun $r.Key }}{{ $last = "bad" }}{{ if eq .State "ok" }}{{ $last = "ok" }}{{ end }}{{ end }}
<span class="cc-slot" data-cc="{{ $r.Key }}" data-folder="{{ $f.ID }}" data-job="{{ $r.ID }}" data-name="{{ $r.Name }}" data-label="{{ if $named }}{{ $label }}{{ end }}" data-last="{{ $last }}"{{ if $.CanRun $f $r }} data-can{{ end }}>
{{- if $run -}}
<a class="cc{{ if $named }} named{{ end }} running" href="/runs/{{ $run.ID }}" data-attach="{{ $run.ID }}" title="Attach to {{ $f.ID }} / {{ $r.Name }}, running"><svg class="ic"><use href="#i-eye"/></svg>{{ if $named }}<span>{{ $label }}</span>{{ end }}</a>
{{- else if $.CanRun $f $r -}}
<button type="button" class="cc{{ if $named }} named{{ end }} {{ $last }}" data-run data-folder="{{ $f.ID }}" data-job="{{ $r.ID }}" data-title="{{ $f.ID }} / {{ $r.Name }}" title="Run {{ $f.ID }} / {{ $r.Name }}"><svg class="ic"><use href="#i-play"/></svg>{{ if $named }}<span>{{ $label }}</span>{{ end }}</button>
{{- else -}}
<a class="cc{{ if $named }} named{{ end }} idle {{ $last }}" href="{{ $.Href $f.ID }}/{{ $r.ID }}" title="{{ $f.ID }} / {{ $r.Name }}: your role may not run it"><i class="shebang" aria-hidden="true">#!</i>{{ if $named }}<span>{{ $label }}</span>{{ end }}</a>
{{- end -}}
</span>
{{- end }}

{{/* controlmore is the rest of a folder's jobs behind one button: a menu
     of them, each with its own action, live as the slots are; nothing
     for no jobs. (list $ folder jobs) */}}
{{- define "controlmore" }}
{{- $ := index . 0 }}{{ $f := index . 1 }}{{ $jobs := index . 2 }}
{{- with $jobs }}
<details class="cc-more">
	<summary class="cc" title="{{ len . }} more job{{ if ne (len .) 1 }}s{{ end }} of {{ $f.ID }}" aria-label="{{ len . }} more jobs of {{ $f.ID }}"><svg class="ic"><use href="#i-more"/></svg></summary>
	<div class="cc-menu" role="menu">{{ range . }}{{ template "controljob" (list $ $f . true) }}{{ end }}</div>
</details>
{{- end }}
{{- end }}

{{/* controljobs are a folder's jobs in one cell, as a Files list's
     directory row holds them: the first three by name, default first,
     the rest under Other. (list $ folder) */}}
{{- define "controljobs" }}
{{- $ := index . 0 }}{{ $f := index . 1 }}{{ $jobs := $f.Jobs }}
{{- if $jobs }}
<div class="cc-row">
	{{- if gt (len $jobs) 3 }}
	{{- range slice $jobs 0 3 }}{{ template "controljob" (list $ $f . true) }}{{ end }}
	{{- template "controlmore" (list $ $f (slice $jobs 3)) }}
	{{- else }}
	{{- range $jobs }}{{ template "controljob" (list $ $f . true) }}{{ end }}
	{{- end }}
</div>
{{- end }}
{{- end }}
