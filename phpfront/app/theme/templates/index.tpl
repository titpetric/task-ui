<!DOCTYPE html>
<html lang="en">
{eval $data = $chrome}
{include components/head.tpl}
<body data-roles='{chrome.roles}' data-workspaces='{chrome.workspaces}'{if $chrome.user} data-me="{chrome.user.name}"{/if} data-user-menu="{if $chrome.can['user/edit']}edit{elseif $chrome.can['user/view']}view{/if}">
{include components/icons.tpl}
{if $chrome.user}
<a class="skip-link" href="#main">Skip to content</a>
{eval $data = $chrome}
{include components/header.tpl}
	<div class="shell">
		<div class="menu-backdrop" id="menu-backdrop"></div>
		<aside class="side" id="side">
{eval $data = $chrome.sidebar}
{include components/sidebar-menu.tpl}
{eval $data = $chrome.folders}
{include components/sidebar-folders.tpl}
		</aside>
		<main id="main" tabindex="-1">
{eval $data = $chrome}
{include components/crumbs.tpl}
{include components/alerts.tpl}
{content|raw}
		</main>
{eval $data = $chrome.recents}
{include components/recents.tpl}
	</div>
{include components/dialog.tpl}
{include components/confirm.tpl}
{include components/voice.tpl}
{else}
	<header class="top login-top">
		<h1><a href="{base}/"><span>TaskUI</span>{include components/logo.tpl}</a></h1>
	</header>
{content|raw}
{/if}
	<div class="toasts" id="toasts" role="status" aria-live="polite"></div>
{eval $data = $chrome}
{include components/footer.tpl}
</body>
</html>
