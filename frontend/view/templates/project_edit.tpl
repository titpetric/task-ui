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
	<div class="form-foot">
		<a class="btn" href="/settings/projects">Cancel</a>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
</form>
{{ end }}
