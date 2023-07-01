<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
	<meta http-equiv="x-ua-compatible" content="ie=edge">

	<title>Task UI</title>

	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@5.2.3/dist/css/bootstrap.min.css" integrity="sha256-wLz3iY/cO4e6vKZ4zRmo4+9XDpMcgKOvv/zEU3OMlRo=" crossorigin="anonymous">
	<link href="/static/index.css" rel="stylesheet"/>
</head>
<body>

{{ template "icons" . }}

<div class="container">
	<div class="row">
		<div class="col-12 col-lg-3">
			{{ template "sidebar" . }}
		</div>
		<div class="col-12 col-lg-9 p-1">
			    <div class="d-flex align-items-center flex-shrink-0 p-3 pb-0 link-dark text-decoration-none">
			      <svg class="bi me-2 d-inline-block d-lg-none" width="30" height="24"><use xlink:href="#task"></use></svg>
			      <span class="fs-5 me-4 fw-semibold">Current: {{ .Current.Task }}</span>

				{{ if .Current.Flags.Interactive }}<span class="badge bg-danger">Interactive</span>{{ end }}
				{{ if not .Current.Flags.Interactive }}<span class="badge bg-success">Non-interactive</span>{{ end }}
			    </div>


			<ul class="nav nav-tabs mt-2" id="myTab" role="tablist">
			  <li class="nav-item" role="presentation">
			    <button class="nav-link active" id="home-tab" data-bs-toggle="tab" data-bs-target="#home" type="button" role="tab" aria-controls="home" aria-selected="true">Runner</button>
			  </li>
			  <li class="nav-item" role="presentation">
			    <button class="nav-link" id="history-tab" data-bs-toggle="tab" data-bs-target="#history" type="button" role="tab" aria-controls="history" aria-selected="false">History</button>
			  </li>
			  <li class="nav-item" role="presentation">
			    <button class="nav-link" id="current-tab" data-bs-toggle="tab" data-bs-target="#current" type="button" role="tab" aria-controls="current" aria-selected="false">Current task</button>
			  </li>
			  <li class="nav-item" role="presentation">
			    <button class="nav-link" id="tasks-tab" data-bs-toggle="tab" data-bs-target="#tasks" type="button" role="tab" aria-controls="tasks" aria-selected="false">All Tasks</button>
			  </li>
			</ul>

			<div class="tab-content" id="myTabContent">
			  <div class="tab-pane fade p-2 show active" id="home" role="tabpanel" aria-labelledby="home-tab">

				<table class="table table-sm w-auto">
				<tbody>
					<tr>
						<th>Description:</th>
						<td>{{ .Current.Description }}</td>
					</tr>
				</tbody>
				</table>

				<div id="container-terminal" class="container-terminal">
					<div class="terminal xterm xterm-theme-default">
						<p class="mb-0">This is a web based terminal.<br/>To run the task target, press the Run button below.</p>
					</div>
				</div>
				<a class="btn btn-primary mt-2" onclick="runner.exec('container-terminal', {{ .Current.Task }}, {{ .Current.Flags.Interactive }})">Run</a>

			  </div>
			  <div class="tab-pane fade p-2" id="history" role="tabpanel" aria-labelledby="history-tab">
                                <table class="table table-sm table-striped">
				<thead>
					<tr>
					<th>Time of run</th>
					<th>Actions</th>
					</tr>
				</thead>
				<tbody>
				{{ range .Current.History }}
						<tr>
							<td>{{ .Datetime }}</td>
							<td>
								<a onclick="runner.history('container-{{ .ID }}', '{{ .ID }}')" class="btn btn-primary btn-outline btn-sm">View log</a>
								<a onclick="runner.history('container-{{ .ID }}', '{{ .ID }}', true)" class="btn btn-primary btn-outline btn-sm">Playback</a>
							</td>
						</tr>
						<tr>
							<td colspan="2">
								<div id="container-{{ .ID }}" class="container-terminal">
									<div class="terminal xterm xterm-theme-default"></div>
								</div>
							</td>
						</tr>
				{{ end }}
					</tbody>
				</table>
			  </div>
			  <div class="tab-pane fade p-2" id="current" role="tabpanel" aria-labelledby="current-tab">
				<code><pre>
				{{- .Current | json -}}
				</pre></code>
			  </div>
			  <div class="tab-pane fade p-2" id="tasks" role="tabpanel" aria-labelledby="tasks-tab">
				<code><pre>
				{{- .Tasks | json -}}
				</pre></code>
			  </div>
			</div>

		</div>
	</div>
</div>

</div>

<link rel="stylesheet" href="//cdn.rawgit.com/sourcelair/xterm.js/ed13218f/dist/xterm.css"/>

<script src="https://cdn.jsdelivr.net/npm/bootstrap@5.2.3/dist/js/bootstrap.min.js" integrity="sha256-m81NDyncZVbr7v9E6qCWXwx/cwjuWDlHCMzi9pjMobA=" crossorigin="anonymous"></script>
<script src="//cdn.rawgit.com/sourcelair/xterm.js/ed13218f/dist/xterm.js"></script>
<script src="/static/index.js"></script>

</body>
</html>