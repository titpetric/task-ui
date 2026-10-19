{{ define "content" }}
{{- with $k := .Key }}
<div class="page-head">
	<div>
		<h2>{{ $k.Name }}</h2>
		<p>{{ if $k.Comment }}{{ $k.Comment }}, {{ end }}<span class="dim">{{ $k.Type }}</span> {{ $k.Fingerprint }}. Pick the roles whose runs get it: a run started by someone with the role has it in its ssh agent, beside the keys of its project's .ssh/ folder.</p>
	</div>
</div>

<form class="form" data-api="/api/keys/{{ $k.Name }}" data-method="PUT" data-next="/settings/keys?notice=Policies+saved">
	{{- template "keypolicies" (index $.KeyPolicies $k.Name) }}
	<div class="form-foot">
		<a class="btn" href="/settings/keys">Cancel</a>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
</form>
{{- end }}
{{ end }}
