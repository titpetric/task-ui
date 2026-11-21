{if $data.can['prompt/edit'] && $data.can['prompt/view'] && $data.can['note/view'] && $data.claude}	<header class="top" data-prompt-folder="{data.prompt.folder}" data-prompt-dir="{data.prompt.dir}" data-prompt-run="{data.prompt.run}" data-prompt-label="{data.prompt.label}" data-prompt-href="{data.prompt.href}">
		<button type="button" class="btn-icon menu-toggle" id="menu-toggle" aria-label="Menu" aria-controls="side" aria-expanded="false"><i class="ic ic-menu" aria-hidden="true"></i></button>
		<h1><a href="{data.base}/"><span>TaskUI</span>{include components/logo.tpl}</a></h1>
		<form class="header-search" method="post" action="{base}/search/ask" role="search">
			<input type="search" name="q" aria-label="Search the documentation and job targets" autocomplete="off" spellcheck="false" placeholder="Search docs and jobs">
		</form>
{else}	<header class="top">
		<button type="button" class="btn-icon menu-toggle" id="menu-toggle" aria-label="Menu" aria-controls="side" aria-expanded="false"><i class="ic ic-menu" aria-hidden="true"></i></button>
		<h1><a href="{data.base}/"><span>TaskUI</span>{include components/logo.tpl}</a></h1>
		<span class="spacer"></span>
{/if}		{include components/notes-bell.tpl}
		<details class="user" id="user-menu">
			<summary aria-label="{data.user.name}, {data.user.role}" title="{data.user.name}"><span class="avatar role-{data.user.role}">{data.user.initials}</span></summary>
			<div class="user-menu">
				<div class="who"><b>{data.user.name}</b><span>{data.user.role}</span></div>
				<a href="{data.base}/users/{data.user.name}"><i class="ic ic-users" aria-hidden="true"></i>Profile</a>
				<a href="{data.base}/keys"><i class="ic ic-key" aria-hidden="true"></i>SSH keys</a>
				{if $data.settings}
				<a href="{data.base}/settings"><i class="ic ic-settings" aria-hidden="true"></i>Settings</a>
				{/if}
				{if $data.loginRequired}
				<form method="post" action="/logout"><button type="submit"><i class="ic ic-logout" aria-hidden="true"></i>Logout</button></form>
				{else}
				<div class="who"><span>auto login, no logout</span></div>
				{/if}
			</div>
		</details>
	</header>
