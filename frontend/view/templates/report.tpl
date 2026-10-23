{{ define "content" }}
{{- with .Report }}{{ $rep := . }}{{ $run := $.Run }}{{ $about := $.ReportOf }}
<div class="page-head">
	<div>
		<h2>{{ $about.Name }} <span class="badge">report</span></h2>
		<p>{{ if $rep.IsSummary }}Claude's summary of the run{{ else }}Claude's answer about the run{{ end }} of {{ stamp $about.StartedAt }} by {{ $about.UserName }}{{ if ended $about.State }}, which ended {{ $about.State }}{{ end }}{{ with $rep.By }}; asked by {{ . }}{{ end }}{{ with $rep.At }}, written {{ ago . }}{{ end }}.</p>
	</div>
	<div class="actions">
		<a class="btn" href="/runs/{{ $about.ID }}">Run</a>
		{{- if ne $about.ID $run.ID }}<a class="btn" href="/runs/{{ $run.ID }}">Claude's run</a>{{ end }}
		{{- with $.RunJobHref $about.Folder $about.Dir $about.Job }}<a class="btn" href="{{ . }}">Job</a>{{ end }}
	</div>
</div>

{{- if not $rep.IsSummary }}
<section class="block">
	<h3>Question</h3>
	<div class="markdown-doc report-question">{{ markdown $rep.Question }}</div>
</section>
{{- end }}

<section class="block">
	<h3>{{ if $rep.IsSummary }}Summary{{ else }}Answer{{ end }}</h3>
	<article class="markdown-doc report-doc">{{ markdown $rep.Body }}</article>
</section>
{{- end }}
{{ end }}
