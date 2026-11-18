<?php $_v=&$this->vars;?>
<!DOCTYPE html>
<html lang="en">
<?php $_v['data'] = $_v['chrome'];?>

<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<meta name="robots" content="noindex, nofollow">
	<meta name="color-scheme" content="dark light">
	<meta name="theme-color" content="#f7f5f0" media="(prefers-color-scheme: light)">
	<meta name="theme-color" content="#07090c" media="(prefers-color-scheme: dark)">
	<title><?php if($_v['data']['title']){ echo htmlspecialchars($_v['data']['title'], ENT_QUOTES);?>
 | <?php }?>TaskUI</title>
	<link rel="icon" href="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/icon.svg" type="image/svg+xml">
	<link rel="stylesheet" href="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/ext/xterm.css?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
">
	<link rel="stylesheet" href="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/mdedit/mdedit.min.css?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
">
	<link rel="stylesheet" href="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/task-ui.css?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
">
	<link rel="stylesheet" href="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/voice/voice.css?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
">
<?php if(!empty($_v['data']['styles']))foreach($_v['data']['styles'] as $_v['style']){?>	<link rel="stylesheet" href="<?php echo htmlspecialchars($_v['style'], ENT_QUOTES);?>
">
<?php }?></head>

<body data-roles='<?php echo htmlspecialchars($_v['chrome']['roles'], ENT_QUOTES);?>
' data-workspaces='<?php echo htmlspecialchars($_v['chrome']['workspaces'], ENT_QUOTES);?>
'<?php if($_v['chrome']['user']){?>
 data-me="<?php echo htmlspecialchars($_v['chrome']['user']['name'], ENT_QUOTES);?>
"<?php }?><?php if($_v['chrome']['can']['artifact/edit']){?>
 data-upload="<?php echo htmlspecialchars($_v['chrome']['upload'], ENT_QUOTES);?>
" data-artifact-edit<?php }?><?php if($_v['chrome']['can']['project/edit']){?>
 data-file-edit<?php }?> data-user-menu="<?php if($_v['chrome']['can']['user/edit']){?>
edit<?php }elseif($_v['chrome']['can']['user/view']){?>
view<?php }?>">
<?php if($_v['chrome']['user']){?>

<a class="skip-link" href="#main">Skip to content</a>
<?php $_v['data'] = $_v['chrome'];if($_v['data']['can']['prompt/edit'] && $_v['data']['can']['prompt/view'] && $_v['data']['can']['note/view'] && $_v['data']['claude']){?>
	<header class="top" data-prompt-folder="<?php echo htmlspecialchars($_v['data']['prompt']['folder'], ENT_QUOTES);?>
" data-prompt-dir="<?php echo htmlspecialchars($_v['data']['prompt']['dir'], ENT_QUOTES);?>
" data-prompt-run="<?php echo htmlspecialchars($_v['data']['prompt']['run'], ENT_QUOTES);?>
" data-prompt-label="<?php echo htmlspecialchars($_v['data']['prompt']['label'], ENT_QUOTES);?>
" data-prompt-href="<?php echo htmlspecialchars($_v['data']['prompt']['href'], ENT_QUOTES);?>
">
		<button type="button" class="btn-icon menu-toggle" id="menu-toggle" aria-label="Menu" aria-controls="side" aria-expanded="false"><i class="ic ic-menu" aria-hidden="true"></i></button>
		<h1><a href="<?php echo htmlspecialchars($_v['data']['base'], ENT_QUOTES);?>
/"><span>TaskUI</span><svg class="logo" viewBox="0 0 32 32" aria-hidden="true"><rect width="32" height="32" rx="3"/><path class="prompt" d="M8 10l6 6-6 6"/><path class="cursor" d="M16 23h8"/></svg>
</a></h1>
		<form class="header-search" method="post" action="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/search/ask" role="search">
			<input type="search" name="q" aria-label="Search the documentation and job targets" autocomplete="off" spellcheck="false" placeholder="Search docs and jobs">
		</form>
<?php }else{?>
	<header class="top">
		<button type="button" class="btn-icon menu-toggle" id="menu-toggle" aria-label="Menu" aria-controls="side" aria-expanded="false"><i class="ic ic-menu" aria-hidden="true"></i></button>
		<h1><a href="<?php echo htmlspecialchars($_v['data']['base'], ENT_QUOTES);?>
/"><span>TaskUI</span><svg class="logo" viewBox="0 0 32 32" aria-hidden="true"><rect width="32" height="32" rx="3"/><path class="prompt" d="M8 10l6 6-6 6"/><path class="cursor" d="M16 23h8"/></svg>
</a></h1>
		<span class="spacer"></span>
<?php }?>		<?php if($_v['chrome']['can']['note/view']){?>
<details class="notes-bell" id="notes-bell" data-notes-bell>
	<summary class="btn-icon bell-icon" aria-label="Chats" aria-haspopup="menu" title="Chats"><i class="ic ic-bell" aria-hidden="true"></i><span class="bell-count" id="notes-bell-count" hidden></span></summary>
	<div class="notes-bell-menu" role="menu">
		<p class="notes-bell-empty faint">No chat yet. Pick someone in the orb to start one.</p>
		<div class="notes-bell-items" id="notes-bell-items"></div>
	</div>
</details>
<?php }?>

		<details class="user" id="user-menu">
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
"><i class="ic ic-users" aria-hidden="true"></i>Profile</a>
				<a href="<?php echo htmlspecialchars($_v['data']['base'], ENT_QUOTES);?>
/keys"><i class="ic ic-key" aria-hidden="true"></i>SSH keys</a>
				<?php if($_v['data']['settings']){?>

				<a href="<?php echo htmlspecialchars($_v['data']['base'], ENT_QUOTES);?>
/settings"><i class="ic ic-settings" aria-hidden="true"></i>Settings</a>
				<?php }?>
				<?php if($_v['data']['loginRequired']){?>

				<form method="post" action="/logout"><button type="submit"><i class="ic ic-logout" aria-hidden="true"></i>Logout</button></form>
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
active<?php }?>"><i class="ic ic-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i><span><?php echo htmlspecialchars($_v['item']['title'], ENT_QUOTES);?>
</span><?php if($_v['item']['pulse']){?>
<i class="dot<?php if($_v['chrome']['running']){?>
 running<?php }?>" data-any-running></i><?php }?></a>
<?php if($_v['item']['flows'] && $_v['chrome']['workflows']){ $_v['data_workflows'] = $_v['chrome']['workflows'];?>
				<div class="flows"><?php if(!empty($_v['data_workflows']))foreach($_v['data_workflows'] as $_v['flow']){ if(!$_v['flow']['hidden']){?>
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
"><i class="ic ic-eye" aria-hidden="true"></i></a>
<?php }elseif($_v['flow']['canRun']){?>
						<button type="button" class="play" data-workflow-run="<?php echo htmlspecialchars($_v['flow']['id'], ENT_QUOTES);?>
"<?php if($_v['flow']['prompted']){?>
 data-workflow-prompt="<?php echo htmlspecialchars($_v['flow']['name'], ENT_QUOTES);?>
"<?php }?> title="Run <?php echo htmlspecialchars($_v['flow']['name'], ENT_QUOTES);?>
" aria-label="Run <?php echo htmlspecialchars($_v['flow']['name'], ENT_QUOTES);?>
"><i class="ic ic-play" aria-hidden="true"></i></button>
<?php }?>
					</div>
<?php } }?>
				</div>
<?php } }?>
			</nav>

<?php $_v['data'] = $_v['chrome']['folders'];?>

			<nav class="folders">
<?php if(!empty($_v['data']))foreach($_v['data'] as $_v['group']){ if($_v['group']['workspace']){ $_v['data_workspace'] = $_v['group']; if($_v['data_workspace']['top']){?>

<details class="workspace<?php if($_v['data_workspace']['active']){?>
 active<?php }?>" data-project="<?php echo htmlspecialchars($_v['data_workspace']['id'], ENT_QUOTES);?>
" open>
	<summary<?php if($_v['data_workspace']['here']){?>
 class="active"<?php }?> title="<?php echo htmlspecialchars($_v['data_workspace']['title'], ENT_QUOTES);?>
">
		<i class="ic ic-chev chev" aria-hidden="true"></i>
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
<?php if(!empty($_v['data_workspace']['projects']))foreach($_v['data_workspace']['projects'] as $_v['project']){?><?php $_v['data_project'] = $_v['project'];?>
				<div class="project<?php if($_v['data_project']['active']){?>
 active<?php }?>" data-project="<?php echo htmlspecialchars($_v['data_project']['id'], ENT_QUOTES);?>
">
					<a class="section<?php if($_v['data_project']['here']){?>
 active<?php }?>" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_project']['href'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['data_project']['id'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['data_project']['title'], ENT_QUOTES);?>
">
<?php $_v['data_icon'] = $_v['data_project'];?>
<i class="ic ic-<?php echo htmlspecialchars($_v['data_icon']['icon'], ENT_QUOTES); if($_v['data_icon']['dirty']){?>
 dirty<?php }?>"<?php if($_v['data_icon']['dirty']){?>
 title="<?php echo htmlspecialchars($_v['data_icon']['changed'], ENT_QUOTES);?>
"<?php }?>></i><span><?php echo htmlspecialchars($_v['data_project']['name'], ENT_QUOTES);?>
</span>
<?php if($_v['data_project']['bad']){?>
<small class="bad" title="<?php echo htmlspecialchars($_v['data_project']['badTitle'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_project']['badge'], ENT_QUOTES);?>
</small><?php }elseif($_v['data_project']['counted']){?>
<small><?php echo htmlspecialchars($_v['data_project']['count'], ENT_QUOTES);?>
</small><?php }?>
					</a>
				</div>
<?php }?>
</details>
<?php }else{if(!empty($_v['data_workspace']['projects']))foreach($_v['data_workspace']['projects'] as $_v['project']){?><?php $_v['data_project'] = $_v['project'];?>
				<div class="project<?php if($_v['data_project']['active']){?>
 active<?php }?>" data-project="<?php echo htmlspecialchars($_v['data_project']['id'], ENT_QUOTES);?>
">
					<a class="section<?php if($_v['data_project']['here']){?>
 active<?php }?>" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_project']['href'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['data_project']['id'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['data_project']['title'], ENT_QUOTES);?>
">
<?php $_v['data_icon'] = $_v['data_project'];?>
<i class="ic ic-<?php echo htmlspecialchars($_v['data_icon']['icon'], ENT_QUOTES); if($_v['data_icon']['dirty']){?>
 dirty<?php }?>"<?php if($_v['data_icon']['dirty']){?>
 title="<?php echo htmlspecialchars($_v['data_icon']['changed'], ENT_QUOTES);?>
"<?php }?>></i><span><?php echo htmlspecialchars($_v['data_project']['name'], ENT_QUOTES);?>
</span>
<?php if($_v['data_project']['bad']){?>
<small class="bad" title="<?php echo htmlspecialchars($_v['data_project']['badTitle'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_project']['badge'], ENT_QUOTES);?>
</small><?php }elseif($_v['data_project']['counted']){?>
<small><?php echo htmlspecialchars($_v['data_project']['count'], ENT_QUOTES);?>
</small><?php }?>
					</a>
				</div>
<?php } }}else{ $_v['data_project'] = $_v['group'];?>
				<div class="project<?php if($_v['data_project']['active']){?>
 active<?php }?>" data-project="<?php echo htmlspecialchars($_v['data_project']['id'], ENT_QUOTES);?>
">
					<a class="section<?php if($_v['data_project']['here']){?>
 active<?php }?>" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_project']['href'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['data_project']['id'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['data_project']['title'], ENT_QUOTES);?>
">
<?php $_v['data_icon'] = $_v['data_project'];?>
<i class="ic ic-<?php echo htmlspecialchars($_v['data_icon']['icon'], ENT_QUOTES); if($_v['data_icon']['dirty']){?>
 dirty<?php }?>"<?php if($_v['data_icon']['dirty']){?>
 title="<?php echo htmlspecialchars($_v['data_icon']['changed'], ENT_QUOTES);?>
"<?php }?>></i><span><?php echo htmlspecialchars($_v['data_project']['name'], ENT_QUOTES);?>
</span>
<?php if($_v['data_project']['bad']){?>
<small class="bad" title="<?php echo htmlspecialchars($_v['data_project']['badTitle'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_project']['badge'], ENT_QUOTES);?>
</small><?php }elseif($_v['data_project']['counted']){?>
<small><?php echo htmlspecialchars($_v['data_project']['count'], ENT_QUOTES);?>
</small><?php }?>
					</a>
				</div>

<?php } }else{?>

				<div class="empty">No folders yet. <?php if($_v['chrome']['can']['project/edit']){?>
<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/settings/projects/new">Add a git repository</a><?php }else{?>
Add a git repository in Settings<?php }?>. Put a folder with jobs under the root.</div>
<?php }?>
			</nav>

		</aside>
		<main id="main" tabindex="-1">
<?php $_v['data'] = $_v['chrome'];if($_v['data']['crumbs']){?>

			<div class="crumbs"><a href="<?php echo htmlspecialchars($_v['data']['base'], ENT_QUOTES);?>
/">Dashboard</a><?php if(!empty($_v['data']['crumbs']))foreach($_v['data']['crumbs'] as $_v['crumb']){?><span><?php if($_v['crumb']['href']){?>
<a href="<?php echo htmlspecialchars($_v['data']['base'], ENT_QUOTES); echo htmlspecialchars($_v['crumb']['href'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['crumb']['text'], ENT_QUOTES);?>
</a><?php }else{ echo htmlspecialchars($_v['crumb']['text'], ENT_QUOTES); }?></span><?php }?></div>
<?php }if($_v['data']['notice']){?>

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
</b><small><?php if($_v['entry']['icon']){?>
<i class="ic ic-<?php echo htmlspecialchars($_v['entry']['icon'], ENT_QUOTES);?>
 lead" aria-hidden="true"></i><?php }?><?php echo htmlspecialchars($_v['entry']['place'], ENT_QUOTES);?>
</small></span>
<?php if($_v['entry']['children']){?>
						<span class="badge" title="<?php if($_v['entry']['thread']){?>
Replies<?php }else{?>
Runs it started<?php }?>"><?php echo htmlspecialchars($_v['entry']['children'], ENT_QUOTES);?>
</span><?php } if(!$_v['entry']['ended']){?>
						<i class="ic ic-<?php echo htmlspecialchars($_v['entry']['act'], ENT_QUOTES);?>
 act" aria-hidden="true"></i><?php }?>
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
" class="recent more" data-recents-more hidden><i class="ic ic-list" aria-hidden="true"></i><span class="what"><b>All runs</b><small></small></span></a>
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
			<button type="button" class="btn-icon" id="run-mic" aria-haspopup="dialog" title="Dictate into the run" aria-label="Dictate into the run" hidden><i class="ic ic-mic" aria-hidden="true"></i></button>
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
		<div class="run-body preview-body"><img id="preview-image" alt="" draggable="false"><div id="preview-doc" class="preview-doc" hidden></div><audio id="preview-audio" class="preview-audio" controls preload="metadata" hidden></audio><iframe id="preview-frame" class="preview-frame" title="The page the run wrote" sandbox referrerpolicy="no-referrer" hidden></iframe></div>
		<div class="run-foot">
			<span class="hint"><select id="preview-version" aria-label="Version" hidden></select><input type="text" id="preview-commit" placeholder="Commit message" aria-label="Commit message" maxlength="200" autocomplete="off" hidden></span>
			<span class="preview-actions" id="preview-actions"></span>
			<button type="button" class="btn" id="preview-edit" hidden>Edit</button>
			<button type="button" class="btn" id="preview-push" hidden>Push</button>
			<button type="button" class="btn" id="preview-cancel" hidden>Cancel</button>
			<button type="button" class="btn btn-primary" id="preview-save" hidden>Save</button>
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

	<button type="button" class="voice-orb" id="voice-orb" aria-haspopup="dialog" title="Dictate" aria-label="Dictate" hidden><i class="ic ic-mic" aria-hidden="true"></i></button>
	<dialog class="voice on-dark" id="voice-overlay" aria-label="Dictate"<?php if($_v['chrome']['claude']){?>
 data-claude<?php }?><?php if($_v['chrome']['can']['note/edit']){?>
 data-note-edit<?php }?><?php if($_v['chrome']['contacts']){?>
 data-chat<?php }?>>
		<div class="voice-who">
<?php if($_v['chrome']['agents'] && $_v['chrome']['claude']){?>
			<label for="voice-agent">Talking to</label>
			<select id="voice-agent" class="voice-agent" aria-label="Talking to">
				<option value=""><?php echo htmlspecialchars($_v['chrome']['user']['name'], ENT_QUOTES);?>
 (you)</option>
<?php if(!empty($_v['chrome']['agents']))foreach($_v['chrome']['agents'] as $_v['agent']){?>				<option value="<?php echo htmlspecialchars($_v['agent'], ENT_QUOTES);?>
">@<?php echo htmlspecialchars($_v['agent'], ENT_QUOTES);?>
</option>
<?php }?>			</select>
<?php }?><?php if($_v['chrome']['contacts']){?>
			<label for="voice-chat">Chat</label>
			<select id="voice-chat" class="voice-chat" aria-label="Chat with">
				<option value="">No one</option>
<?php if(!empty($_v['chrome']['contacts']))foreach($_v['chrome']['contacts'] as $_v['contact']){?>				<option value="<?php echo htmlspecialchars($_v['contact']['name'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['contact']['name'], ENT_QUOTES);?>
</option>
<?php }?>			</select>
<?php }?>			<button type="button" class="voice-wake" id="voice-wake" aria-pressed="false" title="Start recording when you say a name" aria-label="Start recording when you say a name" hidden><i class="ic ic-pulse" aria-hidden="true"></i>Wake</button>
		</div>
		<button type="button" class="voice-stage" id="voice-stage" aria-pressed="false" title="Record" aria-label="Record" autofocus>
			<span class="voice-well" aria-hidden="true"></span>
			<svg class="voice-field" viewBox="0 0 320 320" aria-hidden="true" focusable="false">
				<defs>
					<filter id="voice-gravity" x="-20%" y="-20%" width="140%" height="140%" color-interpolation-filters="sRGB">
						<feImage id="voice-well-map" x="0" y="0" width="100%" height="100%" preserveAspectRatio="none" result="field"/>
						<feDisplacementMap in="SourceGraphic" in2="field" scale="54" xChannelSelector="R" yChannelSelector="G"/>
					</filter>
				</defs>
			</svg>
			<canvas class="voice-avatar" id="voice-avatar" width="256" height="256" aria-hidden="true"></canvas>
			<span class="voice-rec" aria-hidden="true"><i></i>Recording</span>
		</button>
		<p class="voice-said" id="voice-said" role="status" aria-live="polite"></p>
		<p class="voice-where" id="voice-where"></p>
		<div class="voice-quiz" id="voice-quiz" hidden>
			<p class="voice-quiz-count" id="voice-quiz-count"></p>
			<p class="voice-quiz-ask" id="voice-quiz-ask" role="status" aria-live="polite"></p>
			<div class="voice-quiz-picks" id="voice-quiz-picks" role="group" aria-labelledby="voice-quiz-ask"></div>
			<p class="voice-quiz-help" id="voice-quiz-help"></p>
			<div class="voice-quiz-acts">
				<button type="button" class="btn" id="voice-quiz-cancel">Cancel</button>
				<button type="button" class="btn btn-primary" id="voice-quiz-submit">Submit</button>
			</div>
		</div>
		<div class="voice-acts">
			<button type="button" class="voice-act discard" id="voice-discard" title="Back" aria-label="Back"><i class="ic ic-back" aria-hidden="true"></i></button>
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
/theme/ext/xterm.js?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
" defer></script>
	<script src="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/ext/addon-fit.js?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
" defer></script>
	<script src="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/mdedit/mdedit.global.js?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
" defer></script>
	<script src="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/task-ui.js?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
" defer></script>
	<script src="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/voice/orbs.js?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
" defer></script>
	<script src="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/voice/voice.js?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
" defer></script>
<?php if(!empty($_v['data']['scripts']))foreach($_v['data']['scripts'] as $_v['script']){?>	<script src="<?php echo htmlspecialchars($_v['script'], ENT_QUOTES);?>
" defer></script>
<?php }?>

</body>
</html>
