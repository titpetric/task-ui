{{ define "content" }}
{{- with $k := .Key }}
<div class="page-head">
	<div>
		<h2>Edit {{ $k.Name }}</h2>
		<p>{{ if $k.Comment }}{{ $k.Comment }}, {{ end }}<span class="dim">{{ $k.Kind }}</span> {{ $k.Fingerprint }}. Choose whose runs get it: runs started by a checked user, or by anyone with a checked role, have it in their ssh agent. Users see their keys under SSH keys in their menu.</p>
	</div>
</div>

{{- $next := index $.Form "next" }}
<form class="form" data-api="/api/keys/{{ $k.Name }}" data-method="PUT" data-next="{{ $next }}?notice=Links+saved#key-{{ $k.Name }}">
	{{- template "keypolicies" (list $k.Roles $k.Users $.Users) }}
	<div class="field">
		<label class="check"><input type="checkbox" name="git"{{ if $k.Git }} checked{{ end }}> git pulls and clones with this key</label>
		<small class="faint">One key at a time; the remotes trust its public half.</small>
	</div>
	<div class="form-foot">
		<a class="btn" href="{{ $next }}">Cancel</a>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
</form>
{{- end }}
{{ end }}
