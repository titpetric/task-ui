{{ define "content" }}
{{- $p := .Folder }}
<div class="page-head">
	<div>
		<h2>Configure {{ $p.ID }}</h2>
		<p>What the instance adds to the folder: a title, who may run it, and which sources to read. The files in the folder stay as they are.</p>
	</div>
</div>

<form class="form" data-api="/api/projects/{{ $p.ID }}" data-method="PUT" data-next="/settings/projects?notice=Saved">
	<div class="field">
		<label for="pe-title">Title</label>
		<input type="text" id="pe-title" name="title" value="{{ index .Form "title" }}" autocomplete="off">
	</div>
	<div class="field">
		<label for="pe-description">Description</label>
		<input type="text" id="pe-description" name="description" value="{{ index .Form "description" }}" autocomplete="off">
	</div>
	<div class="field">
		<label for="pe-allow">Who may run <span class="faint">user names or roles, comma separated; empty means every runner</span></label>
		<input type="text" id="pe-allow" name="allow" data-type="list" value="{{ index .Form "allow" }}" placeholder="ops, release" spellcheck="false" autocomplete="off">
	</div>
	<div class="field">
		<label for="pe-sources">Sources <span class="faint">files to read, comma separated; empty detects them</span></label>
		<input type="text" id="pe-sources" name="sources" data-type="list" value="{{ index .Form "sources" }}" placeholder="atkins.yml, Makefile" spellcheck="false" autocomplete="off">
	</div>
	<div class="field">
		<label for="pe-hours">Hours <span class="faint">when the jobs may run, one rule per line; the last rule for a day wins; empty is always</span></label>
		<textarea id="pe-hours" name="hours" rows="3" spellcheck="false" autocomplete="off" placeholder="mon-fri 9-17&#10;fri 9-13">{{ index .Form "hours" }}</textarea>
	</div>
	{{- if $p.Jobs }}
	{{- $allow := $.Project.JobAllowMap }}{{ $hours := $.Project.JobHoursMap }}
	<div class="field">
		<label>Per job <span class="faint">who may run it and when, over the folder's; empty inherits</span></label>
		<div class="table-wrap">
		<table class="table cards">
			<thead><tr><th>Job</th><th>Who may run</th><th>Hours</th></tr></thead>
			<tbody>
			{{- range $p.Jobs }}
			<tr>
				<td class="key"><label for="ja-{{ .ID }}">{{ .Name }}</label></td>
				<td data-label="Who may run"><input type="text" id="ja-{{ .ID }}" name="job_allow.{{ .ID }}" data-type="list" value="{{ join (index $allow .ID) ", " }}" placeholder="ops, release" spellcheck="false" autocomplete="off"></td>
				<td data-label="Hours"><input type="text" name="job_hours.{{ .ID }}" value="{{ index $hours .ID }}" placeholder="mon-fri 9-17" spellcheck="false" autocomplete="off"></td>
			</tr>
			{{- end }}
			</tbody>
		</table>
		</div>
	</div>
	{{- end }}
	<div class="form-foot">
		<a class="btn" href="/settings/projects">Cancel</a>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
</form>
{{ end }}
