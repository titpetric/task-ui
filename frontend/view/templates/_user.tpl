{{/* user takes (list $ name): a user's avatar and name, with the [...]
     menu next to them: the profile, for yourself or with user/view; the
     runs by the user; their edit page, with user/edit. userHTML in
     task-ui.js draws the same for the rows a script adds. */}}
{{ define "user" }}{{ $p := index . 0 }}{{ $name := index . 1 }}<span class="user-ref">{{ template "avatar" (list $p $name) }} <span class="user-name">{{ $name }}</span>{{ template "user-menu" (list $p $name) }}</span>{{ end }}

{{ define "user-menu" }}{{ $p := index . 0 }}{{ $name := index . 1 -}}
{{ template "menu" (menu (print "More on " $name)
	(and ($p.CanProfile $name) (item "users" "Profile" "href" (print "/users/" $name)))
	(item "list" "Runs by this user" "href" (print "/runs?user=" (urlquery $name)))
	(and ($p.CanPermissions $name) (item "edit" "Edit" "href" (print "/users/" $name "/edit")))) }}
{{- end }}

{{/* user-options takes (list users selected [short]): an option per
     user, name and role, or the name alone when short, for a select that
     picks a user: who a schedule runs as, who a key is for. */}}
{{ define "user-options" }}{{ $selected := index . 1 }}{{ $short := gt (len .) 2 }}{{ range index . 0 }}
			<option value="{{ .Name }}"{{ if eq .Name $selected }} selected{{ end }}>{{ .Name }}{{ if not $short }} ({{ .EffectiveRole }}){{ end }}</option>
{{- end }}{{ end }}

{{/* runas takes (list $ id): the Run as field of a schedule form, for an
     admin, on the admin to start. */}}
{{ define "runas" }}{{ $p := index . 0 }}{{ if $p.Users }}
	<div class="field">
		<label for="{{ index . 1 }}">Run as</label>
		<select id="{{ index . 1 }}" name="user">
			{{- template "user-options" (list $p.Users $p.User.Name) }}
		</select>
	</div>
{{- end }}{{ end }}

{{/* runas-row takes (list $ schedule): who a schedule runs as, with the
     user menu; an admin picks another user, and the schedule saves on
     the change. */}}
{{ define "runas-row" }}{{ $p := index . 0 }}{{ with index . 1 }}{{ if $p.Users -}}
<form class="run-as" data-api="/api/schedules/{{ .ID }}" data-method="PUT" data-next="reload">{{ template "avatar" (list $p .UserName) }} <select name="user" data-submit aria-label="Run as">
			{{- template "user-options" (list $p.Users .UserName true) }}
		</select>{{ template "user-menu" (list $p .UserName) }}</form>
{{- else }}{{ template "user" (list $p .UserName) }}{{ end }}{{ end }}{{ end }}

{{/* folder-pick takes (list choices all label): the folders a user
     sees, All projects or the boxes of Custom, in the user forms; label
     heads the field, left out where the block's heading says it. */}}
{{ define "folder-pick" }}{{ $all := index . 1 -}}
	<div class="field" data-folder-pick>
		{{- if index . 2 }}
		<span class="label">Folders</span>
		{{- end }}
		<label class="check"><input type="radio" name="scope" value="all"{{ if $all }} checked{{ end }}> All projects</label>
		<label class="check"><input type="radio" name="scope" value="custom"{{ if not $all }} checked{{ end }}> Custom</label>
		<div class="folder-choices">
			{{- range index . 0 }}
			<label class="check"><input type="checkbox" name="folders" value="{{ .ID }}" data-type="set"{{ if .Checked }} checked{{ end }}{{ if $all }} disabled{{ end }}> <span class="mono">{{ .ID }}</span></label>
			{{- else }}
			<span class="faint">No folders yet.</span>
			{{- end }}
		</div>
	</div>
{{- end }}

{{/* matrix takes (list $ matrix): a policy, a row for each section with
     a cell for each rule, then the application permissions. A cell
     cycles inherit, allow (✓) and deny (✕) on a click; Save puts the
     cells that changed, each a hidden field that turns on when its cell
     changes. With Effective, a muted mark after each cell says whether
     the user may do it by the role's policy and their own. */}}
{{ define "matrix" }}{{ $m := index . 1 -}}
<form class="matrix-form{{ if $m.Fixed }} fixed{{ end }}"{{ if $m.Editable }} data-api="{{ $m.API }}" data-method="PUT" data-next="reload" data-matrix{{ end }}>
	<div class="table-wrap">
	<table class="table matrix{{ if $m.Effective }} effective{{ end }}">
		<thead><tr><th>Section</th>{{ range $m.Rules }}<th>{{ .Title }}</th>{{ end }}</tr></thead>
		<tbody>
		{{- range $m.Sections }}
			<tr><th scope="row">{{ .Title }}</th>{{ range .Cells }}<td>{{ template "tri" (list $m .) }}</td>{{ end }}</tr>
		{{- end }}
			<tr class="group"><th colspan="{{ $m.Columns }}">Application</th></tr>
		{{- range $m.Permissions }}
			<tr><th scope="row">{{ .Title }}</th><td colspan="{{ sub $m.Columns 1 }}">{{ template "tri" (list $m .) }}</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
	{{- if $m.Editable }}
	<div class="form-foot">
		<span class="faint" data-matrix-changed>No changes yet.</span>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
	{{- end }}
</form>
{{- end }}

{{/* tri takes (list matrix cell): one cell of a matrix. */}}
{{ define "tri" }}{{ $m := index . 0 }}{{ $c := index . 1 }}{{ $label := printf "%s: %s" $c.Title (or $c.Effect "inherit") -}}
<span class="tri-cell"><button type="button" class="tri" data-tri="{{ $c.Effect }}" data-title="{{ $c.Title }}" title="{{ $label }}" aria-label="{{ $label }}"{{ if not $m.Editable }} disabled{{ end }}>{{ if eq $c.Effect "allow" }}✓{{ else if eq $c.Effect "deny" }}✕{{ end }}</button>
{{- if $m.Editable }}<input type="hidden" name="{{ $c.Path }}" value="{{ $c.Effect }}" data-was="{{ $c.Effect }}" disabled>{{ end }}
{{- if and $m.Effective $c.Checked }}<span class="eff {{ if $c.Allowed }}ok{{ else }}bad{{ end }}" title="Result: {{ if $c.Allowed }}allowed{{ else }}denied{{ end }}">{{ if $c.Allowed }}✓{{ else }}✕{{ end }}</span>{{ end }}</span>
{{- end }}

{{/* nextrun takes a schedule: when it runs next, the day and the time
     apart so a narrow column wraps between them; a dash for one turned
     off. */}}
{{ define "nextrun" }}{{ if .IsEnabled }}{{ $at := split (stamp .NextRunAt) " " }}{{ range $at }}<span class="nowrap">{{ . }}</span> {{ end }}{{ else }}<span class="faint">-</span>{{ end }}{{ end }}
