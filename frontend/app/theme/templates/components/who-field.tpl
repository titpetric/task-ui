<div class="whofield{if $data_whofield.compact} compact{/if}" data-who-empty="{data_whofield.empty}">
	<div class="checks who-roles">
		<span class="faint">Roles</span>{foreach $data_whofield.roles as $role}
		<label class="check"><input type="checkbox" name="{data_whofield.name}" value="{role.name}" data-type="{data_whofield.type}"{if $role.checked} checked{/if}> {role.name} <small class="count">· {role.count}</small></label>{/foreach}
	</div>
	<div class="checks who-users">
		<span class="faint">Users</span>
		<ul class="who-list">{foreach $data_whofield.users as $user}
		<li><label class="check who-user" data-role="{user.role}"><input type="checkbox" name="{data_whofield.userName}" value="{user.name}" data-type="{data_whofield.type}"{if $user.checked} checked{/if}> {user.name} <small class="role">{user.role}</small> <small class="who-state"></small></label></li>{/foreach}{foreach $data_whofield.extra as $other}
		<li><label class="check who-user" data-role=""><input type="checkbox" name="{data_whofield.userName}" value="{other}" data-type="{data_whofield.type}" checked> {other} <small class="role">no such user</small> <small class="who-state"></small></label></li>{/foreach}
		</ul>
	</div>{if !$data_whofield.compact}
	<p class="dim who-summary"></p>{/if}
</div>