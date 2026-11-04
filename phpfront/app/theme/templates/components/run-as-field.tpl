{* input scope: data_runas. See run-as-field.md. *}{if $data_runas.options.users}	<div class="field">
		<label for="{data_runas.id}">Runs as</label>
		<select id="{data_runas.id}" name="user">
{eval $data_options = $data_runas.options}{include components/user-options.tpl}		</select>
		<small class="faint">The keys are the instance keys the user gets; the job's workspace adds its own keys.</small>
	</div>
{/if}