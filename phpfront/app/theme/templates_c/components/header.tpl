<?php $_v=&$this->vars; if($_v['data']['can']['prompt/edit'] && $_v['data']['can']['prompt/view'] && $_v['data']['can']['note/view'] && $_v['data']['claude']){?>
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
