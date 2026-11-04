{* The page that edits one user. The scope is {data}; the chrome is {chrome}. *}
{eval $data_head = $data.head}
{include components/page-head.tpl}

<section class="block" id="roles">
	<h3>Roles <small>what {data.name} is</small></h3>
{if $data.roles.edits}	<form class="form" data-api="{data.api}" data-method="PUT" data-next="{data.roles.next}">
		<div class="field">
			<label class="check"><input type="checkbox" name="admin"{if $data.roles.admin} checked{/if}> Administrative privileges (all)</label>
			<small class="faint">Every permission, for any role; the last admin keeps it.</small>
		</div>
		<div class="field">
			<label for="u-role">Role <span class="faint">without the admin flag</span></label>
			<select id="u-role" name="role">
{foreach $data.roles.options as $option}				<option value="{option}"{if $data.roles.role === $option} selected{/if}>{option}</option>
{/foreach}			</select>
		</div>
		<div class="field">
			<label class="check"><input type="checkbox" name="mcp"{if $data.roles.mcp} checked{/if}> MCP agent user</label>
			<small class="faint">A token reaches the MCP endpoint; an agent learns of work from the notes that name the user.</small>
		</div>
		<div class="field">
			<label class="check"><input type="checkbox" name="agent_reply"{if $data.roles.reply} checked{/if}> Run instance Claude when mentioned</label>
			<small class="faint">A note naming {data.name} starts the Notes agent as them; needs MCP agent user.</small>
		</div>
		<div class="form-foot">
			<a class="btn" href="{data.view}">Cancel</a>
			<button class="btn btn-primary" type="submit">Save</button>
		</div>
	</form>
{else}	<p>{if $data.roles.admin}Administrative privileges (all). {/if}Role: {data.roles.role}.{if $data.roles.mcp} MCP agent user.{/if}{if $data.roles.reply} Runs instance Claude when mentioned.{/if}</p>
{/if}</section>

<section class="block" id="folders">
	<h3>Folders <small>what {data.name} can see</small></h3>
{if $data.folders.edits}	<form class="form" data-api="{data.api}" data-method="PUT" data-next="{data.folders.next}">
{eval $data_pick = $data.folders.pick}{include components/folder-pick.tpl}		<div class="form-foot">
			<a class="btn" href="{data.view}">Cancel</a>
			<button class="btn btn-primary" type="submit">Save</button>
		</div>
	</form>
{elseif $data.folders.list}	<p>{foreach $data.folders.list as $index => $folder}{if $index}, {/if}<span class="mono">{folder}</span>{/foreach}</p>
{else}	<p>All folders.</p>
{/if}</section>

<section class="block" id="permissions">
	<h3>Permissions <small>{data.name}'s own rules, before the role's</small></h3>
	<p>Role: {if $data.permissions.links}<a href="{base}/settings/permissions#role-{data.permissions.role}">{data.permissions.role}</a>{else}{data.permissions.role}{/if}.{if $data.permissions.matrix.fixed} Admins can do everything; this can't change.{/if}</p>
{if !$data.permissions.matrix.fixed}	<p>An allow or a deny here is {data.name}'s own and decides the rule for them. Inherit, the empty box, leaves it to the role's rules, where a deny denies the action even where another rule allows it. The mark after each box shows the result.</p>
{/if}{eval $data_matrix = $data.permissions.matrix}{include components/matrix.tpl}
</section>
{if $data.keys}{eval $data_user_keys = $data.keys}{include components/user-keys.tpl}{/if}
{eval $data_mcp = $data.mcp}{include components/mcp-access.tpl}
