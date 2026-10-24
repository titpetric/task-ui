{{- /* The control buttons of a folder's jobs, the dashboard's control
     centre's. */}}

{{/* controltable is the control centre's table: a row per folder, a
     column per job name that repeats, the rest under other. On a
     folder's page every row holds Named buttons: its label, the source's file
     linked to where it runs or the group's kind, then its jobs as named
     buttons; a table of Named rows has no head. With menus, each row
     ends with a [...] of what the buttons leave out.
     (list $ control menus) */}}
{{- define "controltable" }}
{{- $ := index . 0 }}{{ $c := index . 1 }}{{ $menus := index . 2 }}
	<div class="table-wrap">
	<table class="table control col-hover">
		{{- $places := false }}{{ range $c.Rows }}{{ if not .Named }}{{ $places = true }}{{ end }}{{ end }}
		{{- if or $places (not $c.Rows) }}
		<thead><tr><th>Folder</th>{{ range $c.Columns }}<th class="col"><span>{{ . }}</span></th>{{ end }}<th class="col"><span>other</span></th>{{ if $menus }}<th></th>{{ end }}</tr></thead>
		{{- end }}
		<tbody>
		{{- range $row := $c.Rows }}
			<tr{{ if $row.Named }} class="named" data-span{{ end }}>
				{{- if $row.Named }}
				<td class="key nowrap"{{ with $row.Title }} title="{{ . }}"{{ end }}>{{ if $row.Place }}<a class="badge" href="{{ $.Href $row.Folder.ID }}{{ with $row.Folder.Rel }}/{{ . }}{{ end }}">{{ $row.Label }}</a>{{ else }}<span class="badge">{{ $row.Label }}</span>{{ end }}</td>
				<td colspan="{{ add (len $c.Columns) 1 }}">{{ if and $row.Empty (not $row.Extra) }}<span class="faint">{{ $row.Empty }}</span>{{ else }}<div class="cc-row">{{ range $row.Extra }}{{ template "controljob" (list $ $row.Folder . true) }}{{ end }}</div>{{ end }}</td>
				{{- else }}
				<td class="key nowrap"{{ with $row.Title }} title="{{ . }}"{{ end }}><a href="{{ $.Href $row.Folder.ID }}{{ with $row.Folder.Rel }}/{{ . }}{{ end }}">{{ or $row.Label $row.Folder.ID }}</a></td>
				{{- if and $row.Empty (not $row.Jobs) }}
				<td class="faint" colspan="{{ add (len $c.Columns) 1 }}">{{ $row.Empty }}</td>
				{{- else }}
				{{- range $i, $cell := $row.Cells }}
				{{- /* One job is a bare button; jobs that share a name carry their kinds. */}}
				<td class="col" data-label="{{ index $c.Columns $i }}">{{ if $cell }}<div>{{ range $cell }}{{ template "controljob" (list $ $row.Folder . (gt (len $cell) 1)) }}{{ end }}</div>{{ end }}</td>
				{{- end }}
				<td class="col other" data-label="Other">
					{{- template "controlmore" (list $ $row.Folder $row.Extra) }}
				</td>
				{{- end }}
				{{- end }}
				{{- if $menus }}
				<td class="actions">{{ template "menu" ($.ControlMenu $row) }}</td>
				{{- end }}
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="{{ add (len $c.Columns) 2 }}">No jobs yet. Add a project from a git repository in Settings, or put a folder with jobs under the root.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
{{- end }}

{{/* controljob is a job's button in the control centre: Run, an eye on
     the run in progress, or a #! to its page for a user who may not run
     it. In the other column the button carries the job's name. The
     button and the #! take the colour of the job's last run: ok, bad for
     failed, stopped or killed, never without one. Its title says what the
     job does; an interactive job's button on a folder's page shows a
     terminal; a job that takes args opens a field for them first. The
     slot around it says what the script needs to draw it again as runs
     start and end. */}}
{{- define "controljob" }}
{{- $ := index . 0 }}{{ $f := index . 1 }}{{ $r := index . 2 }}{{ $named := index . 3 }}
{{- $run := $.RunningRun $r.Key }}
{{- /* When two sources share a name, the kind tells them apart, as in the sidebar. */}}
{{- $label := $r.Name }}{{ if ne $r.ID $r.Name }}{{ $label = $r.ID }}{{ end }}
{{- $last := "never" }}{{ with $.LastRun $r.Key }}{{ $last = "bad" }}{{ if eq .State "ok" }}{{ $last = "ok" }}{{ end }}{{ end }}
{{- $what := print $f.ID (or (and $r.Dir (print "/" $r.Dir)) "") " / " $r.Name }}
{{- $desc := "" }}{{ if and $r.Description (ne $r.Description $r.Name) }}{{ $desc = $r.Description }}{{ end }}
{{- $confirm := $.RunConfirm $r }}
{{- /* On a folder's page a job that takes a terminal shows one in place of the play. */}}
{{- $term := and $r.Interactive $.JobGroups }}
<span class="cc-slot" data-cc="{{ $r.Key }}" data-folder="{{ $f.ID }}" data-job="{{ $r.ID }}" data-name="{{ $r.Name }}" data-label="{{ if $named }}{{ $label }}{{ end }}" data-last="{{ $last }}"{{ if $.CanRun $f $r }} data-can{{ end }}{{ with $r.Dir }} data-dir="{{ . }}"{{ end }}{{ with $desc }} data-desc="{{ . }}"{{ end }}{{ with $r.ArgsCommand }} data-args="{{ . }}"{{ end }}{{ with $confirm }} data-confirm="{{ . }}"{{ end }}{{ if eq $r.Kind "file" }} data-path="{{ $r.Source }}"{{ end }}{{ if $term }} data-interactive{{ end }}>
{{- if $run -}}
<a class="cc{{ if $named }} named{{ end }} running" href="/runs/{{ $run.ID }}" data-attach="{{ $run.ID }}" title="Attach to {{ $what }}, running{{ with $run.Title }}: {{ . }}{{ end }}"><svg class="ic"><use href="#i-eye"/></svg>{{ if $named }}<span>{{ $label }}</span>{{ end }}</a>
{{- else if $.CanRun $f $r -}}
<button type="button" class="cc{{ if $named }} named{{ end }} {{ $last }}" data-run data-folder="{{ $f.ID }}"{{ with $r.Dir }} data-dir="{{ . }}"{{ end }} data-job="{{ $r.ID }}" data-title="{{ $what }}"{{ with $confirm }} data-confirm="{{ . }}" data-confirm-ok="Run {{ $r.Name }}"{{ end }}{{ with $r.ArgsCommand }} data-args="{{ . }}"{{ end }} title="Run {{ $what }}{{ with $desc }}: {{ . }}{{ end }}{{ with $r.ArgsCommand }}, with arguments after {{ . }}{{ end }}{{ if $term }}, in a terminal{{ end }}"><svg class="ic"><use href="#i-{{ if $term }}terminal{{ else }}play{{ end }}"/></svg>{{ if $named }}<span>{{ $label }}</span>{{ end }}</button>
{{- else -}}
<a class="cc{{ if $named }} named{{ end }} idle {{ $last }}" href="{{ $.JobHref $f $r }}" title="{{ $what }}{{ with $desc }}: {{ . }}{{ end }}: no permission to run it"><i class="shebang" aria-hidden="true">#!</i>{{ if $named }}<span>{{ $label }}</span>{{ end }}</a>
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
