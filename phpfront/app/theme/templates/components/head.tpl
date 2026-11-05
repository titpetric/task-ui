<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<meta name="robots" content="noindex, nofollow">
	<title>{if $data.title}{data.title} | {/if}TaskUI</title>
	<link rel="icon" href="{data.assets}/theme/icon.svg" type="image/svg+xml">
	<link rel="stylesheet" href="{data.assets}/theme/vendor/xterm.css?v={data.assetVersion}">
	<link rel="stylesheet" href="{data.assets}/theme/task-ui.css?v={data.assetVersion}">
{foreach $data.styles as $style}	<link rel="stylesheet" href="{style}">
{/foreach}</head>
