{if $data_tooling.scopes}<div class="tooling">
{foreach $data_tooling.scopes as $scope}	<section class="tooling-scope">
		<p class="tooling-sum"><span class="tooling-db" title="{scope.database}">&lrm;{scope.database}</span><span><span class="key">{scope.name}</span>{if $scope.repository} <span class="faint">{scope.repository}</span>{/if} · {scope.sum}</span></p>
		<div class="tooling-charts">
{foreach $scope.charts as $chart}			<figure class="{chart.class}">
				<figcaption class="tooling-title"><span>{chart.title}</span><small>{chart.stats}</small></figcaption>
				<div class="{chart.bars.class}" style="{chart.bars.style}" role="img" aria-label="{chart.bars.label}">
{foreach $chart.bars.slots as $bar}{if $bar.segments}					<span aria-label="{bar.title}">{foreach $bar.segments as $part}{if $part.title}<b class="{part.class}" style="{part.style}" data-tip="{part.title}"></b>{else}<b></b>{/if}{/foreach}</span>
{else}					<span class="{bar.class}" data-tip="{bar.title}" aria-label="{bar.title}"></span>
{/if}{/foreach}				</div>
				<div class="activity-axis">{foreach $chart.ticks as $tick}<span style="left: {tick.at}%">{tick.label}</span>{/foreach}</div>
				<ul class="tooling-legend">
{if $chart.legend}{foreach $chart.legend as $item}					<li><span class="swatch {item.color}"></span><span class="key" title="{item.name}">{item.name}</span><span class="nowrap">{item.total}</span><small class="faint nowrap">{item.runs}</small></li>
{/foreach}{else}					<li class="empty">Nothing ran in this window.</li>
{/if}				</ul>
			</figure>
{/foreach}		</div>
{if $scope.latest}		<div class="tooling-latest">
			<h4>Latest commands</h4>
			<ol>
{foreach $scope.latest as $ran}				<li><time datetime="{ran.stamp}" title="{ran.stamp}">{ran.when}</time><span class="swatch {ran.color}"></span><span class="key" title="{ran.title}">{ran.command}{if $ran.job} <span class="faint">{ran.job}</span>{/if}</span><span class="nowrap">{ran.took}</span>{if $ran.failed}<span class="bad" title="exit {ran.exit}">failed</span>{else}<span class="ok">ok</span>{/if}</li>
{/foreach}			</ol>
		</div>
{/if}{if $scope.commands}		<details class="tooling-details">
			<summary><svg class="ic chev"><use href="#i-chev"/></svg>{scope.toggle}</summary>
			<div class="table-wrap">
			<table class="table cards">
				<thead><tr><th>Command</th><th>Tool</th><th>Runs</th><th>Total</th><th>Average</th><th>Slowest</th><th>Failed</th><th>Last</th></tr></thead>
				<tbody>
{foreach $scope.commands as $row}					<tr>
						<td class="key">{row.command}</td>
						<td data-label="Tool">{row.tool}</td>
						<td data-label="Runs">{row.runs}</td>
						<td class="nowrap" data-label="Total">{row.total}</td>
						<td class="nowrap" data-label="Average">{row.average}</td>
						<td class="nowrap" data-label="Slowest">{row.max}</td>
						<td data-label="Failed">{if $row.failures}<span class="bad">{row.failures}</span>{else}0{/if}</td>
						<td class="nowrap" data-label="Last">{row.last}</td>
					</tr>
{/foreach}				</tbody>
			</table>
			</div>
		</details>
{/if}	</section>
{/foreach}</div>{/if}
