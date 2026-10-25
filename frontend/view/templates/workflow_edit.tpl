{{ define "content" }}
{{- with $w := .Workflow }}
<div class="page-head">
	<div>
		<h2>Edit {{ $w.Name }}</h2>
		<p>Change the name and the steps: fold a step open to edit it, drag it by its grip to move it. Earlier runs keep the steps they ran with.</p>
	</div>
	<div class="actions">
		{{ template "menu" (menu (print "Actions on " $w.Name)
			(item "flow" "Open" "href" (print "/workflows/" $w.ID))
			$.StepMoveItems
			(danger "trash" "Delete" "data-delete" (print "/api/workflows/" $w.ID) "data-confirm" (print "Delete the workflow " $w.Name "? Its runs stay in the history.") "data-confirm-ok" "Delete" "data-next" "/workflows")
		) }}
	</div>
</div>

{{ template "stepeditor" (list $ $w "/api/workflows/" "PUT" (print "/workflows/" $w.ID) "Save") }}

{{- /* The workflow's schedules, with the form that adds one: the
     schedules page's create form less the job, the args, the triage
     and Run as, which a workflow run has no use for: it runs as the
     workflow's user. */}}
<section class="block" id="schedule">
	{{- $as := "the user who adds the schedule" }}{{ if $w.RunAs }}{{ $as = print $w.RunAs ", the user set on the workflow" }}{{ end }}
	{{- template "blockhead" (head "Schedule" (print "Runs the workflow on a cron expression, as " $as ". Each run needs permission to run every step at that time.") (and (not $.Schedules) "Not scheduled.")) }}
	{{- template "schedtable" (list $ $.Schedules (or $w.CanRun $.User.CanAdmin) $w.Name) }}
	<form class="form inline" data-api="/api/schedules" data-method="POST" data-next="reload">
		<input type="hidden" name="workflow" value="{{ $w.ID }}">
		<div class="field">
			<label for="sch-expr">Expression</label>
			<input type="text" id="sch-expr" name="expression" placeholder="0 3 * * *  or  @daily" spellcheck="false" autocomplete="off" required>
		</div>
		<div class="form-foot">
			<button class="btn btn-primary" type="submit"><svg class="ic"><use href="#i-plus"/></svg>Add schedule</button>
		</div>
	</form>
</section>
{{- end }}
{{ end }}
