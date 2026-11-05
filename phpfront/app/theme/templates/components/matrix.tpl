<form class="matrix-form{if $data_matrix.fixed} fixed{/if}"{if $data_matrix.editable} data-api="{data_matrix.api}" data-method="PUT" data-next="reload" data-notice="Permissions saved" data-matrix{/if}>
	<div class="table-wrap">
	<table class="table matrix{if $data_matrix.effective} effective{/if}">
		<thead><tr><th>Section</th>{foreach $data_matrix.rules as $rule}<th>{rule.title}</th>{/foreach}</tr></thead>
		<tbody>
{foreach $data_matrix.sections as $section}			<tr><th scope="row">{section.title}</th>{foreach $section.cells as $cell}<td>{eval $data_cell = $cell}{include components/tri.tpl}</td>{/foreach}</tr>
{/foreach}			<tr class="group"><th colspan="{data_matrix.columns}">Application</th></tr>
{foreach $data_matrix.permissions as $cell}			<tr><th scope="row">{cell.title}</th><td colspan="{data_matrix.span}">{eval $data_cell = $cell}{include components/tri.tpl}</td></tr>
{/foreach}		</tbody>
	</table>
	</div>
{if $data_matrix.editable}	<div class="form-foot">
		<span class="faint" data-matrix-changed>No changes yet.</span>
		<button class="btn btn-primary" type="submit">Save permissions</button>
	</div>
{/if}</form>
