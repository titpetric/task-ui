{* input scope: none. See confirm.md. *}
	<dialog class="confirm" id="workflow-input-dialog" aria-labelledby="workflow-input-title">
		<form method="dialog" data-workflow-input>
			<h3 id="workflow-input-title">Run</h3>
			<p>Optional input for the workflow's prompt. Claude reads it after the prompt, under the line <code>Input:</code>.</p>
			<textarea name="input" rows="3" spellcheck="false" aria-label="Input for the prompt"></textarea>
			<div class="form-foot">
				<button type="button" class="btn" data-workflow-input-cancel>Cancel</button>
				<button type="submit" class="btn btn-primary">Run</button>
			</div>
		</form>
	</dialog>
	<dialog class="confirm" id="confirm-dialog" aria-labelledby="confirm-title">
		<h3 id="confirm-title">Are you sure?</h3>
		<p id="confirm-text"></p>
		<div class="form-foot">
			<button type="button" class="btn" id="confirm-cancel">Cancel</button>
			<button type="button" class="btn btn-danger" id="confirm-ok">Delete</button>
		</div>
	</dialog>
