{{ define "content" }}
{{- with $k := .Key }}
<div class="page-head">
	<div>
		<h2>{{ $k.Name }}</h2>
		<p>{{ if $k.Comment }}{{ $k.Comment }}, {{ end }}<span class="dim">{{ $k.Kind }}</span> {{ $k.Fingerprint }}. Choose whose runs get it: runs started by a checked user, or by anyone with a checked role, have it in their ssh agent. Users see their keys under SSH keys in their menu.</p>
	</div>
</div>

<form class="form" data-api="/api/keys/{{ $k.Name }}" data-method="PUT" data-next="/settings/keys?notice=Links+saved">
	{{- template "keypolicies" (list $k.Roles $k.Users $.Users) }}
	<label class="check"><input type="checkbox" name="git"{{ if $k.Git }} checked{{ end }}> git pulls and clones with this key <span class="faint">one key at a time; the remotes trust its public half</span></label>
	<div class="form-foot">
		<a class="btn" href="/settings/keys">Cancel</a>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
</form>
{{- end }}
{{ end }}
