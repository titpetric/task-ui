{* input scope: data. See header.md. *}
	<header class="top">
		<button type="button" class="btn-icon menu-toggle" id="menu-toggle" aria-label="Menu" aria-controls="side" aria-expanded="false"><svg class="ic"><use href="#i-menu"/></svg></button>
		<h1><a href="{data.base}/"><span>TaskUI</span>{include components/logo.tpl}</a></h1>
		<span class="spacer"></span>
		<details class="user" id="user-menu">
			<summary aria-label="{data.user.name}, {data.user.role}" title="{data.user.name}"><span class="avatar role-{data.user.role}">{data.user.initials}</span></summary>
			<div class="user-menu">
				<div class="who"><b>{data.user.name}</b><span>{data.user.role}</span></div>
				<a href="{data.base}/users/{data.user.name}"><svg class="ic"><use href="#i-users"/></svg>Profile</a>
				<a href="{data.base}/keys"><svg class="ic"><use href="#i-key"/></svg>SSH keys</a>
				{if $data.settings}
				<a href="{data.base}/settings"><svg class="ic"><use href="#i-settings"/></svg>Settings</a>
				{/if}
				<form method="post" action="/logout"><button type="submit"><svg class="ic"><use href="#i-logout"/></svg>Logout</button></form>
			</div>
		</details>
	</header>
