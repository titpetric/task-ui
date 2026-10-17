{{ define "content" }}
{{- $u := .Account }}
<div class="page-head">
	<div>
		<h2>{{ .Title }}</h2>
		<p>Limit the folders {{ $u.Name }} sees, or show them all. The name, the password and the role come from the password file and the role lists.</p>
	</div>
</div>

<form class="form" data-api="/api/users/{{ $u.ID }}" data-method="PUT" data-next="/settings/users?notice=Folders+saved">
	<div class="field">
		<label for="u-name">Name</label>
		<input type="text" id="u-name" value="{{ $u.Name }}" readonly>
	</div>
	<div class="field">
		<label for="u-role">Role <span class="faint">from TASK_UI_ADMINS and TASK_UI_VIEWERS</span></label>
		<input type="text" id="u-role" value="{{ $u.Role }}" readonly>
	</div>
	<div class="field">
		<label for="u-folders">Folders <span class="faint">comma separated; empty shows all</span></label>
		<input type="text" id="u-folders" name="folders" data-type="list" value="{{ join $u.FolderList ", " }}" placeholder="infra, website" spellcheck="false" autocomplete="off" autofocus>
	</div>
	<div class="form-foot">
		<a class="btn" href="/settings/users">Cancel</a>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
</form>
{{ end }}
