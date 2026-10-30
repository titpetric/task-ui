<?php $_v=&$this->vars;?>
<section class="refinement">
	<form class="form refinement-start" method="post" action="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/threads">
		<label for="refinement-want">What to design</label>
		<p class="faint" id="refinement-want-help">One paragraph, and the thread starts from it. Every comment after it is a Claude run over the artifacts as they stand.</p>
		<textarea id="refinement-want" name="want" rows="5" aria-describedby="refinement-want-help" placeholder="A database schema with strict separation of user, user_login_method and user_login_method_password." required></textarea>
		<div class="form-foot">
			<button class="btn btn-primary" type="submit">Start a thread</button>
		</div>
	</form>

	<?php if($_v['threads']){?>

	<div class="table-wrap">
		<table class="table">
			<thead><tr><th>Thread</th><th class="num">Comments</th><th>Started</th><th></th></tr></thead>
			<tbody>
			<?php if(!empty($_v['threads']))foreach($_v['threads'] as $_v['thread']){?>
				<tr>
					<td class="key"><a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/threads/<?php echo htmlspecialchars($_v['thread']['id'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['thread']['title'], ENT_QUOTES);?>
</a></td>
					<td class="num mono"><?php echo htmlspecialchars($_v['thread']['comments'], ENT_QUOTES);?>
</td>
					<td class="nowrap faint"><?php echo htmlspecialchars($_v['thread']['created_at'], ENT_QUOTES);?>
</td>
					<td class="actions"><a class="btn btn-sm" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/threads/<?php echo htmlspecialchars($_v['thread']['id'], ENT_QUOTES);?>
">Open</a></td>
				</tr>
			<?php }?>
			</tbody>
		</table>
	</div>
	<?php }else{?>

	<p class="faint">No threads yet.</p>
	<?php }?>
</section>
