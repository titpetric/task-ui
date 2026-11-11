{eval $data_head = $data.head}
{include components/page-head.tpl}
<section class="block instance-section" id="general">
	<h3>General</h3>
	{include components/instance-head.tpl}
		<tr><td class="key">Base link</td><td class="value mono" data-label="Value">{data.site}</td><td class="var" data-label="Variable"><code>TASK_UI_URL</code></td><td class="desc" data-label="About"><span>{if $data.instance.url}The link pages and MCP clients use.{else}Empty: task-ui reads the link from each request.{/if}</span></td></tr>
		<tr><td class="key">Version</td><td class="value mono ro" data-label="Value">{chrome.version}</td><td class="var" data-label="Variable"><span class="faint">none</span></td><td class="desc" data-label="About"><span>The build that serves this page.</span></td></tr>
		<tr><td class="key">Listen</td><td class="value mono" data-label="Value">{data.instance.listen}</td><td class="var" data-label="Variable"><code>TASK_UI_LISTEN</code></td><td class="desc" data-label="About"><span>The address the server listens on.</span></td></tr>
		<tr><td class="key">Time zone</td><td class="value mono ro" data-label="Value">{if $data.instance.time_zone}{data.instance.time_zone}{else}<span class="faint">local</span>{/if}</td><td class="var" data-label="Variable"><code>TZ</code></td><td class="desc" data-label="About"><span>The clock of the server, which the hours of a folder and the schedules read.</span></td></tr>
		<tr><td class="key">Sign in</td><td class="value" data-label="Value">{if $data.instance.auto_login}<span class="badge bad">auto login</span> {eval $data_user = $data.signin}{include components/user.tpl}{elseif $chrome.loginRequired}<span class="badge ok">required</span>{else}<span class="badge">open</span>{/if}</td><td class="var" data-label="Variable"><code>TASK_UI_AUTOLOGIN</code></td><td class="desc" data-label="About"><span>Auto login signs everyone in as one user; for demos only.</span></td></tr>
		<tr><td class="key">Session lifetime</td><td class="value mono" data-label="Value">{if $data.instance.session_ttl}{data.instance.session_ttl}{else}<span class="faint">default</span>{/if}</td><td class="var" data-label="Variable"><code>TASK_UI_SESSION_TTL</code></td><td class="desc" data-label="About"><span>How far ahead of its last use a session stays signed in: one in use keeps this window ahead of it.</span></td></tr>
		<tr><td class="key">Session ceiling</td><td class="value mono" data-label="Value">{if $data.instance.session_max}{data.instance.session_max}{else}<span class="faint">default</span>{/if}</td><td class="var" data-label="Variable"><code>TASK_UI_SESSION_MAX</code></td><td class="desc" data-label="About"><span>How long a session can be extended from the sign in. At or below the lifetime nothing slides and a session ends one lifetime after signing in.</span></td></tr>
		<tr><td class="key">Password file</td><td class="value mono" data-label="Value">{if $data.instance.htpasswd}{data.instance.htpasswd}{else}<span class="faint">none</span>{/if}</td><td class="var" data-label="Variable"><span class="faint">none</span></td><td class="desc" data-label="About"><span>Who can sign in: <code>.htpasswd</code> in the data directory.</span></td></tr>
		<tr><td class="key">Roles</td><td class="value" data-label="Value"><a href="{data.users}">Users</a></td><td class="var" data-label="Variable"><span class="faint">none</span></td><td class="desc" data-label="About"><span>The admin flag, the role and MCP of each user live on their profile, and <code>task-ui user</code> sets them.</span></td></tr>
		<tr><td class="key">MCP</td><td class="value mono" data-label="Value">{data.site}/api/mcp</td><td class="var" data-label="Variable"><span class="faint">none</span></td><td class="desc" data-label="About"><span>Where MCP clients connect, with a user's token.</span></td></tr>
	{include components/instance-foot.tpl}
</section>

<section class="block instance-section" id="runs">
	<h3>Runs</h3>
	{include components/instance-head.tpl}
{if $data.instance.running}		<tr><td class="key">Parallel runs</td><td class="value mono ro" data-label="Value">{data.instance.backlog.parallel}<small class="faint">{data.instance.backlog.going} going, {data.instance.backlog.waiting} waiting</small></td><td class="var" data-label="Variable"><span class="faint">computed</span></td><td class="desc" data-label="About"><span>One run per CPU of the machine; a run over that waits, queued, in order.</span></td></tr>
		<tr><td class="key">Automatic backlog</td><td class="value mono ro" data-label="Value">{data.instance.backlog.auto}</td><td class="var" data-label="Variable"><code>backlog_length_auto</code></td><td class="desc" data-label="About"><span>20 runs per CPU: the length of the line while an admin sets no backlog length.</span></td></tr>
		<tr><td class="key">Backlog length</td><td class="value" data-label="Value">
{if $data.admin}			<form class="inline-setting" data-api="/api/instance" data-method="PUT" data-next="{data.backlog}">
				<input type="text" name="backlog_length" value="{if $data.instance.backlog.configured}{data.instance.backlog.configured}{/if}" placeholder="{data.instance.backlog.auto}" inputmode="numeric" autocomplete="off" spellcheck="false" aria-label="Backlog length, empty for the automatic {data.instance.backlog.auto}">
				<button class="btn btn-sm" type="submit">Save</button>
			</form>
{else}			<span class="mono">{if $data.instance.backlog.configured}{data.instance.backlog.configured}{else}<span class="faint">automatic</span>{/if}</span>
{/if}			<small class="faint">in effect: {data.instance.backlog.length}</small></td>
			<td class="var" data-label="Variable"><code>backlog_length</code></td><td class="desc" data-label="About"><span>How many runs may wait for a slot, up to 100000; empty is the automatic length. A start past it answers <code>429</code>.</span></td></tr>
{/if}		<tr><td class="key">Keep attachable</td><td class="value mono" data-label="Value">{if $data.instance.keep}{data.instance.keep}{else}<span class="faint">default</span>{/if}</td><td class="var" data-label="Variable"><code>TASK_UI_KEEP</code></td><td class="desc" data-label="About"><span>How long a finished run stays in memory to attach to; its row and recording stay.</span></td></tr>
		<tr><td class="key">Stop grace</td><td class="value mono" data-label="Value">{if $data.instance.grace}{data.instance.grace}{else}<span class="faint">default</span>{/if}</td><td class="var" data-label="Variable"><code>TASK_UI_GRACE</code></td><td class="desc" data-label="About"><span>How long a stopped run gets before the kill.</span></td></tr>
		<tr><td class="key">Artifact history</td><td class="value mono" data-label="Value">{if $data.instance.artifact_history}{data.instance.artifact_history} versions{else}every version{/if}</td><td class="var" data-label="Variable"><code>TASK_UI_ARTIFACT_HISTORY</code></td><td class="desc" data-label="About"><span>How many versions the artifact history keeps before it squashes to one.</span></td></tr>
		<tr><td class="key">Retention</td><td class="value" data-label="Value"><a href="{data.retention}">Data retention</a></td><td class="var" data-label="Variable"><span class="faint">none</span></td><td class="desc" data-label="About"><span>How long runs, recordings, prompt threads, workflow runs, artifact history and the secret read log stay.</span></td></tr>
	{include components/instance-foot.tpl}
</section>

{if $data.instance.parent_env}<section class="block instance-section" id="environment">
	<h3>Parent environment</h3>
	<p class="section-intro"><a href="https://github.com/titpetric/task-ui/blob/main/docs/guide/settings.md#secrets-in-scripts">Secrets in scripts</a></p>
{if $data.admin}	<form class="form parent-env" data-api="/api/instance" data-method="PUT" data-next="{data.forward}" data-parent-env>
		<div class="field" role="radiogroup" aria-labelledby="parent-env-label">
			<span class="label" id="parent-env-label">Forward <span class="faint">applies to the next run</span></span>
{foreach $data.modes as $mode}			<label class="check"><input type="radio" name="parent_env" value="{mode.value}"{if $mode.checked} checked{/if}> <strong>{mode.label}</strong></label>
			<small class="faint">{mode.note}</small>
{/foreach}		</div>
		<div class="form-foot">
			<button class="btn btn-primary" type="submit">Save</button>
		</div>
	</form>
	<p class="faint">Base variables: {foreach $data.base as $item}{item.sep}<code>{item.name}</code>{/foreach}. A name holding {foreach $data.words as $item}{item.sep}<code>{item.name}</code>{/foreach} is sensitive.</p>
{if $data.env}	<div class="table-wrap">
	<table class="table env-table cards">
		<thead><tr><th>Variable</th><th>Forwarded</th><th>Flag</th></tr></thead>
		<tbody>
{foreach $data.env as $row}		<tr data-env-none="{row.none}" data-env-auto="{row.auto}"><td class="mono">{row.name}</td>
			<td data-label="Forwarded" data-env-forwarded>{if $row.forwarded}<span class="badge ok">yes</span>{else}<span class="badge">no</span>{/if}</td>
			<td data-label="Flag">{if $row.sensitive}<span class="badge bad" title="The name looks like a credential's">sensitive</span>{/if}</td></tr>
{/foreach}		</tbody>
	</table>
	</div>
{/if}{else}	<p>In effect: <strong>{data.instance.parent_env}</strong>; an admin changes it here.</p>
{/if}</section>
{/if}
<section class="block instance-section" id="paths">
	<h3>Paths</h3>
	{include components/instance-head.tpl}
		<tr><td class="key">Workspace folder</td><td class="value mono" data-label="Value">{data.instance.root}</td><td class="var" data-label="Variable"><code>TASK_UI_WORKSPACE</code></td><td class="desc" data-label="About"><span>The folders with jobs to run, by default <code>workspaces</code> in the data directory.</span></td></tr>
		<tr><td class="key">Discovery</td><td class="value" data-label="Value">{if $data.instance.discover}<span class="badge ok">on</span>{else}<span class="badge">off</span>{/if}</td><td class="var" data-label="Variable"><code>TASK_UI_DISCOVER</code></td><td class="desc" data-label="About"><span>List every folder under the root with something to run.</span></td></tr>
		<tr><td class="key">Data</td><td class="value mono" data-label="Value">{data.instance.data_dir}</td><td class="var" data-label="Variable"><code>TASK_UI_DATA</code></td><td class="desc" data-label="About"><span>The database, recordings and artifacts.</span></td></tr>
		<tr><td class="key">Data on the host</td><td class="value mono" data-label="Value">{if $data.instance.host_data}{data.instance.host_data}{else}<span class="faint">the same</span>{/if}</td><td class="var" data-label="Variable"><code>TASK_UI_HOST_DATA</code></td><td class="desc" data-label="About"><span>The data directory as the host's docker daemon sees it; runs get it, and their own directory in <code>TASK_UI_HOST_PWD</code>.</span></td></tr>
		<tr><td class="key">Key folder</td><td class="value mono" data-label="Value">{data.instance.ssh_dir}</td><td class="var" data-label="Variable"><code>TASK_UI_SSH_DIR</code></td><td class="desc" data-label="About"><span>Every key of the instance, and the backups of rotated keys in <code>backup/</code>.</span></td></tr>
		<tr><td class="key">Compose project</td><td class="value mono" data-label="Value">{if $data.instance.compose_project}{data.instance.compose_project}{else}<span class="faint">none</span>{/if}</td><td class="var" data-label="Variable"><code>TASK_UI_COMPOSE_PROJECT</code></td><td class="desc" data-label="About"><span>The docker compose project task-ui runs in, whose up and down ask first.</span></td></tr>
	{include components/instance-foot.tpl}
</section>
{if $data.instance.measured}<section class="block instance-section" id="storage">
	<h3>Storage</h3>
	<div class="table-wrap">
	<table class="table storage-table cards">
		<colgroup><col class="c-item"><col class="c-where"><col class="c-size"><col class="c-files"></colgroup>
		<thead><tr><th>Item</th><th>Where</th><th class="num">Size</th><th class="num">Files</th></tr></thead>
		<tbody>
{foreach $data.storage.rows as $row}		<tr><td class="key">{row.name}</td><td class="where mono" data-label="Where">{row.where}</td><td class="num mono" data-label="Size">{if $row.pending}<span class="faint pending" aria-label="not measured yet"></span>{else}{row.size}{/if}</td><td class="num mono" data-label="Files">{if $row.pending}<span class="faint pending" aria-label="not measured yet"></span>{else}{row.files}{/if}</td></tr>
{/foreach}		</tbody>
		<tfoot><tr><th>Total</th><td></td><td class="num mono" data-label="Size">{if $data.storage.total.pending}<span class="faint pending" aria-label="not measured yet"></span>{else}{data.storage.total.size}{/if}</td><td class="num mono" data-label="Files">{if $data.storage.total.pending}<span class="faint pending" aria-label="not measured yet"></span>{else}{data.storage.total.files}{/if}</td></tr></tfoot>
	</table>
	</div>
</section>
{/if}
<section class="block instance-section" id="tooling">
	<h3>Tooling</h3>
{if $data.admin}	<form class="inline" data-api="/api/instance/tools/check" data-method="POST" data-next="{data.checked}"><button class="btn btn-sm" type="submit">Check all</button></form>
{/if}{foreach $data.tooling as $group}	<h4>{group.title}</h4>
	<div class="table-wrap">
	<table class="table toolset-table cards">
		<colgroup><col class="c-tool"><col class="c-version"><col class="c-path"><col class="c-use"><col class="c-actions"></colgroup>
		<thead><tr><th>Tool</th><th>Version</th><th>Path</th><th>Used for</th><th></th></tr></thead>
		<tbody>
{foreach $group.tools as $tool}		<tr><td class="key mono" data-label="Tool">{tool.name}</td><td class="mono" data-label="Version">{if $tool.version}{tool.version}{else}<span class="faint">unknown</span>{/if}</td><td class="path mono" data-label="Path">{tool.path}</td><td class="desc" data-label="Used for">{tool.description}</td><td class="actions">{if $data.admin}<form class="inline" data-api="/api/instance/tools/{tool.id}/check" data-method="POST" data-next="{data.probed}"><button class="btn btn-sm" type="submit">Check</button></form>{/if}</td></tr>
{/foreach}		</tbody>
	</table>
	</div>
{/foreach}</section>
{if $data.keys.shown}<section class="block instance-section" id="keys">
	<h3>Keys <small>in the key folder</small></h3>
	<p class="section-intro"><a href="{data.keys.every}">Every key</a></p>
	<div class="table-wrap fit">
	<table class="table instance-keys key-table cards">
		{include components/key-head.tpl}
		<tbody>
{if $data.keys.encrypt}{eval $data_key = $data.keys.encrypt}{include components/key-row.tpl}{else}			<tr><td class="key">Encryption key</td><td colspan="4" data-label="Key"><span class="badge bad">missing</span> <span class="faint">The integrations' settings and the secrets need one.</span></td>
				<td class="actions">{if $data.keys.edits}<form class="inline" data-api="/api/keys/encrypt/rotate" data-method="POST" data-next="{data.keys.next}" data-confirm="Generate the instance's encryption key?" data-confirm-ok="Generate"><button class="btn btn-sm" type="submit">Generate</button></form> <a class="btn btn-sm" href="{data.keys.upload}">Upload</a>{/if}</td></tr>
{/if}{if $data.keys.git}{eval $data_key = $data.keys.git}{include components/key-row.tpl}{else}			<tr><td class="key">Git key</td><td colspan="4" data-label="Key"><span class="faint">none: git uses ssh's own</span></td>
				<td class="actions">{if $data.keys.edits}<a class="btn btn-sm" href="{data.keys.every}">Choose</a>{/if}</td></tr>
{/if}		</tbody>
	</table>
	</div>
</section>
{/if}
<section class="block instance-section" id="integrations">
	<h3>Integrations</h3>
	{include components/instance-head.tpl}
		<tr><td class="key">Inventory</td><td class="value" data-label="Value">{if $data.instance.beszel}<span class="badge ok">on</span>{else}<span class="badge">off</span>{/if}</td><td class="var" data-label="Variable"><span class="faint">none</span></td><td class="desc" data-label="About"><span>Hosts from the Beszel <a href="{data.integrations}">integration</a>.</span></td></tr>
	{include components/instance-foot.tpl}
</section>
{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
