<?php $_v=&$this->vars;?>
<section class="refinement">
	<p class="refinement-back"><a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
">&larr; Every thread</a></p>

	<?php if($_v['report']){?>

	<div class="block refinement-report">
		<h3>The design <small>as the last run left it</small></h3>
		<?php echo $_v['report'];?>

	</div>
	<?php }else{?>

	<p class="faint">The first run has not written its report yet. It appears here once the run ends.</p>
	<?php }?>

	
	<p class="refinement-files">
		<a class="btn btn-sm" href="/artifacts?path=<?php echo htmlspecialchars($_v['thread']['id'], ENT_QUOTES);?>
">The files of this thread</a>
		<?php if($_v['scratchpad']){?>
<a class="btn btn-sm" href="/artifacts/files/<?php echo htmlspecialchars($_v['thread']['id'], ENT_QUOTES);?>
/SCRATCHPAD.md">The scratch pad</a><?php }?>
	</p>

	<ol class="refinement-thread">
	<?php if(!empty($_v['entries']))foreach($_v['entries'] as $_v['entry']){?>
		<?php if($_v['entry']['folded_count']){?>

		<li class="refinement-folded">
			<details>
				<summary><?php echo htmlspecialchars($_v['entry']['folded_count'], ENT_QUOTES);?>
 folded into the request below</summary>
				<ol>
				<?php if(!empty($_v['entry']['folded']))foreach($_v['entry']['folded'] as $_v['comment']){?>
					<li class="refinement-comment" data-run-state="<?php echo htmlspecialchars($_v['comment']['state'], ENT_QUOTES);?>
">
						<div class="refinement-comment-head">
							<strong><?php echo htmlspecialchars($_v['comment']['user'], ENT_QUOTES);?>
</strong>
							<span class="faint"><?php echo htmlspecialchars($_v['comment']['created_at'], ENT_QUOTES);?>
</span>
							<?php if($_v['comment']['note_id']){?>
<a class="badge" href="/notes#<?php echo htmlspecialchars($_v['comment']['note_id'], ENT_QUOTES);?>
">the run</a><?php }?>
							<?php if($_v['comment']['correction']){?>
<span class="badge">rule</span><?php }?>
						</div>
						<div class="refinement-body"><?php echo $_v['comment']['html'];?>
</div>
					</li>
				<?php }?>
				</ol>
			</details>
		</li>
		<?php }?>
		<li class="refinement-comment" data-run-state="<?php echo htmlspecialchars($_v['entry']['comment']['state'], ENT_QUOTES);?>
">
			<div class="refinement-comment-head">
				<strong><?php echo htmlspecialchars($_v['entry']['comment']['user'], ENT_QUOTES);?>
</strong>
				<span class="faint"><?php echo htmlspecialchars($_v['entry']['comment']['created_at'], ENT_QUOTES);?>
</span>
				<?php if($_v['entry']['comment']['note_id']){?>
<a class="badge" href="/notes#<?php echo htmlspecialchars($_v['entry']['comment']['note_id'], ENT_QUOTES);?>
">the run</a><?php }?>
				<?php if($_v['entry']['applied']){?>
<span class="badge signal">applied</span><?php }?>
				<?php if($_v['entry']['comment']['correction']){?>
<span class="badge">rule</span><?php }?>
				
				<form class="refinement-delete" method="post" action="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/threads/<?php echo htmlspecialchars($_v['thread']['id'], ENT_QUOTES);?>
/comments/<?php echo htmlspecialchars($_v['entry']['comment']['id'], ENT_QUOTES);?>
/delete"
					<?php if($_v['entry']['comment']['id'] == $_v['design']){?>
data-confirm="Delete the design, and the whole thread with it? Its rules and its runs go. The files stay in the artifacts, under <?php echo htmlspecialchars($_v['thread']['id'], ENT_QUOTES);?>
."<?php }else{?>
data-confirm="Delete this request? The next Apply all will not carry it. The files it produced stay."<?php }?>>
					<button class="btn-icon" type="submit" title="Delete" aria-label="Delete"><svg class="ic"><use href="#i-trash"/></svg></button>
				</form>
			</div>
			<div class="refinement-body"><?php echo $_v['entry']['comment']['html'];?>
</div>
		</li>
	<?php }?>
	</ol>

	
	<form class="form refinement-apply" method="post" action="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/threads/<?php echo htmlspecialchars($_v['thread']['id'], ENT_QUOTES);?>
/apply">
		<p class="faint">Folds every request above into one and runs it over the files as they stand, so the result satisfies all of them together rather than one after another.</p>
		<button class="btn" type="submit"><svg class="ic"><use href="#i-collapse"/></svg>Apply all</button>
	</form>

	<form class="form refinement-comment-form" method="post" action="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/threads/<?php echo htmlspecialchars($_v['thread']['id'], ENT_QUOTES);?>
/comments">
		<label for="refinement-comment">Refine it</label>
		<p class="faint" id="refinement-comment-help">One change per comment. Each one starts a run over the artifacts as they stand now.</p>
		<textarea id="refinement-comment" name="comment" rows="3" aria-describedby="refinement-comment-help" placeholder="use ULID as primary keys" required></textarea>
		<div class="form-foot">
			<button class="btn btn-primary" type="submit">Send</button>
		</div>
	</form>
</section>
