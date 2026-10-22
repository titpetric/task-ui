{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Not found</h2>
		<p>{{ .Error }} Someone may have renamed it in the pipeline file, or you don't have access to it. Pick a folder on the left.</p>
	</div>
</div>
{{ end }}
