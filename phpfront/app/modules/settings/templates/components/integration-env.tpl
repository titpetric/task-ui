{* input scope: data_env. See integration-env.md. *}						<p class="faint integration-env-note">A set API key sends every variable below; an empty value takes its default.</p>
						<div class="table-wrap integration-env">
							<table class="table">
								<thead><tr><th>Variable</th><th>Value</th></tr></thead>
								<tbody>
{foreach $data_env as $row}								<tr>
									<td><b class="mono">{row.env}</b><div class="faint">{row.help}</div></td>
{if $row.source}									<td><span class="faint">the API key</span>
{elseif $row.empty}									<td><span class="faint">empty</span>
{else}									<td>
									<input type="text" id="{row.id}" name="{row.name}" value="{row.value}" aria-label="{row.env}" autocomplete="off" spellcheck="false">
{/if}									</td>
								</tr>
{/foreach}								</tbody>
							</table>
						</div>
