<?php $_v=&$this->vars;?>

<!DOCTYPE html>
<html lang="en">
<?php $_v['data'] = $_v['chrome'];?>


<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<meta name="robots" content="noindex, nofollow">
	<title><?php if($_v['data']['title']){ echo htmlspecialchars($_v['data']['title'], ENT_QUOTES);?>
 | <?php }?>TaskUI</title>
	<link rel="icon" href="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/icon.svg" type="image/svg+xml">
	<link rel="stylesheet" href="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/vendor/xterm.css?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
">
	<link rel="stylesheet" href="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/task-ui.css?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
">
<?php if(!empty($_v['data']['styles']))foreach($_v['data']['styles'] as $_v['style']){?>	<link rel="stylesheet" href="<?php echo htmlspecialchars($_v['style'], ENT_QUOTES);?>
">
<?php }?></head>

<body data-roles='<?php echo htmlspecialchars($_v['chrome']['roles'], ENT_QUOTES);?>
' data-workspaces='<?php echo htmlspecialchars($_v['chrome']['workspaces'], ENT_QUOTES);?>
'<?php if($_v['chrome']['user']){?>
 data-me="<?php echo htmlspecialchars($_v['chrome']['user']['name'], ENT_QUOTES);?>
"<?php }?> data-user-menu="<?php if($_v['chrome']['can']['user/edit']){?>
edit<?php }elseif($_v['chrome']['can']['user/view']){?>
view<?php }?>">

<svg xmlns="http://www.w3.org/2000/svg" style="display:none">
	<symbol id="i-project" viewBox="0 0 16 16"><circle cx="4.5" cy="3.5" r="1.9" fill="none" stroke="currentColor" stroke-width="1.3"/><circle cx="4.5" cy="12.5" r="1.9" fill="none" stroke="currentColor" stroke-width="1.3"/><circle cx="11.5" cy="5.5" r="1.9" fill="none" stroke="currentColor" stroke-width="1.3"/><path d="M4.5 5.4v5.2M11.5 7.4c0 2.6-2.2 3.1-5.2 3.9" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/></symbol>
	<symbol id="i-folder" viewBox="0 0 16 10"><path d="M.65 1.6A1 1 0 0 1 1.65.6h3.6l1.3 1.3h7.8a1 1 0 0 1 1 1v5.5a1 1 0 0 1-1 1H1.65a1 1 0 0 1-1-1z" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linejoin="round"/></symbol>
	<symbol id="i-folder-open" viewBox="0 0 16 10"><path d="M.65 8.4V1.6A1 1 0 0 1 1.65.6h3.6l1.3 1.3h6.8a1 1 0 0 1 1 1v.7" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linejoin="round"/><path d="M2.6 4.1h12.2a.6.6 0 0 1 .57.8l-1.2 3.9a1 1 0 0 1-.96.7H1.3a.6.6 0 0 1-.57-.8l1-3.9a1 1 0 0 1 .87-.7z" fill="currentColor" opacity=".85"/></symbol>
	<symbol id="i-grid" viewBox="0 0 16 16"><rect x="2" y="2" width="5" height="5" rx="1" fill="none" stroke="currentColor" stroke-width="1.2"/><rect x="9" y="2" width="5" height="5" rx="1" fill="none" stroke="currentColor" stroke-width="1.2"/><rect x="2" y="9" width="5" height="5" rx="1" fill="none" stroke="currentColor" stroke-width="1.2"/><rect x="9" y="9" width="5" height="5" rx="1" fill="none" stroke="currentColor" stroke-width="1.2"/></symbol>
	<symbol id="i-list" viewBox="0 0 16 16"><path d="M5 4h9M5 8h9M5 12h9" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/><circle cx="2.5" cy="4" r="1" fill="currentColor"/><circle cx="2.5" cy="8" r="1" fill="currentColor"/><circle cx="2.5" cy="12" r="1" fill="currentColor"/></symbol>
	<symbol id="i-box" viewBox="0 0 16 16"><path d="M2 5.5 8 2.5l6 3v5l-6 3-6-3zM2 5.5l6 3 6-3M8 8.5v5" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linejoin="round"/></symbol>
	<symbol id="i-flow" viewBox="0 0 16 16"><circle cx="3" cy="4" r="1.6" fill="none" stroke="currentColor" stroke-width="1.2"/><circle cx="13" cy="8" r="1.6" fill="none" stroke="currentColor" stroke-width="1.2"/><circle cx="3" cy="12" r="1.6" fill="none" stroke="currentColor" stroke-width="1.2"/><path d="M4.6 4h3.4a2 2 0 0 1 2 2v.4M10 9.6v.4a2 2 0 0 1-2 2H4.6M10 6.4 11.4 8" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-prompt" viewBox="0 0 16 16"><path d="M2.5 3.5a1 1 0 0 1 1-1h9a1 1 0 0 1 1 1v6.5a1 1 0 0 1-1 1H7l-3 2.5v-2.5h-.5a1 1 0 0 1-1-1z" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linejoin="round"/><path d="M5.5 5.5 7 7 5.5 8.5M8.5 8.5h2" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-summary" viewBox="0 0 16 16"><path d="M11.5 2.5h-7l4 5.5-4 5.5h7" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-grip" viewBox="0 0 16 16"><circle cx="6" cy="3.5" r="1.1" fill="currentColor"/><circle cx="10" cy="3.5" r="1.1" fill="currentColor"/><circle cx="6" cy="8" r="1.1" fill="currentColor"/><circle cx="10" cy="8" r="1.1" fill="currentColor"/><circle cx="6" cy="12.5" r="1.1" fill="currentColor"/><circle cx="10" cy="12.5" r="1.1" fill="currentColor"/></symbol>
	<symbol id="i-plus" viewBox="0 0 16 16"><path d="M8 3v10M3 8h10" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/></symbol>
	<symbol id="i-server" viewBox="0 0 16 16"><rect x="2.5" y="2.5" width="11" height="4.5" rx="1" fill="none" stroke="currentColor" stroke-width="1.2"/><rect x="2.5" y="9" width="11" height="4.5" rx="1" fill="none" stroke="currentColor" stroke-width="1.2"/><circle cx="5" cy="4.75" r=".8" fill="currentColor"/><circle cx="5" cy="11.25" r=".8" fill="currentColor"/></symbol>
	<symbol id="i-plug" viewBox="0 0 16 16"><path d="M6 1.5v3M10 1.5v3M4 4.5h8v3a4 4 0 0 1-8 0zM8 11.5v3" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-note" viewBox="0 0 16 16"><path d="M3 2.5h10v7.5l-3.5 3.5H3zM13 10H9.5v3.5M5.5 5.5h5M5.5 8h3" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linejoin="round" stroke-linecap="round"/></symbol>
	<symbol id="i-clock" viewBox="0 0 16 16"><circle cx="8" cy="8" r="6" fill="none" stroke="currentColor" stroke-width="1.2"/><path d="M8 4.5V8l2.5 1.5" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></symbol>
	<symbol id="i-pulse" viewBox="0 0 16 16"><path d="M1.5 8h3L6 4.5 8.5 11.5 10.5 8h4" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-users" viewBox="0 0 16 16"><circle cx="6" cy="5.5" r="2.5" fill="none" stroke="currentColor" stroke-width="1.2"/><path d="M1.5 13.5c0-2.5 2-4 4.5-4s4.5 1.5 4.5 4M11 3.5a2.5 2.5 0 0 1 0 4.5M12 9.5c1.5.4 2.5 1.8 2.5 4" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></symbol>
	<symbol id="i-settings" viewBox="0 0 16 16"><circle cx="8" cy="8" r="2.2" fill="none" stroke="currentColor" stroke-width="1.2"/><path d="M8 1.5v2M8 12.5v2M1.5 8h2M12.5 8h2M3.4 3.4l1.4 1.4M11.2 11.2l1.4 1.4M3.4 12.6l1.4-1.4M11.2 4.8l1.4-1.4" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></symbol>
	<symbol id="i-logout" viewBox="0 0 16 16"><path d="M6 2.5H3.5a1 1 0 0 0-1 1v9a1 1 0 0 0 1 1H6M10 11l3-3-3-3M13 8H6" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-copy" viewBox="0 0 16 16"><rect x="5.5" y="5.5" width="8" height="8" rx="1" fill="none" stroke="currentColor" stroke-width="1.2"/><path d="M10.5 5.5v-2a1 1 0 0 0-1-1h-6a1 1 0 0 0-1 1v6a1 1 0 0 0 1 1h2" fill="none" stroke="currentColor" stroke-width="1.2"/></symbol>
	<symbol id="i-rotate" viewBox="0 0 16 16"><path d="M13 8a5 5 0 1 1-1.5-3.6M13 2.5v3h-3" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-trash" viewBox="0 0 16 16"><path d="M2.5 4h11M6 4V2.5h4V4M4 4l.7 9.5h6.6L12 4M6.5 6.5v5M9.5 6.5v5" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-edit" viewBox="0 0 16 16"><path d="M11.5 2.5l2 2L6 12H4v-2zM10 4l2 2" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linejoin="round"/><path d="M2.5 13.5h11" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></symbol>
	<symbol id="i-key" viewBox="0 0 16 16"><circle cx="5.5" cy="10.5" r="3" fill="none" stroke="currentColor" stroke-width="1.2"/><path d="M7.7 8.3 13.5 2.5M11 5l2 2M9.5 6.5l1.5 1.5" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></symbol>
	<symbol id="i-secret" viewBox="0 0 16 16"><rect x="3" y="7" width="10" height="7" rx="1" fill="none" stroke="currentColor" stroke-width="1.2"/><path d="M5.5 7V5a2.5 2.5 0 0 1 5 0v2" fill="none" stroke="currentColor" stroke-width="1.2"/></symbol>
	<symbol id="i-file" viewBox="0 0 16 16"><path d="M4 1.5h5l3.5 3.5v9.5h-8.5zM9 1.5V5h3.5" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linejoin="round"/></symbol>
	<symbol id="i-script" viewBox="0 0 16 16"><path d="M5 4 1.8 8 5 12M11 4l3.2 4L11 12M9.4 2.5l-2.8 11" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-terminal" viewBox="0 0 16 16"><rect x="1.5" y="2.5" width="13" height="11" rx="1.5" fill="none" stroke="currentColor" stroke-width="1.2"/><path d="M4.5 6l2 2-2 2M8 10.5h3.5" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-play" viewBox="0 0 16 16"><path d="M4.5 2.5v11l8-5.5z" fill="currentColor"/></symbol>
	<symbol id="i-stop" viewBox="0 0 16 16"><rect x="3.5" y="3.5" width="9" height="9" rx="1" fill="currentColor"/></symbol>
	<symbol id="i-pin" viewBox="0 0 16 16"><path d="M6 1.5h4M7 1.5v4.2L4.5 8.5h7L9 5.7V1.5M8 8.5v6" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-eye" viewBox="0 0 16 16"><path d="M1.5 8s2.5-4.5 6.5-4.5S14.5 8 14.5 8 12 12.5 8 12.5 1.5 8 1.5 8z" fill="none" stroke="currentColor" stroke-width="1.2"/><circle cx="8" cy="8" r="2" fill="none" stroke="currentColor" stroke-width="1.2"/></symbol>
	<symbol id="i-menu" viewBox="0 0 16 16"><path d="M2.5 4h11M2.5 8h11M2.5 12h11" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/></symbol>
	<symbol id="i-expand" viewBox="0 0 16 16"><path d="M9.5 2.5h4v4M13.5 2.5 9 7M6.5 13.5h-4v-4M2.5 13.5 7 9" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-collapse" viewBox="0 0 16 16"><path d="M13 7H9V3M9 7l4.5-4.5M3 9h4v4M7 9l-4.5 4.5" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-more" viewBox="0 0 16 16"><circle cx="3.5" cy="8" r="1.3" fill="currentColor"/><circle cx="8" cy="8" r="1.3" fill="currentColor"/><circle cx="12.5" cy="8" r="1.3" fill="currentColor"/></symbol>
	<symbol id="i-chev" viewBox="0 0 16 16"><path d="M6 3.5 10.5 8 6 12.5" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-up" viewBox="0 0 16 16"><path d="M8 13.5v-11M3.5 7 8 2.5 12.5 7" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-down" viewBox="0 0 16 16"><path d="M8 2.5v11M3.5 9 8 13.5 12.5 9" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-download" viewBox="0 0 16 16"><path d="M8 2.5v8M4.5 7 8 10.5 11.5 7M2.5 13.5h11" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-upload" viewBox="0 0 16 16"><path d="M8 10.5v-8M4.5 6 8 2.5 11.5 6M2.5 13.5h11" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
</svg>

<?php if($_v['chrome']['user']){?>

<a class="skip-link" href="#main">Skip to content</a>
<?php $_v['data'] = $_v['chrome'];?>


<?php if($_v['data']['can']['prompt/edit']){?>
	<header class="top" data-prompt-folder="<?php echo htmlspecialchars($_v['data']['prompt']['folder'], ENT_QUOTES);?>
" data-prompt-dir="<?php echo htmlspecialchars($_v['data']['prompt']['dir'], ENT_QUOTES);?>
" data-prompt-run="<?php echo htmlspecialchars($_v['data']['prompt']['run'], ENT_QUOTES);?>
" data-prompt-label="<?php echo htmlspecialchars($_v['data']['prompt']['label'], ENT_QUOTES);?>
" data-prompt-href="<?php echo htmlspecialchars($_v['data']['prompt']['href'], ENT_QUOTES);?>
">
		<button type="button" class="btn-icon menu-toggle" id="menu-toggle" aria-label="Menu" aria-controls="side" aria-expanded="false"><svg class="ic"><use href="#i-menu"/></svg></button>
		
		<h1><a href="<?php echo htmlspecialchars($_v['data']['base'], ENT_QUOTES);?>
/"><span>TaskUI</span></a><button type="button" class="logo-toggle" id="header-toggle" aria-label="Prompt" aria-controls="header-answer" aria-expanded="false"><svg class="logo" viewBox="0 0 32 32" aria-hidden="true"><rect width="32" height="32" rx="3"/><path class="prompt" d="M8 10l6 6-6 6"/><path class="cursor" d="M16 23h8"/></svg>
</button></h1>
		<input type="text" class="header-prompt" id="header-prompt" aria-label="Ask Claude" autocomplete="off" spellcheck="false">
<?php if($_v['data']['runAs']['users']){?>
		
		<span class="header-runas"><select class="header-prompt-user" id="header-prompt-user" aria-label="Runs as"><?php $_v['data_options'] = $_v['data']['runAs'];if(!empty($_v['data_options']['users']))foreach($_v['data_options']['users'] as $_v['option']){?>			<option value="<?php echo htmlspecialchars($_v['option']['name'], ENT_QUOTES);?>
"<?php if($_v['option']['name'] == $_v['data_options']['selected']){?>
 selected<?php }?> data-keys="<?php echo htmlspecialchars($_v['option']['keys'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['option']['name'], ENT_QUOTES); if(!$_v['data_options']['short']){?>
 (<?php echo htmlspecialchars($_v['option']['role'], ENT_QUOTES);?>
)<?php }?> · <?php echo htmlspecialchars($_v['option']['said'], ENT_QUOTES);?>
</option>
<?php }?>		</select></span>
<?php }?>		<div class="header-answer" id="header-answer" hidden></div>
<?php }else{?>
	<header class="top">
		<button type="button" class="btn-icon menu-toggle" id="menu-toggle" aria-label="Menu" aria-controls="side" aria-expanded="false"><svg class="ic"><use href="#i-menu"/></svg></button>
		<h1><a href="<?php echo htmlspecialchars($_v['data']['base'], ENT_QUOTES);?>
/"><span>TaskUI</span><svg class="logo" viewBox="0 0 32 32" aria-hidden="true"><rect width="32" height="32" rx="3"/><path class="prompt" d="M8 10l6 6-6 6"/><path class="cursor" d="M16 23h8"/></svg>
</a></h1>
		<span class="spacer"></span>
<?php }?>		<details class="user" id="user-menu">
			<summary aria-label="<?php echo htmlspecialchars($_v['data']['user']['name'], ENT_QUOTES);?>
, <?php echo htmlspecialchars($_v['data']['user']['role'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['data']['user']['name'], ENT_QUOTES);?>
"><span class="avatar role-<?php echo htmlspecialchars($_v['data']['user']['role'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data']['user']['initials'], ENT_QUOTES);?>
</span></summary>
			<div class="user-menu">
				<div class="who"><b><?php echo htmlspecialchars($_v['data']['user']['name'], ENT_QUOTES);?>
</b><span><?php echo htmlspecialchars($_v['data']['user']['role'], ENT_QUOTES);?>
</span></div>
				<a href="<?php echo htmlspecialchars($_v['data']['base'], ENT_QUOTES);?>
/users/<?php echo htmlspecialchars($_v['data']['user']['name'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-users"/></svg>Profile</a>
				<a href="<?php echo htmlspecialchars($_v['data']['base'], ENT_QUOTES);?>
/keys"><svg class="ic"><use href="#i-key"/></svg>SSH keys</a>
				<?php if($_v['data']['settings']){?>

				<a href="<?php echo htmlspecialchars($_v['data']['base'], ENT_QUOTES);?>
/settings"><svg class="ic"><use href="#i-settings"/></svg>Settings</a>
				<?php }?>
				<?php if($_v['data']['loginRequired']){?>

				<form method="post" action="/logout"><button type="submit"><svg class="ic"><use href="#i-logout"/></svg>Logout</button></form>
				<?php }else{?>

				<div class="who"><span>auto login, no logout</span></div>
				<?php }?>
			</div>
		</details>
	</header>

	<div class="shell">
		<div class="menu-backdrop" id="menu-backdrop"></div>
		<aside class="side" id="side">
<?php $_v['data'] = $_v['chrome']['sidebar'];?>


			<nav class="menu">
<?php if(!empty($_v['data']))foreach($_v['data'] as $_v['item']){?>
				<a href="<?php echo htmlspecialchars($_v['item']['href'], ENT_QUOTES);?>
" class="<?php if($_v['item']['active']){?>
active<?php }?>"><svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
"/></svg><span><?php echo htmlspecialchars($_v['item']['title'], ENT_QUOTES);?>
</span><?php if($_v['item']['pulse']){?>
<i class="dot<?php if($_v['chrome']['running']){?>
 running<?php }?>" data-any-running></i><?php }?></a>
<?php if($_v['item']['flows'] && $_v['chrome']['workflows']){ $_v['data_workflows'] = $_v['chrome']['workflows'];?>
				<div class="flows">
<?php if(!empty($_v['data_workflows']))foreach($_v['data_workflows'] as $_v['flow']){ if(!$_v['flow']['hidden']){?>
					<div class="flow<?php if($_v['flow']['active']){?>
 active<?php }?>">
						<i class="shebang<?php if($_v['flow']['running']){?>
 running<?php }?>" aria-hidden="true">#!</i>
						<a class="name" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/workflows/<?php echo htmlspecialchars($_v['flow']['id'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['flow']['title'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['flow']['name'], ENT_QUOTES);?>
</a>
<?php if($_v['flow']['running']){?>
						<a class="play" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/workflows/<?php echo htmlspecialchars($_v['flow']['id'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['flow']['name'], ENT_QUOTES);?>
, running" aria-label="Watch <?php echo htmlspecialchars($_v['flow']['name'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-eye"/></svg></a>
<?php }elseif($_v['flow']['canRun']){?>
						<button type="button" class="play" data-workflow-run="<?php echo htmlspecialchars($_v['flow']['id'], ENT_QUOTES);?>
"<?php if($_v['flow']['prompted']){?>
 data-workflow-prompt="<?php echo htmlspecialchars($_v['flow']['name'], ENT_QUOTES);?>
"<?php }?> title="Run <?php echo htmlspecialchars($_v['flow']['name'], ENT_QUOTES);?>
" aria-label="Run <?php echo htmlspecialchars($_v['flow']['name'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-play"/></svg></button>
<?php }?>
					</div>
<?php } }?>
				</div>
<?php } }?>
			</nav>

<?php $_v['data'] = $_v['chrome']['folders'];?>


			<nav class="folders">
<?php if(!empty($_v['data']))foreach($_v['data'] as $_v['group']){ if($_v['group']['workspace']){ $_v['data_workspace'] = $_v['group'];if($_v['data_workspace']['top']){?>

<details class="workspace<?php if($_v['data_workspace']['active']){?>
 active<?php }?>" data-project="<?php echo htmlspecialchars($_v['data_workspace']['id'], ENT_QUOTES);?>
" open>
	<summary<?php if($_v['data_workspace']['here']){?>
 class="active"<?php }?> title="<?php echo htmlspecialchars($_v['data_workspace']['title'], ENT_QUOTES);?>
">
		<svg class="ic chev"><use href="#i-chev"/></svg>
		<a class="section<?php if($_v['data_workspace']['here']){?>
 active<?php }?>" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_workspace']['href'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['data_workspace']['id'], ENT_QUOTES);?>
">
			<span><?php echo htmlspecialchars($_v['data_workspace']['name'], ENT_QUOTES);?>
</span>
<?php if($_v['data_workspace']['bad']){?>
<small class="bad" title="<?php echo htmlspecialchars($_v['data_workspace']['badTitle'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_workspace']['badge'], ENT_QUOTES);?>
</small><?php }else{?>
<small><?php echo htmlspecialchars($_v['data_workspace']['count'], ENT_QUOTES);?>
</small><?php }?>
		</a>
	</summary>
<?php if($_v['data_workspace']['jobs']){?>

	<div class="jobs top">
<?php $_v['data_jobs'] = $_v['data_workspace'];if(!empty($_v['data_jobs']['jobs']))foreach($_v['data_jobs']['jobs'] as $_v['job']){?>
<div class="job<?php if($_v['job']['pinned']){?>
 pinned<?php }?><?php if($_v['job']['active']){?>
 active<?php }?>" data-job-key="<?php echo htmlspecialchars($_v['job']['key'], ENT_QUOTES);?>
">
	<i class="shebang<?php if($_v['job']['running']){?>
 running<?php }?>" aria-hidden="true">#!</i>
	<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['job']['href'], ENT_QUOTES);?>
"<?php if($_v['job']['hint']){?>
 title="<?php echo htmlspecialchars($_v['job']['hint'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['job']['kind']){?>
<i><?php echo htmlspecialchars($_v['job']['kind'], ENT_QUOTES);?>
:</i><?php }?><?php echo htmlspecialchars($_v['job']['label'], ENT_QUOTES);?>
</a>
<?php if($_v['job']['running']){?>

	<button type="button" class="play" data-attach="<?php echo htmlspecialchars($_v['job']['attach'], ENT_QUOTES);?>
" title="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
, running" aria-label="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-eye"/></svg></button>
<?php }elseif($_v['job']['can']){?>

	<button type="button" class="play" data-run data-folder="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
"<?php if($_v['job']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['job']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['job']['job'], ENT_QUOTES);?>
"<?php if($_v['job']['args']){?>
 data-run-args="<?php echo htmlspecialchars($_v['job']['args'], ENT_QUOTES);?>
"<?php }?> data-title="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
 / <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" title="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" aria-label="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-play"/></svg></button>
<?php }?>
</div>
<?php }else{?>

<div class="empty"><?php echo htmlspecialchars($_v['data_jobs']['empty'], ENT_QUOTES);?>
</div>
<?php }?>

	</div>
<?php } if(!empty($_v['data_workspace']['projects']))foreach($_v['data_workspace']['projects'] as $_v['project']){?><?php $_v['data_project'] = $_v['project'];if($_v['data_project']['top']){?>

				<details class="project<?php if($_v['data_project']['active']){?>
 active<?php }?>" data-project="<?php echo htmlspecialchars($_v['data_project']['id'], ENT_QUOTES);?>
" open>
					<summary<?php if($_v['data_project']['here']){?>
 class="active"<?php }?> title="<?php echo htmlspecialchars($_v['data_project']['title'], ENT_QUOTES);?>
">
						<a class="section<?php if($_v['data_project']['here']){?>
 active<?php }?>" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_project']['href'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['data_project']['id'], ENT_QUOTES);?>
">
<?php $_v['data_icon'] = $_v['data_project'];?>
<svg class="ic<?php if($_v['data_icon']['dirty']){?>
 dirty<?php }?>"><?php if($_v['data_icon']['dirty']){?>
<title><?php echo htmlspecialchars($_v['data_icon']['changed'], ENT_QUOTES);?>
</title><?php }?><use href="#i-<?php echo htmlspecialchars($_v['data_icon']['icon'], ENT_QUOTES);?>
"/></svg><span><?php echo htmlspecialchars($_v['data_project']['name'], ENT_QUOTES);?>
</span>
<?php if($_v['data_project']['bad']){?>
<small class="bad" title="<?php echo htmlspecialchars($_v['data_project']['badTitle'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_project']['badge'], ENT_QUOTES);?>
</small><?php }elseif($_v['data_project']['counted']){?>
<small><?php echo htmlspecialchars($_v['data_project']['count'], ENT_QUOTES);?>
</small><?php }?>
						</a>
						<svg class="ic chev"><use href="#i-chev"/></svg>
					</summary>
<?php if($_v['data_project']['jobs']){?>

					<div class="jobs top">
<?php $_v['data_jobs'] = $_v['data_project'];if(!empty($_v['data_jobs']['jobs']))foreach($_v['data_jobs']['jobs'] as $_v['job']){?>
<div class="job<?php if($_v['job']['pinned']){?>
 pinned<?php }?><?php if($_v['job']['active']){?>
 active<?php }?>" data-job-key="<?php echo htmlspecialchars($_v['job']['key'], ENT_QUOTES);?>
">
	<i class="shebang<?php if($_v['job']['running']){?>
 running<?php }?>" aria-hidden="true">#!</i>
	<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['job']['href'], ENT_QUOTES);?>
"<?php if($_v['job']['hint']){?>
 title="<?php echo htmlspecialchars($_v['job']['hint'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['job']['kind']){?>
<i><?php echo htmlspecialchars($_v['job']['kind'], ENT_QUOTES);?>
:</i><?php }?><?php echo htmlspecialchars($_v['job']['label'], ENT_QUOTES);?>
</a>
<?php if($_v['job']['running']){?>

	<button type="button" class="play" data-attach="<?php echo htmlspecialchars($_v['job']['attach'], ENT_QUOTES);?>
" title="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
, running" aria-label="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-eye"/></svg></button>
<?php }elseif($_v['job']['can']){?>

	<button type="button" class="play" data-run data-folder="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
"<?php if($_v['job']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['job']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['job']['job'], ENT_QUOTES);?>
"<?php if($_v['job']['args']){?>
 data-run-args="<?php echo htmlspecialchars($_v['job']['args'], ENT_QUOTES);?>
"<?php }?> data-title="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
 / <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" title="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" aria-label="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-play"/></svg></button>
<?php }?>
</div>
<?php }else{?>

<div class="empty"><?php echo htmlspecialchars($_v['data_jobs']['empty'], ENT_QUOTES);?>
</div>
<?php }?>

					</div>
<?php } if(!empty($_v['data_project']['nested']))foreach($_v['data_project']['nested'] as $_v['folder']){?><?php $_v['data_folder'] = $_v['folder'];?>

<details class="folder<?php if($_v['data_folder']['active']){?>
 active<?php }?>" data-folder="<?php echo htmlspecialchars($_v['data_folder']['id'], ENT_QUOTES);?>
"<?php if($_v['data_folder']['open']){?>
 open<?php }?>>
	<summary title="<?php echo htmlspecialchars($_v['data_folder']['title'], ENT_QUOTES);?>
">
		<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_folder']['href'], ENT_QUOTES);?>
"><?php $_v['data_icon'] = $_v['data_folder'];?>
<svg class="ic<?php if($_v['data_icon']['dirty']){?>
 dirty<?php }?>"><?php if($_v['data_icon']['dirty']){?>
<title><?php echo htmlspecialchars($_v['data_icon']['changed'], ENT_QUOTES);?>
</title><?php }?><use href="#i-<?php echo htmlspecialchars($_v['data_icon']['icon'], ENT_QUOTES);?>
"/></svg><span><?php if($_v['data_folder']['dir']){?>
<i><?php echo htmlspecialchars($_v['data_folder']['dir'], ENT_QUOTES);?>
/</i><?php }?><b><?php echo htmlspecialchars($_v['data_folder']['name'], ENT_QUOTES);?>
</b></span></a>
		<i class="dot"></i>
<?php if($_v['data_folder']['bad']){?>
<small class="bad" title="<?php echo htmlspecialchars($_v['data_folder']['badTitle'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_folder']['badge'], ENT_QUOTES);?>
</small><?php }else{?>
<small><?php echo htmlspecialchars($_v['data_folder']['count'], ENT_QUOTES);?>
</small><?php }?>
		<svg class="ic chev"><use href="#i-chev"/></svg>
	</summary>
	<div class="jobs">
<?php $_v['data_jobs'] = $_v['data_folder'];if(!empty($_v['data_jobs']['jobs']))foreach($_v['data_jobs']['jobs'] as $_v['job']){?>
<div class="job<?php if($_v['job']['pinned']){?>
 pinned<?php }?><?php if($_v['job']['active']){?>
 active<?php }?>" data-job-key="<?php echo htmlspecialchars($_v['job']['key'], ENT_QUOTES);?>
">
	<i class="shebang<?php if($_v['job']['running']){?>
 running<?php }?>" aria-hidden="true">#!</i>
	<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['job']['href'], ENT_QUOTES);?>
"<?php if($_v['job']['hint']){?>
 title="<?php echo htmlspecialchars($_v['job']['hint'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['job']['kind']){?>
<i><?php echo htmlspecialchars($_v['job']['kind'], ENT_QUOTES);?>
:</i><?php }?><?php echo htmlspecialchars($_v['job']['label'], ENT_QUOTES);?>
</a>
<?php if($_v['job']['running']){?>

	<button type="button" class="play" data-attach="<?php echo htmlspecialchars($_v['job']['attach'], ENT_QUOTES);?>
" title="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
, running" aria-label="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-eye"/></svg></button>
<?php }elseif($_v['job']['can']){?>

	<button type="button" class="play" data-run data-folder="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
"<?php if($_v['job']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['job']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['job']['job'], ENT_QUOTES);?>
"<?php if($_v['job']['args']){?>
 data-run-args="<?php echo htmlspecialchars($_v['job']['args'], ENT_QUOTES);?>
"<?php }?> data-title="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
 / <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" title="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" aria-label="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-play"/></svg></button>
<?php }?>
</div>
<?php }else{?>

<div class="empty"><?php echo htmlspecialchars($_v['data_jobs']['empty'], ENT_QUOTES);?>
</div>
<?php }?>

	</div>
</details>
<?php }?>
				</details>
<?php }else{if(!empty($_v['data_project']['nested']))foreach($_v['data_project']['nested'] as $_v['folder']){?><?php $_v['data_folder'] = $_v['folder'];?>

<details class="folder<?php if($_v['data_folder']['active']){?>
 active<?php }?>" data-folder="<?php echo htmlspecialchars($_v['data_folder']['id'], ENT_QUOTES);?>
"<?php if($_v['data_folder']['open']){?>
 open<?php }?>>
	<summary title="<?php echo htmlspecialchars($_v['data_folder']['title'], ENT_QUOTES);?>
">
		<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_folder']['href'], ENT_QUOTES);?>
"><?php $_v['data_icon'] = $_v['data_folder'];?>
<svg class="ic<?php if($_v['data_icon']['dirty']){?>
 dirty<?php }?>"><?php if($_v['data_icon']['dirty']){?>
<title><?php echo htmlspecialchars($_v['data_icon']['changed'], ENT_QUOTES);?>
</title><?php }?><use href="#i-<?php echo htmlspecialchars($_v['data_icon']['icon'], ENT_QUOTES);?>
"/></svg><span><?php if($_v['data_folder']['dir']){?>
<i><?php echo htmlspecialchars($_v['data_folder']['dir'], ENT_QUOTES);?>
/</i><?php }?><b><?php echo htmlspecialchars($_v['data_folder']['name'], ENT_QUOTES);?>
</b></span></a>
		<i class="dot"></i>
<?php if($_v['data_folder']['bad']){?>
<small class="bad" title="<?php echo htmlspecialchars($_v['data_folder']['badTitle'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_folder']['badge'], ENT_QUOTES);?>
</small><?php }else{?>
<small><?php echo htmlspecialchars($_v['data_folder']['count'], ENT_QUOTES);?>
</small><?php }?>
		<svg class="ic chev"><use href="#i-chev"/></svg>
	</summary>
	<div class="jobs">
<?php $_v['data_jobs'] = $_v['data_folder'];if(!empty($_v['data_jobs']['jobs']))foreach($_v['data_jobs']['jobs'] as $_v['job']){?>
<div class="job<?php if($_v['job']['pinned']){?>
 pinned<?php }?><?php if($_v['job']['active']){?>
 active<?php }?>" data-job-key="<?php echo htmlspecialchars($_v['job']['key'], ENT_QUOTES);?>
">
	<i class="shebang<?php if($_v['job']['running']){?>
 running<?php }?>" aria-hidden="true">#!</i>
	<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['job']['href'], ENT_QUOTES);?>
"<?php if($_v['job']['hint']){?>
 title="<?php echo htmlspecialchars($_v['job']['hint'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['job']['kind']){?>
<i><?php echo htmlspecialchars($_v['job']['kind'], ENT_QUOTES);?>
:</i><?php }?><?php echo htmlspecialchars($_v['job']['label'], ENT_QUOTES);?>
</a>
<?php if($_v['job']['running']){?>

	<button type="button" class="play" data-attach="<?php echo htmlspecialchars($_v['job']['attach'], ENT_QUOTES);?>
" title="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
, running" aria-label="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-eye"/></svg></button>
<?php }elseif($_v['job']['can']){?>

	<button type="button" class="play" data-run data-folder="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
"<?php if($_v['job']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['job']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['job']['job'], ENT_QUOTES);?>
"<?php if($_v['job']['args']){?>
 data-run-args="<?php echo htmlspecialchars($_v['job']['args'], ENT_QUOTES);?>
"<?php }?> data-title="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
 / <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" title="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" aria-label="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-play"/></svg></button>
<?php }?>
</div>
<?php }else{?>

<div class="empty"><?php echo htmlspecialchars($_v['data_jobs']['empty'], ENT_QUOTES);?>
</div>
<?php }?>

	</div>
</details>
<?php } } }?>
</details>
<?php }else{if(!empty($_v['data_workspace']['projects']))foreach($_v['data_workspace']['projects'] as $_v['project']){?><?php $_v['data_project'] = $_v['project'];if($_v['data_project']['top']){?>

				<details class="project<?php if($_v['data_project']['active']){?>
 active<?php }?>" data-project="<?php echo htmlspecialchars($_v['data_project']['id'], ENT_QUOTES);?>
" open>
					<summary<?php if($_v['data_project']['here']){?>
 class="active"<?php }?> title="<?php echo htmlspecialchars($_v['data_project']['title'], ENT_QUOTES);?>
">
						<a class="section<?php if($_v['data_project']['here']){?>
 active<?php }?>" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_project']['href'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['data_project']['id'], ENT_QUOTES);?>
">
<?php $_v['data_icon'] = $_v['data_project'];?>
<svg class="ic<?php if($_v['data_icon']['dirty']){?>
 dirty<?php }?>"><?php if($_v['data_icon']['dirty']){?>
<title><?php echo htmlspecialchars($_v['data_icon']['changed'], ENT_QUOTES);?>
</title><?php }?><use href="#i-<?php echo htmlspecialchars($_v['data_icon']['icon'], ENT_QUOTES);?>
"/></svg><span><?php echo htmlspecialchars($_v['data_project']['name'], ENT_QUOTES);?>
</span>
<?php if($_v['data_project']['bad']){?>
<small class="bad" title="<?php echo htmlspecialchars($_v['data_project']['badTitle'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_project']['badge'], ENT_QUOTES);?>
</small><?php }elseif($_v['data_project']['counted']){?>
<small><?php echo htmlspecialchars($_v['data_project']['count'], ENT_QUOTES);?>
</small><?php }?>
						</a>
						<svg class="ic chev"><use href="#i-chev"/></svg>
					</summary>
<?php if($_v['data_project']['jobs']){?>

					<div class="jobs top">
<?php $_v['data_jobs'] = $_v['data_project'];if(!empty($_v['data_jobs']['jobs']))foreach($_v['data_jobs']['jobs'] as $_v['job']){?>
<div class="job<?php if($_v['job']['pinned']){?>
 pinned<?php }?><?php if($_v['job']['active']){?>
 active<?php }?>" data-job-key="<?php echo htmlspecialchars($_v['job']['key'], ENT_QUOTES);?>
">
	<i class="shebang<?php if($_v['job']['running']){?>
 running<?php }?>" aria-hidden="true">#!</i>
	<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['job']['href'], ENT_QUOTES);?>
"<?php if($_v['job']['hint']){?>
 title="<?php echo htmlspecialchars($_v['job']['hint'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['job']['kind']){?>
<i><?php echo htmlspecialchars($_v['job']['kind'], ENT_QUOTES);?>
:</i><?php }?><?php echo htmlspecialchars($_v['job']['label'], ENT_QUOTES);?>
</a>
<?php if($_v['job']['running']){?>

	<button type="button" class="play" data-attach="<?php echo htmlspecialchars($_v['job']['attach'], ENT_QUOTES);?>
" title="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
, running" aria-label="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-eye"/></svg></button>
<?php }elseif($_v['job']['can']){?>

	<button type="button" class="play" data-run data-folder="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
"<?php if($_v['job']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['job']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['job']['job'], ENT_QUOTES);?>
"<?php if($_v['job']['args']){?>
 data-run-args="<?php echo htmlspecialchars($_v['job']['args'], ENT_QUOTES);?>
"<?php }?> data-title="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
 / <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" title="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" aria-label="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-play"/></svg></button>
<?php }?>
</div>
<?php }else{?>

<div class="empty"><?php echo htmlspecialchars($_v['data_jobs']['empty'], ENT_QUOTES);?>
</div>
<?php }?>

					</div>
<?php } if(!empty($_v['data_project']['nested']))foreach($_v['data_project']['nested'] as $_v['folder']){?><?php $_v['data_folder'] = $_v['folder'];?>

<details class="folder<?php if($_v['data_folder']['active']){?>
 active<?php }?>" data-folder="<?php echo htmlspecialchars($_v['data_folder']['id'], ENT_QUOTES);?>
"<?php if($_v['data_folder']['open']){?>
 open<?php }?>>
	<summary title="<?php echo htmlspecialchars($_v['data_folder']['title'], ENT_QUOTES);?>
">
		<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_folder']['href'], ENT_QUOTES);?>
"><?php $_v['data_icon'] = $_v['data_folder'];?>
<svg class="ic<?php if($_v['data_icon']['dirty']){?>
 dirty<?php }?>"><?php if($_v['data_icon']['dirty']){?>
<title><?php echo htmlspecialchars($_v['data_icon']['changed'], ENT_QUOTES);?>
</title><?php }?><use href="#i-<?php echo htmlspecialchars($_v['data_icon']['icon'], ENT_QUOTES);?>
"/></svg><span><?php if($_v['data_folder']['dir']){?>
<i><?php echo htmlspecialchars($_v['data_folder']['dir'], ENT_QUOTES);?>
/</i><?php }?><b><?php echo htmlspecialchars($_v['data_folder']['name'], ENT_QUOTES);?>
</b></span></a>
		<i class="dot"></i>
<?php if($_v['data_folder']['bad']){?>
<small class="bad" title="<?php echo htmlspecialchars($_v['data_folder']['badTitle'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_folder']['badge'], ENT_QUOTES);?>
</small><?php }else{?>
<small><?php echo htmlspecialchars($_v['data_folder']['count'], ENT_QUOTES);?>
</small><?php }?>
		<svg class="ic chev"><use href="#i-chev"/></svg>
	</summary>
	<div class="jobs">
<?php $_v['data_jobs'] = $_v['data_folder'];if(!empty($_v['data_jobs']['jobs']))foreach($_v['data_jobs']['jobs'] as $_v['job']){?>
<div class="job<?php if($_v['job']['pinned']){?>
 pinned<?php }?><?php if($_v['job']['active']){?>
 active<?php }?>" data-job-key="<?php echo htmlspecialchars($_v['job']['key'], ENT_QUOTES);?>
">
	<i class="shebang<?php if($_v['job']['running']){?>
 running<?php }?>" aria-hidden="true">#!</i>
	<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['job']['href'], ENT_QUOTES);?>
"<?php if($_v['job']['hint']){?>
 title="<?php echo htmlspecialchars($_v['job']['hint'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['job']['kind']){?>
<i><?php echo htmlspecialchars($_v['job']['kind'], ENT_QUOTES);?>
:</i><?php }?><?php echo htmlspecialchars($_v['job']['label'], ENT_QUOTES);?>
</a>
<?php if($_v['job']['running']){?>

	<button type="button" class="play" data-attach="<?php echo htmlspecialchars($_v['job']['attach'], ENT_QUOTES);?>
" title="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
, running" aria-label="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-eye"/></svg></button>
<?php }elseif($_v['job']['can']){?>

	<button type="button" class="play" data-run data-folder="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
"<?php if($_v['job']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['job']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['job']['job'], ENT_QUOTES);?>
"<?php if($_v['job']['args']){?>
 data-run-args="<?php echo htmlspecialchars($_v['job']['args'], ENT_QUOTES);?>
"<?php }?> data-title="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
 / <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" title="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" aria-label="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-play"/></svg></button>
<?php }?>
</div>
<?php }else{?>

<div class="empty"><?php echo htmlspecialchars($_v['data_jobs']['empty'], ENT_QUOTES);?>
</div>
<?php }?>

	</div>
</details>
<?php }?>
				</details>
<?php }else{if(!empty($_v['data_project']['nested']))foreach($_v['data_project']['nested'] as $_v['folder']){?><?php $_v['data_folder'] = $_v['folder'];?>

<details class="folder<?php if($_v['data_folder']['active']){?>
 active<?php }?>" data-folder="<?php echo htmlspecialchars($_v['data_folder']['id'], ENT_QUOTES);?>
"<?php if($_v['data_folder']['open']){?>
 open<?php }?>>
	<summary title="<?php echo htmlspecialchars($_v['data_folder']['title'], ENT_QUOTES);?>
">
		<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_folder']['href'], ENT_QUOTES);?>
"><?php $_v['data_icon'] = $_v['data_folder'];?>
<svg class="ic<?php if($_v['data_icon']['dirty']){?>
 dirty<?php }?>"><?php if($_v['data_icon']['dirty']){?>
<title><?php echo htmlspecialchars($_v['data_icon']['changed'], ENT_QUOTES);?>
</title><?php }?><use href="#i-<?php echo htmlspecialchars($_v['data_icon']['icon'], ENT_QUOTES);?>
"/></svg><span><?php if($_v['data_folder']['dir']){?>
<i><?php echo htmlspecialchars($_v['data_folder']['dir'], ENT_QUOTES);?>
/</i><?php }?><b><?php echo htmlspecialchars($_v['data_folder']['name'], ENT_QUOTES);?>
</b></span></a>
		<i class="dot"></i>
<?php if($_v['data_folder']['bad']){?>
<small class="bad" title="<?php echo htmlspecialchars($_v['data_folder']['badTitle'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_folder']['badge'], ENT_QUOTES);?>
</small><?php }else{?>
<small><?php echo htmlspecialchars($_v['data_folder']['count'], ENT_QUOTES);?>
</small><?php }?>
		<svg class="ic chev"><use href="#i-chev"/></svg>
	</summary>
	<div class="jobs">
<?php $_v['data_jobs'] = $_v['data_folder'];if(!empty($_v['data_jobs']['jobs']))foreach($_v['data_jobs']['jobs'] as $_v['job']){?>
<div class="job<?php if($_v['job']['pinned']){?>
 pinned<?php }?><?php if($_v['job']['active']){?>
 active<?php }?>" data-job-key="<?php echo htmlspecialchars($_v['job']['key'], ENT_QUOTES);?>
">
	<i class="shebang<?php if($_v['job']['running']){?>
 running<?php }?>" aria-hidden="true">#!</i>
	<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['job']['href'], ENT_QUOTES);?>
"<?php if($_v['job']['hint']){?>
 title="<?php echo htmlspecialchars($_v['job']['hint'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['job']['kind']){?>
<i><?php echo htmlspecialchars($_v['job']['kind'], ENT_QUOTES);?>
:</i><?php }?><?php echo htmlspecialchars($_v['job']['label'], ENT_QUOTES);?>
</a>
<?php if($_v['job']['running']){?>

	<button type="button" class="play" data-attach="<?php echo htmlspecialchars($_v['job']['attach'], ENT_QUOTES);?>
" title="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
, running" aria-label="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-eye"/></svg></button>
<?php }elseif($_v['job']['can']){?>

	<button type="button" class="play" data-run data-folder="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
"<?php if($_v['job']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['job']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['job']['job'], ENT_QUOTES);?>
"<?php if($_v['job']['args']){?>
 data-run-args="<?php echo htmlspecialchars($_v['job']['args'], ENT_QUOTES);?>
"<?php }?> data-title="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
 / <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" title="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" aria-label="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-play"/></svg></button>
<?php }?>
</div>
<?php }else{?>

<div class="empty"><?php echo htmlspecialchars($_v['data_jobs']['empty'], ENT_QUOTES);?>
</div>
<?php }?>

	</div>
</details>
<?php } } } }}else{ $_v['data_project'] = $_v['group'];if($_v['data_project']['top']){?>

				<details class="project<?php if($_v['data_project']['active']){?>
 active<?php }?>" data-project="<?php echo htmlspecialchars($_v['data_project']['id'], ENT_QUOTES);?>
" open>
					<summary<?php if($_v['data_project']['here']){?>
 class="active"<?php }?> title="<?php echo htmlspecialchars($_v['data_project']['title'], ENT_QUOTES);?>
">
						<a class="section<?php if($_v['data_project']['here']){?>
 active<?php }?>" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_project']['href'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['data_project']['id'], ENT_QUOTES);?>
">
<?php $_v['data_icon'] = $_v['data_project'];?>
<svg class="ic<?php if($_v['data_icon']['dirty']){?>
 dirty<?php }?>"><?php if($_v['data_icon']['dirty']){?>
<title><?php echo htmlspecialchars($_v['data_icon']['changed'], ENT_QUOTES);?>
</title><?php }?><use href="#i-<?php echo htmlspecialchars($_v['data_icon']['icon'], ENT_QUOTES);?>
"/></svg><span><?php echo htmlspecialchars($_v['data_project']['name'], ENT_QUOTES);?>
</span>
<?php if($_v['data_project']['bad']){?>
<small class="bad" title="<?php echo htmlspecialchars($_v['data_project']['badTitle'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_project']['badge'], ENT_QUOTES);?>
</small><?php }elseif($_v['data_project']['counted']){?>
<small><?php echo htmlspecialchars($_v['data_project']['count'], ENT_QUOTES);?>
</small><?php }?>
						</a>
						<svg class="ic chev"><use href="#i-chev"/></svg>
					</summary>
<?php if($_v['data_project']['jobs']){?>

					<div class="jobs top">
<?php $_v['data_jobs'] = $_v['data_project'];if(!empty($_v['data_jobs']['jobs']))foreach($_v['data_jobs']['jobs'] as $_v['job']){?>
<div class="job<?php if($_v['job']['pinned']){?>
 pinned<?php }?><?php if($_v['job']['active']){?>
 active<?php }?>" data-job-key="<?php echo htmlspecialchars($_v['job']['key'], ENT_QUOTES);?>
">
	<i class="shebang<?php if($_v['job']['running']){?>
 running<?php }?>" aria-hidden="true">#!</i>
	<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['job']['href'], ENT_QUOTES);?>
"<?php if($_v['job']['hint']){?>
 title="<?php echo htmlspecialchars($_v['job']['hint'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['job']['kind']){?>
<i><?php echo htmlspecialchars($_v['job']['kind'], ENT_QUOTES);?>
:</i><?php }?><?php echo htmlspecialchars($_v['job']['label'], ENT_QUOTES);?>
</a>
<?php if($_v['job']['running']){?>

	<button type="button" class="play" data-attach="<?php echo htmlspecialchars($_v['job']['attach'], ENT_QUOTES);?>
" title="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
, running" aria-label="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-eye"/></svg></button>
<?php }elseif($_v['job']['can']){?>

	<button type="button" class="play" data-run data-folder="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
"<?php if($_v['job']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['job']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['job']['job'], ENT_QUOTES);?>
"<?php if($_v['job']['args']){?>
 data-run-args="<?php echo htmlspecialchars($_v['job']['args'], ENT_QUOTES);?>
"<?php }?> data-title="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
 / <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" title="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" aria-label="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-play"/></svg></button>
<?php }?>
</div>
<?php }else{?>

<div class="empty"><?php echo htmlspecialchars($_v['data_jobs']['empty'], ENT_QUOTES);?>
</div>
<?php }?>

					</div>
<?php } if(!empty($_v['data_project']['nested']))foreach($_v['data_project']['nested'] as $_v['folder']){?><?php $_v['data_folder'] = $_v['folder'];?>

<details class="folder<?php if($_v['data_folder']['active']){?>
 active<?php }?>" data-folder="<?php echo htmlspecialchars($_v['data_folder']['id'], ENT_QUOTES);?>
"<?php if($_v['data_folder']['open']){?>
 open<?php }?>>
	<summary title="<?php echo htmlspecialchars($_v['data_folder']['title'], ENT_QUOTES);?>
">
		<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_folder']['href'], ENT_QUOTES);?>
"><?php $_v['data_icon'] = $_v['data_folder'];?>
<svg class="ic<?php if($_v['data_icon']['dirty']){?>
 dirty<?php }?>"><?php if($_v['data_icon']['dirty']){?>
<title><?php echo htmlspecialchars($_v['data_icon']['changed'], ENT_QUOTES);?>
</title><?php }?><use href="#i-<?php echo htmlspecialchars($_v['data_icon']['icon'], ENT_QUOTES);?>
"/></svg><span><?php if($_v['data_folder']['dir']){?>
<i><?php echo htmlspecialchars($_v['data_folder']['dir'], ENT_QUOTES);?>
/</i><?php }?><b><?php echo htmlspecialchars($_v['data_folder']['name'], ENT_QUOTES);?>
</b></span></a>
		<i class="dot"></i>
<?php if($_v['data_folder']['bad']){?>
<small class="bad" title="<?php echo htmlspecialchars($_v['data_folder']['badTitle'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_folder']['badge'], ENT_QUOTES);?>
</small><?php }else{?>
<small><?php echo htmlspecialchars($_v['data_folder']['count'], ENT_QUOTES);?>
</small><?php }?>
		<svg class="ic chev"><use href="#i-chev"/></svg>
	</summary>
	<div class="jobs">
<?php $_v['data_jobs'] = $_v['data_folder'];if(!empty($_v['data_jobs']['jobs']))foreach($_v['data_jobs']['jobs'] as $_v['job']){?>
<div class="job<?php if($_v['job']['pinned']){?>
 pinned<?php }?><?php if($_v['job']['active']){?>
 active<?php }?>" data-job-key="<?php echo htmlspecialchars($_v['job']['key'], ENT_QUOTES);?>
">
	<i class="shebang<?php if($_v['job']['running']){?>
 running<?php }?>" aria-hidden="true">#!</i>
	<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['job']['href'], ENT_QUOTES);?>
"<?php if($_v['job']['hint']){?>
 title="<?php echo htmlspecialchars($_v['job']['hint'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['job']['kind']){?>
<i><?php echo htmlspecialchars($_v['job']['kind'], ENT_QUOTES);?>
:</i><?php }?><?php echo htmlspecialchars($_v['job']['label'], ENT_QUOTES);?>
</a>
<?php if($_v['job']['running']){?>

	<button type="button" class="play" data-attach="<?php echo htmlspecialchars($_v['job']['attach'], ENT_QUOTES);?>
" title="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
, running" aria-label="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-eye"/></svg></button>
<?php }elseif($_v['job']['can']){?>

	<button type="button" class="play" data-run data-folder="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
"<?php if($_v['job']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['job']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['job']['job'], ENT_QUOTES);?>
"<?php if($_v['job']['args']){?>
 data-run-args="<?php echo htmlspecialchars($_v['job']['args'], ENT_QUOTES);?>
"<?php }?> data-title="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
 / <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" title="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" aria-label="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-play"/></svg></button>
<?php }?>
</div>
<?php }else{?>

<div class="empty"><?php echo htmlspecialchars($_v['data_jobs']['empty'], ENT_QUOTES);?>
</div>
<?php }?>

	</div>
</details>
<?php }?>
				</details>
<?php }else{if(!empty($_v['data_project']['nested']))foreach($_v['data_project']['nested'] as $_v['folder']){?><?php $_v['data_folder'] = $_v['folder'];?>

<details class="folder<?php if($_v['data_folder']['active']){?>
 active<?php }?>" data-folder="<?php echo htmlspecialchars($_v['data_folder']['id'], ENT_QUOTES);?>
"<?php if($_v['data_folder']['open']){?>
 open<?php }?>>
	<summary title="<?php echo htmlspecialchars($_v['data_folder']['title'], ENT_QUOTES);?>
">
		<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_folder']['href'], ENT_QUOTES);?>
"><?php $_v['data_icon'] = $_v['data_folder'];?>
<svg class="ic<?php if($_v['data_icon']['dirty']){?>
 dirty<?php }?>"><?php if($_v['data_icon']['dirty']){?>
<title><?php echo htmlspecialchars($_v['data_icon']['changed'], ENT_QUOTES);?>
</title><?php }?><use href="#i-<?php echo htmlspecialchars($_v['data_icon']['icon'], ENT_QUOTES);?>
"/></svg><span><?php if($_v['data_folder']['dir']){?>
<i><?php echo htmlspecialchars($_v['data_folder']['dir'], ENT_QUOTES);?>
/</i><?php }?><b><?php echo htmlspecialchars($_v['data_folder']['name'], ENT_QUOTES);?>
</b></span></a>
		<i class="dot"></i>
<?php if($_v['data_folder']['bad']){?>
<small class="bad" title="<?php echo htmlspecialchars($_v['data_folder']['badTitle'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_folder']['badge'], ENT_QUOTES);?>
</small><?php }else{?>
<small><?php echo htmlspecialchars($_v['data_folder']['count'], ENT_QUOTES);?>
</small><?php }?>
		<svg class="ic chev"><use href="#i-chev"/></svg>
	</summary>
	<div class="jobs">
<?php $_v['data_jobs'] = $_v['data_folder'];if(!empty($_v['data_jobs']['jobs']))foreach($_v['data_jobs']['jobs'] as $_v['job']){?>
<div class="job<?php if($_v['job']['pinned']){?>
 pinned<?php }?><?php if($_v['job']['active']){?>
 active<?php }?>" data-job-key="<?php echo htmlspecialchars($_v['job']['key'], ENT_QUOTES);?>
">
	<i class="shebang<?php if($_v['job']['running']){?>
 running<?php }?>" aria-hidden="true">#!</i>
	<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['job']['href'], ENT_QUOTES);?>
"<?php if($_v['job']['hint']){?>
 title="<?php echo htmlspecialchars($_v['job']['hint'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['job']['kind']){?>
<i><?php echo htmlspecialchars($_v['job']['kind'], ENT_QUOTES);?>
:</i><?php }?><?php echo htmlspecialchars($_v['job']['label'], ENT_QUOTES);?>
</a>
<?php if($_v['job']['running']){?>

	<button type="button" class="play" data-attach="<?php echo htmlspecialchars($_v['job']['attach'], ENT_QUOTES);?>
" title="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
, running" aria-label="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-eye"/></svg></button>
<?php }elseif($_v['job']['can']){?>

	<button type="button" class="play" data-run data-folder="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
"<?php if($_v['job']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['job']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['job']['job'], ENT_QUOTES);?>
"<?php if($_v['job']['args']){?>
 data-run-args="<?php echo htmlspecialchars($_v['job']['args'], ENT_QUOTES);?>
"<?php }?> data-title="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
 / <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" title="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" aria-label="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-play"/></svg></button>
<?php }?>
</div>
<?php }else{?>

<div class="empty"><?php echo htmlspecialchars($_v['data_jobs']['empty'], ENT_QUOTES);?>
</div>
<?php }?>

	</div>
</details>
<?php } }} }else{?>

				<div class="empty">No folders yet. <?php if($_v['chrome']['can']['project/edit']){?>
<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/settings/projects/new">Add a git repository</a><?php }else{?>
Add a git repository in Settings<?php }?>, or put a folder with jobs under the root.</div>
<?php }?>
			</nav>

		</aside>
		<main id="main" tabindex="-1">
<?php $_v['data'] = $_v['chrome'];?>


<?php if($_v['data']['crumbs']){?>

			<div class="crumbs"><a href="<?php echo htmlspecialchars($_v['data']['base'], ENT_QUOTES);?>
/">Dashboard</a><?php if(!empty($_v['data']['crumbs']))foreach($_v['data']['crumbs'] as $_v['crumb']){?><span><?php if($_v['crumb']['href']){?>
<a href="<?php echo htmlspecialchars($_v['data']['base'], ENT_QUOTES); echo htmlspecialchars($_v['crumb']['href'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['crumb']['text'], ENT_QUOTES);?>
</a><?php }else{ echo htmlspecialchars($_v['crumb']['text'], ENT_QUOTES); }?></span><?php }?></div>
<?php }?>


<?php if($_v['data']['notice']){?>

			<div class="alert ok" role="status"><?php echo htmlspecialchars($_v['data']['notice'], ENT_QUOTES);?>
</div>
<?php } if($_v['data']['error']){?>

			<div class="alert bad" role="alert"><?php echo htmlspecialchars($_v['data']['error'], ENT_QUOTES);?>
</div>
<?php }echo $_v['content'];?>

		</main>
<?php $_v['data'] = $_v['chrome']['recents'];?>


		<aside class="rail" aria-label="Recents">
			<div class="rail-inner">
				<div class="label side-label"><a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data']['href'], ENT_QUOTES);?>
">Recents</a></div>
				<nav class="recents" id="recents">
<?php if(!empty($_v['data']['groups']))foreach($_v['data']['groups'] as $_v['group']){?>
					<div class="recent-group"><?php echo htmlspecialchars($_v['group']['label'], ENT_QUOTES);?>
</div>
<?php if(!empty($_v['group']['entries']))foreach($_v['group']['entries'] as $_v['entry']){?>
					<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['entry']['href'], ENT_QUOTES);?>
" class="recent<?php if($_v['entry']['flow']){?>
 flow<?php }?>" title="<?php echo htmlspecialchars($_v['entry']['title'], ENT_QUOTES);?>
">
						<i class="dot <?php echo htmlspecialchars($_v['entry']['state'], ENT_QUOTES);?>
"></i>
						<span class="what"><b><?php echo htmlspecialchars($_v['entry']['name'], ENT_QUOTES);?>
</b><small><?php if($_v['entry']['flow']){?>
<svg class="ic lead"><use href="#i-flow"/></svg><?php }?><?php echo htmlspecialchars($_v['entry']['place'], ENT_QUOTES);?>
</small></span>
<?php if(!$_v['entry']['ended']){?>
						<svg class="ic act"><use href="#i-<?php echo htmlspecialchars($_v['entry']['act'], ENT_QUOTES);?>
"/></svg><?php }?>
						<?php $_v['data_avatar'] = $_v['entry']['avatar'];?>
<span class="avatar sm<?php if($_v['data_avatar']['role']){?>
 role-<?php echo htmlspecialchars($_v['data_avatar']['role'], ENT_QUOTES); }?>" title="<?php echo htmlspecialchars($_v['data_avatar']['name'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_avatar']['initials'], ENT_QUOTES);?>
</span>
					</a>
<?php } } if(!$_v['data']['groups']){?>

					<div class="empty">Runs started by anyone show up here.</div>
<?php }?>
					<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data']['href'], ENT_QUOTES);?>
" class="recent more" data-recents-more hidden><svg class="ic"><use href="#i-list"/></svg><span class="what"><b>All runs</b><small></small></span></a>
				</nav>
			</div>
		</aside>

	</div>

	<dialog class="run" id="run-dialog" aria-labelledby="run-title">
		<div class="run-head">
			<span class="label" id="run-kind">Run</span>
			<h3 id="run-title"></h3>
			<span class="run-subtitle" id="run-subtitle" hidden></span>
			<span class="tag rec" id="run-rec" hidden>Recording</span>
			<span class="tag input-on" id="run-input" hidden>Input on</span>
			<span class="spacer"></span>
			<span class="dim" id="run-who"></span>
			<span class="elapsed" id="run-elapsed"></span>
			<span class="state" id="run-state"></span>
			<button type="button" class="btn btn-sm" id="run-replay" title="Replay the recording" aria-label="Replay the recording" hidden>|&gt;</button>
			<a class="btn btn-sm" id="run-page" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/runs" hidden>Open run page</a>
		</div>
		<div class="run-body">
			<div class="terminal-host" id="run-terminal"></div>
		</div>
		<p class="run-notice" id="run-notice" role="status"></p>
		<div class="run-reason" id="run-reason" hidden>
			<label for="run-reason-text">Why does this need to run outside its hours? task-ui saves your reason with the run.</label>
			<input type="text" id="run-reason-text" placeholder="customer is waiting on the hotfix" autocomplete="off" maxlength="200">
		</div>
		<div class="run-foot">
			<nav class="run-chain" id="run-chain" aria-label="The runs of the chain" hidden>
				<button type="button" class="btn btn-sm" id="run-prev" title="The run before this one">‹ Previous</button>
				<span class="run-pos" id="run-pos"></span>
				<button type="button" class="btn btn-sm" id="run-next" title="The run after this one">Next ›</button>
			</nav>
			<a class="btn btn-sm" id="run-report" href="#" title="Claude's answer or summary, as a document" hidden>Report</a>
			<button type="button" class="btn" id="run-close" aria-label="Close">Close</button>
			<span class="hint" id="run-hint"></span>
			<button type="button" class="btn" id="run-attach" hidden>Attach</button>
			<button type="button" class="btn" id="run-again" hidden>Run again</button>
			<button type="button" class="btn" id="run-take" aria-pressed="false" title="Send your keystrokes to the run, to answer a prompt it waits on" hidden>Take input</button>
			<button type="button" class="btn" id="run-stop" hidden>Stop</button>
			<button type="button" class="btn btn-danger" id="run-kill" hidden>Kill</button>
			<button type="button" class="btn" id="run-pause" hidden>Pause</button>
			<button type="button" class="btn" id="run-restart" hidden>Restart</button>
			<button type="button" class="btn btn-primary" id="run-force" hidden>Run force</button>
		</div>
	</dialog>

	<dialog class="run preview" id="preview-dialog" aria-labelledby="preview-title">
		<div class="run-head">
			<span class="label" id="preview-label">Artifact</span>
			<h3 id="preview-title"></h3>
		</div>
		<div class="run-body preview-body"><img id="preview-image" alt=""><div id="preview-doc" class="preview-doc" hidden></div></div>
		<div class="run-foot">
			<span class="hint"><select id="preview-version" aria-label="Version" hidden></select></span>
			<a class="btn" id="preview-download" href="#" download>Download</a>
			<span class="preview-actions" id="preview-actions"></span>
			<button type="button" class="btn" id="preview-close" aria-label="Close">Close</button>
		</div>
	</dialog>

	<dialog class="run pipeline-window" id="pipeline-dialog" aria-labelledby="pipeline-title">
		<div class="run-head">
			<span class="label">Workflow run</span>
			<h3 id="pipeline-title"></h3>
		</div>
		<div class="run-body pipeline-body" id="pipeline-body"></div>
		<div class="run-foot">
			<span class="hint"></span>
			<a class="btn" id="pipeline-page" href="#">Open page</a>
			<button type="button" class="btn" id="pipeline-close" aria-label="Close">Close</button>
		</div>
	</dialog>

	<dialog class="run preview run-summary" id="summary-dialog" aria-labelledby="summary-title">
		<div class="run-head">
			<span class="label">Summary</span>
			<h3 id="summary-title"></h3>
		</div>
		<div class="run-body preview-body"><div id="summary-text" class="preview-doc markdown"></div></div>
		<div class="run-foot">
			<span class="hint" id="summary-meta"></span>
			<a class="btn" id="summary-run" href="#">View run</a>
			<button type="button" class="btn" id="summary-close" aria-label="Close">Close</button>
		</div>
	</dialog>

	<dialog class="run source" id="source-dialog" aria-labelledby="source-title">
		<div class="run-head">
			<span class="label">Source</span>
			<h3 id="source-title"></h3>
		</div>
		<div class="run-body"><pre class="code" id="source-code"></pre></div>
		<div class="run-foot">
			<span class="hint"></span>
			<button type="button" class="btn" id="source-close" aria-label="Close">Close</button>
		</div>
	</dialog>


	<dialog class="confirm" id="workflow-input-dialog" aria-labelledby="workflow-input-title">
		<form method="dialog" data-workflow-input>
			<h3 id="workflow-input-title">Run</h3>
			<p>Optional input for the workflow's prompt. Claude reads it after the prompt, under the line <code>Input:</code>.</p>
			<textarea name="input" rows="3" spellcheck="false" aria-label="Input for the prompt"></textarea>
			<div class="form-foot">
				<button type="button" class="btn" data-workflow-input-cancel>Cancel</button>
				<button type="submit" class="btn btn-primary">Run</button>
			</div>
		</form>
	</dialog>
	<dialog class="confirm" id="confirm-dialog" aria-labelledby="confirm-title">
		<h3 id="confirm-title">Are you sure?</h3>
		<p id="confirm-text"></p>
		<div class="form-foot">
			<button type="button" class="btn" id="confirm-cancel">Cancel</button>
			<button type="button" class="btn btn-danger" id="confirm-ok">Delete</button>
		</div>
	</dialog>

<?php }else{?>

	<header class="top login-top">
		<h1><a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/"><span>TaskUI</span><svg class="logo" viewBox="0 0 32 32" aria-hidden="true"><rect width="32" height="32" rx="3"/><path class="prompt" d="M8 10l6 6-6 6"/><path class="cursor" d="M16 23h8"/></svg>
</a></h1>
	</header>
<?php echo $_v['content'];}?>
	<div class="toasts" id="toasts" role="status" aria-live="polite"></div>
<?php $_v['data'] = $_v['chrome'];?>


	<footer>
		<span class="credit">
			<a class="wordmark" href="https://github.com/titpetric/task-ui">task-ui</a>
			<span class="sep" aria-hidden="true">|</span>
			<span>pipelines by <a href="https://github.com/titpetric">@titpetric</a></span>
			<span class="sep" aria-hidden="true">|</span>
			<a href="https://blacksky.digital/products/task-ui">blacksky.digital</a>
		</span>
	</footer>
	<script src="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/vendor/xterm.js?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
" defer></script>
	<script src="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/vendor/addon-fit.js?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
" defer></script>
	<script src="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/task-ui.js?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
" defer></script>
<?php if(!empty($_v['data']['scripts']))foreach($_v['data']['scripts'] as $_v['script']){?>	<script src="<?php echo htmlspecialchars($_v['script'], ENT_QUOTES);?>
" defer></script>
<?php }?>

</body>
</html>
