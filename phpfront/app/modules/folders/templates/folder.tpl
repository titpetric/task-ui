{eval $data_head = $data.head}
{include components/page-head.tpl}
{if $data.findings}{eval $data_findings = $data.findings}{include components/pipeline-alert.tpl}
{/if}{if $data.git}{eval $data_git = $data.git}{include components/git-line.tpl}
{/if}{if $data.error}<div class="alert bad"><b>Could not read a source.</b> {data.error}</div>
{/if}{if $data.control}<section class="block" data-live>
	<h3>Jobs <small>a row per job source, then files, skills and terminals</small></h3>
{eval $data_control = $data.control}{include components/control-table.tpl}</section>
{/if}{if $data.nothing}<div class="alert">task-ui found nothing to run in this folder: no pipeline file, Makefile, package scripts or scripts folder.</div>
{/if}{if $data.prompt}{eval $data_prompt = $data.prompt}{include components/prompt.tpl}
{/if}{if $data.files}<section class="block">
	<h3>Files <small>{data.files.note}</small></h3>
{eval $data_files = $data.files.panel}{include components/file-table.tpl}</section>
{/if}{if $data.home}<section class="block folder-home" id="home">
	<h3>Home <small>{data.home.note}</small></h3>
{eval $data_files = $data.home.panel}{include components/file-table.tpl}</section>
{/if}{if $data.detail}<div class="grid-2">
<section class="block">
	<h3>Folder</h3>
	<div class="table-wrap">
	<table class="table table-kv">
		<tbody>
			<tr><th>Path</th><td><svg class="ic"><use href="#i-folder"/></svg> {data.detail.path}</td></tr>
{foreach $data.detail.sources as $source}			<tr><th>{source.kind}</th><td><svg class="ic"><use href="#i-file"/></svg> {source.file}{if $source.command} <span class="dim">runs with</span> <code>{source.command} &lt;target&gt;</code>{/if}{if $source.shows}{eval $data_source = $source.shows}{include components/show-source.tpl}{/if}</td></tr>
{/foreach}{if $data.detail.checkout}			<tr><th>Source</th><td><div class="mono-break">{data.detail.checkout.remote}</div><div class="dim">{data.detail.checkout.branch} <span class="faint">{data.detail.checkout.commit}</span> {data.detail.checkout.subject}{if $data.detail.checkout.ago} <span class="faint">{data.detail.checkout.ago}</span>{/if}</div><div class="faint">pulled before every run</div></td></tr>
{/if}{if $data.detail.space}			<tr><th>Checkout size</th><td>{eval $data_space = $data.detail.space}{include components/space.tpl}			</td></tr>
{/if}			<tr><th>Who can run</th><td>{data.detail.who}</td></tr>
			<tr><th>Hours</th><td>{eval $data_hours = $data.detail.hours}{include components/hours.tpl}</td></tr>
{if $data.detail.includes}			<tr><th>Includes</th><td>{foreach $data.detail.includes as $read}<div><svg class="ic"><use href="#i-file"/></svg> {read.file}{eval $data_source = $read.shows}{include components/show-source.tpl}</div>{/foreach}</td></tr>
{/if}			<tr><th>Scheduled</th><td>{if $data.detail.count}{data.detail.count} <a href="{data.detail.page}">{data.detail.schedules}</a>{else}<span class="faint">none</span>, <a href="{data.detail.page}">add one</a>{/if}</td></tr>
		</tbody>
	</table>
	</div>
</section>
</div>
{/if}
