{* input scope: data. See alerts.md. *}
{if $data.notice}
			<div class="alert ok" role="status">{data.notice}</div>
{/if}
{if $data.error}
			<div class="alert bad" role="alert">{data.error}</div>
{/if}
