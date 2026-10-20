{{ define "content" }}
{{- with $k := .Key }}
<div class="page-head">
	<div>
		<h2>{{ $k.Name }}</h2>
		<p>{{ if $k.Comment }}{{ $k.Comment }}, {{ end }}<span class="dim">{{ $k.Type }}</span> {{ $k.Fingerprint }}. Pick whose runs get it: a run started by someone with a checked role, or by a checked user, has it in its ssh agent. A user sees the keys given to them under SSH keys in their menu.</p>
	</div>
</div>

<form class="form" data-api="/api/keys/{{ $k.Name }}" data-method="PUT" data-next="/settings/keys?notice=Policies+saved">
	{{- template "keypolicies" (list (index $.KeyPolicies $k.Name) (index $.KeyUsers $k.Name) $.Users) }}
	<div class="form-foot">
		<a class="btn" href="/settings/keys">Cancel</a>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
</form>
{{- end }}
{{ end }}
