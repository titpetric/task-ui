{if $data.can['prompt/edit']}	<header class="top" data-prompt-folder="{data.prompt.folder}" data-prompt-dir="{data.prompt.dir}" data-prompt-run="{data.prompt.run}" data-prompt-label="{data.prompt.label}" data-prompt-href="{data.prompt.href}">
		<button type="button" class="btn-icon menu-toggle" id="menu-toggle" aria-label="Menu" aria-controls="side" aria-expanded="false"><svg class="ic"><use href="#i-menu"/></svg></button>
		<h1><a href="{data.base}/"><span>TaskUI</span></a><button type="button" class="logo-toggle" id="header-toggle" aria-label="Prompt" aria-controls="header-answer" aria-expanded="false">{include components/logo.tpl}</button></h1>
		<input type="text" class="header-prompt" id="header-prompt" aria-label="Ask Claude" autocomplete="off" spellcheck="false">
{if $data.runAs.users}
		<span class="header-runas"><select class="header-prompt-user" id="header-prompt-user" aria-label="Runs as">{eval $data_options = $data.runAs}{include components/user-options.tpl}		</select></span>
{/if}		<div class="header-answer" id="header-answer" hidden></div>
{else}	<header class="top">
		<button type="button" class="btn-icon menu-toggle" id="menu-toggle" aria-label="Menu" aria-controls="side" aria-expanded="false"><svg class="ic"><use href="#i-menu"/></svg></button>
		<h1><a href="{data.base}/"><span>TaskUI</span>{include components/logo.tpl}</a></h1>
		<span class="spacer"></span>
{/if}		<details class="user" id="user-menu">
			<summary aria-label="{data.user.name}, {data.user.role}" title="{data.user.name}"><span class="avatar role-{data.user.role}">{data.user.initials}</span></summary>
			<div class="user-menu">
				<div class="who"><b>{data.user.name}</b><span>{data.user.role}</span></div>
				<a href="{data.base}/users/{data.user.name}"><svg class="ic"><use href="#i-users"/></svg>Profile</a>
				<a href="{data.base}/keys"><svg class="ic"><use href="#i-key"/></svg>SSH keys</a>
				{if $data.settings}
				<a href="{data.base}/settings"><svg class="ic"><use href="#i-settings"/></svg>Settings</a>
				{/if}
				{if $data.loginRequired}
				<form method="post" action="/logout"><button type="submit"><svg class="ic"><use href="#i-logout"/></svg>Logout</button></form>
				{else}
				<div class="who"><span>auto login, no logout</span></div>
				{/if}
			</div>
		</details>
	</header>
