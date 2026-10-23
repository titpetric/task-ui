{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Add workflow</h2>
		<p>Name the workflow, then add its steps in the order they run: commands, prompts for Claude and a summary. Each step starts when the one before it ends ok.</p>
	</div>
</div>

{{ template "stepeditor" (list $ nil "/api/workflows" "POST" "/workflows" "Create") }}
{{ end }}
