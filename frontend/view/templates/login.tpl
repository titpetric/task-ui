{{ define "content" }}
<section class="login">
	<h2>Sign in</h2>
	{{- if .Error }}
	<p class="login-error" role="alert">{{ .Error }}</p>
	{{- end }}
	<form method="post" action="/login">
		<input type="hidden" name="back" value="{{ .Back }}">
		<div class="field">
			<label for="username">Username</label>
			<input type="text" name="username" id="username" value="{{ index .Form "username" }}" autocomplete="username" autofocus required>
		</div>
		<div class="field">
			<label for="password">Password</label>
			<input type="password" name="password" id="password" autocomplete="current-password" required>
		</div>
		<button type="submit">Sign in</button>
	</form>
</section>
{{ end }}
