{{/* projecttable lists projects, (list page entries): the table of the
     Projects page and of a workspace's page, the same rows on both:
     the folder, its sources and jobs, the checkout's remote and branch,
     the last commit, the size on disk when the page measures it, and the
     row's [...] with Open, Project settings and Remove checkout. A page
     with no entries draws its blockhead's empty line instead. */}}
{{- define "projecttable" }}{{ $p := index . 0 }}
<div class="table-wrap">
<table class="table projects">
	<thead><tr><th>Folder</th><th>Sources</th><th>Origin</th><th class="hide-sm">Last commit</th>{{ if $p.Spaces }}<th class="hide-sm">Size</th>{{ end }}<th class="actions" aria-label="Actions"></th></tr></thead>
	<tbody>
	{{- range index . 1 }}
		<tr>
			<td class="key nowrap">{{ if .Sources }}<a href="{{ $p.Href .ID }}"><svg class="ic"><use href="#i-folder"/></svg> {{ .ID }}</a>{{ else }}<svg class="ic"><use href="#i-folder"/></svg> {{ .ID }}{{ end }}</td>
			<td class="nowrap">{{ if .Error }}<span class="badge bad">error</span> {{ end }}{{ if .Sources }}{{ join .Sources ", " }} <span class="faint">{{ .Jobs }} jobs</span>{{ else }}<span class="faint">nothing to run</span>{{ end }}</td>
			<td class="mono-break">{{ with .Checkout }}{{ .Remote }} <span class="badge">{{ .Branch }}</span>{{ else }}<span class="faint">local folder</span>{{ end }}</td>
			<td class="desc hide-sm">{{ with .Checkout }}<span class="faint">{{ .Commit }}</span> {{ .Subject }}{{ if not .When.IsZero }} <span class="faint">{{ ago .When }}</span>{{ end }}{{ end }}</td>
			{{- if $p.Spaces }}<td class="hide-sm">{{ template "spacemini" index $p.Spaces .ID }}</td>{{ end }}
			<td class="actions">{{ template "menu" (menu (print "Actions on " .ID)
				(and .Sources (item "folder-open" "Open" "href" ($p.Href .ID)))
				(item "settings" "Project settings" "href" (print "/settings/projects/" .ID))
				(and .Checkout ($p.Can "project/delete") (danger "trash" "Remove checkout" "data-delete" (print "/api/projects/" .ID) "data-confirm" (removeConfirm .ID .Checkout.Remote) "data-confirm-ok" "Remove checkout" "data-next" "reload"))
			) }}</td>
		</tr>
	{{- end }}
	</tbody>
</table>
</div>
{{- end }}
