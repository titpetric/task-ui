{{ define "content" }}
<div class="page-head">
	<div>
		<h2>{{ .Title }}</h2>
		<p>Adds a user to the password file. They can sign in right away. They start as a viewer, the lowest role; an admin sets the role and the flags on their profile.</p>
	</div>
</div>

<form class="form" data-api="/api/users" data-method="POST" data-next="/settings/users?notice=User+created">
	<div class="field">
		<label for="u-name">Name</label>
		<input type="text" id="u-name" name="name" required pattern="[a-z0-9][a-z0-9\-]{0,31}" spellcheck="false" autocomplete="off" autofocus>
		<small class="faint">Lowercase letters, digits and dashes; it also names their SSH keys.</small>
	</div>
	<div class="field">
		<label for="u-password">Password <span class="faint">at least 8 characters</span></label>
		<input type="password" id="u-password" name="password" required minlength="8" autocomplete="new-password">
		<small class="faint">Stored as a bcrypt hash.</small>
	</div>
	{{- template "folder-pick" (list .FolderChoices true true) }}
	<div class="form-foot">
		<a class="btn" href="/settings/users">Cancel</a>
		<button class="btn btn-primary" type="submit">Add user</button>
	</div>
</form>
{{ end }}
