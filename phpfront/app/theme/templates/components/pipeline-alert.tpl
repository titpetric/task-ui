{* input scope: data_findings. See pipeline-alert.md. *}
<div class="alert bad pipeline-alert" role="alert">
	<b>{data_findings.pipelines} {data_findings.says} errors.</b> atkins lists no jobs of a pipeline its linter refuses; fix the file and they return.
	<ul class="pipeline-errors">
{foreach $data_findings.shown as $finding}{eval $data_finding = $finding}{include components/pipeline-error.tpl}
{/foreach}	</ul>
{if $data_findings.rest}	<details class="pipeline-more">
		<summary><svg class="ic chev"><use href="#i-chev"/></svg>Show {data_findings.rest} more</summary>
		<ul class="pipeline-errors">
{foreach $data_findings.more as $finding}{eval $data_finding = $finding}{include components/pipeline-error.tpl}
{/foreach}		</ul>
	</details>
{/if}</div>
