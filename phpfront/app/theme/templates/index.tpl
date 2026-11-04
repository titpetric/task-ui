{* The page shell. Each component is handed its own slice of the chrome as
   {data} immediately before it is included; components/<name>.md states the
   scope each one reads.

   The shell reads four of the chrome itself, for the body: roles and
   workspaces, each the JSON of a map the script parses out of its attribute;
   user, for the name the script answers as the reader; and can, for which
   profile page a mention opens. *}
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
