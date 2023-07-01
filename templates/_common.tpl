{{ define "nav" }}
<div class="p-1 side-nav">
<b>Task targets</b>
<ul>
{{ range .Tasks }}
	<li><a href="/{{ .Task }}">{{ .Task }}</a></li>
{{ end }}
</ul>
</div>
{{ end }}

{{ define "sidebar" }}

<div class="d-flex flex-column">

    <a href="/" class="d-flex align-items-center flex-shrink-0 p-3 link-dark text-decoration-none border-bottom">
      <svg class="bi me-2" width="30" height="24"><use xlink:href="#task"></use></svg>
      <span class="fs-5 fw-semibold">Task UI</span>
    </a>

    <div class="list-group list-group-flush border-bottom scrollarea">

{{ range $idx, $task := .Tasks }}

      <a href="/{{ $task.Task }}" class="list-group-item list-group-item-action py-3 lh-tight {{ if IsCurrent $task }} active{{ end }}"> <!-- .active, aria-current="true" -->
        <div class="d-flex w-100 align-items-center justify-content-between">
	<span class="mb-1"><b>{{ .Task }}</b>: {{ .Description }}</span>
          <span>
            <svg class="bi me-2" width="24" height="24"><use xlink:href="#ellipsis"></use></svg>
	  </span>
        </div>
      </a>

{{ end }}

    </div>
  </div>

{{ end }}

{{ define "icons" }}

<svg xmlns="http://www.w3.org/2000/svg" style="display: none;">
<symbol id="task" viewBox="0 0 375 375">
  <title>Task</title>
  <path fill="#29beb0" d="M 187.570312 190.933594 L 187.570312 375 L 30.070312 279.535156 L 30.070312 95.464844 Z"/>
  <path fill="#69d2c8" d="M 187.570312 190.933594 L 187.570312 375 L 345.070312 279.535156 L 345.070312 95.464844 Z"/>
  <path fill="#94dfd8" d="M 187.570312 190.933594 L 30.070312 95.464844 L 187.570312 0 L 345.070312 95.464844 Z"/>
</symbol>
<symbol id="ellipsis" viewBox="0 0 512 512">
  <title>Navigation ellipsis</title>
  <circle cx="256" cy="256" r="26"/>
  <circle cx="346" cy="256" r="26"/>
  <circle cx="166" cy="256" r="26"/>
  <path d="M448,256c0-106-86-192-192-192S64,150,64,256s86,192,192,192S448,362,448,256Z" style="fill:none;stroke:#000000;stroke-miterlimit:10;stroke-width:32px"/>
</symbol>
</svg>

{{ end }}
