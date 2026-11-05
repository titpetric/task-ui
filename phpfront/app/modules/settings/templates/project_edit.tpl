{eval $data_head = $data.head}
{include components/page-head.tpl}

<form class="form" data-api="{data.api}" data-method="PUT" data-next="{data.next}">
	<div class="field">
		<label for="pe-title">Title</label>
		<input type="text" id="pe-title" name="title" value="{data.form.title}" autocomplete="off">
	</div>
	<div class="field">
		<label for="pe-description">Description</label>
		<input type="text" id="pe-description" name="description" value="{data.form.description}" autocomplete="off">
	</div>
	<div class="field" role="group" aria-labelledby="pe-allow-label">
		<span class="label" id="pe-allow-label">Who can run</span>{eval $data_whofield = $data.allow}{include components/who-field.tpl}
		<small class="faint">With nothing checked, everyone who may run can.</small>
	</div>
{if $data.move}	<div class="field">
		<label for="pe-workspace">Workspace</label>
		<select id="pe-workspace" name="workspace" data-move="{data.move.id}" data-move-from="{data.move.from}" data-move-left="{data.move.left}">
{foreach $data.move.options as $option}			<option value="{option.value}"{if $option.selected} selected{/if}>{option.label}</option>
{/foreach}		</select>
		<small class="faint">Picking another workspace moves the checkout there, with the users' folder access; earlier runs stay under {data.move.id}.</small>
	</div>
{/if}	<div class="field">
		<label for="pe-sources">Sources</label>
		<input type="text" id="pe-sources" name="sources" data-type="list" value="{data.form.sources}" placeholder="atkins.yml, Makefile" spellcheck="false" autocomplete="off">
		<small class="faint">Files to read, comma separated; leave empty to detect them.</small>
	</div>
{eval $data_recursive = $data.recursive}{include components/recursive-field.tpl}
	<div class="field">
		<label class="check"><input type="checkbox" id="pe-sidebar" name="sidebar"{if $data.sidebar} checked{/if}> Show in sidebar</label>
		<small class="faint">Off hides the project and its folders from the sidebar; their pages still work.</small>
	</div>
{if $data.jobs}	<div class="field">
		<span class="label">Per job</span>
		<div class="table-wrap">
		<table class="table cards">
			<thead><tr><th>Job</th><th>Who can run</th></tr></thead>
			<tbody>
{foreach $data.jobs as $job}			<tr>
				<td class="key">{job.name}</td>
				<td data-label="Who can run">{eval $data_whofield = $job.who}{include components/who-field.tpl}</td>
			</tr>
{/foreach}			</tbody>
		</table>
		</div>
		<small class="faint">Who can run each job and when, on top of the folder's settings; leave unchecked and empty to use the folder's.</small>
	</div>
{/if}	<div class="form-foot">
		<a class="btn" href="{data.back}">Cancel</a>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
</form>
{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
