{eval $data_head = $data.head}
{include components/page-head.tpl}

<section class="block key-overview">
	<h3>Scopes <small>the keys of the instance and each workspace</small></h3>
	<div class="table-wrap">
	<table class="table key-purposes">
		<thead><tr><th>Scope</th><th>Variables</th><th>Encryption key</th><th>SSH keys</th></tr></thead>
		<tbody>
			<tr class="key-instance" id="instance">
				<td><b>Instance</b><div class="faint">every run, by policy</div></td>
				<td><a href="{base}/settings/integrations">integrations</a><div class="faint">their settings</div></td>
				<td>
{if $data.scopes.instance.set}					<a class="badge ok" href="{data.scopes.instance.anchor}" title="Its [...] in Every key rotates it">set</a>{if $data.scopes.instance.backup} <span class="backup-active" title="The previous material serves beside the key until you delete the backup">BACKUP ACTIVE: {data.scopes.instance.backup}</span>{/if}
					<div class="key-fp faint" title="{data.scopes.instance.name}">{data.scopes.instance.fingerprint}</div>
{else}					<span class="badge bad">missing</span> <span class="faint">task-ui creates it on start</span>
{/if}				</td>
				<td>
{if $data.scopes.instance.count}					<a href="#instance-ssh">{data.scopes.instance.count} key{if $data.scopes.instance.count !== 1}s{/if}</a>
					<div class="key-names">{foreach $data.scopes.instance.chips as $chip}<a class="key-chip" href="{chip.href}" title="{chip.title}">{chip.name}</a>{/foreach}</div>
{else}					<span class="faint">none: runs get no SSH key from the instance</span>
{/if}				</td>
			</tr>
{foreach $data.scopes.workspaces as $ws}			<tr id="ws-{ws.folder}">
				<td>
					<a href="{ws.href}"><b>{ws.title}</b></a>{if $ws.badge} <span class="badge">{ws.badge}</span>{/if}
					<div class="faint key-projects">{if $ws.projects}{ws.projects}{else}no projects{/if}</div>
				</td>
				<td>{if $ws.secrets}<a href="{ws.secretsHref}">{ws.secrets} variable{if $ws.secrets !== 1}s{/if}</a>{else}<a class="faint" href="{ws.secretsHref}">none</a>{/if}</td>
				<td>
					<span class="faint">the instance's</span>
				</td>
				<td>
{if $ws.ssh}					<a href="{ws.keysHref}">{ws.count} key{if $ws.count !== 1}s{/if}</a>
					<div class="key-names">{foreach $ws.ssh as $chip}<a class="key-chip" href="{chip.href}" title="{chip.title}">{chip.name}</a>{/foreach}</div>
{else}					<a class="faint" href="{ws.keysHref}">none</a>
					<div class="faint">Its runs get only the instance SSH keys their policies grant.</div>
{/if}				</td>
			</tr>
{/foreach}		</tbody>
	</table>
	</div>
</section>

<section class="block" id="instance-ssh">
{eval $data_block = $data.block}{include components/block-head.tpl}
{if $data.keys}	<p class="dim">Every key lives in the key folder as <code>&lt;name&gt;.key</code> and <code>&lt;name&gt;.pub</code>, the name ending in its algorithm and size. An instance SSH key goes to the runs of the roles and users it links, through their ssh agent, and the inventory signs in to hosts with it; a workspace's SSH key goes to the runs of that workspace. The git key, which a key's page sets, pulls and clones the checkouts. An encryption key never leaves task-ui.</p>
{/if}{eval $data_keys = $data.keys}{include components/key-table.tpl}</section>

<section class="block">
	<h3>Who gets which key <small>through the ssh agent of their runs</small></h3>
	<div class="table-wrap">
	<table class="table key-holders">
		<thead><tr><th>Policy or user</th><th>Keys</th></tr></thead>
		<tbody>
{foreach $data.holders as $holder}			<tr>
{if $holder.user}				<td class="nowrap">{eval $data_user = $holder.user}{include components/user.tpl} <span class="faint">{holder.role}</span></td>
{else}				<td class="nowrap"><span class="badge">{holder.role}</span> <span class="faint">every {holder.role}</span></td>
{/if}				<td>{foreach $holder.direct as $chip}<a class="key-chip" href="{chip.href}" title="{chip.title}">{chip.name}</a>{/foreach}{foreach $holder.viaRole as $chip}<a class="key-chip faint" href="{chip.href}" title="{chip.title}">{chip.name} <small>via role</small></a>{/foreach}{if !$holder.direct && !$holder.viaRole}<span class="faint">none</span>{/if}
				</td>
			</tr>
{/foreach}		</tbody>
	</table>
	</div>
</section>
{foreach $data.folders as $folder}<div class="alert signal"><b>{folder.project}</b> has private keys in <b>{folder.dir}</b>: {folder.keys}. Its runs don't get them. Upload a key the project needs as an instance key, then link it to roles or users.</div>
{/foreach}{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
