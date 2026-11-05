<div class="run-split">
	<button type="button" class="btn btn-primary btn-sm" data-run data-folder="{data_run.folder}"{if $data_run.dir} data-dir="{data_run.dir}"{/if} data-job="{data_run.job}" data-title="{data_run.what}"{if $data_run.args} data-args="{data_run.args}" title="Run with arguments after {data_run.args}"{/if}><svg class="ic"><use href="#i-play"/></svg>Run{if $data_run.takesArgs}...{/if}</button>
{if $data_run.menu}	<details class="more-menu run-menu">
		<summary class="btn btn-primary btn-sm run-caret" aria-haspopup="menu" aria-expanded="false" title="More ways to run {data_run.name}" aria-label="More ways to run {data_run.name}"><svg class="ic"><use href="#i-down"/></svg></summary>
		<div class="more-list">
			<div class="more-items" role="menu" aria-label="Run options">
{if $data_run.summarize}				<button type="button" class="more-item" role="menuitem" data-run data-summarize data-folder="{data_run.folder}"{if $data_run.dir} data-dir="{data_run.dir}"{/if} data-job="{data_run.job}" data-title="{data_run.what}"{if $data_run.args} data-args="{data_run.args}"{/if} title="Run and summarize"><svg class="ic"><use href="#i-summary"/></svg><span>Run and summarize</span></button>
{/if}{if $data_run.force}				<button type="button" class="more-item danger" role="menuitem" data-run data-force data-folder="{data_run.folder}"{if $data_run.dir} data-dir="{data_run.dir}"{/if} data-job="{data_run.job}" data-title="{data_run.place}" title="Run out of hours, with a reason"><svg class="ic"><use href="#i-play"/></svg><span>Run force</span></button>
{/if}			</div>
		</div>
	</details>
{/if}</div>
