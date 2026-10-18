{{ define "content" }}
<div class="page-head">
	<div>
		<h2>New workflow</h2>
		<p>A name and its steps, in the order they run: a job of a folder, or another workflow. Each step starts when the one before it ended ok.</p>
	</div>
</div>

{{ template "stepeditor" (list $ nil "/api/workflows" "POST" "/workflows" "Create") }}
{{ end }}
