<?php $_v=&$this->vars;$_v['data_head'] = $_v['data']['head'];?>


<div class="page-head">
	<div>
		<h2><?php if($_v['data_head']['icon']){?>
<svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['data_head']['icon'], ENT_QUOTES);?>
"/></svg> <?php }?><?php echo htmlspecialchars($_v['data_head']['title'], ENT_QUOTES); if($_v['data_head']['badge']){?>
 <span class="badge<?php if($_v['data_head']['tone']){?>
 <?php echo htmlspecialchars($_v['data_head']['tone'], ENT_QUOTES); }?>"><?php echo htmlspecialchars($_v['data_head']['badge'], ENT_QUOTES);?>
</span><?php }?><?php if(!empty($_v['data_head']['tags']))foreach($_v['data_head']['tags'] as $_v['tag']){?> <span class="badge"><?php echo htmlspecialchars($_v['tag'], ENT_QUOTES);?>
</span><?php }?><?php if($_v['data_head']['note']){?>
 <span class="term-title" title="<?php echo htmlspecialchars($_v['data_head']['note'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_head']['note'], ENT_QUOTES);?>
</span><?php }?></h2>
<?php if($_v['data_head']['lead'] || $_v['data_head']['leads']){?>
		<p><?php echo $_v['data_head']['lead'];?>
</p>
<?php }?><?php if($_v['data_head']['faint']){?>
		<p class="faint"><?php echo htmlspecialchars($_v['data_head']['faint'], ENT_QUOTES);?>
</p>
<?php }?>	</div>
<?php if($_v['data_head']['show']){?>
	<div class="actions"<?php if($_v['data_head']['live']){?>
 data-actions<?php }?>>
<?php if($_v['data_head']['run']){?>
		<?php $_v['data_run'] = $_v['data_head']['run'];?>

<div class="run-split">
	<button type="button" class="btn btn-primary btn-sm" data-run data-folder="<?php echo htmlspecialchars($_v['data_run']['folder'], ENT_QUOTES);?>
"<?php if($_v['data_run']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_run']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['data_run']['job'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_run']['what'], ENT_QUOTES);?>
"<?php if($_v['data_run']['args']){?>
 data-args="<?php echo htmlspecialchars($_v['data_run']['args'], ENT_QUOTES);?>
" title="Run with arguments after <?php echo htmlspecialchars($_v['data_run']['args'], ENT_QUOTES);?>
"<?php }?>><svg class="ic"><use href="#i-play"/></svg>Run<?php if($_v['data_run']['takesArgs']){?>
...<?php }?></button>
<?php if($_v['data_run']['menu']){?>
	<details class="more-menu run-menu">
		<summary class="btn btn-primary btn-sm run-caret" aria-haspopup="menu" aria-expanded="false" title="More ways to run <?php echo htmlspecialchars($_v['data_run']['name'], ENT_QUOTES);?>
" aria-label="More ways to run <?php echo htmlspecialchars($_v['data_run']['name'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-down"/></svg></summary>
		<div class="more-list">
			<div class="more-items" role="menu" aria-label="Run options">
<?php if($_v['data_run']['summarize']){?>
				<button type="button" class="more-item" role="menuitem" data-run data-summarize data-folder="<?php echo htmlspecialchars($_v['data_run']['folder'], ENT_QUOTES);?>
"<?php if($_v['data_run']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_run']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['data_run']['job'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_run']['what'], ENT_QUOTES);?>
"<?php if($_v['data_run']['args']){?>
 data-args="<?php echo htmlspecialchars($_v['data_run']['args'], ENT_QUOTES);?>
"<?php }?> title="Run and summarize"><svg class="ic"><use href="#i-summary"/></svg><span>Run and summarize</span></button>
<?php }?><?php if($_v['data_run']['force']){?>
				<button type="button" class="more-item danger" role="menuitem" data-run data-force data-folder="<?php echo htmlspecialchars($_v['data_run']['folder'], ENT_QUOTES);?>
"<?php if($_v['data_run']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_run']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['data_run']['job'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_run']['place'], ENT_QUOTES);?>
" title="Run out of hours, with a reason"><svg class="ic"><use href="#i-play"/></svg><span>Run force</span></button>
<?php }?>			</div>
		</div>
	</details>
<?php }?></div>

<?php }?><?php if($_v['data_head']['buttons']['items']){?>
		<?php $_v['data_buttons'] = $_v['data_head']['buttons'];if(!empty($_v['data_buttons']['items']))foreach($_v['data_buttons']['items'] as $_v['button']){?><?php if($_v['button']['bare']){?>
	<button type="button" class="btn-icon"<?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
"/></svg></button>
<?php }elseif($_v['button']['faint']){?>
	<span class="faint run-denied"><?php echo htmlspecialchars($_v['button']['faint'], ENT_QUOTES);?>
</span>
<?php }elseif($_v['button']['href']){?>
	<a class="btn<?php if($_v['data_buttons']['small'] || $_v['button']['small']){?>
 btn-sm<?php }?><?php if($_v['button']['primary']){?>
 btn-primary<?php }?><?php if($_v['button']['danger']){?>
 btn-danger<?php }?>" href="<?php echo htmlspecialchars($_v['button']['href'], ENT_QUOTES);?>
"<?php if($_v['button']['download']){?>
 download<?php }?><?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><?php if($_v['button']['icon']){?>
<svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
"/></svg><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
</a>
<?php }else{?>
	<button type="button" class="btn<?php if($_v['data_buttons']['small'] || $_v['button']['small']){?>
 btn-sm<?php }?><?php if($_v['button']['primary']){?>
 btn-primary<?php }?><?php if($_v['button']['danger']){?>
 btn-danger<?php }?>"<?php if($_v['button']['disabled']){?>
 disabled<?php if($_v['button']['note']){?>
 title="<?php echo htmlspecialchars($_v['button']['note'], ENT_QUOTES);?>
"<?php }?><?php }?><?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><?php if($_v['button']['icon']){?>
<svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
"/></svg><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
</button>
<?php }?><?php }}?><?php if($_v['data_head']['actions']){?>
		<?php $_v['data_menu'] = $_v['data_head']['actions'];?>

		
<?php if(!$_v['data_menu']['empty']){?>
<details class="more-menu">
	<summary class="btn-icon" aria-haspopup="menu" aria-expanded="false" title="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-more"/></svg></summary>
	<div class="more-list">
<?php if($_v['data_menu']['details']){?>
		<dl class="more-details">
<?php if(!empty($_v['data_menu']['details']))foreach($_v['data_menu']['details'] as $_v['detail']){?>			<dt><?php echo htmlspecialchars($_v['detail']['label'], ENT_QUOTES);?>
</dt><dd<?php if($_v['detail']['mono']){?>
 class="mono"<?php }?><?php if($_v['detail']['title']){?>
 title="<?php echo htmlspecialchars($_v['detail']['title'], ENT_QUOTES);?>
"<?php }?>><?php echo htmlspecialchars($_v['detail']['value'], ENT_QUOTES);?>
</dd>
<?php }?>		</dl>
<?php }?>		<div class="more-items" role="menu" aria-label="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
">
<?php if(!empty($_v['data_menu']['groups']))foreach($_v['data_menu']['groups'] as $_v['index'] => $_v['group']){?><?php if($_v['index']){?>
			<hr class="more-sep">
<?php }?><?php if(!empty($_v['group']))foreach($_v['group'] as $_v['item']){?><?php if($_v['item']['href']){?>
			<a class="more-item<?php if($_v['item']['danger']){?>
 danger<?php }?>" role="menuitem" href="<?php echo htmlspecialchars($_v['item']['href'], ENT_QUOTES);?>
"<?php if($_v['item']['download']){?>
 download<?php }?><?php if($_v['item']['attrs']){?>
 <?php echo $_v['item']['attrs']; }?>><svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
"/></svg><span><?php echo htmlspecialchars($_v['item']['label'], ENT_QUOTES);?>
</span><?php if($_v['item']['note']){?>
<small class="faint"><?php echo htmlspecialchars($_v['item']['note'], ENT_QUOTES);?>
</small><?php }?></a>
<?php }else{?>
			<button type="button" class="more-item<?php if($_v['item']['danger']){?>
 danger<?php }?>" role="menuitem"<?php if($_v['item']['disabled']){?>
 disabled<?php }?><?php if($_v['item']['attrs']){?>
 <?php echo $_v['item']['attrs']; }?>><svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
"/></svg><span><?php echo htmlspecialchars($_v['item']['label'], ENT_QUOTES);?>
</span><?php if($_v['item']['note']){?>
<small class="faint"><?php echo htmlspecialchars($_v['item']['note'], ENT_QUOTES);?>
</small><?php }?></button>
<?php }?><?php }?><?php }?>
		</div>
	</div>
</details><?php }}?>	</div>
<?php }?></div>


<section class="refinement">
<?php if($_v['data']['visual']){?>
	<div class="refinement-output"><?php echo $_v['data']['output'];?>
</div>
<?php if($_v['data']['thinking']){?>
	<details class="refinement-thinking">
		<summary>show thinking</summary>
		<div class="refinement-report"><?php echo $_v['data']['thinking'];?>
</div>
	</details>
<?php }?><?php }elseif($_v['data']['thinking']){?>
	<div class="block refinement-report">
		<h3>The design <small>as the last run left it</small></h3>
		<?php echo $_v['data']['thinking'];?>

	</div>
<?php }else{?>
	<p class="faint">The first run has not written its report yet. It appears here once the run ends.</p>
<?php }?>
	<ol class="refinement-thread">
<?php if(!empty($_v['data']['entries']))foreach($_v['data']['entries'] as $_v['entry']){?><?php if($_v['entry']['folded_count']){?>
		<li class="refinement-folded">
			<details>
				<summary><?php echo htmlspecialchars($_v['entry']['folded_count'], ENT_QUOTES);?>
 folded into the request below</summary>
				<ol>
<?php if(!empty($_v['entry']['folded']))foreach($_v['entry']['folded'] as $_v['comment']){?>					<li class="refinement-comment">
						<div class="refinement-comment-head">
							<strong><?php echo htmlspecialchars($_v['comment']['user'], ENT_QUOTES);?>
</strong>
							<span class="faint"><?php echo htmlspecialchars($_v['comment']['created_at'], ENT_QUOTES);?>
</span>
							<?php if($_v['comment']['run']){?>
<span class="state" data-run-id="<?php echo htmlspecialchars($_v['comment']['run'], ENT_QUOTES);?>
"></span><?php }?><?php if($_v['comment']['note_id']){?>
<a class="badge" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/notes#<?php echo htmlspecialchars($_v['comment']['note_id'], ENT_QUOTES);?>
">the run</a><?php }?>
							<?php if($_v['comment']['correction']){?>
<span class="badge">rule</span><?php }?>
						</div>
						<div class="refinement-body"><?php echo $_v['comment']['html'];?>
</div>
					</li>
<?php }?>				</ol>
			</details>
		</li>
<?php }?>		<li class="refinement-comment">
			<div class="refinement-comment-head">
				<strong><?php echo htmlspecialchars($_v['entry']['comment']['user'], ENT_QUOTES);?>
</strong>
				<span class="faint"><?php echo htmlspecialchars($_v['entry']['comment']['created_at'], ENT_QUOTES);?>
</span>
				<?php if($_v['entry']['comment']['run']){?>
<span class="state" data-run-id="<?php echo htmlspecialchars($_v['entry']['comment']['run'], ENT_QUOTES);?>
"></span><?php }?><?php if($_v['entry']['comment']['note_id']){?>
<a class="badge" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/notes#<?php echo htmlspecialchars($_v['entry']['comment']['note_id'], ENT_QUOTES);?>
">the run</a><?php }?>
				<?php if($_v['entry']['applied']){?>
<span class="badge signal">applied</span><?php }?>
				<?php if($_v['entry']['comment']['correction']){?>
<span class="badge">rule</span><?php }?>
				
				<form class="refinement-delete" method="post" action="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/refinement/threads/<?php echo htmlspecialchars($_v['data']['thread']['id'], ENT_QUOTES);?>
/comments/<?php echo htmlspecialchars($_v['entry']['comment']['id'], ENT_QUOTES);?>
/delete"
					<?php if($_v['entry']['comment']['id'] == $_v['data']['design']){?>
data-confirm="Delete the design, and the whole thread with it? Its rules and its runs go. The files stay in the artifacts, under <?php echo htmlspecialchars($_v['data']['thread']['id'], ENT_QUOTES);?>
."<?php }else{?>
data-confirm="Delete this request? The next Apply all will not carry it. The files it produced stay."<?php }?>>
					<button class="btn-icon" type="submit" title="Delete" aria-label="Delete"><svg class="ic"><use href="#i-trash"/></svg></button>
				</form>
			</div>
			<div class="refinement-body"><?php echo $_v['entry']['comment']['html'];?>
</div>
		</li>
<?php }?>	</ol>

	<form class="form refinement-apply" method="post" action="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/refinement/threads/<?php echo htmlspecialchars($_v['data']['thread']['id'], ENT_QUOTES);?>
/apply">
		<p class="faint">Folds every request above into one and runs it over the files as they stand, so the result satisfies all of them together rather than one after another.</p>
		<button class="btn" type="submit"><svg class="ic"><use href="#i-collapse"/></svg>Apply all</button>
	</form>

	<form class="form refinement-fork" method="post" action="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/refinement/threads/<?php echo htmlspecialchars($_v['data']['thread']['id'], ENT_QUOTES);?>
/fork">
		<label for="refinement-fork">Fork</label>
		<p class="faint" id="refinement-fork-help">A new thread on another subject, carrying every rule above. Say what it is and what it should produce; the rules come along and the files do not.</p>
		<textarea id="refinement-fork" name="want" rows="2" aria-describedby="refinement-fork-help" placeholder="A deployment topology diagram: the hosts, what runs on each and the links between them." required></textarea>
		<div class="form-foot">
			<button class="btn" type="submit"><svg class="ic"><use href="#i-flow"/></svg>Fork</button>
		</div>
	</form>

	<form class="form refinement-comment-form" method="post" action="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/refinement/threads/<?php echo htmlspecialchars($_v['data']['thread']['id'], ENT_QUOTES);?>
/comments">
		<label for="refinement-comment">Refine it</label>
		<p class="faint" id="refinement-comment-help">One change per comment. Each one starts a run over the artifacts as they stand now.</p>
		<textarea id="refinement-comment" name="comment" rows="3" aria-describedby="refinement-comment-help" placeholder="use ULID as primary keys" required></textarea>
		<div class="form-foot">
			<button class="btn btn-primary" type="submit">Send</button>
		</div>
	</form>
</section>
