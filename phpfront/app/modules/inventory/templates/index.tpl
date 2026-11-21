{eval $data_head = $data.head}
{include components/page-head.tpl}
{if $data.off}<p class="empty-state">The hosts below are the inventory's own. Beszel isn't set up, so no host has measures.{if $data.integrations} Add its URL and token under <a href="{base}/settings/integrations#beszel">Settings, Integrations</a>.{/if}</p>
{/if}{if $data.form}<section class="block inventory-run">
	<h3>Run on the hosts</h3>
	<form data-inventory-run>
		<div class="run-commands">
{foreach $data.form.commands as $command}			<button type="button" class="btn btn-sm" data-command="{command}">{command}</button>
{/foreach}		</div>
		<textarea name="command" rows="2" required spellcheck="false" placeholder="a shell command to run on every picked host" aria-label="The command"></textarea>
		<div class="run-hosts" data-host-picker>
			<div class="host-tools">
				<span class="faint" data-host-count></span>
				<button type="button" class="btn btn-sm" data-hosts="all">All</button>
				<button type="button" class="btn btn-sm" data-hosts="none">None</button>
				<button type="button" class="btn btn-sm" data-hosts="invert">Invert</button>
			</div>
			<div class="host-picks">
{foreach $data.form.picks as $pick}				<label class="host-pick"><input type="checkbox" name="hosts" value="{pick.host}" checked><span>{pick.login}</span></label>
{/foreach}			</div>
		</div>
		<div class="form-foot note"><span></span><span class="actions">{if $data.form.summarize}<button class="btn" type="submit" name="summarize" value="1"><i class="ic ic-note" aria-hidden="true"></i>Run and summarize</button>{/if}<button class="btn btn-primary" type="submit"><i class="ic ic-play" aria-hidden="true"></i>Run</button></span></div>
	</form>
		<div class="inventory-results" data-inventory-results></div>
{if $data.prompt}{eval $data_prompt = $data.prompt}{include components/prompt.tpl}{/if}</section>
{/if}<section class="block" data-hosts-live>
{if $data.cluster}<div class="cluster">
	<h3 class="section-title">Cluster <small>{data.cluster.count} host{if $data.cluster.count !== 1}s{/if}, {data.cluster.up} up{if $data.cluster.down}, <span class="bad">{data.cluster.down} down</span>{/if}</small></h3>
	<div class="pools">
{foreach $data.cluster.pools as $pool}		<div class="pool">
			<div class="pool-head">
				<span class="pool-name">{pool.name}</span>
				<span class="pool-figure">{pool.figure}</span>{if $pool.share}<span class="pool-pct">{pool.pct}%</span>{/if}
			</div>
			<div class="pool-bar" role="img" aria-label="{pool.name}: {pool.figure}">{foreach $pool.segments as $segment}<span class="seg c{segment.color}" style="flex-grow: {segment.width}" title="{segment.title}"><i style="width: {segment.fill}%"></i></span>{/foreach}
			</div>
		</div>
{/foreach}	</div>
	<ul class="pool-legend">
{foreach $data.cluster.hosts as $host}		<li{if $host.left} class="left"{/if}><i class="swatch{if !$host.left} c{host.color}{/if}"></i>{host.name}</li>
{/foreach}	</ul>
{if $data.cluster.left}	<p class="faint pool-note">{data.cluster.left} host{if $data.cluster.left !== 1}s are down or send{else} is down or sends{/if} no measures, greyed in the legend and left out of the pools.</p>
{/if}{if $data.cluster.charts}	<div class="charts">{foreach $data.cluster.charts as $chart}{eval $data_line_chart = $chart}{include components/line-chart.tpl}{/foreach}	</div>
{/if}</div>
{/if}<h3>Hosts</h3>
<div class="table-wrap">
<table class="table hosts">
	<thead><tr><th>Host</th>{if $data.logins}<th>SSH</th>{/if}<th class="center">Cores</th><th>CPU</th><th colspan="2">Memory <small>GB</small></th><th colspan="2">Disk <small>GB</small></th><th class="actions" aria-label="Actions"></th></tr></thead>
	<tbody>
{foreach $data.rows as $row}		<tr{if $row.open} class="host-row" data-host-row="{row.name}" data-host-url="{row.href}" tabindex="0" aria-expanded="false"{/if}>
			<td class="key nowrap" title="{row.host}, {row.status}">{if $row.open}<i class="ic ic-chev chev" aria-hidden="true"></i>{/if}<i class="ic ic-server host-state {if $row.up}up{else}down{/if}" role="img" aria-label="{row.status}"></i> <a href="{row.href}">{row.name}</a></td>
{if $data.logins}			<td class="nowrap">{if $row.login}{eval $data_copy = $row.copy}{include components/copyable.tpl}{else}<span class="faint">-</span>{/if}</td>
{/if}			<td class="center">{if $row.cpus}{row.cpus}{else}<span class="faint">-</span>{/if}</td>
			<td>{eval $data_meter = $row.cpu}{include components/meter.tpl}			</td>
			<td>{eval $data_meter = $row.memory}{include components/meter.tpl}			</td>
			<td class="size faint nowrap">{if $row.measured}{row.memoryUsed} / {row.memoryTotal}{/if}</td>
			<td>{eval $data_meter = $row.disk}{include components/meter.tpl}			</td>
			<td class="size faint nowrap">{if $row.measured}{row.diskUsed} / {row.diskTotal}{/if}</td>
			<td class="actions">{if $row.drop}<form class="inline inventory-drop" method="post" action="{row.drop}" data-confirm="{row.asked}"><button class="btn-icon danger" type="submit" title="Remove from the inventory" aria-label="Remove {row.name} from the inventory"><i class="ic ic-trash" aria-hidden="true"></i></button></form>{/if}{eval $data_menu = $row.menu}{include components/menu.tpl}			</td>
		</tr>
{else}		<tr class="empty"><td colspan="9">No hosts yet.{if $data.add} Add one above.{/if}</td></tr>
{/foreach}	</tbody>
</table>
</div>
<p class="faint live-stamp">Updated {data.now}</p>
</section>
