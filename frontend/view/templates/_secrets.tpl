{{- /* secretgroup is a workspace's block of secrets, (list page group
     inline): its head with Add secret, then its secrets, those every
     folder in it gets first, then by project. inline opens Add secret
     and Replace the value in the page's dialogs, secretdialogs. */}}
{{- define "secretgroup" }}
{{- $ := index . 0 }}{{ $g := index . 1 }}{{ $inline := index . 2 }}
{{- $add := item "plus" "Add secret" "disabled" "The instance has no encryption key yet" }}
{{- if and $g.Key $inline }}{{ $add = item "plus" "Add secret" "data-dialog-open" "secret-add-dialog" }}{{ else if $g.Key }}{{ $add = item "plus" "Add secret" "href" (print "/settings/secrets/new?scope=" $g.Shared) }}{{ end }}
<section class="block secret-group" id="ws-{{ $g.Workspace }}">
	{{- $note := print (len $g.Secrets) " secret" (or (and (ne (len $g.Secrets) 1) "s") "") }}{{ if ne $g.Title $g.Workspace }}{{ $note = print $g.Workspace ", " $note }}{{ end }}
	{{- template "blockhead" (head $g.Title $note (and (not $g.Secrets) (print "No secrets in " $g.Workspace " yet.")) (and ($.Can "secret/edit") (not $inline) $add)) }}
	{{- if not $g.Key }}
	<div class="alert signal secret-nokey"><svg class="ic"><use href="#i-secret"/></svg> The instance has no encryption key, so it stores no secret. <a href="/settings/instance#keys">Instance</a></div>
	{{- end }}
	{{- if $g.Secrets }}
	<div class="table-wrap">
	<table class="table secrets">
		<thead><tr><th>Name</th><th>Scope</th><th>Updated</th><th>Reads</th><th>Last read</th><th class="actions" aria-label="Actions"></th></tr></thead>
		{{- with $g.SharedSecrets }}
		<tbody>
			<tr class="secret-scope"><th colspan="6">Every folder in {{ $g.Title }}</th></tr>
			{{- range . }}{{ template "secretrow" (list $ $g . $inline) }}{{ end }}
		</tbody>
		{{- end }}
		{{- range $g.Projects }}
		<tbody>
			<tr class="secret-scope"><th colspan="6"><a href="{{ $.Href .Folder }}"><svg class="ic"><use href="#i-folder"/></svg> {{ .Folder }}</a></th></tr>
			{{- range .Secrets }}{{ template "secretrow" (list $ $g . $inline) }}{{ end }}
		</tbody>
		{{- end }}
	</table>
	</div>
	{{- end }}
</section>
{{- end }}

{{- /* secretrow is a secret's row, (list page group secret inline): its
     name, its scope, when it changed, how many runs read it, linking to
     its read log, the last run that did, and replace and delete, each
     behind its rule. The value never shows. inline opens Replace the
     value in the page's dialog. */}}
{{- define "secretrow" }}
{{- $p := index . 0 }}{{ $g := index . 1 }}{{ $s := index . 2 }}{{ $inline := index . 3 }}
			<tr>
				<td class="key">{{ $s.Name }}</td>
				<td class="scope faint">{{ $s.Scope }}</td>
				<td class="nowrap" title="{{ stamp $s.UpdatedAt }}">{{ ago $s.UpdatedAt }}</td>
				<td class="nowrap">{{ with index $p.SecretReadCounts $s.ID }}<a href="/settings/secrets/{{ $s.ID }}/reads" title="The runs that got {{ $s.Name }}">{{ . }}</a>{{ else }}<span class="faint">0</span>{{ end }}</td>
				{{- $last := index $p.SecretLastReads $s.ID }}
				<td>{{ if $last.RunID }}<span title="{{ stamp $last.CreatedAt }}">{{ ago $last.CreatedAt }}</span> <a href="/runs/{{ $last.RunID }}">{{ $last.Folder }}/{{ with $last.Dir }}{{ . }}/{{ end }}{{ $last.Job }}</a>{{ else }}<span class="faint">never</span>{{ end }}</td>
				<td class="actions">{{ template "menu" (menu (print "Actions on " $s.Name)
					(detail "Changed" (stamp $s.UpdatedAt))
					(and ($p.Can "secret/edit") (not $inline) (item "edit" "Replace the value" "href" (print "/settings/secrets/edit?scope=" (urlquery $s.Scope) "&name=" (urlquery $s.Name))))
					(and ($p.Can "secret/edit") $inline (item "edit" "Replace the value" "data-dialog-open" "secret-replace-dialog" "data-fill-scope" $s.Scope "data-fill-name" $s.Name))
					(item "list" "Reads" "href" (print "/settings/secrets/" $s.ID "/reads"))
					(and ($p.Can "secret/delete") (danger "trash" "Delete" "data-delete" (print "/api/secrets/" $s.ID) "data-confirm" (print "Delete " $s.Name " from " (or (and (eq $s.Scope $g.Shared) (print "every folder in " $g.Workspace)) $s.Scope) "? Runs stop getting it.") "data-confirm-ok" "Delete" "data-next" "reload"))
				) }}</td>
			</tr>
{{- end }}

{{- /* secretfields are the fields of a secret's form, (list secret
     groups): its scope, a workspace's or one of its projects', its name,
     read only once it has one, and its value, empty: a value never
     shows again. */}}
{{- define "secretfields" }}
{{- $s := index . 0 }}{{ $groups := index . 1 }}
	<div class="field">
		<label for="s-scope">Scope</label>
		<select id="s-scope" name="scope">
			{{- range $groups }}
			{{- $off := not .Key }}
			<optgroup label="{{ .Title }}{{ if $off }}: no encryption key{{ end }}"{{ if $off }} disabled{{ end }}>
				<option value="{{ .Shared }}"{{ if eq $s.Scope .Shared }} selected{{ end }}{{ if $off }} disabled{{ end }}>every folder in {{ .Title }}</option>
				{{- range .Folders }}
				<option value="{{ . }}"{{ if eq $s.Scope . }} selected{{ end }}{{ if $off }} disabled{{ end }}>{{ . }}</option>
				{{- end }}
			</optgroup>
			{{- end }}
		</select>
	</div>
	<div class="field">
		<label for="s-name">Name</label>
		<input type="text" id="s-name" name="name" value="{{ $s.Name }}" pattern="[A-Za-z_][A-Za-z0-9_]*" spellcheck="false" autocomplete="off" required{{ if $s.Name }} readonly{{ end }}>
	</div>
	<div class="field">
		<label for="s-content">Value</label>
		<textarea id="s-content" name="content" rows="3" spellcheck="false" autocomplete="off" required></textarea>
	</div>
{{- end }}

{{- /* secretdialogs are the dialogs of a page that edits secrets in
     place, (list page next): Add secret, the fields of secretfields
     for the page's groups, and Replace the value, which takes the scope
     and the name from the item that opens it. Both return to next. */}}
{{- define "secretdialogs" }}
{{- $p := index . 0 }}{{ $next := index . 1 }}
<dialog class="confirm" id="secret-add-dialog" aria-labelledby="secret-add-title">
	<form class="form" data-api="/api/secrets" data-method="POST" data-next="{{ $next }}?notice=Secret+saved">
		<h3 id="secret-add-title">Add secret</h3>
		<p>task-ui encrypts the value and never shows it again. Saving an existing name replaces its value.</p>
		{{- template "secretfields" (list $p.Secret $p.SecretGroups) }}
		<div class="form-foot">
			<button type="button" class="btn" data-dialog-close>Cancel</button>
			<button class="btn btn-primary" type="submit">Save</button>
		</div>
	</form>
</dialog>
<dialog class="confirm" id="secret-replace-dialog" aria-labelledby="secret-replace-title">
	<form class="form" data-api="/api/secrets" data-method="POST" data-next="{{ $next }}?notice=Secret+saved">
		<h3 id="secret-replace-title">Replace <span class="mono" data-fill-text="name"></span></h3>
		<p>The new value replaces the old one for <span class="mono" data-fill-text="scope"></span>; neither shows again.</p>
		<input type="hidden" name="scope" value="">
		<input type="hidden" name="name" value="">
		<div class="field">
			<label for="sr-content">Value</label>
			<textarea id="sr-content" name="content" rows="3" spellcheck="false" autocomplete="off" required></textarea>
		</div>
		<div class="form-foot">
			<button type="button" class="btn" data-dialog-close>Cancel</button>
			<button class="btn btn-primary" type="submit">Save</button>
		</div>
	</form>
</dialog>
{{- end }}
