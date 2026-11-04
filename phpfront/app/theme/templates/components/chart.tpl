{* input scope: data_chart. See chart.md. *}
{if $data_chart.charts}<div class="activity">
	<p class="activity-sum">{data_chart.sum}{if $data_chart.failed} · <span class="bad">{data_chart.failed}</span>{/if}{if $data_chart.median} · {data_chart.median}{/if}<span class="activity-key"><i class="ok"></i>runs <i class="line avg"></i>average <i class="line max"></i>slowest <i class="line min"></i>fastest</span></p>
	<div class="activity-charts">
{foreach $data_chart.charts as $chart}		<figure class="{chart.class}">
			<figcaption class="activity-title"><span>{chart.title}</span>{if $chart.stats.text}<span class="faint" title="{chart.stats.title}">{chart.stats.text}</span>{/if}</figcaption>
			<div class="activity-bars" style="--n: {chart.n}" role="img" aria-label="{chart.label}">
{foreach $chart.bars as $bar}				<span class="{bar.class}" data-tip="{bar.tip}" aria-label="{bar.tip}">{if $bar.other}<i class="other" style="{bar.other}"></i>{/if}{if $bar.failed}<i class="bad" style="{bar.failed}"></i>{/if}{if $bar.ok}<i class="ok" style="{bar.ok}"></i>{/if}</span>
{/foreach}			</div>
{if $chart.lines}			<div class="activity-line durations" style="--n: {chart.n}">
				<svg viewBox="0 0 100 100" preserveAspectRatio="none" role="img" aria-label="{chart.lines.label}">
					<path class="max" d="{chart.lines.max}" vector-effect="non-scaling-stroke"></path>
					<path class="min" d="{chart.lines.min}" vector-effect="non-scaling-stroke"></path>
					<path class="avg" d="{chart.lines.avg}" vector-effect="non-scaling-stroke"></path>
				</svg>
				<div class="activity-hits">
{foreach $chart.bars as $bar}					<span class="{bar.class}" data-tip="{bar.hit}" aria-label="{bar.hit}"></span>
{/foreach}				</div>
			</div>
{/if}			<div class="activity-axis">{foreach $chart.ticks as $tick}<span style="left: {tick.at}%">{tick.label}</span>{/foreach}</div>
		</figure>
{/foreach}	</div>
{if $data_chart.durations}	<figure class="activity-chart durations">
		<div class="activity-line">
			<svg viewBox="0 0 100 100" preserveAspectRatio="none" role="img" aria-label="{data_chart.durations.label}">
				<path class="wash" d="{data_chart.durations.area}"></path>
{foreach $data_chart.durations.points as $point}{if $point.failed}				<line class="bad" x1="{point.at}" y1="0" x2="{point.at}" y2="100" vector-effect="non-scaling-stroke"></line>
{/if}{/foreach}				<path class="line" d="{data_chart.durations.line}" vector-effect="non-scaling-stroke"></path>
			</svg>
{foreach $data_chart.durations.points as $point}			<span class="{point.class}" style="{point.style}" data-tip="{point.title}" aria-label="{point.title}"></span>
{/foreach}			<span class="activity-longest">{data_chart.durations.longest}</span>
		</div>
		<figcaption><span>{data_chart.durations.caption}</span></figcaption>
	</figure>
{/if}</div>{/if}
