{{ define "content" }}
<section class="login">
	<h2>Sign in</h2>
	<p>This instance runs runbooks.<br>Sign in to see projects and history.</p>
	{{- if .Error }}
	<div class="alert bad" role="alert">{{ .Error }}</div>
	{{- end }}
	<form method="post" action="/login">
		<input type="hidden" name="back" value="{{ .Back }}">
		<div class="field">
			<label for="username">Username</label>
			<input type="text" name="username" id="username" value="{{ .Username }}" autocomplete="username" autofocus required>
		</div>
		<div class="field">
			<label for="password">Password</label>
			<input type="password" name="password" id="password" autocomplete="current-password" required>
		</div>
		<button type="submit" class="btn btn-primary">Sign in</button>
	</form>
</section>
{{ end }}
