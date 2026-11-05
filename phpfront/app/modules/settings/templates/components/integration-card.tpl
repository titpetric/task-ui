		<li class="wf-step integration-step kind-{data_card.id}" id="{data_card.id}">
			<span class="wf-node" aria-hidden="true">{data_card.number}</span>
			<div class="wf-card">
				<div class="wf-card-head">
					<button type="button" class="wf-toggle" data-card-toggle aria-expanded="false">
						<span class="wf-kind"><svg class="ic" aria-hidden="true"><use href="#i-plug"/></svg>Integration</span>
						<span class="wf-line">{data_card.title}</span>{if $data_card.configured}<span class="badge ok">set up</span>{else}<span class="badge">not set up</span>{/if}
						<svg class="ic wf-chev" aria-hidden="true"><use href="#i-chev"/></svg>
					</button>
					{eval $data_menu = $data_card.menu}{include components/menu.tpl}
				</div>
				<div class="wf-card-body">
					<div class="dim desc integration-desc">{data_card.desc|raw}</div>
					<form class="form integration-form" data-api="{data_card.api}" data-method="PUT" data-next="{data_card.next}">
						<div class="wf-fields">
{foreach $data_card.fields as $field}{eval $data_field = $field}{include components/integration-field.tpl}{/foreach}						</div>
{if $data_card.env}{eval $data_env = $data_card.env}{include components/integration-env.tpl}{/if}						<details class="integration-key">
							<summary>Get the {data_card.credential}</summary>
							<ol class="steps">
{foreach $data_card.steps as $step}								<li>{step|raw}</li>
{/foreach}							</ol>
{if $data_card.hook}							<p class="dim">{data_card.hook|raw}
							</p>
{/if}						</details>
						<div class="form-foot">
{if $data_card.runs}							<button type="button" class="btn" data-run data-folder="internal" data-job="@forgejo-sync" data-run-args='["--dry-run"]' data-title="internal / forgejo-sync --dry-run" title="Run @forgejo-sync --dry-run: the report alone, nothing changes">Dry run</button>
							<button type="button" class="btn" data-run data-folder="internal" data-job="@forgejo-sync" data-title="internal / forgejo-sync" title="Run @forgejo-sync now: workspaces, homes, projects and hooks">Sync now</button>
{/if}							<button class="btn btn-primary" type="submit" disabled>Save</button>
						</div>
					</form>
{if $data_card.sync}					<p class="dim integration-sync">{if $data_card.last}Last sync {eval $data_state = $data_card.last.state}{include components/state.tpl} {data_card.last.ago} by {data_card.last.who}: <a href="{data_card.last.href}">the report</a>.
{else}No sync ran yet: <b>Dry run</b> shows what one would do.
{/if}					</p>
{/if}				</div>
			</div>
		</li>
