{{/* projecttable lists projects, (list page entries): the table of the
     Projects page and of a workspace's page, the same rows on both:
     the folder, its sources and jobs, the checkout's remote, cut to its
     column and whole on a hover, and branch, the last commit with its
     subject on a hover, the size on disk when the page measures it, and
     the row's [...] with Open, Project settings, Pull now on a checkout
     whose git skill the user may run, and Remove checkout. A page with
     no entries draws its blockhead's empty line instead. */}}
{{- define "projecttable" }}{{ $p := index . 0 }}
<div class="table-wrap">
<table class="table projects cards">
	<thead><tr><th>Folder</th><th>Sources</th><th>Origin</th><th class="hide-sm">Last commit</th>{{ if $p.Spaces }}<th class="hide-sm">Size</th>{{ end }}<th class="actions" aria-label="Actions"></th></tr></thead>
	<tbody>
	{{- range index . 1 }}
		<tr>
			<td class="key nowrap"><a href="{{ $p.Href .ID }}"><svg class="ic"><use href="#i-folder"/></svg> {{ .ID }}</a></td>
			<td class="nowrap" data-label="Sources">{{ if .Error }}<span class="badge bad">error</span> {{ end }}{{ if .Sources }}{{ join .Sources ", " }} <span class="faint">{{ .Jobs }} jobs</span>{{ else }}<span class="faint">nothing to run</span>{{ end }}</td>
			<td class="origin" data-label="Origin">{{ with .Checkout }}<span class="remote mono" title="{{ .Remote }}">{{ .Remote }}</span> <span class="badge">{{ .Branch }}</span>{{ else }}<span class="faint">local folder</span>{{ end }}</td>
			<td class="nowrap hide-sm" data-label="Last commit">{{ with .Checkout }}<span class="mono" title="{{ .Subject }}">{{ .Commit }}</span>{{ if not .When.IsZero }} <span class="faint">{{ ago .When }}</span>{{ end }}{{ end }}</td>
			{{- if $p.Spaces }}<td class="hide-sm" data-label="Size">{{ template "spacemini" index $p.Spaces .ID }}</td>{{ end }}
			{{- $pull := index $p.Pulls .ID }}
			<td class="actions">{{ template "menu" (menu (print "Actions on " .ID)
				(item "folder-open" "Open" "href" ($p.Href .ID))
				(item "settings" "Project settings" "href" (print "/settings/projects/" .ID))
				(and $pull (item "download" "Pull now" "data-run" "" "data-folder" .ID "data-dir" $pull.Dir "data-job" $pull.ID "data-title" (print .ID " / " $pull.Name)))
				(and .Checkout ($p.Can "project/delete") (danger "trash" "Remove checkout" "data-delete" (print "/api/projects/" .ID) "data-confirm" (removeConfirm .ID .Checkout.Remote) "data-confirm-ok" "Remove checkout" "data-next" "reload"))
			) }}</td>
		</tr>
	{{- end }}
	</tbody>
</table>
</div>
{{- end }}
