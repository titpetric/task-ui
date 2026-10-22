{{ define "content" }}
<div class="page-head">
	<div>
		<h2>New workflow</h2>
		<p>Name the workflow and add its steps in the order they run. A step is a job of a folder or another workflow. Each step starts when the one before it ends ok.</p>
	</div>
</div>

{{ template "stepeditor" (list $ nil "/api/workflows" "POST" "/workflows" "Create") }}
{{ end }}
