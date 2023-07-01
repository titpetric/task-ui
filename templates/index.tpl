<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
	<meta http-equiv="x-ua-compatible" content="ie=edge">

	<title>Task UI</title>

	<script src="https://cdn.jsdelivr.net/npm/bootstrap@5.2.3/dist/js/bootstrap.min.js" integrity="sha256-m81NDyncZVbr7v9E6qCWXwx/cwjuWDlHCMzi9pjMobA=" crossorigin="anonymous"></script>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@5.2.3/dist/css/bootstrap.min.css" integrity="sha256-wLz3iY/cO4e6vKZ4zRmo4+9XDpMcgKOvv/zEU3OMlRo=" crossorigin="anonymous">
</head>
<body>

{{ template "icons" . }}

<div class="container">
	<div class="row">
		<div class="col-12">
			{{ template "sidebar" . }}
		</div>
	</div>
</div>

<!--

	<div class="content p-9">
{{ range $idx, $task := .Tasks }}
		<div class="card">
			<div class="card-body">
				<h3>{{ $task.Task }} - {{ $task.Description }}</h3>

				<a href="#" onclick="runner.exec({{ printf "terminal-%d" $idx }}, {{ .Task }}, {{ .Flags.Interactive }})" class="btn btn-primary">Run: task {{ .Task }}</a>

				<div id="terminal-{{ $idx }}" class="container-terminal"></div>

			</div>
		</div>
{{ end }}
	</div>

-->

<pre style="display: none">
{{ .Tasks | json }}
</pre>

<link rel="stylesheet" href="//cdn.rawgit.com/sourcelair/xterm.js/ed13218f/dist/xterm.css"/>
<script src="//cdn.rawgit.com/sourcelair/xterm.js/ed13218f/dist/xterm.js"></script>

<script src="/static/index.js"></script>
<link href="/static/index.css" rel="stylesheet"/>

</body>
</html>