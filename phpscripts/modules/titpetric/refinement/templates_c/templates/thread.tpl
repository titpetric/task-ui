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
	</p>

	<ol class="refinement-thread">
	<?php if(!empty($_v['comments']))foreach($_v['comments'] as $_v['comment']){?>
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
<span class="badge">correction</span><?php }?>
			</div>
			<p><?php echo htmlspecialchars($_v['comment']['body'], ENT_QUOTES);?>
</p>
		</li>
	<?php }?>
	</ol>

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
