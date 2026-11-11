{eval $data_head = $data.head}
{include components/page-head.tpl}

<div class="alert">
	Admins can do everything; you can't change that. <b>Allow All</b> sets every rule in its row, and a rule you set on its own overrides it. Folder access and each job's allow list still limit what a user can see and run.
</div>
{foreach $data.matrices as $matrix}<section class="block matrix-role" id="{matrix.id}">
	<h3><span class="avatar sm role-{matrix.title}" aria-hidden="true">{matrix.initials}</span> {matrix.title} {if $matrix.fixed}<small>can do everything</small>{elseif !$matrix.editable}<small>read only</small>{/if}</h3>
	<p class="role-members">
		<span class="label">Users</span>{foreach $matrix.users as $user} <a class="key-chip" href="{user.href}">{user.name}</a>{/foreach}{if !$matrix.users} <span class="faint">none</span>{/if}{if $data.addUser} <a href="{data.addUser}">Add user</a>{/if}
	</p>{if $data.viewsKeys}
	<p class="role-members">
		<span class="label">Keys</span>{foreach $matrix.keys as $key} <a class="key-chip" href="{key.href}">{key.name}</a>{/foreach}{if !$matrix.keys} <span class="faint">none</span>{/if}{if $data.generate} <a href="{data.generate}">Generate key</a>{/if}
	</p>{/if}{eval $data_matrix = $matrix.matrix}{include components/matrix.tpl}
</section>
{/foreach}{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
