<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<meta name="robots" content="noindex, nofollow">
	<meta name="color-scheme" content="dark light">
	<meta name="theme-color" content="#f7f5f0" media="(prefers-color-scheme: light)">
	<meta name="theme-color" content="#07090c" media="(prefers-color-scheme: dark)">
	<title>{if $data.title}{data.title} | {/if}TaskUI</title>
	<link rel="icon" href="{data.assets}/theme/icon.svg" type="image/svg+xml">
	<link rel="stylesheet" href="{data.assets}/theme/vendor/xterm.css?v={data.assetVersion}">
	<link rel="stylesheet" href="{data.assets}/theme/vendor/mdedit.css?v={data.assetVersion}">
	<link rel="stylesheet" href="{data.assets}/theme/task-ui.css?v={data.assetVersion}">
	<link rel="stylesheet" href="{data.assets}/theme/voice/voice.css?v={data.assetVersion}">
{foreach $data.styles as $style}	<link rel="stylesheet" href="{style}">
{/foreach}</head>
